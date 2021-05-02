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

package federation

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

const (
	// UsernameKey is for account usernames.
	UsernameKey = "username"
	// UsersBasePath is the base path for serving information about Users eg https://example.org/users
	UsersBasePath = "/" + util.UsersPath
	// UsersBasePathWithID is just the users base path with the Username key in it.
	// Use this anywhere you need to know the username of the user being queried.
	// Eg https://example.org/users/:username
	UsersBasePathWithID = UsersBasePath + "/:" + UsernameKey
)

// ActivityPubAcceptHeaders represents the Accept headers mentioned here:
// https://www.w3.org/TR/activitypub/#retrieving-objects
var ActivityPubAcceptHeaders = []string{
	`application/activity+json`,
	`application/ld+json; profile="https://www.w3.org/ns/activitystreams"`,
}

// Module implements the FederationAPIModule interface
type Module struct {
	federator federation.Federator
	config    *config.Config
	db        db.DB
	tc        typeutils.TypeConverter
	log       *logrus.Logger
}

// New returns a new auth module
func New(db db.DB, federator federation.Federator, tc typeutils.TypeConverter, config *config.Config, log *logrus.Logger) apimodule.FederationAPIModule {
	return &Module{
		federator: federator,
		config:    config,
		db:        db,
		tc:        tc,
		log:       log,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *Module) Route(s router.Router) error {
	s.AttachHandler(http.MethodGet, UsersBasePathWithID, m.UsersGETHandler)
	return nil
}

// CreateTables populates necessary tables in the given DB
func (m *Module) CreateTables(db db.DB) error {
	// models := []interface{}{
	// 	&gtsmodel.MediaAttachment{},
	// }

	// for _, m := range models {
	// 	if err := db.CreateTable(m); err != nil {
	// 		return fmt.Errorf("error creating table: %s", err)
	// 	}
	// }
	return nil
}
