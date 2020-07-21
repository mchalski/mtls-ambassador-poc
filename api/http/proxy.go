// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/mendersoftware/go-lib-micro/log"
	"github.com/pkg/errors"

	"github.com/mendersoftware/mtls-ambassador/app"
	"github.com/mendersoftware/mtls-ambassador/client/mender"
)

const (
	UrlDevauthAuthReq = "/api/devices/v1/authentication/auth_requests"
)

var (
	l = log.NewEmpty()
)

// ProxyController proxies device API requests to Mender
// while handling automatic device preauth on POST /auth_requests
type ProxyController struct {
	app   app.App
	proxy Proxy
}

type Proxy interface {
	Redirect(w http.ResponseWriter, r *http.Request)
}

type proxy struct {
	proxy *httputil.ReverseProxy
}

func NewProxy(menderUrl string) (*proxy, error) {
	l.Infof("creating proxy with url %s", menderUrl)
	u, err := url.Parse(menderUrl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create proxy")
	}

	l.Infof("proxy scheme: %s, host: %s", u.Scheme, u.Host)
	director := func(req *http.Request) {
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
	}

	l.Info("creating proxy: ok")
	return &proxy{
		proxy: &httputil.ReverseProxy{Director: director},
	}, nil
}

func (p *proxy) Redirect(w http.ResponseWriter, r *http.Request) {
	l.Debug("proxy redirection called")
	p.proxy.ServeHTTP(w, r)
}

func NewProxyController(app app.App, proxy Proxy) *ProxyController {

	return &ProxyController{
		app:   app,
		proxy: proxy,
	}
}

func (pc *ProxyController) Any(c *gin.Context) {
	if c.Request.URL.Path == UrlDevauthAuthReq {
		l.Debug("auth request intercepted")

		l.Debug("parsing auth request")
		authreq, raw, err := parseAuthReq(c.Request)
		if err != nil {
			l.Errorf("parsing auth request failed: %s", err.Error())
			c.Writer.WriteHeader(http.StatusBadRequest)
			return
		}
		l.Debug("parsing auth request: ok")

		l.Debug("verifying client cert")
		err = pc.app.VerifyClientCert(c,
			c.Request.TLS.PeerCertificates,
			authreq,
			raw,
			c.Request.Header.Get("X-MEN-Signature"))
		if err != nil {
			l.Errorf("verifying client cert failed: %s", err.Error())
			c.Writer.WriteHeader(http.StatusBadRequest)
			return
		}
		l.Debug("verifying client cert: ok")

		l.Debug("preauthorizing")
		err = pc.app.Preauth(c, authreq)
		if err != nil && err != app.ErrPreauthConflict {
			l.Errorf("preauthorization failed: %s", err.Error())
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err == app.ErrPreauthConflict {
			l.Info("preauthorization conflict detected, but it's ok, proceeding ")
		}

		l.Debug("preauthorizing: ok")
	}

	l.Debug("proxying...")
	pc.proxy.Redirect(c.Writer, c.Request)
}

// parseAuthReq parses the auth request and for convenience
// returns also the raw body for further verification
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
