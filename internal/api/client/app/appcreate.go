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

package app

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// AppsPOSTHandler should be served at https://example.org/api/v1/apps
// It is equivalent to: https://docs.joinmastodon.org/methods/apps/
func (m *Module) AppsPOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "AppsPOSTHandler")
	l.Trace("entering AppsPOSTHandler")

	authed, err := oauth.Authed(c, false, false, false, false)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	form := &model.ApplicationCreateRequest{}
	if err := c.ShouldBind(form); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// permitted length for most fields
	formFieldLen := 64
	// redirect can be a bit bigger because we probably need to encode data in the redirect uri
	formRedirectLen := 512

	// check lengths of fields before proceeding so the user can't spam huge entries into the database
	if len(form.ClientName) > formFieldLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("client_name must be less than %d bytes", formFieldLen)})
		return
	}
	if len(form.Website) > formFieldLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("website must be less than %d bytes", formFieldLen)})
		return
	}
	if len(form.RedirectURIs) > formRedirectLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("redirect_uris must be less than %d bytes", formRedirectLen)})
		return
	}
	if len(form.Scopes) > formFieldLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("scopes must be less than %d bytes", formFieldLen)})
		return
	}

	mastoApp, err := m.processor.AppCreate(authed, form)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// done, return the new app information per the spec here: https://docs.joinmastodon.org/methods/apps/
	c.JSON(http.StatusOK, mastoApp)
}
