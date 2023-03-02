package main

import (
	"flag"
	"fmt"

	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "github.com/salvatore-081/curt/docs"
	"github.com/salvatore-081/curt/internal"
	"github.com/salvatore-081/curt/internal/controllers"
	"github.com/salvatore-081/curt/internal/middlewares"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/teris-io/shortid"
)

// @title Curt API
// @version 1.1.0
// @contact.name Salvatore Emilio
// @contact.url http://salvatoreemilio.it
// @contact.email @info@salvatoreemilio.it
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host curt.salvatoreemilio.it
// @BasePath /
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
// @securitydefinitions.apikey X-API-Key
// @in header
// @name X-API-Key
func main() {
	port := flag.String("PORT", "8080", "server port")
	logLevel := flag.String("LOG_LEVEL", "MISSING", "log level")
	xAPIKey := flag.String("X_API_KEY", "", "X-API-Key")
	host := flag.String("HOST", "http://localhost:8080", "host")

	flag.Parse()

	logOutput := zerolog.ConsoleWriter{Out: os.Stdout}
	logOutput.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("|%s|", i))
	}

	logOutput.FormatTimestamp = func(i interface{}) string {
		return ""
	}

	log.Logger = zerolog.New(logOutput)

	log.Info().Str("service", "CURT").Msg("starting curt")

	var e error
	var l zerolog.Level

	if *logLevel == "MISSING" {
		log.Info().Str("service", "CURT").Msg("missing log_level, defaulting to DEBUG")
		l = 0
	} else {
		l, e = zerolog.ParseLevel(strings.ToLower(*logLevel))
		if e != nil {
			log.Info().Str("service", "CURT").Err(e).Msg(fmt.Sprintf("unknown log_level: %s, defaulting to DEBUG", *logLevel))
			l = 0
		}
	}
	zerolog.SetGlobalLevel(l)

	var r internal.Resolver
	e = r.Create(*host, *xAPIKey)
	if e != nil {
		log.Fatal().Str("service", "badgerDB").Err(e).Msg("")
	}
	defer r.Close()

	sid, e := shortid.New(1, shortid.DefaultABC, 2342)
	if e != nil {
		log.Fatal().Str("service", "ID").Err(e).Msg("")
	}
	shortid.SetDefault(sid)

	gin.SetMode(gin.ReleaseMode)

	g := gin.New()

	g.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	g.Use(middlewares.GinLoggerMiddleware())

	controllers.C(g.Group("/c"), &r)
	controllers.Status(g.Group("/status"), &r)

	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL(*host+"/swagger/doc.json")))

	log.Info().Str("service", "CURT").Msg("listening and serving HTTP on port " + *port)

	g.Run(":" + *port)
}
