// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package http

import (
	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/go-lib-micro/log"

	"github.com/mendersoftware/mtls-ambassador/app"
)

const (
	ApiUrlStatus = "/status"
	ApiUrlProxy  = "/api/devices/*path"
)

func NewRouter(app app.App, proxy Proxy) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()

	l := log.NewEmpty()

	router := gin.New()

	//TODO use custom Mender-compliant logger
	router.Use(routerLogger(l))
	router.Use(gin.Recovery())

	status := NewStatusController()
	router.GET(ApiUrlStatus, status.GetStatus)

	proxyController := NewProxyController(app, proxy)
	router.Any(ApiUrlProxy, proxyController.Any)

	return router, nil
}
