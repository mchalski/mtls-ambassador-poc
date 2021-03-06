// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mender "github.com/mendersoftware/mtls-ambassador/client/mender"
	mock "github.com/stretchr/testify/mock"

	x509 "crypto/x509"
)

// App is an autogenerated mock type for the App type
type App struct {
	mock.Mock
}

// Preauth provides a mock function with given fields: ctx, req
func (_m *App) Preauth(ctx context.Context, req *mender.AuthReq) error {
	ret := _m.Called(ctx, req)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *mender.AuthReq) error); ok {
		r0 = rf(ctx, req)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// VerifyClientCert provides a mock function with given fields: ctx, certs, req, bodyRaw, bodySignature
func (_m *App) VerifyClientCert(ctx context.Context, certs []*x509.Certificate, req *mender.AuthReq, bodyRaw []byte, bodySignature string) error {
	ret := _m.Called(ctx, certs, req, bodyRaw, bodySignature)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []*x509.Certificate, *mender.AuthReq, []byte, string) error); ok {
		r0 = rf(ctx, certs, req, bodyRaw, bodySignature)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
