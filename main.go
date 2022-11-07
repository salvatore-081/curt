package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salvatore-081/curt/internal/middlewares"
	"github.com/salvatore-081/curt/pkg/models"
	"github.com/teris-io/shortid"
)

func main() {
	port := flag.String("PORT", "8080", "server port")
	logLevel := flag.String("LOG_LEVEL", "MISSING", "log level")
	apiKey := flag.String("API_KEY", "", "api key")
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

	dbOptions := badger.DefaultOptions("./data")
	dbOptions = dbOptions.WithLogger(middlewares.BadgerLogger{})
	db, e := badger.Open(dbOptions)
	if e != nil {
		log.Fatal().Str("service", "DB").Err(e).Msg("")
	}
	defer db.Close()

	sid, e := shortid.New(1, shortid.DefaultABC, 2342)
	if e != nil {
		log.Fatal().Str("service", "ID").Err(e).Msg("")
	}
	shortid.SetDefault(sid)

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

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "api_key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(middlewares.GinLoggerMiddleware())

	r.GET("/c", func(c *gin.Context) {
		log.Info().Msg("/c")
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

		curts := []models.Curt{}

		e := db.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.AllVersions = false
			opts.PrefetchSize = 10
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				var ttl *uint16
				var expiresAt *uint64
				if item.ExpiresAt() > 0 {
					ttl = new(uint16)
					expiresAt = new(uint64)
					*expiresAt = item.ExpiresAt()
					*ttl = uint16(time.Until(time.Unix(int64(*expiresAt), 0)).Hours())
				}
				k := item.Key()
				e := item.Value(func(v []byte) error {
					curts = append(curts, models.Curt{Url: string(v), Key: string(k), Curt: fmt.Sprintf("%sc/%s", *host, k), TTL: ttl, ExpiresAt: expiresAt})
					return nil
				})
				if e != nil {
					return e
				}
			}
			return nil
		})

		if e == nil {
			c.JSON(http.StatusOK, curts)
			return
		}

		switch e {
		default:
			c.JSON(http.StatusInternalServerError, map[string]string{
				"message": e.Error(),
			})
		}

	})

	r.GET("/c/:key", func(c *gin.Context) {
		log.Info().Msg("/c/:key")

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

	r.POST("/c/", func(c *gin.Context) {
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

		key, e := shortid.Generate()
		if e != nil {
			c.JSON(http.StatusInternalServerError, map[string]string{
				"message": e.Error(),
			})
			return
		}

		e = db.Update(func(txn *badger.Txn) error {
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
			response := map[string]string{
				"key":  key,
				"curt": *host + "/c/" + key,
				"url":  body.Url,
			}
			if body.TTL != nil {
				response["TTL"] = fmt.Sprintf("%d", *body.TTL)
			}
			c.JSON(http.StatusCreated, response)
			return
		}

		switch e {
		default:
			c.JSON(http.StatusInternalServerError, map[string]string{
				"message": e.Error(),
			})
		}
	})

	log.Info().Str("service", "CURT").Msg("listening and serving HTTP on port " + *port)
	r.Run(":" + *port)
}
