package http

import (
	"github.com/gin-gonic/gin"
)

const (
	ApiUrlStatus = "/status"
)

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	//TODO use custom Mender-compliant logger
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	status := NewStatusController()
	router.GET(ApiUrlStatus, status.GetStatus)

	return router
}
