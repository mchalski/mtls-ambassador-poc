// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// AuthProvider is an autogenerated mock type for the AuthProvider type
type AuthProvider struct {
	mock.Mock
}

// GetToken provides a mock function with given fields:
func (_m *AuthProvider) GetToken() (string, error) {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
