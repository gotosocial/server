// Code generated by mockery v2.7.4. DO NOT EDIT.

package model

import mock "github.com/stretchr/testify/mock"

// MockSmallMeta is an autogenerated mock type for the SmallMeta type
type MockSmallMeta struct {
	mock.Mock
}

// GetAspect provides a mock function with given fields:
func (_m *MockSmallMeta) GetAspect() float64 {
	ret := _m.Called()

	var r0 float64
	if rf, ok := ret.Get(0).(func() float64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}

// GetHeight provides a mock function with given fields:
func (_m *MockSmallMeta) GetHeight() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetSize provides a mock function with given fields:
func (_m *MockSmallMeta) GetSize() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetWidth provides a mock function with given fields:
func (_m *MockSmallMeta) GetWidth() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}
