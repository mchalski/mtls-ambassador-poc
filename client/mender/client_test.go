// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package mender

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	c := NewClient("https://hosted.mender.io", false)
	assert.NotNil(t, c)
	assert.Equal(t, "https://hosted.mender.io", c.baseUrl)
	assert.NotNil(t, c.c.Timeout)
}

func TestClientLogin(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string

		insecure bool
		ret      int

		out error
	}{
		{
			name: "ok",
			ret:  200,
		},
		{
			name:     "ok, insecure",
			insecure: true,
			ret:      200,
		},
		{
			name: "unauthorized",
			ret:  401,
			out:  ErrUnauthorized,
		},
		{
			name: "internal",
			ret:  500,
			out:  errors.New("unexpected response from login: HTTP 500\nerror response"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(*testing.T) {

			f := mockLogin(t, tc.ret)

			// to test insecure https - create a https server (self-signed)
			s := mockServer("/api/management/v1/useradm/auth/login", tc.insecure, f)

			defer s.Close()

			c := NewClient(s.URL, tc.insecure)

			tok, err := c.Login(context.TODO(), "foo", "bar")

			if tc.out == nil {
				assert.Equal(t, "atoken", tok)
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, tc.out, err.Error())
			}

		})
	}
}

func TestClientPreauth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string

		insecure bool

		idData       string
		idDataStruct map[string]interface{}
		pubkey       string
		token        string

		ret int

		out error
	}{
		{
			name: "ok",

			idData: `{"mac": "00:01:02:03"}`,
			pubkey: "key",
			token:  "token",
			idDataStruct: map[string]interface{}{
				"mac": "00:01:02:03",
			},

			ret: 201,
		},
		{
			name: "ok, insecure",

			insecure: true,
			idData:   `{"mac": "00:01:02:03"}`,
			pubkey:   "key",
			token:    "token",
			idDataStruct: map[string]interface{}{
				"mac": "00:01:02:03",
			},

			ret: 201,
		},
		{
			name:   "invalid json input",
			idData: `{mac: 00:01:02:03}`,
			pubkey: "key",
			token:  "token",

			out: errors.New("invalid character 'm' looking for beginning of object key string"),
		},
		{
			name: "conflict",

			idData: `{"mac": "00:01:02:03"}`,
			pubkey: "key",
			token:  "token",
			idDataStruct: map[string]interface{}{
				"mac": "00:01:02:03",
			},

			ret: 409,
			out: ErrPreauthConflict,
		},
		{
			name: "internal",

			idData: `{"mac": "00:01:02:03"}`,
			pubkey: "key",
			token:  "token",
			idDataStruct: map[string]interface{}{
				"mac": "00:01:02:03",
			},

			ret: 500,
			out: errors.New("unexpected response from preauth: HTTP 500\nerror response"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(*testing.T) {
			f := mockPreauth(t, tc.idDataStruct, tc.pubkey, tc.token, tc.ret)

			// to test insecure https - create a https server (self-signed)
			s := mockServer("/api/management/v2/devauth/devices", tc.insecure, f)
			defer s.Close()

			c := NewClient(s.URL, tc.insecure)
			err := c.Preauth(context.TODO(), tc.idData, tc.pubkey, tc.token)

			if tc.out == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.out.Error())
			}

		})
	}
}

func mockServer(url string, https bool, handleFun func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc(url, handleFun)

	var srv *httptest.Server
	if https {
		srv = httptest.NewTLSServer(handler)
	} else {
		srv = httptest.NewServer(handler)
	}
	return srv
}

func mockLogin(t *testing.T, ret int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		u, p, ok := r.BasicAuth()
		assert.Equal(t, "foo", u)
		assert.Equal(t, "bar", p)
		assert.Equal(t, true, ok)

		if ret == http.StatusOK {
			w.Header().Set("Content-Type", "application/jwt")
			w.Write([]byte("atoken"))
		} else {
			w.WriteHeader(ret)
			w.Write([]byte("error response"))
		}
	}
}

func mockPreauth(t *testing.T, iddata map[string]interface{}, key, token string, ret int) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tok := r.Header.Get("Authorization")
		tok = strings.Split(tok, " ")[1]

		assert.Equal(t, token, tok)

		defer r.Body.Close()
		jd := json.NewDecoder(r.Body)

		var pr PreauthReq
		err := jd.Decode(&pr)
		assert.NoError(t, err)

		assert.Equal(t, iddata, pr.IdData)
		assert.Equal(t, key, pr.PubKey)

		w.WriteHeader(ret)
		if ret != http.StatusCreated {
			w.Write([]byte("error response"))
		}
	}
}
