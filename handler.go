package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mtls-ambassador/mender"
)

func NewHandler(apiClient *mender.Client, c *Config, mgmtToken string) *gin.Engine {
	r := gin.Default()

	// just a dummy healthcheck
	r.GET("/ping", handlePing)

	// actual Mender device API proxy
	handleMenderApi := setupMenderApiHandler(apiClient, c, mgmtToken)
	r.Any("/api/devices/*path", handleMenderApi)

	return r
}

func handlePing(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func setupMenderApiHandler(apiClient *mender.Client, config *Config, mgmtToken string) func(c *gin.Context) {
	// standard reverse proxy deals with all of the 1:1 api calls redirection
	director := func(req *http.Request) {
		req.URL.Scheme = "https"
		req.URL.Host = config.MenderBackend
	}
	proxy := &httputil.ReverseProxy{Director: director}

	// we have to just intercept POST /auth_requests to preauth the device first
	handler := func(c *gin.Context) {
		if c.Request.URL.Path == "/api/devices/v1/authentication/auth_requests" {
			log.Println("intercepted POST /auth_requests")
			authreq, raw, err := parseAuthReq(c.Request)
			if err != nil {

				log.Printf("failed to parse auth request: %v\n", err)
				c.Writer.WriteHeader(http.StatusBadRequest)
				return
			}

			log.Println("client cert details:")
			tlsPubkey := examineClientCert(c.Request)

			log.Println("verifying client key")
			err = verifyClientKey(tlsPubkey, authreq, raw, c.Request.Header.Get("X-MEN-Signature"))
			if err != nil {
				log.Printf("client key verification error: %v\n", err)
				c.Writer.WriteHeader(http.StatusBadRequest)
				return
			}
			log.Println("verifying client key: success")

			log.Println("preauthorizing")
			err = apiClient.Preauth(
				authreq.IdData,
				authreq.PubKey,
				config.MenderBackend,
				mgmtToken,
			)
			if err != nil {
				if mender.ErrIsPreauthConflict(err) {
					log.Printf("preauthorizing: conflict detected (but it's ok) %v\n", err)
				} else {
					log.Printf("preauthorizing: fatal error %v\n", err)
					c.Writer.WriteHeader(http.StatusInternalServerError)
				}
			} else {
				log.Println("preauthorizing: success")
			}

			log.Println("proxying auth request to Mender")
		}

		// rely on the proxy to forward the actual request
		proxy.ServeHTTP(c.Writer, c.Request)
	}

	return handler
}

// parseAuthReq parses the auth request and for convenience
// return also the raw read body for further verification
func parseAuthReq(r *http.Request) (*mender.AuthReq, []byte, error) {
	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	authreq := mender.AuthReq{}
	err = json.Unmarshal(data, &authreq)

	if err != nil {
		return nil, nil, err
	}

	// restore request body - will be redirected further
	r.Body = ioutil.NopCloser(bytes.NewReader(data))

	return &authreq, data, nil
}

// examineClientCert inspects incoming TLS cert and retrieves the pubkey
// returns: public key, error
func examineClientCert(r *http.Request) interface{} {
	// technically, more than 1 cert can be present - dump some infoon all
	for _, cert := range r.TLS.PeerCertificates {
		dumpCert(cert)
	}

	// really we're expecting just one cert, get its key
	return r.TLS.PeerCertificates[0].PublicKey.(*rsa.PublicKey)

}

//verifyClientKey has some ideas on additionally verifying the incoming tls key:
// - compare its string against the one in auth request
// - verify the signature
// - (could also verify the signature with the auth req key)
// in short, it should do what deviceauth does with each auth request
func verifyClientKey(key interface{}, authReq *mender.AuthReq, authReqRaw []byte, signature string) error {
	keystr, err := SerializePubKey(key)
	if err != nil {
		return err
	}

	log.Printf("client key: %s\n", keystr)

	if keystr == authReq.PubKey {
		log.Printf("client key matches auth req key")
	} else {
		log.Printf("error: client key doesn't match auth request key: %s\n", authReq.PubKey)
		return err
	}

	return VerifyAuthReqSign(signature, key, authReqRaw)
}
