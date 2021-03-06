package visibility

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *filter) StatusHometimelineable(targetStatus *gtsmodel.Status, timelineOwnerAccount *gtsmodel.Account) (bool, error) {
	l := f.log.WithFields(logrus.Fields{
		"func":     "StatusHometimelineable",
		"statusID": targetStatus.ID,
	})

	// status owner should always be able to see their own status in their timeline so we can return early if this is the case
	if timelineOwnerAccount != nil && targetStatus.AccountID == timelineOwnerAccount.ID {
		return true, nil
	}

	v, err := f.StatusVisible(targetStatus, timelineOwnerAccount)
	if err != nil {
		return false, fmt.Errorf("StatusHometimelineable: error checking visibility of status with id %s: %s", targetStatus.ID, err)
	}

	if !v {
		l.Debug("status is not hometimelineable because it's not visible to the requester")
		return false, nil
	}

	// Don't timeline a status whose parent hasn't been dereferenced yet or can't be dereferenced.
	// If we have the reply to URI but don't have an ID for the replied-to account or the replied-to status in our database, we haven't dereferenced it yet.
	if targetStatus.InReplyToURI != "" && (targetStatus.InReplyToID == "" || targetStatus.InReplyToAccountID == "") {
		return false, nil
	}

	// if a status replies to an ID we know in the database, we need to make sure we also follow the replied-to status owner account
	if targetStatus.InReplyToID != "" {
		// pin the reply to status on to this status if it hasn't been done already
		if targetStatus.GTSReplyToStatus == nil {
			rs := &gtsmodel.Status{}
			if err := f.db.GetByID(targetStatus.InReplyToID, rs); err != nil {
				return false, fmt.Errorf("StatusHometimelineable: error getting replied to status with id %s: %s", targetStatus.InReplyToID, err)
			}
			targetStatus.GTSReplyToStatus = rs
		}

		// pin the reply to account on to this status if it hasn't been done already
		if targetStatus.GTSReplyToAccount == nil {
			ra := &gtsmodel.Account{}
			if err := f.db.GetByID(targetStatus.InReplyToAccountID, ra); err != nil {
				return false, fmt.Errorf("StatusHometimelineable: error getting replied to account with id %s: %s", targetStatus.InReplyToAccountID, err)
			}
			targetStatus.GTSReplyToAccount = ra
		}

		// if it's a reply to the timelineOwnerAccount, we don't need to check if the timelineOwnerAccount follows itself, just return true, they can see it
		if targetStatus.AccountID == timelineOwnerAccount.ID {
			return true, nil
		}

		// the replied-to account != timelineOwnerAccount, so make sure the timelineOwnerAccount follows the replied-to account
		follows, err := f.db.Follows(timelineOwnerAccount, targetStatus.GTSReplyToAccount)
		if err != nil {
			return false, fmt.Errorf("StatusHometimelineable: error checking follow from account %s to account %s: %s", timelineOwnerAccount.ID, targetStatus.InReplyToAccountID, err)
		}

		// we don't want to timeline a reply to a status whose owner isn't followed by the requesting account
		if !follows {
			return false, nil
		}
	}

	return true, nil
}
