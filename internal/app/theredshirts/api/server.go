package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

const server_root_path = "/server"

func initServerInterface(group *echo.Group, api *EchoApi) {
	group.GET("/ping", api.ping)
}

func (api *EchoApi) ping(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Ping server")
	return context.NoContent(http.StatusNoContent)
}
