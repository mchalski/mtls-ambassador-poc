package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
)

// predefined paths - mount in container to override
const (
	// regular server's SSL cert
	ServerCert = "/etc/mtls/certs/server/server.crt"
	ServerKey  = "/etc/mtls/certs/server/server.key"
	// tenant's CA cert, used to verify client certs in mTLS handshake
	TenantPEM = "/etc/mtls/certs/tenant-ca/tenant.ca.pem"

	// internally we're always on 8080 - override via docker/k8s
	Port = "8080"
)

type Server struct {
	server http.Server
}

func NewServer(c *Config, h http.Handler) (*Server, error) {
	pool, err := certPool()
	if err != nil {
		return nil, err
	}

	// custom TLSConfig - enables client cert verification against a custom CA
	server := http.Server{
		Addr:    ":" + Port,
		Handler: h,
		TLSConfig: &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	}

	return &Server{
		server: server,
	}, nil
}

func (s *Server) Run() error {
	return s.server.ListenAndServeTLS(ServerCert, ServerKey)
}

// certPool prepares a custom cert pool with tenant's CA - to verify client certs against
func certPool() (*x509.CertPool, error) {
	tenantPEM, err := ioutil.ReadFile(TenantPEM)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(tenantPEM)
	return caCertPool, nil
}
