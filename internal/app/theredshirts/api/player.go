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
	PlayerCreate struct {
		ID       uuid.UUID              `param:"playerId" validate:"required"`
		Name     string                 `json:"name" validate:"required"`
		LobbyId  uuid.UUID              `json:"lobby_id" validate:"required"`
		Password string                 `json:"password"`
		Payload  map[string]interface{} `json:"payload"`
	}

	PlayerUpdate struct {
		ID      uuid.UUID              `param:"playerId" validate:"required"`
		Name    string                 `json:"name" validate:"required"`
		Payload map[string]interface{} `json:"payload"`
	}

	PlayerDelete struct {
		ID uuid.UUID `param:"playerId" validate:"required"`
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
	logger.Debug("Create player")

	createPlayer, err := bindCreatePlayerDTO(context)
	if err != nil {
		logger.Warnf("Error while binding player to create: %v", err)
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
	logger.Debug("Update player")

	updatePlayer, err := bindUpdatePlayerDTO(context)
	if err != nil {
		logger.Warnf("Error while binding player to update: %v", err)
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
	logger.Debug("Delete player")

	deletePlayer, err := bindDeletePlayerDTO(context)
	if err != nil {
		logger.Warnf("Error while binding player to delete: %v", err)
		return echo.ErrBadRequest
	}

	if err = api.core.DeletePlayer(customContext, deletePlayer.ID); err != nil {
		logger.Warnf("Error while player leaving lobyy: %v", err)
		return echo.ErrInternalServerError
	}

	return context.NoContent(http.StatusNoContent)
}

func bindCreatePlayerDTO(context echo.Context) (createPlayer *PlayerCreate, err error) {
	createPlayer = new(PlayerCreate)
	if err := context.Bind(createPlayer); err != nil {
		return nil, fmt.Errorf("could not bind lobby, %v", err)
	}
	if err := context.Validate(createPlayer); err != nil {
		return nil, fmt.Errorf("could not validate lobby, %v", err)
	}

	return createPlayer, nil
}

func bindUpdatePlayerDTO(context echo.Context) (updatePlayer *PlayerUpdate, err error) {
	updatePlayer = new(PlayerUpdate)
	if err := context.Bind(updatePlayer); err != nil {
		return nil, fmt.Errorf("could not bind lobby, %v", err)
	}
	if err := context.Validate(updatePlayer); err != nil {
		return nil, fmt.Errorf("could not validate lobby, %v", err)
	}

	return updatePlayer, nil
}

func bindDeletePlayerDTO(context echo.Context) (deletePlayer *PlayerDelete, err error) {
	deletePlayer = new(PlayerDelete)
	if err := context.Bind(deletePlayer); err != nil {
		return nil, fmt.Errorf("could not bind lobby, %v", err)
	}
	if err := context.Validate(deletePlayer); err != nil {
		return nil, fmt.Errorf("could not validate lobby, %v", err)
	}

	return deletePlayer, nil
}

func mapCreatePlayerToPlayer(player *PlayerCreate) *core.Player {
	return &core.Player{ID: player.ID, LobbyId: player.LobbyId, Name: player.Name, Payload: player.Payload}
}

func mapUpdatePlayerToCorePlayer(player *PlayerUpdate) *core.Player {
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
