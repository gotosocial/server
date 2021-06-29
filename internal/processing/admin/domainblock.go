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

package admin

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func (p *processor) DomainBlockCreate(account *gtsmodel.Account, form *apimodel.DomainBlockCreateRequest) (*apimodel.DomainBlock, gtserror.WithCode) {
	// first check if we already have a block -- if err == nil we already had a block so we can skip a whole lot of work
	domainBlock := &gtsmodel.DomainBlock{}
	err := p.db.GetWhere([]db.Where{{Key: "domain", Value: form.Domain, CaseInsensitive: true}}, domainBlock)
	if err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			// something went wrong in the DB
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("DomainBlockCreate: db error checking for existence of domain block %s: %s", form.Domain, err))
		}

		// there's no block for this domain yet so create one
		// note: we take a new ulid from timestamp here in case we need to sort blocks
		blockID, err := id.NewULID()
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("DomainBlockCreate: error creating id for new domain block %s: %s", form.Domain, err))
		}

		domainBlock = &gtsmodel.DomainBlock{
			ID:                 blockID,
			Domain:             form.Domain,
			CreatedByAccountID: account.ID,
			PrivateComment:     form.PrivateComment,
			PublicComment:      form.PublicComment,
			Obfuscate:          form.Obfuscate,
		}

		// put the new block in the database
		if err := p.db.Put(domainBlock); err != nil {
			if _, ok := err.(db.ErrAlreadyExists); !ok {
				// there's a real error creating the block
				return nil, gtserror.NewErrorInternalError(fmt.Errorf("DomainBlockCreate: db error putting new domain block %s: %s", form.Domain, err))
			}
		}

		// process the side effects of the domain block asynchronously since it might take a little while
		go p.domainBlockProcessSideEffects(domainBlock) // TODO: add this to a queuing system so it can retry/resume
	}

	mastoDomainBlock, err := p.tc.DomainBlockToMasto(domainBlock)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("DomainBlockCreate: error converting domain block to frontend/masto representation %s: %s", form.Domain, err))
	}

	return mastoDomainBlock, nil
}

func (p *processor) domainBlockProcessSideEffects(block *gtsmodel.DomainBlock) {
	l := p.log.WithFields(logrus.Fields{
		"func": "domainBlockProcessSideEffects",
		"domain": block.Domain,
	})

	l.Debug("processing domain block side effects")

	// if we have an instance entry for this domain, update it with the new block ID and clear all fields
	instance := &gtsmodel.Instance{}
	if err := p.db.GetWhere([]db.Where{{Key: "domain", Value: block.Domain, CaseInsensitive: true}}, instance); err == nil {
		instance.Title = ""
		instance.UpdatedAt = time.Now()
		instance.SuspendedAt = time.Now()
		instance.DomainBlockID = block.ID
		instance.ShortDescription = ""
		instance.Description = ""
		instance.Terms = ""
		instance.ContactEmail = ""
		instance.ContactAccountUsername = ""
		instance.ContactAccountID = ""
		instance.Version = ""
		if err := p.db.UpdateByID(instance.ID, instance); err != nil {
			l.Errorf("domainBlockProcessSideEffects: db error updating instance: %s", err)
		}
		l.Debug("instance entry updated")
	}

	// if we have an instance account for this instance, delete it
	if err := p.db.DeleteWhere([]db.Where{{Key: "username", Value: block.Domain, CaseInsensitive: true}}, &gtsmodel.Account{}); err != nil {
		l.Errorf("domainBlockProcessSideEffects: db error removing instance account: %s", err)
	}

	// TODO: delete accounts through the normal account deletion system (which should also delete media + posts + remove posts from timelines)
	
}