// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package app

import (
	"context"
	"errors"
	"testing"

	"github.com/mendersoftware/mtls-ambassador/client/mender"
	mmender "github.com/mendersoftware/mtls-ambassador/client/mender/mocks"

	"github.com/stretchr/testify/assert"
)

func TestAuthProviderNew(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string

		user string
		pass string

		clientToken string
		clientErr   error

		outErr error
	}{
		{
			name: "ok",
			user: "user",
			pass: "password",

			clientToken: "sometoken",
		},
		{
			name: "error, unauthorized",
			user: "user",
			pass: "password",

			clientErr: mender.ErrUnauthorized,
			outErr:    ErrUnauthorized,
		},
		{
			name: "error, internal",
			user: "user",
			pass: "password",

			clientErr: errors.New("internal error"),
			outErr:    errors.New("internal error"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(*testing.T) {
			client := &mmender.Client{}

			client.On("Login",
				context.TODO(),
				tc.user,
				tc.pass).
				Return(tc.clientToken, tc.clientErr)

			ap, err := NewAuthProvider(client, tc.user, tc.pass)

			if tc.outErr != nil {
				assert.Nil(t, ap)
				assert.EqualError(t, err, tc.outErr.Error())
			} else {
				assert.NotNil(t, ap)
				assert.Equal(t, tc.clientToken, ap.token)
				assert.Equal(t, tc.user, ap.user)
				assert.Equal(t, tc.pass, ap.pass)
			}

			client.AssertExpectations(t)
		})
	}
}

func TestAuthProviderGetToken(t *testing.T) {
	client := &mmender.Client{}

	client.On("Login",
		context.TODO(),
		"user",
		"pass").
		Return("token", nil)

	ap, err := NewAuthProvider(client, "user", "pass")
	assert.NoError(t, err)

	tok, _ := ap.GetToken()
	assert.Equal(t, "token", tok)
}
