package api

import (
	"net/http"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/labstack/echo/v4"
)

const server_root_path = "/server"

func initServerInterface(group *echo.Group, api *EchoApi) {
	group.GET("/ping", api.ping)
}

func (api *EchoApi) ping(context echo.Context) error {
	logger := context.Get(context_key).(*util.Context).Logger
	logger.Debug("Ping server")
	return context.NoContent(http.StatusNoContent)
}
