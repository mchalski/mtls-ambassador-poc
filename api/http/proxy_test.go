// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package http

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mtls-ambassador/app"
	mapp "github.com/mendersoftware/mtls-ambassador/app/mocks"
	"github.com/mendersoftware/mtls-ambassador/client/mender"
)

func TestProxyController(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string

		inUrl  string
		inBody []byte
		inHdr  map[string]string
		// corresponds to inBody, if valid
		authReq *mender.AuthReq

		appPreauthErr error
		appVerifyErr  error

		willProxy   bool
		proxyStatus int

		outStatus int
	}{
		{
			name:   "ok, auth request 1",
			inUrl:  "/api/devices/v1/authentication/auth_requests",
			inBody: []byte(`{"id_data": "{\"sn\": \"0001\"}", "pubkey": "foo", "tenant_token": "token"}`),
			inHdr: map[string]string{
				"X-MEN-Signature": "signature",
				"X-MEN-RequestID": "reqid",
			},
			authReq: &mender.AuthReq{
				IdData:      `{"sn": "0001"}`,
				PubKey:      "foo",
				TenantToken: "token",
			},

			willProxy: true,

			proxyStatus: 200,

			outStatus: 200,
		},
		{
			name:   "ok, auth request 2",
			inUrl:  "/api/devices/v1/authentication/auth_requests",
			inBody: []byte(`{"id_data": "{\"mac\": \"00:00:00:01\"}", "pubkey": "bar", "tenant_token": "token-bar"}`),
			inHdr: map[string]string{
				"X-MEN-Signature": "signature",
				"X-MEN-RequestID": "reqid",
			},
			authReq: &mender.AuthReq{
				IdData:      `{"mac": "00:00:00:01"}`,
				PubKey:      "bar",
				TenantToken: "token-bar",
			},
			willProxy: true,

			proxyStatus: 200,

			outStatus: 200,
		},
		{
			name:   "ok, auth request, with preauth conflict",
			inUrl:  "/api/devices/v1/authentication/auth_requests",
			inBody: []byte(`{"id_data": "{\"sn\": \"0001\"}", "pubkey": "foo", "tenant_token": "token"}`),
			inHdr: map[string]string{
				"X-MEN-Signature": "signature",
				"X-MEN-RequestID": "reqid",
			},
			authReq: &mender.AuthReq{
				IdData:      `{"sn": "0001"}`,
				PubKey:      "foo",
				TenantToken: "token",
			},

			appPreauthErr: app.ErrPreauthConflict,

			willProxy: true,

			proxyStatus: 200,

			outStatus: 200,
		},
		{
			name:   "error, auth request, parse",
			inUrl:  "/api/devices/v1/authentication/auth_requests",
			inBody: []byte(`{"id_data": {\"mac\": \"00:00:00:01\"}, "pubkey": "bar", "tenant_token": "token-bar"}`),
			inHdr: map[string]string{
				"X-MEN-Signature": "signature",
				"X-MEN-RequestID": "reqid",
			},

			willProxy: false,

			outStatus: 400,
		},
		{
			name:   "error, auth request, verify cert",
			inUrl:  "/api/devices/v1/authentication/auth_requests",
			inBody: []byte(`{"id_data": "{\"sn\": \"0001\"}", "pubkey": "foo", "tenant_token": "token"}`),
			inHdr: map[string]string{
				"X-MEN-Signature": "signature",
				"X-MEN-RequestID": "reqid",
			},
			authReq: &mender.AuthReq{
				IdData:      `{"sn": "0001"}`,
				PubKey:      "foo",
				TenantToken: "token",
			},

			appVerifyErr: errors.New("cert verification failed"),

			willProxy: false,

			outStatus: 400,
		},
		{
			name:   "ok, other device api",
			inUrl:  "/api/devices/not/an/auth/req",
			inBody: []byte(`{"foo": "foo-val", "bar": "bar-val"}`),
			inHdr: map[string]string{
				"X-MEN-RequestID": "reqid",
				"X-MEN-Other":     "other",
			},

			willProxy: true,

			proxyStatus: 200,

			outStatus: 200,
		},
		{
			name:   "error, request outside of device api",
			inUrl:  "/api/management/not/a/device/api",
			inBody: []byte(`{"foo": "foo-val", "bar": "bar-val"}`),
			inHdr: map[string]string{
				"X-MEN-RequestID": "reqid",
			},

			willProxy: false,

			outStatus: 404,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(*testing.T) {

			app := &mapp.App{}

			if tc.authReq != nil {
				app.On("VerifyClientCert",
					mock.AnythingOfType("*gin.Context"),
					mock.AnythingOfType("[]*x509.Certificate"),
					tc.authReq,
					tc.inBody,
					tc.inHdr["X-MEN-Signature"]).
					Return(tc.appVerifyErr)
				if tc.appVerifyErr == nil {
					app.On("Preauth",
						mock.AnythingOfType("*gin.Context"),
						tc.authReq).
						Return(tc.appPreauthErr)
				}
			}

			proxy := &mockProxy{
				url:    tc.inUrl,
				body:   tc.inBody,
				hdr:    tc.inHdr,
				status: tc.proxyStatus,
				t:      t,
			}

			r, err := NewRouter(app, proxy)
			assert.NoError(t, err)
			server := mockServer(tc.inUrl, r.ServeHTTP)

			defer server.Close()

			t.Logf("FRONTEND URL %v\n", server.URL[len("https://"):])
			req, _ := http.NewRequest("POST", server.URL+tc.inUrl, bytes.NewReader(tc.inBody))

			for k, v := range tc.inHdr {
				req.Header.Add(k, v)
			}

			client := server.Client()

			res, err := client.Do(req)
			assert.NoError(t, err)

			assert.Equal(t, tc.outStatus, res.StatusCode)

			if tc.willProxy {
				assert.True(t, proxy.called)
			}

			app.AssertExpectations(t)
		})
	}
}

func mockServer(url string, handleFun func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc(url, handleFun)

	srv := httptest.NewUnstartedServer(handler)
	srv.StartTLS()

	return srv
}

type mockProxy struct {
	url    string
	body   []byte
	hdr    map[string]string
	status int

	called bool

	t *testing.T
}

func (mp *mockProxy) Redirect(w http.ResponseWriter, r *http.Request) {
	for k, v := range mp.hdr {
		assert.Equal(mp.t, r.Header.Get(k), v)
	}

	assert.Equal(mp.t, mp.url, r.URL.Path)

	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	assert.NoError(mp.t, err)

	assert.Equal(mp.t, mp.body, data)

	w.WriteHeader(mp.status)

	mp.called = true
}
