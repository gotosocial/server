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

package security

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// UserAgentBlock blocks requests with undesired, empty, or invalid user-agent strings.
func (m *Module) UserAgentBlock(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func": "UserAgentBlock",
	})

	ua := c.Request.UserAgent()
	if ua == "" {
		l.Debug("aborting request because there's no user-agent set")
		c.AbortWithStatus(http.StatusTeapot)
		return
	}

	if strings.Contains(strings.ToLower(ua), strings.ToLower("friendica")) {
		l.Debugf("aborting request with user-agent %s because it contains 'friendica'", ua)
		c.AbortWithStatus(http.StatusTeapot)
		return
	}
}
