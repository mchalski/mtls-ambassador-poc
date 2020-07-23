// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package app

import (
	"context"

	"github.com/mendersoftware/mtls-ambassador/client/mender"
)

type AuthProvider interface {
	GetToken() (string, error)
}

type authProvider struct {
	client mender.Client
	token  string
	user   string
	pass   string
}

func NewAuthProvider(client mender.Client, user, pass string) (*authProvider, error) {
	l.Infof("logging in with user %s", user)

	token, err := client.Login(context.TODO(), user, pass)

	switch err {
	case nil:
		l.Info("logging in: ok")
		l.Debugf("token: %s", token)
		return &authProvider{
			client: client,
			user:   user,
			pass:   pass,
			token:  token,
		}, nil
	case mender.ErrUnauthorized:
		return nil, ErrUnauthorized
	default:
		return nil, err
	}
}

func (ap *authProvider) GetToken() (string, error) {
	return ap.token, nil
}
