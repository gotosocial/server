/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package account

import (
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *processor) FollowCreate(requestingAccount *gtsmodel.Account, form *apimodel.AccountFollowRequest) (*apimodel.Relationship, gtserror.WithCode) {
	// if there's a block between the accounts we shouldn't create the request ofc
	blocked, err := p.db.Blocked(requestingAccount.ID, form.TargetAccountID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}
	if blocked {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("accountfollowcreate: block exists between accounts"))
	}

	// make sure the target account actually exists in our db
	targetAcct := &gtsmodel.Account{}
	if err := p.db.GetByID(form.TargetAccountID, targetAcct); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("accountfollowcreate: account %s not found in the db: %s", form.TargetAccountID, err))
		}
	}

	// check if a follow exists already
	follows, err := p.db.Follows(requestingAccount, targetAcct)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("accountfollowcreate: error checking follow in db: %s", err))
	}
	if follows {
		// already follows so just return the relationship
		return p.RelationshipGet(requestingAccount, form.TargetAccountID)
	}

	// check if a follow exists already
	followRequested, err := p.db.FollowRequested(requestingAccount, targetAcct)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("accountfollowcreate: error checking follow request in db: %s", err))
	}
	if followRequested {
		// already follow requested so just return the relationship
		return p.RelationshipGet(requestingAccount, form.TargetAccountID)
	}

	// make the follow request
	newFollowID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	fr := &gtsmodel.FollowRequest{
		ID:              newFollowID,
		AccountID:       requestingAccount.ID,
		TargetAccountID: form.TargetAccountID,
		ShowReblogs:     true,
		URI:             util.GenerateURIForFollow(requestingAccount.Username, p.config.Protocol, p.config.Host, newFollowID),
		Notify:          false,
	}
	if form.Reblogs != nil {
		fr.ShowReblogs = *form.Reblogs
	}
	if form.Notify != nil {
		fr.Notify = *form.Notify
	}

	// whack it in the database
	if err := p.db.Put(fr); err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("accountfollowcreate: error creating follow request in db: %s", err))
	}

	// if it's a local account that's not locked we can just straight up accept the follow request
	if !targetAcct.Locked && targetAcct.Domain == "" {
		if _, err := p.db.AcceptFollowRequest(requestingAccount.ID, form.TargetAccountID); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("accountfollowcreate: error accepting folow request for local unlocked account: %s", err))
		}
		// return the new relationship
		return p.RelationshipGet(requestingAccount, form.TargetAccountID)
	}

	// otherwise we leave the follow request as it is and we handle the rest of the process asynchronously
	p.fromClientAPI <- gtsmodel.FromClientAPI{
		APObjectType:   gtsmodel.ActivityStreamsFollow,
		APActivityType: gtsmodel.ActivityStreamsCreate,
		GTSModel:       fr,
		OriginAccount:  requestingAccount,
		TargetAccount:  targetAcct,
	}

	// return whatever relationship results from this
	return p.RelationshipGet(requestingAccount, form.TargetAccountID)
}
