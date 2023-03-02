package middlewares

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salvatore-081/curt/pkg/models"
)

func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		c.Next()
		t := time.Since(now).String()

		var event *zerolog.Event
		statusCode := c.Writer.Status()

		switch {
		case statusCode >= 400 && statusCode < 500:
			event = log.Warn()
		case statusCode >= 500:
			event = log.Error()
		default:
			event = log.Info()
		}

		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery
		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		for _, e := range c.Errors {
			event = event.Err(e.Err)
		}

		event.Str("service", "API").Str("method", c.Request.Method).Str("path", path).Int("status", statusCode).Str("client ip", c.ClientIP()).Str("response time", t).Msg("")
	}
}

func GinAuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var header models.Header

		if len(apiKey) > 0 {
			e := c.ShouldBindHeader(&header)
			if e != nil {
				c.JSON(http.StatusBadRequest, map[string]string{
					"message": e.Error(),
				})
				c.Abort()
				return
			}

			if header.ApiKey != apiKey {
				c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "wrong api_key",
				})
				c.Abort()
				return
			}
		}
	}
}
