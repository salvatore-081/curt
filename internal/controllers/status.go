package controllers

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/salvatore-081/curt/internal"
	"github.com/salvatore-081/curt/internal/middlewares"
	"github.com/salvatore-081/curt/pkg/models"
)

func Status(g *gin.RouterGroup, r *internal.Resolver) {
	Health(g, r)
	About(g, r)
}

// @Tags status
// @Summary Health check
// @Produce  plain/text
// @Success 200 {string} string	"OK"
// @Failure 500 {object} models.GenericError
// @Router /status/health [get]
// @Security X-API-Key
func Health(g *gin.RouterGroup, r *internal.Resolver) {
	g.GET("/health", middlewares.GinAuthMiddleware(r.XAPIKey), func(c *gin.Context) {
		c.JSON(200, "OK")
	})
}

// @Tags status
// @Summary About
// @Produce  json
// @Success 200 {object} []models.Module
// @Failure 500 {object} models.GenericError
// @Router /status/about [get]
// @Security X-API-Key
func About(g *gin.RouterGroup, r *internal.Resolver) {
	g.GET("/about", middlewares.GinAuthMiddleware(r.XAPIKey), func(c *gin.Context) {
		info, ok := debug.ReadBuildInfo()

		if !ok {
			c.JSON(http.StatusInternalServerError,
				models.GenericError{
					Message: "unable to read build info",
				})
			return
		}

		modules := []models.Module{{
			Path:    "Curt",
			Version: "1.1.0-rc.3",
		}}

		for _, module := range info.Deps {
			modules = append(modules, models.Module{
				Path:    module.Path,
				Version: module.Version,
			})
		}

		c.JSON(http.StatusOK, modules)
	})
}
