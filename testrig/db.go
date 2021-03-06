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

package testrig

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/pg"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

var testModels []interface{} = []interface{}{
	&gtsmodel.Account{},
	&gtsmodel.Application{},
	&gtsmodel.Block{},
	&gtsmodel.DomainBlock{},
	&gtsmodel.EmailDomainBlock{},
	&gtsmodel.Follow{},
	&gtsmodel.FollowRequest{},
	&gtsmodel.MediaAttachment{},
	&gtsmodel.Mention{},
	&gtsmodel.Status{},
	&gtsmodel.StatusFave{},
	&gtsmodel.StatusBookmark{},
	&gtsmodel.StatusMute{},
	&gtsmodel.Tag{},
	&gtsmodel.User{},
	&gtsmodel.Emoji{},
	&gtsmodel.Instance{},
	&gtsmodel.Notification{},
	&gtsmodel.RouterSession{},
	&oauth.Token{},
	&oauth.Client{},
}

// NewTestDB returns a new initialized, empty database for testing
func NewTestDB() db.DB {
	config := NewTestConfig()
	l := logrus.New()
	l.SetLevel(logrus.TraceLevel)
	testDB, err := pg.NewPostgresService(context.Background(), config, l)
	if err != nil {
		panic(err)
	}
	return testDB
}

// StandardDBSetup populates a given db with all the necessary tables/models for perfoming tests.
func StandardDBSetup(db db.DB) {
	for _, m := range testModels {
		if err := db.CreateTable(m); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestTokens() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestClients() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestApplications() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestUsers() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestAccounts() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestAttachments() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestStatuses() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestEmojis() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestTags() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestFaves() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestFollows() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	for _, v := range NewTestNotifications() {
		if err := db.Put(v); err != nil {
			panic(err)
		}
	}

	if err := db.CreateInstanceAccount(); err != nil {
		panic(err)
	}

	if err := db.CreateInstanceInstance(); err != nil {
		panic(err)
	}
}

// StandardDBTeardown drops all the standard testing tables/models from the database to ensure it's clean for the next test.
func StandardDBTeardown(db db.DB) {
	for _, m := range testModels {
		if err := db.DropTable(m); err != nil {
			panic(err)
		}
	}
}
