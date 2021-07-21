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

package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CallbackGETHandler parses a token from an external auth provider.
func (m *Module) CallbackGETHandler(c *gin.Context) {
	state := c.Query(callbackStateParam)
	code := c.Query(callbackCodeParam)

	claims, err := m.idp.HandleCallback(c.Request.Context(), state, code)
	if err != nil {
		c.String(http.StatusForbidden, err.Error())
		return
	}

	c.JSON(http.StatusOK, claims)
}
