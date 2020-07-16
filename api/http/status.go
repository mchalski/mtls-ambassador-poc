package http

import (
	"github.com/gin-gonic/gin"
)

type StatusController struct{}

func NewStatusController() *StatusController {
	return &StatusController{}
}

func (sc *StatusController) GetStatus(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
