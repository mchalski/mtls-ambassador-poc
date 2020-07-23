// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type Server struct {
	server   *http.Server
	certFile string
	keyFile  string
}

func NewServer(h http.Handler,
	srvCertFile,
	srvKeyFile,
	tenantCACertFile string,
	port string) (*Server, error) {
	l.Infof("creating server with cert %s and key %s", srvCertFile, srvKeyFile)

	pool, err := certPool(tenantCACertFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create server")
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

	l.Info("creating server: ok")
	return &Server{
		server:   &server,
		certFile: srvCertFile,
		keyFile:  srvKeyFile,
	}, nil
}

func (s *Server) Run() error {
	l.Info("running...")
	return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
}

// certPool prepares a custom cert pool with tenant's CA - to verify client certs against
func certPool(tenantCACertFile string) (*x509.CertPool, error) {
	l.Infof("creating cert pool with tenant CA cert: %s", tenantCACertFile)

	tenantPEM, err := ioutil.ReadFile(
		tenantCACertFile,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cert pool")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(tenantPEM)

	l.Info("creating cert pool: ok")

	return caCertPool, nil
}
