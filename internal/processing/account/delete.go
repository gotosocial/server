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
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func (p *processor) Delete(account *gtsmodel.Account) error {
	l := p.log.WithFields(logrus.Fields{
		"func":     "Delete",
		"username": account.Username,
	})

	l.Debug("beginning account delete process")

	// TODO in this function:
	// 1. Delete account's application(s), clients, and oauth tokens
	// 2. Delete account's blocks
	// 3. Delete account's emoji
	// 4. Delete account's follow requests
	// 5. Delete account's follows
	// 6. Delete account's media attachments
	// 7. Delete account's mentions
	// 8. Delete account's notifications
	// 9. Delete account's polls
	// 10. Delete account's statuses
	// 11. Delete account's bookmarks
	// 12. Delete account's faves
	// 13. Delete account's mutes
	// 14. Delete account's streams
	// 15. Delete account's tags
	// 16. Delete account's user
	// 17. Delete account's timeline
	// 18. Delete account itself

	// 1. Delete account's application(s), clients, and oauth tokens
	// we only need to do this step for local account since remote ones won't have any tokens or applications on our server
	if account.Domain == "" {
		// see if we can get a user for this account
		u := &gtsmodel.User{}
		if err := p.db.GetWhere([]db.Where{{Key: "account_id", Value: account.ID}}, u); err == nil {
			// we got one! select all tokens with the user's ID
			tokens := []*oauth.Token{}
			if err := p.db.GetWhere([]db.Where{{Key: "user_id", Value: u.ID}}, &tokens); err == nil {
				// we have some tokens to delete
				for _, t := range tokens {
					// delete client(s) associated with this token
					if err := p.db.DeleteByID(t.ClientID, &oauth.Client{}); err != nil {
						l.Errorf("error deleting oauth client: %s", err)
					}
					// delete application(s) associated with this token
					if err := p.db.DeleteWhere([]db.Where{{Key: "client_id", Value: t.ClientID}}, &gtsmodel.Application{}); err != nil {
						l.Errorf("error deleting application: %s", err)
					}
					// delete the token itself
					if err := p.db.DeleteByID(t.ID, t); err != nil {
						l.Errorf("error deleting oauth token: %s", err)
					}
				}
			}
		}
	}

	// 2. Delete account's blocks
	// first delete any blocks that this account created
	if err := p.db.DeleteWhere([]db.Where{{Key: "account_id", Value: account.ID}}, &[]*gtsmodel.Block{}); err != nil {
		l.Errorf("error deleting blocks created by account: %s", err)
	}

	// now delete any blocks that target this account
	if err := p.db.DeleteWhere([]db.Where{{Key: "target_account_id", Value: account.ID}}, &[]*gtsmodel.Block{}); err != nil {
		l.Errorf("error deleting blocks targeting account: %s", err)
	}

	// 3. Delete account's emoji
	// nothing to do here

	// 4. Delete account's follow requests
	// first delete any follow requests that this account created
	if err := p.db.DeleteWhere([]db.Where{{Key: "account_id", Value: account.ID}}, &[]*gtsmodel.FollowRequest{}); err != nil {
		l.Errorf("error deleting follow requests created by account: %s", err)
	}

	// now delete any follow requests that target this account
	if err := p.db.DeleteWhere([]db.Where{{Key: "target_account_id", Value: account.ID}}, &[]*gtsmodel.FollowRequest{}); err != nil {
		l.Errorf("error deleting follow requests targeting account: %s", err)
	}

	// 5. Delete account's follows
	// first delete any follows that this account created
	if err := p.db.DeleteWhere([]db.Where{{Key: "account_id", Value: account.ID}}, &[]*gtsmodel.Follow{}); err != nil {
		l.Errorf("error deleting follows created by account: %s", err)
	}

	// now delete any follows that target this account
	if err := p.db.DeleteWhere([]db.Where{{Key: "target_account_id", Value: account.ID}}, &[]*gtsmodel.Follow{}); err != nil {
		l.Errorf("error deleting follows targeting account: %s", err)
	}

	return nil
}
