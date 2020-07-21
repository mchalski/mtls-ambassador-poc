// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package app

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"errors"

	"github.com/mendersoftware/mtls-ambassador/client/mender"
	"github.com/mendersoftware/mtls-ambassador/utils"
)

var (
	ErrKeyMismatch     = errors.New("certificate key and auth request key don't match")
	ErrCertNum         = errors.New("need at least one client certificate")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrPreauthConflict = errors.New("preauth conflict")
)

type App interface {
	Preauth(ctx context.Context, req *mender.AuthReq) error
	VerifyClientCert(ctx context.Context,
		certs []*x509.Certificate,
		req *mender.AuthReq,
		bodyRaw []byte,
		bodySignature string) error
}

type app struct {
	apiClient    mender.Client
	authProvider AuthProvider
}

func NewApp(apiClient mender.Client, auth AuthProvider) *app {
	return &app{
		apiClient:    apiClient,
		authProvider: auth,
	}
}

func (app *app) VerifyClientCert(ctx context.Context, certs []*x509.Certificate,
	req *mender.AuthReq,
	bodyRaw []byte,
	bodySignature string) error {

	if len(certs) == 0 {
		return ErrCertNum
	}

	certKey := certs[0].PublicKey.(*rsa.PublicKey)

	certKeyStr, err := utils.SerializePubKey(certKey)
	if err != nil {
		return err
	}

	if certKeyStr != req.PubKey {
		return ErrKeyMismatch
	}

	return utils.VerifyAuthReqSign(bodySignature, certKey, bodyRaw)
}

func (app *app) Preauth(ctx context.Context, req *mender.AuthReq) error {

	token, err := app.authProvider.GetToken()
	if err != nil {
		return err
	}

	err = app.apiClient.Preauth(
		ctx,
		req.IdData,
		req.PubKey,
		token)

	if err == mender.ErrPreauthConflict {
		return ErrPreauthConflict
	}

	return err
}
