package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ServerCert = "/etc/mtls-ping/certs/server/server.crt"
	ServerKey  = "/etc/mtls-ping/certs/server/server.key"
	TenantPEM  = "/etc/mtls-ping/certs/tenant-ca/tenant.ca.pem"
)

func main() {
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
