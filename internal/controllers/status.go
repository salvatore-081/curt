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
	g.GET("/am-i-up", middlewares.GinAuthMiddleware(r.ApiKey), func(c *gin.Context) {
		c.JSON(200, "OK")
	})

	g.GET("/about", middlewares.GinAuthMiddleware(r.ApiKey), func(c *gin.Context) {
		info, ok := debug.ReadBuildInfo()

		if !ok {
			c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "unable to read build info",
			})
			return
		}

		modules := []models.Module{{
			Path:    "Curt",
			Version: "1.1.0-RC.1",
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
