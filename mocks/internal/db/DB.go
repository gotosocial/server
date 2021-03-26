// Code generated by mockery v2.7.4. DO NOT EDIT.

package mocks

import (
	context "context"

	mastotypes "github.com/gotosocial/gotosocial/pkg/mastotypes"

	mock "github.com/stretchr/testify/mock"

	model "github.com/gotosocial/gotosocial/internal/db/model"

	net "net"

	pub "github.com/go-fed/activity/pub"
)

// DB is an autogenerated mock type for the DB type
type DB struct {
	mock.Mock
}

// AccountToMastoSensitive provides a mock function with given fields: account
func (_m *DB) AccountToMastoSensitive(account *model.Account) (*mastotypes.Account, error) {
	ret := _m.Called(account)

	var r0 *mastotypes.Account
	if rf, ok := ret.Get(0).(func(*model.Account) *mastotypes.Account); ok {
		r0 = rf(account)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*mastotypes.Account)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.Account) error); ok {
		r1 = rf(account)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateTable provides a mock function with given fields: i
func (_m *DB) CreateTable(i interface{}) error {
	ret := _m.Called(i)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(i)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteByID provides a mock function with given fields: id, i
func (_m *DB) DeleteByID(id string, i interface{}) error {
	ret := _m.Called(id, i)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, interface{}) error); ok {
		r0 = rf(id, i)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteWhere provides a mock function with given fields: key, value, i
func (_m *DB) DeleteWhere(key string, value interface{}, i interface{}) error {
	ret := _m.Called(key, value, i)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, interface{}, interface{}) error); ok {
		r0 = rf(key, value, i)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DropTable provides a mock function with given fields: i
func (_m *DB) DropTable(i interface{}) error {
	ret := _m.Called(i)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(i)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Federation provides a mock function with given fields:
func (_m *DB) Federation() pub.Database {
	ret := _m.Called()

	var r0 pub.Database
	if rf, ok := ret.Get(0).(func() pub.Database); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(pub.Database)
		}
	}

	return r0
}

// GetAccountByUserID provides a mock function with given fields: userID, account
func (_m *DB) GetAccountByUserID(userID string, account *model.Account) error {
	ret := _m.Called(userID, account)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *model.Account) error); ok {
		r0 = rf(userID, account)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAll provides a mock function with given fields: i
func (_m *DB) GetAll(i interface{}) error {
	ret := _m.Called(i)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(i)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByID provides a mock function with given fields: id, i
func (_m *DB) GetByID(id string, i interface{}) error {
	ret := _m.Called(id, i)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, interface{}) error); ok {
		r0 = rf(id, i)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetFollowersByAccountID provides a mock function with given fields: accountID, followers
func (_m *DB) GetFollowersByAccountID(accountID string, followers *[]model.Follow) error {
	ret := _m.Called(accountID, followers)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *[]model.Follow) error); ok {
		r0 = rf(accountID, followers)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetFollowingByAccountID provides a mock function with given fields: accountID, following
func (_m *DB) GetFollowingByAccountID(accountID string, following *[]model.Follow) error {
	ret := _m.Called(accountID, following)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *[]model.Follow) error); ok {
		r0 = rf(accountID, following)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetLastStatusForAccountID provides a mock function with given fields: accountID, status
func (_m *DB) GetLastStatusForAccountID(accountID string, status *model.Status) error {
	ret := _m.Called(accountID, status)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *model.Status) error); ok {
		r0 = rf(accountID, status)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetStatusesByAccountID provides a mock function with given fields: accountID, statuses
func (_m *DB) GetStatusesByAccountID(accountID string, statuses *[]model.Status) error {
	ret := _m.Called(accountID, statuses)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *[]model.Status) error); ok {
		r0 = rf(accountID, statuses)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetStatusesByTimeDescending provides a mock function with given fields: accountID, statuses, limit
func (_m *DB) GetStatusesByTimeDescending(accountID string, statuses *[]model.Status, limit int) error {
	ret := _m.Called(accountID, statuses, limit)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *[]model.Status, int) error); ok {
		r0 = rf(accountID, statuses, limit)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetWhere provides a mock function with given fields: key, value, i
func (_m *DB) GetWhere(key string, value interface{}, i interface{}) error {
	ret := _m.Called(key, value, i)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, interface{}, interface{}) error); ok {
		r0 = rf(key, value, i)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IsEmailAvailable provides a mock function with given fields: email
func (_m *DB) IsEmailAvailable(email string) error {
	ret := _m.Called(email)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(email)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IsHealthy provides a mock function with given fields: ctx
func (_m *DB) IsHealthy(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IsUsernameAvailable provides a mock function with given fields: username
func (_m *DB) IsUsernameAvailable(username string) error {
	ret := _m.Called(username)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(username)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewSignup provides a mock function with given fields: username, reason, requireApproval, email, password, signUpIP, locale
func (_m *DB) NewSignup(username string, reason string, requireApproval bool, email string, password string, signUpIP net.IP, locale string) (*model.User, error) {
	ret := _m.Called(username, reason, requireApproval, email, password, signUpIP, locale)

	var r0 *model.User
	if rf, ok := ret.Get(0).(func(string, string, bool, string, string, net.IP, string) *model.User); ok {
		r0 = rf(username, reason, requireApproval, email, password, signUpIP, locale)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, bool, string, string, net.IP, string) error); ok {
		r1 = rf(username, reason, requireApproval, email, password, signUpIP, locale)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Put provides a mock function with given fields: i
func (_m *DB) Put(i interface{}) error {
	ret := _m.Called(i)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(i)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Stop provides a mock function with given fields: ctx
func (_m *DB) Stop(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateByID provides a mock function with given fields: id, i
func (_m *DB) UpdateByID(id string, i interface{}) error {
	ret := _m.Called(id, i)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, interface{}) error); ok {
		r0 = rf(id, i)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
