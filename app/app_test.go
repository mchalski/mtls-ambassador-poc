// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package app

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"testing"

	mapp "github.com/mendersoftware/mtls-ambassador/app/mocks"
	"github.com/mendersoftware/mtls-ambassador/client/mender"
	mmender "github.com/mendersoftware/mtls-ambassador/client/mender/mocks"

	"github.com/stretchr/testify/assert"
)

func TestAppPreauth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string

		authReq *mender.AuthReq

		authToken string
		authErr   error

		clientErr error

		outErr error
	}{
		{
			name: "ok",

			authReq: &mender.AuthReq{
				IdData:      `{"sn": "0001"}`,
				TenantToken: "tenanttoken",
				PubKey:      "pubkey",
			},

			authToken: "token",
		},
		{
			name: "error, auth provider",

			authReq: &mender.AuthReq{
				IdData:      `{"sn": "0001"}`,
				TenantToken: "tenanttoken",
				PubKey:      "pubkey",
			},

			authErr: errors.New("no token"),

			outErr: errors.New("no token"),
		},
		{
			name: "error, client, preauth conflict",

			authReq: &mender.AuthReq{
				IdData:      `{"sn": "0001"}`,
				TenantToken: "tenanttoken",
				PubKey:      "pubkey",
			},

			clientErr: mender.ErrPreauthConflict,

			outErr: ErrPreauthConflict,
		},
		{
			name: "error, client, generic",

			authReq: &mender.AuthReq{
				IdData:      `{"sn": "0001"}`,
				TenantToken: "tenanttoken",
				PubKey:      "pubkey",
			},

			clientErr: errors.New("client error"),

			outErr: errors.New("client error"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(*testing.T) {
			ctx := context.TODO()

			authProvider := &mapp.AuthProvider{}
			authProvider.On("GetToken").Return(tc.authToken, tc.authErr)

			client := &mmender.Client{}
			if tc.authErr == nil {
				client.On("Preauth",
					ctx,
					tc.authReq.IdData,
					tc.authReq.PubKey,
					tc.authToken).
					Return(tc.clientErr)
			}

			app := NewApp(client, authProvider)

			err := app.Preauth(ctx, tc.authReq)

			if tc.outErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.outErr.Error())
			}

			authProvider.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}

func TestAppVerify(t *testing.T) {
	t.Parallel()

	const rsaCert = `-----BEGIN CERTIFICATE-----
MIIECzCCAfMCAgTSMA0GCSqGSIb3DQEBCwUAMEwxCzAJBgNVBAYTAlVTMRMwEQYD
VQQIDApTb21lLVN0YXRlMRMwEQYDVQQKDApUZW5hbnQgRm9vMRMwEQYDVQQDDApU
ZW5hbnQgRm9vMB4XDTIwMDYxODEyMjUxMloXDTIxMDYxODEyMjUxMlowSjELMAkG
A1UEBhMCVVMxEzARBgNVBAgMClNvbWUtU3RhdGUxEzARBgNVBAoMClRlbmFudCBG
b28xETAPBgNVBAMMCGRldmljZSAxMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAxex+WqphwqgkDfPWJZZtnXHvvRVhG6j3+q45skFyC8Wa0s3Re8TJIOUK
Xwx6YWrl333zqq+KOyiYPcaosVl+y3IzRkT7hNnyExMFEZi2eygg6SINW4QtNIvT
ASaQqn831QyfkaaQLGl1vqNe262l3uabUAYkaDf9Kaz/RbB5jjCse9d28pvSIlPj
ZGYd7sdj2qVSiOMDWh1tiCdr63Xl/oiGMUMU1qlX9Tv2nRxJfzTLplSJh1C5sT5c
qbv7EfxokWHLFYysSOUQLlINphxwP5Yk1M0n4WaU6a0FQNmqN+EtnSjyjvM8VINn
BMLjOVP7N1siax3e/lKG0bxyWR+ThQIDAQABMA0GCSqGSIb3DQEBCwUAA4ICAQCh
Qevo4vlg/sq31OCYv4Xw7vPrS09nw7cqJAWjKr79RmcPGddsq/AJSmj8FEFNP97s
HzXjrx94lORxTkc7BBlezLYBch1XzUTNDE1Og7OM27f87zZAZ1l9K7nGVI1W2ka8
20P664kfgc+0CxKdKXEFJ1Nklqe0RkEMmZpC37MGH5p9akqsWJ6eqcy7+2cNGdCI
LhFDU85uqI5IYIADZqyUy1igPL15DgPY/E/rTzvN40qXF+1eCNb2quu1vnw/6O5m
cVBFrImPvjxfX3Kj8Ifpeea57RNRUDh7NecE3NIHeZlDrfXPvZS4fyLhKONWmYIy
WKOu62grJhroI7IOty5dZAiYslqDVeoH2ZtdriuMqErwywrxYlqDIXGUlmFRAdk5
BUuJAcEhuyBviL/c0RmQrJzZpw1w1kLlN87QZJyv8ecn+ve100WX4uzF6WChOGLS
qpxEHKqrByr3dmrYQ0j1lJXmM6qw7C5sIiQYOB3MmaPj+8QGrK7BCAPtxCjjfC4I
/JQlGWCebPnqLPdjZMuSX751u/WNE94PHbcJbddzexfgmjxGMJAf1Jj79XjQZQEA
3kjT+msDd+MVs8OMT/bzD3fJT5r1YX7naKJAIWmhICXJGYYkIhfTrFTF5Esxj0nB
a9E3EfLsjJiLB3DlfFTXEYqcnpk2VCSlkLchgC7DRw==
-----END CERTIFICATE-----
`
	const rsaKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxex+WqphwqgkDfPWJZZt
nXHvvRVhG6j3+q45skFyC8Wa0s3Re8TJIOUKXwx6YWrl333zqq+KOyiYPcaosVl+
y3IzRkT7hNnyExMFEZi2eygg6SINW4QtNIvTASaQqn831QyfkaaQLGl1vqNe262l
3uabUAYkaDf9Kaz/RbB5jjCse9d28pvSIlPjZGYd7sdj2qVSiOMDWh1tiCdr63Xl
/oiGMUMU1qlX9Tv2nRxJfzTLplSJh1C5sT5cqbv7EfxokWHLFYysSOUQLlINphxw
P5Yk1M0n4WaU6a0FQNmqN+EtnSjyjvM8VINnBMLjOVP7N1siax3e/lKG0bxyWR+T
hQIDAQAB
-----END PUBLIC KEY-----
`

	const rsaPrivKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAxex+WqphwqgkDfPWJZZtnXHvvRVhG6j3+q45skFyC8Wa0s3R
e8TJIOUKXwx6YWrl333zqq+KOyiYPcaosVl+y3IzRkT7hNnyExMFEZi2eygg6SIN
W4QtNIvTASaQqn831QyfkaaQLGl1vqNe262l3uabUAYkaDf9Kaz/RbB5jjCse9d2
8pvSIlPjZGYd7sdj2qVSiOMDWh1tiCdr63Xl/oiGMUMU1qlX9Tv2nRxJfzTLplSJ
h1C5sT5cqbv7EfxokWHLFYysSOUQLlINphxwP5Yk1M0n4WaU6a0FQNmqN+EtnSjy
jvM8VINnBMLjOVP7N1siax3e/lKG0bxyWR+ThQIDAQABAoIBAC35VBRNVW34zn8r
L4gFnCqhR5W9PJRHOGrTQ3WjfBE98kubIVjFig6JBVK0vEyanxC92fbA1bQOJuba
mV6wsiIhwcVFysK+OVuy5E+FEIYk+RgOH9otJq7496dhxOLFsDtdtkcH0J9wU7hX
jHYsrMXM/TCmbJiEwNqIY7dVWbbsMFINxO+9OVnctoZ6/I8IBKcjoZzmeyEt9923
2E6U4ru4bDYsiYQ3QsziTnn0w7Wh8l7/ILTG5vNuWC2WmSoeecTLAuRbjtdIrvWu
DUA4ONBixL9QMD/i22m7N7u6i/e+wKrISQDsg0Qbv91bO0EXSPcFL69HKEIYThFm
I7coVDUCgYEA+aslJgYB+/9w4MTenTek61A97Bt9MTzmDvoH1rGPXpZ5ltrs08RO
jtMFbZH+ktxwhBvsoeMo590C1Vw4yyr0KD61QUHoeCJZpLJjLe8u4Fs+WPF8wZUl
9E6tSpN+nhl6OPT5EM42wQnSSXhKiSBU0L1GJL/2a6FnUKOt1RzzeBMCgYEAyvFr
ld0tZrDaF0VQ7pyn0MIYjQvXvH5dTpWpfvQomKZyDa5o9aLlxmx5Pa79iC537ofx
N8Yy9PQUDLGTSOnhWzECvvHAoyj1F3ZEBI4oDMMjFz30+YsT49hrak+uoZjdpAdn
XRUuSybsLQ1Iu9SlD7Ma0H3AdNatwsWsLzLa6QcCgYBznuxvNW0J1Hvju4gUasZ3
KwviIcDSYo9v9B5ZMJViinD4iZ4PW+O9hMAIxAmO3YNFyuDE/7vb1KARSsoKXHQB
hzjNZcZQjCfTe8EubovY3qh67CqIQ5f2EdFyred/M/FEGz6Up8r3jqLR32E1K8Hb
gSvQrQ1jPrXnxEUmYmfl/QKBgQCsYFdTmeRgX0M/lN7jbiiUhui3lSGPt32lrDWl
4dlBn88sk8IPMmgdHDH3FNXAgEfaUZmwGCdcLJ2DEqnZut5xyLVeXpWTgMx9OzUW
8XBPNshti3CzLVCdrUu/pyLbm65XDvra84y4xLzCn4/yCvKQ3T6fbNC17Ur2L1TL
WlTarQKBgFOdy9HNCIRO2lBk4lakdOMtb0zC2I8+t5MdW6dSan1brNh9epJyOREg
/cy4Y531nKtBp7cJkf7yiWO1N/1/YL688W2jOSswIx/P24XdAQFJj52nCwlA+69S
c+LLPXwUqRuVoW5DaioEOflB86pL8ezacB1gI/TyQCa5T1VqZd84
-----END RSA PRIVATE KEY-----
`
	cases := []struct {
		name      string
		cert      string
		authReq   *mender.AuthReq
		signature string

		privKey string

		err error
	}{
		{
			name: "ok, rsa",
			cert: rsaCert,
			authReq: &mender.AuthReq{
				IdData:      `{"sn": "0001"}`,
				TenantToken: "tenanttoken",
				PubKey:      rsaKey,
			},
			privKey: rsaPrivKey,
		},
		{
			name: "error, no certs",
			cert: "",
			authReq: &mender.AuthReq{
				IdData:      `{"sn": "0001"}`,
				TenantToken: "tenanttoken",
				PubKey:      rsaKey,
			},
			privKey: rsaPrivKey,

			err: ErrCertNum,
		},
		{
			name: "error, key mismatch",
			cert: rsaCert,
			authReq: &mender.AuthReq{
				IdData:      `{"sn": "0001"}`,
				TenantToken: "tenanttoken",
				PubKey:      "somepubkey",
			},
			privKey: rsaPrivKey,

			err: ErrKeyMismatch,
		},
		{
			name: "error, wrong signature",
			cert: rsaCert,
			authReq: &mender.AuthReq{
				IdData:      `{"sn": "0001"}`,
				TenantToken: "tenanttoken",
				PubKey:      rsaKey,
			},
			privKey:   rsaPrivKey,
			signature: "aW52YWxpZA==",

			err: errors.New("crypto/rsa: verification error"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(*testing.T) {
			ctx := context.TODO()
			client := &mmender.Client{}
			app := NewApp(client, nil)

			certs := []*x509.Certificate{}
			if tc.cert != "" {
				block, _ := pem.Decode([]byte(tc.cert))
				assert.NotNil(t, block)

				cert, err := x509.ParseCertificate(block.Bytes)
				assert.NoError(t, err)

				certs = append(certs, cert)
			}

			authReqRaw, err := json.Marshal(tc.authReq)
			assert.NoError(t, err)

			privPem, _ := pem.Decode([]byte(tc.privKey))
			privKey, err := x509.ParsePKCS1PrivateKey(privPem.Bytes)
			assert.NoError(t, err)

			if tc.signature == "" {
				tc.signature = sign(t, authReqRaw, privKey)
			}

			err = app.VerifyClientCert(ctx,
				certs,
				tc.authReq,
				authReqRaw,
				tc.signature)

			if tc.err == nil {
				assert.NoError(t, err)
			}
		})
	}
}

func sign(t *testing.T, data []byte, key crypto.PrivateKey) string {
	hash := sha256.New()
	if _, err := bytes.NewReader(data).WriteTo(hash); err != nil {
		t.Fatal(err)
	}
	digest := hash.Sum(nil)
	signature, err := rsa.SignPKCS1v15(
		rand.Reader, key.(*rsa.PrivateKey), crypto.SHA256, digest,
	)

	assert.NoError(t, err)
	b64 := make([]byte, base64.StdEncoding.EncodedLen(len(signature)))
	base64.StdEncoding.Encode(b64, signature)

	return string(b64)
}
