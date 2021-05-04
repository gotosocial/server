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

package media

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// MediaCreatePOSTHandler handles requests to create/upload media attachments
func (m *Module) MediaCreatePOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "statusCreatePOSTHandler")
	authed, err := oauth.Authed(c, true, true, true, true) // posting new media is serious business so we want *everything*
	if err != nil {
		l.Debugf("couldn't auth: %s", err)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// extract the media create form from the request context
	l.Tracef("parsing request form: %s", c.Request.Form)
	form := &model.AttachmentRequest{}
	if err := c.ShouldBind(form); err != nil || form == nil {
		l.Debugf("could not parse form from request: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing one or more required form values"})
		return
	}

	// Give the fields on the request form a first pass to make sure the request is superficially valid.
	l.Tracef("validating form %+v", form)
	if err := validateCreateMedia(form, m.config.MediaConfig); err != nil {
		l.Debugf("error validating form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mastoAttachment, err := m.processor.MediaCreate(authed, form)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, mastoAttachment)
}

func validateCreateMedia(form *model.AttachmentRequest, config *config.MediaConfig) error {
	// check there actually is a file attached and it's not size 0
	if form.File == nil || form.File.Size == 0 {
		return errors.New("no attachment given")
	}

	// a very superficial check to see if no size limits are exceeded
	// we still don't actually know which media types we're dealing with but the other handlers will go into more detail there
	maxSize := config.MaxVideoSize
	if config.MaxImageSize > maxSize {
		maxSize = config.MaxImageSize
	}
	if form.File.Size > int64(maxSize) {
		return fmt.Errorf("file size limit exceeded: limit is %d bytes but attachment was %d bytes", maxSize, form.File.Size)
	}

	if len(form.Description) < config.MinDescriptionChars || len(form.Description) > config.MaxDescriptionChars {
		return fmt.Errorf("image description length must be between %d and %d characters (inclusive), but provided image description was %d chars", config.MinDescriptionChars, config.MaxDescriptionChars, len(form.Description))
	}

	// TODO: validate focus here

	return nil
}
