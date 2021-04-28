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

// Package federation provides ActivityPub/federation functionality for GoToSocial
package federation

import (
	"github.com/go-fed/activity/pub"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

// NewFederatingActor returns a go-fed compatible federating actor
func NewFederatingActor(db db.DB, transportController transport.Controller, config *config.Config, log *logrus.Logger) pub.FederatingActor {
	c := NewCommonBehavior(db, log, config, transportController)
	f := NewFederatingProtocol(db, log, config, transportController)
	pubDatabase := db.Federation()
	clock := &Clock{}
	return pub.NewFederatingActor(c, f, pubDatabase, clock)
}
