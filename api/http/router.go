package http

import (
	"github.com/gin-gonic/gin"

	"github.com/mendersoftware/mtls-ambassador-poc/app"
)

const (
	ApiUrlStatus = "/status"
	ApiUrlProxy  = "/api/devices/*path"
)

func NewRouter(app app.App, proxy Proxy) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	//TODO use custom Mender-compliant logger
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	status := NewStatusController()
	router.GET(ApiUrlStatus, status.GetStatus)

	proxyController := NewProxyController(app, proxy)
	router.Any(ApiUrlProxy, proxyController.Any)

	return router, nil
}
