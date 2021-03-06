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
)

func (p *processor) StatusesGet(requestingAccount *gtsmodel.Account, targetAccountID string, limit int, excludeReplies bool, maxID string, pinnedOnly bool, mediaOnly bool) ([]apimodel.Status, gtserror.WithCode) {
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetAccountID, targetAccount); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("no entry found for account id %s", targetAccountID))
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiStatuses := []apimodel.Status{}
	statuses, err := p.db.GetStatusesForAccount(targetAccountID, limit, excludeReplies, maxID, pinnedOnly, mediaOnly)
	if err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return apiStatuses, nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	for _, s := range statuses {
		visible, err := p.filter.StatusVisible(s, requestingAccount)
		if err != nil || !visible {
			continue
		}

		apiStatus, err := p.tc.StatusToMasto(s, requestingAccount)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status to masto: %s", err))
		}

		apiStatuses = append(apiStatuses, *apiStatus)
	}

	return apiStatuses, nil
}
