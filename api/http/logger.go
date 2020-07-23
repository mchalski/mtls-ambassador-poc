// Copyright 2020 Northern.tech AS
//
//    All Rights Reserved

package http

import (
	"fmt"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const typeHTTP = "http"

func routerLogger(logger logrus.FieldLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := math.Ceil(float64(stop.Nanoseconds())) / 1000000.0
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		entry := logger.WithFields(logrus.Fields{
			"clientip":     clientIP,
			"type":         typeHTTP,
			"ts":           start.Round(0),
			"status":       statusCode,
			"responsetime": latency,
			"byteswritten": dataLength,
			"method":       method,
			"path":         path,
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%d %f %s %s %s - %s", statusCode, latency, method, path, clientIP, clientUserAgent)
			if statusCode > 499 {
				entry.Error(msg)
			} else if statusCode > 399 {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}
