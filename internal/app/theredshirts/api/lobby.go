package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

const lobby_root_path = "/lobby"
const lobby_id_param = "lobbyId"
const player_id_param = "playerId"

type (
	LobbyCreate struct {
		ID         uuid.UUID
		Name       string
		Owner      uuid.UUID
		Password   string
		Difficulty string
	}

	LobbyUpdate struct {
		Name       string
		Password   string
		Difficulty string
	}

	Lobby struct {
		ID         uuid.UUID
		Name       string
		Owner      *Player
		Password   string
		Difficulty string
		Players    []*Player
	}

	Player struct {
		ID   uuid.UUID
		Name string
	}
)

func initLobbyInterface(group *echo.Group, api *EchoApi) {
	group.POST("", api.createLobbyId)
	group.PUT(":"+lobby_id_param, api.createLobby)
	group.PATCH(":"+lobby_id_param, api.updateLobby)
	group.DELETE(":"+lobby_id_param, api.deleteLobby)
	group.GET("", api.getAllLobbies)
	group.GET(":"+lobby_id_param, api.getLobby)
	group.PUT(":"+lobby_id_param+"/:"+player_id_param, api.joinLobby)
	group.DELETE(":"+lobby_id_param+"/:"+player_id_param, api.leaveLobby)
}

func (api *EchoApi) createLobbyId(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Create lobby Id")
	return context.String(http.StatusCreated, uuid.NewString())
}

func (api *EchoApi) createLobby(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Create lobby")
	return context.String(http.StatusCreated, uuid.NewString())
}

func (api *EchoApi) updateLobby(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Update lobby")
	return context.String(http.StatusCreated, uuid.NewString())
}

func (api *EchoApi) deleteLobby(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Delete lobby")
	return context.String(http.StatusCreated, uuid.NewString())
}

func (api *EchoApi) getAllLobbies(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Get all lobbies")
	return context.String(http.StatusCreated, uuid.NewString())
}

func (api *EchoApi) getLobby(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Get lobby")
	return context.String(http.StatusCreated, uuid.NewString())
}

func (api *EchoApi) joinLobby(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Join lobby")
	return context.String(http.StatusCreated, uuid.NewString())
}

func (api *EchoApi) leaveLobby(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("LEave lobby")
	return context.String(http.StatusCreated, uuid.NewString())
}
