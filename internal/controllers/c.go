package controllers

import (
	"fmt"
	"net/http"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/gin-gonic/gin"
	"github.com/salvatore-081/curt/internal"
	"github.com/salvatore-081/curt/internal/middlewares"
	"github.com/salvatore-081/curt/pkg/models"
	"github.com/teris-io/shortid"
)

func C(g *gin.RouterGroup, r *internal.Resolver) {
	CGet(g, r)
	CPost(g, r)
	CGetKey(g, r)
}

// @Tags c
// @Summary List all Curt(s)
// @Produce  json
// @Success 200 {object} []models.Curt
// @Failure 500 {object} models.GenericError
// @Router /c [get]
// @Security X-API-Key
func CGet(g *gin.RouterGroup, r *internal.Resolver) {
	g.GET("", middlewares.GinAuthMiddleware(r.XAPIKey), func(c *gin.Context) {
		curts := []models.Curt{}

		e := r.BadgerDB.View(func(txn *badger.Txn) error {
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
					curts = append(curts, models.Curt{Url: string(v), Key: string(k), Curt: fmt.Sprintf("%s/c/%s", r.Host, k), TTL: ttl, ExpiresAt: expiresAt})
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
			c.JSON(http.StatusInternalServerError, models.GenericError{
				Message: e.Error(),
			})
		}

	})
}

// @Tags c
// @Summary Create a new Curt
// @Produce  json
// @Success 200 {object} models.Curt
// @Failure 400,500 {object} models.GenericError
// @Router /c [post]
// @Security X-API-Key
func CPost(g *gin.RouterGroup, r *internal.Resolver) {
	g.POST("", middlewares.GinAuthMiddleware(r.XAPIKey), func(c *gin.Context) {
		var body models.Body
		if e := c.ShouldBindJSON(&body); e != nil {
			c.JSON(http.StatusBadRequest,
				models.GenericError{
					Message: e.Error(),
				})
			return
		}

		key, e := shortid.Generate()
		if e != nil {
			c.JSON(http.StatusInternalServerError,
				models.GenericError{
					Message: e.Error(),
				})
			return
		}

		e = r.BadgerDB.Update(func(txn *badger.Txn) error {
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
				"curt": r.Host + "/c/" + key,
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
			c.JSON(http.StatusInternalServerError,
				models.GenericError{
					Message: e.Error(),
				})
		}
	})
}

// @Tags c
// @Summary Follow a Curt redirect
// @Produce  json
// @Success 301
// @Failure 404,500 {object} models.GenericError
// @Router /c/{key} [get]
// @Param key path string true "Curt Key"
// @Security X-API-Key
func CGetKey(g *gin.RouterGroup, r *internal.Resolver) {
	g.GET("/:key", func(c *gin.Context) {
		var v []byte
		e := r.BadgerDB.View(func(txn *badger.Txn) error {
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
			c.JSON(http.StatusNotFound,
				models.GenericError{
					Message: "not found",
					Details: e.Error(),
				})
		default:
			c.JSON(http.StatusInternalServerError,
				models.GenericError{
					Message: e.Error(),
				})
		}
	})
}
