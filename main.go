package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salvatore-081/curt/internal/middlewares"
	"github.com/salvatore-081/curt/pkg/key"
	"github.com/salvatore-081/curt/pkg/models"
)

func main() {
	port := flag.String("PORT", "8080", "server port")
	logLevel := flag.String("LOG_LEVEL", "MISSING", "log level")
	apiKey := flag.String("API_KEY", "", "api key")

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

	dbOptions := badger.DefaultOptions("./data")
	dbOptions = dbOptions.WithLogger(middlewares.BadgerLogger{})
	db, e := badger.Open(dbOptions)
	if e != nil {
		log.Fatal().Str("service", "DB").Err(e).Msg("")
	}
	defer db.Close()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
		again:
			err := db.RunValueLogGC(0.5)
			if err == nil {
				goto again
			}
		}
	}()

	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(middlewares.GinLoggerMiddleware())

	r.GET("/:key", func(c *gin.Context) {
		var v []byte
		e := db.View(func(txn *badger.Txn) error {
			item, e := txn.Get([]byte(c.Param("key")))
			if e != nil {
				return e
			}

			v, e = item.ValueCopy(nil)
			if e != nil {
				return e
			}
			return nil
		})

		if e == nil {
			c.Redirect(http.StatusMovedPermanently, string(v))
			return
		}

		switch e {
		case badger.ErrKeyNotFound:
			log.Info().Msg("not found")
			c.JSON(http.StatusNotFound, map[string]string{
				"message": "not found",
				"details": e.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, map[string]string{
				"message": e.Error(),
			})
		}
	})

	r.POST("/", func(c *gin.Context) {
		var body models.Body
		var header models.Header

		if len(*apiKey) > 0 {
			e := c.ShouldBindHeader(&header)
			if e != nil {
				c.JSON(http.StatusBadRequest, map[string]string{
					"message": e.Error(),
				})
				return
			}

			if header.ApiKey != *apiKey {
				c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "wrong api_key",
				})
				return
			}
		}

		if e := c.ShouldBindJSON(&body); e != nil {
			c.JSON(http.StatusBadRequest, map[string]string{
				"message": e.Error(),
			})
			return
		}

		key := key.RandStringBytesMaskImprSrcUnsafe(7) // TODO

		e := db.Update(func(txn *badger.Txn) error {
			var entry *badger.Entry
			if body.TTL != nil {
				entry = badger.NewEntry([]byte(key), []byte(body.Url)).WithTTL(time.Hour * time.Duration(*body.TTL))
			} else {
				entry = badger.NewEntry([]byte(key), []byte(body.Url))
			}
			e := txn.SetEntry(entry)
			return e
		})
		if e == nil {
			c.JSON(http.StatusCreated, map[string]string{
				"curt": key,
			})
			return
		}

		switch e {
		default:
			c.JSON(http.StatusInternalServerError, map[string]string{
				"message": e.Error(),
			})
		}
	})

	log.Info().Str("service", "CURT").Msg("listening and serving HTTP on " + *port)
	r.Run(":" + *port)
}
