package util

import (
	log "github.com/sirupsen/logrus"
)

type Context struct {
	CorrelationId string
	Logger        *log.Entry
}
