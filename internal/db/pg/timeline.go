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

package pg

import (
	"sort"

	"github.com/go-pg/pg/v10"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (ps *postgresService) GetHomeTimelineForAccount(accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, error) {
	statuses := []*gtsmodel.Status{}

	q := ps.conn.Model(&statuses)

	q = q.ColumnExpr("status.*").
		Join("LEFT JOIN follows AS f ON f.target_account_id = status.account_id").
		Where("f.account_id = ?", accountID).
		Order("status.id DESC")

	if maxID != "" {
		q = q.Where("status.id < ?", maxID)
	}

	if sinceID != "" {
		q = q.Where("status.id > ?", sinceID)
	}

	if minID != "" {
		q = q.Where("status.id > ?", minID)
	}

	if local {
		q = q.Where("status.local = ?", local)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	err := q.Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, db.ErrNoEntries{}
		}
		return nil, err
	}

	if len(statuses) == 0 {
		return nil, db.ErrNoEntries{}
	}

	return statuses, nil
}

func (ps *postgresService) GetPublicTimelineForAccount(accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, error) {
	statuses := []*gtsmodel.Status{}

	q := ps.conn.Model(&statuses).
		Where("visibility = ?", gtsmodel.VisibilityPublic).
		Where("? IS NULL", pg.Ident("in_reply_to_id")).
		Where("? IS NULL", pg.Ident("in_reply_to_uri")).
		Where("? IS NULL", pg.Ident("boost_of_id")).
		Order("status.id DESC")

	if maxID != "" {
		q = q.Where("status.id < ?", maxID)
	}

	if sinceID != "" {
		q = q.Where("status.id > ?", sinceID)
	}

	if minID != "" {
		q = q.Where("status.id > ?", minID)
	}

	if local {
		q = q.Where("status.local = ?", local)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	err := q.Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, db.ErrNoEntries{}
		}
		return nil, err
	}

	if len(statuses) == 0 {
		return nil, db.ErrNoEntries{}
	}

	return statuses, nil
}

// TODO optimize this query and the logic here, because it's slow as balls -- it takes like a literal second to return with a limit of 20!
// It might be worth serving it through a timeline instead of raw DB queries, like we do for Home feeds.
func (ps *postgresService) GetFavedTimelineForAccount(accountID string, maxID string, minID string, limit int) ([]*gtsmodel.Status, string, string, error) {

	faves := []*gtsmodel.StatusFave{}

	fq := ps.conn.Model(&faves).
		Where("account_id = ?", accountID).
		Order("id DESC")

	if maxID != "" {
		fq = fq.Where("id < ?", maxID)
	}

	if minID != "" {
		fq = fq.Where("id > ?", minID)
	}

	if limit > 0 {
		fq = fq.Limit(limit)
	}

	err := fq.Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, "", "", db.ErrNoEntries{}
		}
		return nil, "", "", err
	}

	if len(faves) == 0 {
		return nil, "", "", db.ErrNoEntries{}
	}

	// map[statusID]faveID -- we need this to sort statuses by fave ID rather than their own ID
	statusesFavesMap := map[string]string{}

	in := []string{}
	for _, f := range faves {
		statusesFavesMap[f.StatusID] = f.ID
		in = append(in, f.StatusID)
	}

	statuses := []*gtsmodel.Status{}
	err = ps.conn.Model(&statuses).Where("id IN (?)", pg.In(in)).Select()
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, "", "", db.ErrNoEntries{}
		}
		return nil, "", "", err
	}

	if len(statuses) == 0 {
		return nil, "", "", db.ErrNoEntries{}
	}

	// arrange statuses by fave ID
	sort.Slice(statuses, func(i int, j int) bool {
		statusI := statuses[i]
		statusJ := statuses[j]
		return statusesFavesMap[statusI.ID] < statusesFavesMap[statusJ.ID]
	})

	nextMaxID := faves[len(faves)-1].ID
	prevMinID := faves[0].ID
	return statuses, nextMaxID, prevMinID, nil
}
