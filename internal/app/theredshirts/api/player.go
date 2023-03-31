package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/core"
	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	player_root_path = "/player"
	player_id_param  = "playerId"
)

type (
	CreatePlayer struct {
		ID       uuid.UUID              `param:"playerId" validate:"required"`
		Name     string                 `json:"name" validate:"required"`
		LobbyId  uuid.UUID              `json:"lobby_id" validate:"required"`
		Password string                 `json:"password"`
		Payload  map[string]interface{} `json:"payload"`
	}

	UpdatePlayer struct {
		ID      uuid.UUID              `param:"playerId" validate:"required"`
		Name    string                 `json:"name" validate:"required"`
		Payload map[string]interface{} `json:"payload"`
	}

	Player struct {
		ID      uuid.UUID              `json:"id" validate:"required"`
		Name    string                 `json:"name" validate:"required"`
		Payload map[string]interface{} `json:"payload"`
	}
)

func initPlayerInterface(group *echo.Group, api *EchoApi) {
	group.PUT("/:"+player_id_param, api.createPlayer)
	group.PATCH("/:"+player_id_param, api.updatePlayer)
	group.DELETE("/:"+player_id_param, api.deletePlayer)
}

func (api *EchoApi) createPlayer(context echo.Context) error {
	customContext := context.Get(context_key).(*util.Context)
	logger := customContext.Logger
	logger.Debug("Join lobby")

	createPlayer, err := bindCreatePlayerDTO(context)
	if err != nil {
		logger.Warnf("Error while binding lobby join: %v", err)
		return echo.ErrBadRequest
	}

	err = api.core.CreatePlayer(customContext, mapCreatePlayerToPlayer(createPlayer), createPlayer.Password)

	if err != nil {
		if errors.Is(err, core.ErrWrongLobbyPassword) {
			logger.Infof("Player enterd wrong lobby password: %v", err)
			return echo.ErrUnauthorized
		}
		logger.Warnf("Error while joining lobby: %v", err)
		return echo.ErrInternalServerError
	}

	return context.NoContent(http.StatusCreated)
}

func (api *EchoApi) updatePlayer(context echo.Context) error {
	customContext := context.Get(context_key).(*util.Context)
	logger := customContext.Logger
	logger.Debug("Join lobby")

	updatePlayer, err := bindUpdatePlayerDTO(context)
	if err != nil {
		logger.Warnf("Error while binding lobby join: %v", err)
		return echo.ErrBadRequest
	}

	err = api.core.UpdatePlayer(customContext, mapUpdatePlayerToCorePlayer(updatePlayer))

	if err != nil {
		if errors.Is(err, core.ErrWrongLobbyPassword) {
			logger.Infof("Player enterd wrong lobby password: %v", err)
			return echo.ErrUnauthorized
		}
		logger.Warnf("Error while joining lobby: %v", err)
		return echo.ErrInternalServerError
	}

	return context.NoContent(http.StatusCreated)
}

func (api *EchoApi) deletePlayer(context echo.Context) error {
	customContext := context.Get(context_key).(*util.Context)
	logger := customContext.Logger
	logger.Debug("Leave lobby")

	playerId, err := getPlayerId(context)
	if err != nil {
		logger.Warnf("Error while binding player id: %v", err)
		return echo.ErrBadRequest
	}

	lobbbyId, err := getLobbyId(context)
	if err != nil {
		logger.Warnf("Error while binding player id: %v", err)
		return echo.ErrBadRequest
	}

	if err = api.core.DeletePlayer(customContext, lobbbyId, playerId); err != nil {
		logger.Warnf("Error while player leaving lobyy: %v", err)
		return echo.ErrInternalServerError
	}

	return context.NoContent(http.StatusNoContent)
}

func getPlayerId(context echo.Context) (uuid.UUID, error) {
	playerId, err := uuid.Parse(context.Param(player_id_param))
	if err != nil {
		return uuid.Nil, fmt.Errorf("error while binding playerId: %v", err)
	}
	return playerId, nil
}

func bindCreatePlayerDTO(context echo.Context) (createPlayer *CreatePlayer, err error) {
	createPlayer = new(CreatePlayer)
	if err := context.Bind(createPlayer); err != nil {
		return nil, fmt.Errorf("could not bind lobby, %v", err)
	}
	if err := context.Validate(createPlayer); err != nil {
		return nil, fmt.Errorf("could not validate lobby, %v", err)
	}

	return createPlayer, nil
}

func bindUpdatePlayerDTO(context echo.Context) (updatePlayer *UpdatePlayer, err error) {
	updatePlayer = new(UpdatePlayer)
	if err := context.Bind(updatePlayer); err != nil {
		return nil, fmt.Errorf("could not bind lobby, %v", err)
	}
	if err := context.Validate(updatePlayer); err != nil {
		return nil, fmt.Errorf("could not validate lobby, %v", err)
	}

	return updatePlayer, nil
}

func mapCreatePlayerToPlayer(player *CreatePlayer) *core.Player {
	return &core.Player{ID: player.ID, LobbyId: player.LobbyId, Name: player.Name, Payload: player.Payload}
}

func mapUpdatePlayerToCorePlayer(player *UpdatePlayer) *core.Player {
	return &core.Player{ID: player.ID, Name: player.Name, Payload: player.Payload}
}

func mapToPlayers(corePlayers []*core.Player) []*Player {
	players := make([]*Player, len(corePlayers))
	for index, player := range corePlayers {
		players[index] = mapToPlayer(player)
	}
	return players
}

func mapToPlayer(player *core.Player) *Player {
	if player == nil {
		return nil
	}
	return &Player{ID: player.ID, Name: player.Name, Payload: player.Payload}
}

func mapToCorePlayer(player *Player) *core.Player {
	if player == nil {
		return nil
	}
	return &core.Player{ID: player.ID, Name: player.Name, Payload: player.Payload}
}
