package controllers

import (
	"fmt"
	"net/http"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/gin-gonic/gin"
	"github.com/salvatore-081/curt/internal"
	"github.com/salvatore-081/curt/pkg/models"
	"github.com/teris-io/shortid"
)

func C(g *gin.RouterGroup, r *internal.Resolver) {
	g.GET("", func(c *gin.Context) {
		var header models.Header

		if len(r.ApiKey) > 0 {
			e := c.ShouldBindHeader(&header)
			if e != nil {
				c.JSON(http.StatusBadRequest, map[string]string{
					"message": e.Error(),
				})
				return
			}

			if header.ApiKey != r.ApiKey {
				c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "wrong api_key",
				})
				return
			}
		}

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
			c.JSON(http.StatusInternalServerError, map[string]string{
				"message": e.Error(),
			})
		}

	})

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

	g.POST("", func(c *gin.Context) {
		var body models.Body
		var header models.Header

		if len(r.ApiKey) > 0 {
			e := c.ShouldBindHeader(&header)
			if e != nil {
				c.JSON(http.StatusBadRequest, map[string]string{
					"message": e.Error(),
				})
				return
			}

			if header.ApiKey != r.ApiKey {
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
			c.JSON(http.StatusInternalServerError, map[string]string{
				"message": e.Error(),
			})
		}
	})
}