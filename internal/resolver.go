package internal

import (
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/rs/zerolog/log"
	"github.com/salvatore-081/curt/internal/middlewares"
)

type Resolver struct {
	Host     string
	ApiKey   string
	BadgerDB *badger.DB
}

func (r *Resolver) Create(host string, apiKey string) (e error) {
	r.Host = host
	r.ApiKey = apiKey

	r.BadgerDB, e = badger.Open(badger.DefaultOptions("./data").WithLogger(middlewares.BadgerLogger{}))
	if e != nil {
		return e
	}

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			e := r.BadgerDB.RunValueLogGC(0.5)
			if e != nil {
				log.Debug().Err(e).Str("service", "badgerDB").Msg("")
			}
		}
	}()

	return nil
}

func (r *Resolver) Close() error {
	return r.BadgerDB.Close()
}
