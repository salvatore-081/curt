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
	"github.com/salvatore-081/curt/internal"
	"github.com/salvatore-081/curt/internal/controllers"
	"github.com/salvatore-081/curt/internal/middlewares"
	"github.com/teris-io/shortid"
)

func main() {
	port := flag.String("PORT", "8080", "server port")
	logLevel := flag.String("LOG_LEVEL", "MISSING", "log level")
	apiKey := flag.String("API_KEY", "", "api_key")
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
	e = r.Create(*host, *apiKey)
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
		AllowHeaders:     []string{"Origin", "api_key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	g.Use(middlewares.GinLoggerMiddleware())

	controllers.C(g.Group("/c"), &r)
	controllers.Status(g.Group("/status"), &r)

	log.Info().Str("service", "CURT").Msg("listening and serving HTTP on port " + *port)

	g.Run(":" + *port)
}
