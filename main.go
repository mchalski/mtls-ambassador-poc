package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	ServerCert = "/etc/mtls-ping/certs/server/server.crt"
	ServerKey  = "/etc/mtls-ping/certs/server/server.key"
	TenantPEM  = "/etc/mtls-ping/certs/tenant-ca/tenant.ca.pem"
)

var (
	MenderBackend = "staging.hosted.mender.io:443"
)

func main() {
	config()

	certPool, err := certPool()
	if err != nil {
		panic(err)
	}

	r := router()

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

func router() *gin.Engine {
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

	r.Any("/api/*path", gin.WrapH(proxy))

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
}
