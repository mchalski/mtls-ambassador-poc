package main

import (
	"bytes"
	"crypto"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mtls-ambassador-poc/mender"
)

const (
	ServerCert = "/etc/mtls/certs/server/server.crt"
	ServerKey  = "/etc/mtls/certs/server/server.key"
	TenantPEM  = "/etc/mtls/certs/tenant-ca/tenant.ca.pem"
)

var (
	MenderBackend   = "staging.hosted.mender.io:443"
	MenderUser      = "mtls@mender.io"
	MenderPass      = ""
	MenderMgmtToken = ""
)

func main() {
	config()

	log.Println("logging in to Mender")
	menderClient := mender.NewClient()
	token, err := menderClient.Login(MenderUser, MenderPass, MenderBackend)

	if err != nil {
		panic(err)
	}

	MenderMgmtToken = token

	log.Println("logging in to Mender: success")

	certPool, err := certPool()
	if err != nil {
		panic(err)
	}

	r := router(menderClient)

	server := http.Server{
		Addr:    ":8080",
		Handler: r,
		TLSConfig: &tls.Config{
			ClientCAs:  certPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	}

	err = server.ListenAndServeTLS(ServerCert, ServerKey)
	if err != nil {
		panic(err)
	}
}

func router(apiClient *mender.Client) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	director := func(req *http.Request) {
		req.URL.Scheme = "https"
		req.URL.Host = MenderBackend
	}
	proxy := &httputil.ReverseProxy{Director: director}

	menderApiHandler := func(c *gin.Context) {
		if c.Request.URL.Path == "/api/devices/v1/authentication/auth_requests" {
			log.Println("intercepting auth request")

			data, err := ioutil.ReadAll(c.Request.Body)
			defer c.Request.Body.Close()

			if err != nil {
				log.Printf("failed to read auth request: %v\n", err)
				c.Writer.WriteHeader(http.StatusBadRequest)
				return
			}

			authreq := mender.AuthReq{}
			err = json.Unmarshal(data, &authreq)
			if err != nil {
				log.Printf("failed to parse auth request: %v\n", err)
				c.Writer.WriteHeader(http.StatusBadRequest)
				return
			}

			c.Request.Body = ioutil.NopCloser(bytes.NewReader(data))

			// peek cert - verify key match
			for _, cert := range c.Request.TLS.PeerCertificates {
				dumpCert(cert)
			}

			// really we're expecting just one cert
			clientKey := c.Request.TLS.PeerCertificates[0].PublicKey.(*rsa.PublicKey)
			clientKeyStr, err := SerializePubKey(clientKey)
			if err != nil {
				log.Printf("client key parsing error: %s\n", err)
				c.Writer.WriteHeader(http.StatusInternalServerError)
			}

			log.Printf("client key: %s\n", clientKeyStr)

			if clientKeyStr == authreq.PubKey {
				log.Printf("client key matches auth req key")
			} else {
				log.Printf("warning: client key doesn't match auth request key: %s\n", authreq.PubKey)
			}

			// verify signature with the client cert
			err = VerifyAuthReqSign(c.Request.Header.Get("X-MEN-Signature"), clientKey, data)
			if err != nil {
				log.Printf("auth req signature invalid!")
			} else {
				log.Printf("auth req signature verified")
			}

			err = apiClient.Preauth(
				authreq.IdData,
				authreq.PubKey,
				MenderBackend,
				MenderMgmtToken,
			)
			if err != nil {
				if mender.ErrIsPreauthConflict(err) {
					log.Printf("preauth conflict detected, proceeding: %v\n", err)
				} else {
					log.Printf("general preauth error: %v\n", err)
					c.Writer.WriteHeader(http.StatusInternalServerError)
				}
			}

		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}

	r.Any("/api/*path", menderApiHandler)

	return r
}

func certPool() (*x509.CertPool, error) {
	tenantPEM, err := ioutil.ReadFile(TenantPEM)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(tenantPEM)
	return caCertPool, nil
}

func config() {
	backend := os.Getenv("MTLS_PING_MENDER_BACKEND")
	if backend != "" {
		MenderBackend = backend
	}

	MenderUser = os.Getenv("MTLS_PING_MENDER_USER")
	if MenderUser == "" {
		panic("provide MTLS_PING_MENDER_USER")
	}

	MenderPass = os.Getenv("MTLS_PING_MENDER_PASS")
	if MenderPass == "" {
		panic("provide MTLS_PING_MENDER_PASS")
	}
}

func dumpCert(cert *x509.Certificate) {
	log.Printf("subject %s\n", cert.Subject.String())
	log.Printf("issuer %s\n", cert.Issuer.String())
}

func SerializePubKey(key interface{}) (string, error) {

	switch key.(type) {
	case *rsa.PublicKey, *dsa.PublicKey, *ecdsa.PublicKey:
		break
	default:
		return "", errors.New("unrecognizable public key type")
	}

	asn1, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return "", err
	}

	out := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1,
	})

	if out == nil {
		return "", err
	}

	return string(out), nil
}

func VerifyAuthReqSign(signature string, pubkey interface{}, content []byte) error {
	hash := sha256.New()
	_, err := bytes.NewReader(content).WriteTo(hash)
	if err != nil {
		return err
	}

	decodedSig, err := base64.StdEncoding.DecodeString(string(signature))
	if err != nil {
		return err
	}

	key := pubkey.(*rsa.PublicKey)

	err = rsa.VerifyPKCS1v15(key, crypto.SHA256, hash.Sum(nil), decodedSig)
	if err != nil {
		return err
	}

	return nil
}
