package message

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) StatusCreate(auth *oauth.Auth, form *apimodel.AdvancedStatusCreateForm) (*apimodel.Status, error) {
	uris := util.GenerateURIsForAccount(auth.Account.Username, p.config.Protocol, p.config.Host)
	thisStatusID := uuid.NewString()
	thisStatusURI := fmt.Sprintf("%s/%s", uris.StatusesURI, thisStatusID)
	thisStatusURL := fmt.Sprintf("%s/%s", uris.StatusesURL, thisStatusID)
	newStatus := &gtsmodel.Status{
		ID:                       thisStatusID,
		URI:                      thisStatusURI,
		URL:                      thisStatusURL,
		Content:                  util.HTMLFormat(form.Status),
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
		Local:                    true,
		AccountID:                auth.Account.ID,
		ContentWarning:           form.SpoilerText,
		ActivityStreamsType:      gtsmodel.ActivityStreamsNote,
		Sensitive:                form.Sensitive,
		Language:                 form.Language,
		CreatedWithApplicationID: auth.Application.ID,
		Text:                     form.Status,
	}

	// check if replyToID is ok
	if err := p.processReplyToID(form, auth.Account.ID, newStatus); err != nil {
		return nil, err
	}

	// check if mediaIDs are ok
	if err := p.processMediaIDs(form, auth.Account.ID, newStatus); err != nil {
		return nil, err
	}

	// check if visibility settings are ok
	if err := p.processVisibility(form, auth.Account.Privacy, newStatus); err != nil {
		return nil, err
	}

	// handle language settings
	if err := p.processLanguage(form, auth.Account.Language, newStatus); err != nil {
		return nil, err
	}

	// handle mentions
	if err := p.processMentions(form, auth.Account.ID, newStatus); err != nil {
		return nil, err
	}

	if err := p.processTags(form, auth.Account.ID, newStatus); err != nil {
		return nil, err
	}

	if err := p.processEmojis(form, auth.Account.ID, newStatus); err != nil {
		return nil, err
	}

	// put the new status in the database, generating an ID for it in the process
	if err := p.db.Put(newStatus); err != nil {
		return nil, err
	}

	// change the status ID of the media attachments to the new status
	for _, a := range newStatus.GTSMediaAttachments {
		a.StatusID = newStatus.ID
		a.UpdatedAt = time.Now()
		if err := p.db.UpdateByID(a.ID, a); err != nil {
			return nil, err
		}
	}

	// return the frontend representation of the new status to the submitter
	mastoStatus, err := p.tc.StatusToMasto(newStatus, auth.Account, auth.Account, nil, newStatus.GTSReplyToAccount, nil)
	if err != nil {
		return nil, err
	}
	return mastoStatus, nil
}

func (p *processor) StatusDelete(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error) {
	l := p.log.WithField("func", "StatusDelete")
	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, fmt.Errorf("error fetching status %s: %s", targetStatusID, err)
	}

	if targetStatus.AccountID != authed.Account.ID {
		return nil, errors.New("status doesn't belong to requesting account")
	}

	l.Trace("going to get relevant accounts")
	relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		return nil, fmt.Errorf("error fetching related accounts for status %s: %s", targetStatusID, err)
	}

	var boostOfStatus *gtsmodel.Status
	if targetStatus.BoostOfID != "" {
		boostOfStatus = &gtsmodel.Status{}
		if err := p.db.GetByID(targetStatus.BoostOfID, boostOfStatus); err != nil {
			return nil, fmt.Errorf("error fetching boosted status %s: %s", targetStatus.BoostOfID, err)
		}
	}

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, authed.Account, authed.Account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostOfStatus)
	if err != nil {
		return nil, fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err)
	}

	if err := p.db.DeleteByID(targetStatus.ID, targetStatus); err != nil {
		return nil, fmt.Errorf("error deleting status from the database: %s", err)
	}

	return mastoStatus, nil
}

func (p *processor) StatusFave(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error) {
	l := p.log.WithField("func", "StatusFave")
	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, fmt.Errorf("error fetching status %s: %s", targetStatusID, err)
	}

	l.Tracef("going to search for target account %s", targetStatus.AccountID)
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetStatus.AccountID, targetAccount); err != nil {
		return nil, fmt.Errorf("error fetching target account %s: %s", targetStatus.AccountID, err)
	}

	l.Trace("going to get relevant accounts")
	relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		return nil, fmt.Errorf("error fetching related accounts for status %s: %s", targetStatusID, err)
	}

	l.Trace("going to see if status is visible")
	visible, err := p.db.StatusVisible(targetStatus, targetAccount, authed.Account, relevantAccounts) // requestingAccount might well be nil here, but StatusVisible knows how to take care of that
	if err != nil {
		return nil, fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err)
	}

	if !visible {
		return nil, errors.New("status is not visible")
	}

	// is the status faveable?
	if !targetStatus.VisibilityAdvanced.Likeable {
		return nil, errors.New("status is not faveable")
	}

	// it's visible! it's faveable! so let's fave the FUCK out of it
	_, err = p.db.FaveStatus(targetStatus, authed.Account.ID)
	if err != nil {
		return nil, fmt.Errorf("error faveing status: %s", err)
	}

	var boostOfStatus *gtsmodel.Status
	if targetStatus.BoostOfID != "" {
		boostOfStatus = &gtsmodel.Status{}
		if err := p.db.GetByID(targetStatus.BoostOfID, boostOfStatus); err != nil {
			return nil, fmt.Errorf("error fetching boosted status %s: %s", targetStatus.BoostOfID, err)
		}
	}

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, targetAccount, authed.Account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostOfStatus)
	if err != nil {
		return nil, fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err)
	}

	return mastoStatus, nil
}

func (p *processor) StatusFavedBy(authed *oauth.Auth, targetStatusID string) ([]*apimodel.Account, error) {
	l := p.log.WithField("func", "StatusFavedBy")

	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, fmt.Errorf("error fetching status %s: %s", targetStatusID, err)
	}

	l.Tracef("going to search for target account %s", targetStatus.AccountID)
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetStatus.AccountID, targetAccount); err != nil {
		return nil, fmt.Errorf("error fetching target account %s: %s", targetStatus.AccountID, err)
	}

	l.Trace("going to get relevant accounts")
	relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		return nil, fmt.Errorf("error fetching related accounts for status %s: %s", targetStatusID, err)
	}

	l.Trace("going to see if status is visible")
	visible, err := p.db.StatusVisible(targetStatus, targetAccount, authed.Account, relevantAccounts) // requestingAccount might well be nil here, but StatusVisible knows how to take care of that
	if err != nil {
		return nil, fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err)
	}

	if !visible {
		return nil, errors.New("status is not visible")
	}

	// get ALL accounts that faved a status -- doesn't take account of blocks and mutes and stuff
	favingAccounts, err := p.db.WhoFavedStatus(targetStatus)
	if err != nil {
		return nil, fmt.Errorf("error seeing who faved status: %s", err)
	}

	// filter the list so the user doesn't see accounts they blocked or which blocked them
	filteredAccounts := []*gtsmodel.Account{}
	for _, acc := range favingAccounts {
		blocked, err := p.db.Blocked(authed.Account.ID, acc.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking blocks: %s", err)
		}
		if !blocked {
			filteredAccounts = append(filteredAccounts, acc)
		}
	}

	// TODO: filter other things here? suspended? muted? silenced?

	// now we can return the masto representation of those accounts
	mastoAccounts := []*apimodel.Account{}
	for _, acc := range filteredAccounts {
		mastoAccount, err := p.tc.AccountToMastoPublic(acc)
		if err != nil {
			return nil, fmt.Errorf("error converting account to api model: %s", err)
		}
		mastoAccounts = append(mastoAccounts, mastoAccount)
	}

	return mastoAccounts, nil
}

func (p *processor) StatusGet(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error) {
	l := p.log.WithField("func", "StatusGet")

	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, fmt.Errorf("error fetching status %s: %s", targetStatusID, err)
	}

	l.Tracef("going to search for target account %s", targetStatus.AccountID)
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetStatus.AccountID, targetAccount); err != nil {
		return nil, fmt.Errorf("error fetching target account %s: %s", targetStatus.AccountID, err)
	}

	l.Trace("going to get relevant accounts")
	relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		return nil, fmt.Errorf("error fetching related accounts for status %s: %s", targetStatusID, err)
	}

	l.Trace("going to see if status is visible")
	visible, err := p.db.StatusVisible(targetStatus, targetAccount, authed.Account, relevantAccounts) // requestingAccount might well be nil here, but StatusVisible knows how to take care of that
	if err != nil {
		return nil, fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err)
	}

	if !visible {
		return nil, errors.New("status is not visible")
	}

	var boostOfStatus *gtsmodel.Status
	if targetStatus.BoostOfID != "" {
		boostOfStatus = &gtsmodel.Status{}
		if err := p.db.GetByID(targetStatus.BoostOfID, boostOfStatus); err != nil {
			return nil, fmt.Errorf("error fetching boosted status %s: %s", targetStatus.BoostOfID, err)
		}
	}

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, targetAccount, authed.Account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostOfStatus)
	if err != nil {
		return nil, fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err)
	}

	return mastoStatus, nil

}

func (p *processor) StatusUnfave(authed *oauth.Auth, targetStatusID string) (*apimodel.Status, error) {
	l := p.log.WithField("func", "StatusUnfave")
	l.Tracef("going to search for target status %s", targetStatusID)
	targetStatus := &gtsmodel.Status{}
	if err := p.db.GetByID(targetStatusID, targetStatus); err != nil {
		return nil, fmt.Errorf("error fetching status %s: %s", targetStatusID, err)
	}

	l.Tracef("going to search for target account %s", targetStatus.AccountID)
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetStatus.AccountID, targetAccount); err != nil {
		return nil, fmt.Errorf("error fetching target account %s: %s", targetStatus.AccountID, err)
	}

	l.Trace("going to get relevant accounts")
	relevantAccounts, err := p.db.PullRelevantAccountsFromStatus(targetStatus)
	if err != nil {
		return nil, fmt.Errorf("error fetching related accounts for status %s: %s", targetStatusID, err)
	}

	l.Trace("going to see if status is visible")
	visible, err := p.db.StatusVisible(targetStatus, targetAccount, authed.Account, relevantAccounts) // requestingAccount might well be nil here, but StatusVisible knows how to take care of that
	if err != nil {
		return nil, fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err)
	}

	if !visible {
		return nil, errors.New("status is not visible")
	}

	// is the status faveable?
	if !targetStatus.VisibilityAdvanced.Likeable {
		return nil, errors.New("status is not faveable")
	}

	// it's visible! it's faveable! so let's unfave the FUCK out of it
	_, err = p.db.UnfaveStatus(targetStatus, authed.Account.ID)
	if err != nil {
		return nil, fmt.Errorf("error unfaveing status: %s", err)
	}

	var boostOfStatus *gtsmodel.Status
	if targetStatus.BoostOfID != "" {
		boostOfStatus = &gtsmodel.Status{}
		if err := p.db.GetByID(targetStatus.BoostOfID, boostOfStatus); err != nil {
			return nil, fmt.Errorf("error fetching boosted status %s: %s", targetStatus.BoostOfID, err)
		}
	}

	mastoStatus, err := p.tc.StatusToMasto(targetStatus, targetAccount, authed.Account, relevantAccounts.BoostedAccount, relevantAccounts.ReplyToAccount, boostOfStatus)
	if err != nil {
		return nil, fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err)
	}

	return mastoStatus, nil
}
