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

package account

import (
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/pkg/mastotypes"
)

func validateCreateAccount(form *mastotypes.AccountCreateRequest, c *config.AccountsConfig, database db.DB) error {
	if !c.OpenRegistration {
		return errors.New("registration is not open for this server")
	}

	if err := util.ValidateSignUpUsername(form.Username); err != nil {
		return err
	}

	if err := util.ValidateEmail(form.Email); err != nil {
		return err
	}

	if err := util.ValidateSignUpPassword(form.Password); err != nil {
		return err
	}

	if !form.Agreement {
		return errors.New("agreement to terms and conditions not given")
	}

	if err := util.ValidateLanguage(form.Locale); err != nil {
		return err
	}

	if err := util.ValidateSignUpReason(form.Reason, c.ReasonRequired); err != nil {
		return err
	}

	if err := database.IsEmailAvailable(form.Email); err != nil {
		return err
	}

	if err := database.IsUsernameAvailable(form.Username); err != nil {
		return err
	}

	return nil
}
