// Code generated by mockery v2.7.4. DO NOT EDIT.

package mocks

import (
	context "context"

	config "github.com/gotosocial/gotosocial/internal/config"

	logrus "github.com/sirupsen/logrus"

	mock "github.com/stretchr/testify/mock"
)

// GTSAction is an autogenerated mock type for the GTSAction type
type GTSAction struct {
	mock.Mock
}

// Execute provides a mock function with given fields: _a0, _a1, _a2
func (_m *GTSAction) Execute(_a0 context.Context, _a1 *config.Config, _a2 *logrus.Logger) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *config.Config, *logrus.Logger) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
