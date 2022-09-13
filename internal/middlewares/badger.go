package middlewares

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

type BadgerLogger struct{}

func (l BadgerLogger) Errorf(s string, i ...interface{}) {
	log.Error().Str("service", "DB").Err(errors.New(strings.TrimSuffix(fmt.Sprintf(s, i...), "\n"))).Msg("")
}

func (l BadgerLogger) Warningf(s string, i ...interface{}) {
	log.Warn().Str("service", "DB").Msg(strings.TrimSuffix(fmt.Sprintf(s, i...), "\n"))
}

func (l BadgerLogger) Infof(s string, i ...interface{}) {
	log.Info().Str("service", "DB").Msg(strings.TrimSuffix(fmt.Sprintf(s, i...), "\n"))
}

func (l BadgerLogger) Debugf(s string, i ...interface{}) {
	log.Debug().Str("service", "DB").Msg(strings.TrimSuffix(fmt.Sprintf(s, i...), "\n"))
}
