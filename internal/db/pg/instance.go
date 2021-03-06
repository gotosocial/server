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
	"github.com/go-pg/pg/v10"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (ps *postgresService) GetUserCountForInstance(domain string) (int, error) {
	q := ps.conn.Model(&[]*gtsmodel.Account{})

	if domain == ps.config.Host {
		// if the domain is *this* domain, just count where the domain field is null
		q = q.Where("? IS NULL", pg.Ident("domain"))
	} else {
		q = q.Where("domain = ?", domain)
	}

	// don't count the instance account or suspended users
	q = q.Where("username != ?", domain).Where("? IS NULL", pg.Ident("suspended_at"))

	return q.Count()
}

func (ps *postgresService) GetStatusCountForInstance(domain string) (int, error) {
	q := ps.conn.Model(&[]*gtsmodel.Status{})

	if domain == ps.config.Host {
		// if the domain is *this* domain, just count where local is true
		q = q.Where("local = ?", true)
	} else {
		// join on the domain of the account
		q = q.Join("JOIN accounts AS account ON account.id = status.account_id").
			Where("account.domain = ?", domain)
	}

	return q.Count()
}

func (ps *postgresService) GetDomainCountForInstance(domain string) (int, error) {
	q := ps.conn.Model(&[]*gtsmodel.Instance{})

	if domain == ps.config.Host {
		// if the domain is *this* domain, just count other instances it knows about
		// exclude domains that are blocked
		q = q.Where("domain != ?", domain).Where("? IS NULL", pg.Ident("suspended_at"))
	} else {
		// TODO: implement federated domain counting properly for remote domains
		return 0, nil
	}

	return q.Count()
}

func (ps *postgresService) GetAccountsForInstance(domain string, maxID string, limit int) ([]*gtsmodel.Account, error) {
	ps.log.Debug("GetAccountsForInstance")

	accounts := []*gtsmodel.Account{}

	q := ps.conn.Model(&accounts).Where("domain = ?", domain).Order("id DESC")

	if maxID != "" {
		q = q.Where("id < ?", maxID)
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

	if len(accounts) == 0 {
		return nil, db.ErrNoEntries{}
	}

	return accounts, nil
}
