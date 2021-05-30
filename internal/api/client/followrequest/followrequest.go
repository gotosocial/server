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

package followrequest

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

const (
	// IDKey is for status UUIDs
	IDKey = "id"
	// BasePath is the base path for serving the follow request API
	BasePath = "/api/v1/follow_requests"
	// BasePathWithID is just the base path with the ID key in it.
	// Use this anywhere you need to know the ID of the follow request being queried.
	BasePathWithID = BasePath + "/:" + IDKey

	// AcceptPath is used for accepting follow requests
	AcceptPath = BasePathWithID + "/authorize"
	// DenyPath is used for denying follow requests
	DenyPath = BasePathWithID + "/reject"
)

// Module implements the ClientAPIModule interface for every related to interacting with follow requests
type Module struct {
	config    *config.Config
	processor processing.Processor
	log       *logrus.Logger
}

// New returns a new follow request module
func New(config *config.Config, processor processing.Processor, log *logrus.Logger) api.ClientModule {
	return &Module{
		config:    config,
		processor: processor,
		log:       log,
	}
}

// Route attaches all routes from this module to the given router
func (m *Module) Route(r router.Router) error {
	r.AttachHandler(http.MethodGet, BasePath, m.FollowRequestGETHandler)
	r.AttachHandler(http.MethodPost, AcceptPath, m.FollowRequestAcceptPOSTHandler)
	r.AttachHandler(http.MethodPost, DenyPath, m.FollowRequestDenyPOSTHandler)
	return nil
}
