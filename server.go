// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
)

type Server struct {
	server   http.Server
	certFile string
	keyFile  string
}

func NewServer(h http.Handler,
	srvCertFile,
	srvKeyFile,
	tenantCACertFile string,
	port string) (*Server, error) {
	pool, err := certPool(tenantCACertFile)
	if err != nil {
		return nil, err
	}

	// custom TLSConfig - enables client cert verification against a custom CA
	server := http.Server{
		Addr:    ":" + port,
		Handler: h,
		TLSConfig: &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	}

	return &Server{
		server:   server,
		certFile: srvCertFile,
		keyFile:  srvKeyFile,
	}, nil
}

func (s *Server) Run() error {
	return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
}

// certPool prepares a custom cert pool with tenant's CA - to verify client certs against
func certPool(tenantCACertFile string) (*x509.CertPool, error) {
	tenantPEM, err := ioutil.ReadFile(
		tenantCACertFile,
	)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(tenantPEM)
	return caCertPool, nil
}
