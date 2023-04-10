package api

import (
	"fmt"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/core"
	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
)

const (
	context_key           = "context"
	correlation_id_header = "X-Correlation-ID"
)

type (
	CustomValidator struct {
		validator *validator.Validate
	}
	EchoApi struct {
		core core.Core
	}
	Api interface {
	}
)

func NewApi() (Api, error) {
	core, err := core.NewCore()
	if err != nil {
		return nil, fmt.Errorf("error while creating core layer: %v", err)
	}

	echoApi := &EchoApi{core: core}
	e := echo.New()
	e.HideBanner = true
	e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
	e.Use(middleware.CORS(), setContextMiddleware, middleware.Recover())
	e.Validator = &CustomValidator{validator: validator.New()}

	serverGroup := e.Group(server_root_path)
	initServerInterface(serverGroup, echoApi)

	lobbyGroup := e.Group(lobby_root_path)
	initLobbyInterface(lobbyGroup, echoApi)

	playerGroup := e.Group(player_root_path)
	initPlayerInterface(playerGroup, echoApi)

	prom := prometheus.NewPrometheus("lobby", nil)
	prom.Use(e)

	address := util.GetEnvWithFallback("ADDRESS", "0.0.0.0")
	port, err := util.GetEnvIntWithFallback("PORT", 1203)
	if err != nil {
		return nil, fmt.Errorf("error while loading port from environment variable: %w", err)
	}
	url := fmt.Sprintf("%s:%d", address, port)
	e.Logger.Fatal(e.Start(url))

	return echoApi, nil
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func setContextMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		correlationId := c.Request().Header.Get(correlation_id_header)
		_, err := uuid.Parse(correlationId)
		if err != nil {
			log.Infof("Correlation id is not from format uuid. Set default correlation id. Error: %v", err)
			correlationId = "WRONG FORMAT"
		}
		logger := log.WithFields(log.Fields{
			correlation_id_header: correlationId,
		})

		c.Set(context_key, &util.Context{CorrelationId: correlationId, Logger: logger})
		return next(c)
	}
}
