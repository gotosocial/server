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

package gtsmodel

import "time"

// StatusFave refers to a 'fave' or 'like' in the database, from one account, targeting the status of another account
type StatusFave struct {
	// id of this fave in the database
	ID string `pg:"type:CHAR(26),pk,notnull,unique"`
	// when was this fave created
	CreatedAt time.Time `pg:"type:timestamp,notnull,default:now()"`
	// id of the account that created ('did') the fave
	AccountID string `pg:"type:CHAR(26),notnull"`
	// id the account owning the faved status
	TargetAccountID string `pg:"type:CHAR(26),notnull"`
	// database id of the status that has been 'faved'
	StatusID string `pg:"type:CHAR(26),notnull"`
	// ActivityPub URI of this fave
	URI string `pg:",notnull"`

	// GTSStatus is the status being interacted with. It won't be put or retrieved from the db, it's just for conveniently passing a pointer around.
	GTSStatus *Status `pg:"-"`
	// GTSTargetAccount is the account being interacted with. It won't be put or retrieved from the db, it's just for conveniently passing a pointer around.
	GTSTargetAccount *Account `pg:"-"`
	// GTSFavingAccount is the account doing the faving. It won't be put or retrieved from the db, it's just for conveniently passing a pointer around.
	GTSFavingAccount *Account `pg:"-"`
}
