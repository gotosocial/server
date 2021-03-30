// Code generated by mockery v2.7.4. DO NOT EDIT.

package media

import (
	mock "github.com/stretchr/testify/mock"
	model "github.com/superseriousbusiness/gotosocial/internal/db/model"
)

// MockMediaHandler is an autogenerated mock type for the MediaHandler type
type MockMediaHandler struct {
	mock.Mock
}

// SetHeaderOrAvatarForAccountID provides a mock function with given fields: img, accountID, headerOrAvi
func (_m *MockMediaHandler) SetHeaderOrAvatarForAccountID(img []byte, accountID string, headerOrAvi string) (*model.MediaAttachment, error) {
	ret := _m.Called(img, accountID, headerOrAvi)

	var r0 *model.MediaAttachment
	if rf, ok := ret.Get(0).(func([]byte, string, string) *model.MediaAttachment); ok {
		r0 = rf(img, accountID, headerOrAvi)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.MediaAttachment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte, string, string) error); ok {
		r1 = rf(img, accountID, headerOrAvi)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
