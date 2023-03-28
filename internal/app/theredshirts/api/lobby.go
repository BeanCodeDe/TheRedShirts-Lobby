package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/core"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

const lobby_root_path = "/lobby"
const lobby_id_param = "lobbyId"
const player_id_param = "playerId"

type (
	LobbyCreate struct {
		Name       string  `json:"name" validate:"required"`
		Owner      *Player `json:"owner" validate:"required"`
		Password   string  `json:"password"`
		Difficulty string  `json:"difficulty" validate:"required"`
	}

	LobbyUpdate struct {
		Name       string `validate:"required"`
		Password   string
		Difficulty string `validate:"required"`
	}

	LobbyJoin struct {
		Name     string `json:"name" validate:"required"`
		Team     string `json:"team" validate:"required"`
		Password string `json:"password"`
	}

	Lobby struct {
		ID         uuid.UUID `json:"id"`
		Name       string    `json:"name"`
		Owner      *Player   `json:"owner"`
		Difficulty string    `json:"difficulty"`
		Players    []*Player `json:"players"`
	}

	Player struct {
		ID   uuid.UUID `json:"id" validate:"required"`
		Name string    `json:"name" validate:"required"`
		Team string    `json:"team" validate:"required"`
	}
)

func initLobbyInterface(group *echo.Group, api *EchoApi) {
	group.POST("", api.createLobbyId)
	group.PUT("/:"+lobby_id_param, api.createLobby)
	group.PATCH("/:"+lobby_id_param, api.updateLobby)
	group.DELETE("/:"+lobby_id_param, api.deleteLobby)
	group.GET("", api.getAllLobbies)
	group.GET("/:"+lobby_id_param, api.getLobby)
	group.PUT("/:"+lobby_id_param+"/player/:"+player_id_param, api.joinLobby)
	group.DELETE("/:"+lobby_id_param+"/player/:"+player_id_param, api.leaveLobby)
}

func (api *EchoApi) createLobbyId(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Create lobby Id")
	return context.String(http.StatusCreated, uuid.NewString())
}

func (api *EchoApi) createLobby(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Create lobby")

	lobby, lobbyId, err := bindLobbyCreationDTO(context)

	if err != nil {
		logger.Warnf("Error while binding lobby: %v", err)
		return echo.ErrBadRequest
	}

	coreLobby := mapLobbyCreateToCoreLobby(lobby, lobbyId)
	err = api.core.CreateLobby(coreLobby)

	if err != nil {
		logger.Warnf("Error while creating lobby: %v", err)
		return echo.ErrInternalServerError
	}

	return context.NoContent(http.StatusCreated)
}

func (api *EchoApi) updateLobby(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Update lobby")

	lobby, lobbyId, err := bindLobbyUpdateDTO(context)

	if err != nil {
		logger.Warnf("Error while binding lobby: %v", err)
		return echo.ErrBadRequest
	}

	coreLobby := mapLobbyUpdateToCoreLobby(lobby, lobbyId)
	err = api.core.UpdateLobby(coreLobby)

	if err != nil {
		logger.Warnf("Error while creating lobby: %v", err)
		return echo.ErrInternalServerError
	}

	return context.NoContent(http.StatusOK)
}

func (api *EchoApi) deleteLobby(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Delete lobby")

	lobbyId, err := getLobbyId(context)
	if err != nil {
		logger.Warnf("Error while binding lobby id: %v", err)
		return echo.ErrBadRequest
	}

	if err := api.core.DeleteLobby(lobbyId); err != nil {
		logger.Warnf("Error while loading lobby: %v", err)
		return echo.ErrInternalServerError
	}

	return context.NoContent(http.StatusOK)
}

func (api *EchoApi) getAllLobbies(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Get all lobbies")

	lobbies, err := api.core.GetLobbies()
	if err != nil {
		logger.Warnf("Error while loading lobby: %v", err)
		return echo.ErrInternalServerError
	}
	return context.JSON(http.StatusOK, mapToLobbies(lobbies))
}

func (api *EchoApi) getLobby(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Get lobby")

	lobbyId, err := getLobbyId(context)
	if err != nil {
		logger.Warnf("Error while binding lobby id: %v", err)
		return echo.ErrBadRequest
	}

	lobby, err := api.core.GetLobby(lobbyId)
	if err != nil {
		logger.Warnf("Error while loading lobby: %v", err)
		return echo.ErrInternalServerError
	}
	return context.JSON(http.StatusOK, mapToLobby(lobby))
}

func (api *EchoApi) joinLobby(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
	logger.Debug("Join lobby")

	joinLobby, lobbyId, playerId, err := bindLobbyJoinDTO(context)
	if err != nil {
		logger.Warnf("Error while binding lobby join: %v", err)
		return echo.ErrBadRequest
	}

	err = api.core.JoinLobby(mapLobbyJoinToCoreJoin(joinLobby, lobbyId, playerId))

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

func (api *EchoApi) leaveLobby(context echo.Context) error {
	logger := context.Get(logger_key).(*log.Entry)
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

	if err = api.core.LeaveLobby(lobbbyId, playerId); err != nil {
		logger.Warnf("Error while player leaving lobyy: %v", err)
		return echo.ErrInternalServerError
	}

	return context.NoContent(http.StatusNoContent)
}

func bindLobbyCreationDTO(context echo.Context) (*LobbyCreate, uuid.UUID, error) {
	var lobby = new(LobbyCreate)
	if err := context.Bind(lobby); err != nil {
		return nil, uuid.Nil, fmt.Errorf("could not bind lobby, %v", err)
	}
	if err := context.Validate(lobby); err != nil {
		return nil, uuid.Nil, fmt.Errorf("could not validate lobby, %v", err)
	}
	lobbyId, err := getLobbyId(context)
	if err != nil {
		return nil, uuid.Nil, err
	}
	return lobby, lobbyId, nil
}

func bindLobbyUpdateDTO(context echo.Context) (*LobbyUpdate, uuid.UUID, error) {
	var lobby = new(LobbyUpdate)
	if err := context.Bind(lobby); err != nil {
		return nil, uuid.Nil, fmt.Errorf("could not bind lobby, %v", err)
	}
	if err := context.Validate(lobby); err != nil {
		return nil, uuid.Nil, fmt.Errorf("could not validate lobby, %v", err)
	}
	lobbyId, err := getLobbyId(context)
	if err != nil {
		return nil, uuid.Nil, err
	}
	return lobby, lobbyId, nil
}

func bindLobbyJoinDTO(context echo.Context) (lobbyJoin *LobbyJoin, lobbyId uuid.UUID, playerId uuid.UUID, err error) {
	lobbyJoin = new(LobbyJoin)
	if err := context.Bind(lobbyJoin); err != nil {
		return nil, uuid.Nil, uuid.Nil, fmt.Errorf("could not bind lobby, %v", err)
	}
	if err := context.Validate(lobbyJoin); err != nil {
		return nil, uuid.Nil, uuid.Nil, fmt.Errorf("could not validate lobby, %v", err)
	}
	lobbyId, err = getLobbyId(context)
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, err
	}

	playerId, err = getPlayerId(context)
	if err != nil {
		return nil, uuid.Nil, uuid.Nil, err
	}
	return lobbyJoin, lobbyId, playerId, nil
}

func getLobbyId(context echo.Context) (uuid.UUID, error) {
	lobbyId, err := uuid.Parse(context.Param(lobby_id_param))
	if err != nil {
		return uuid.Nil, fmt.Errorf("error while binding lobbbyId: %v", err)
	}
	return lobbyId, nil
}

func getPlayerId(context echo.Context) (uuid.UUID, error) {
	playerId, err := uuid.Parse(context.Param(player_id_param))
	if err != nil {
		return uuid.Nil, fmt.Errorf("error while binding playerId: %v", err)
	}
	return playerId, nil
}

func mapLobbyJoinToCoreJoin(lobbyJoin *LobbyJoin, lobbyId uuid.UUID, playerId uuid.UUID) *core.Join {
	return &core.Join{PlayerId: playerId, LobbyID: lobbyId, Name: lobbyJoin.Name, Team: lobbyJoin.Team, Password: lobbyJoin.Password}
}

func mapLobbyCreateToCoreLobby(lobby *LobbyCreate, lobbyId uuid.UUID) *core.Lobby {
	return &core.Lobby{ID: lobbyId, Name: lobby.Name, Owner: mapToCorePlayer(lobby.Owner), Password: lobby.Password, Difficulty: lobby.Difficulty}
}

func mapLobbyUpdateToCoreLobby(lobby *LobbyUpdate, lobbyId uuid.UUID) *core.Lobby {
	return &core.Lobby{ID: lobbyId, Name: lobby.Name, Password: lobby.Password, Difficulty: lobby.Difficulty}
}

func mapToLobby(lobby *core.Lobby) *Lobby {
	return &Lobby{ID: lobby.ID, Name: lobby.Name, Owner: mapToPlayer(lobby.Owner), Difficulty: lobby.Difficulty, Players: mapToPlayers(lobby.Players)}
}

func mapToLobbies(coreLobbies []*core.Lobby) []*Lobby {
	lobbies := make([]*Lobby, len(coreLobbies))
	for index, lobby := range coreLobbies {
		lobbies[index] = mapToLobby(lobby)
	}
	return lobbies
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
	return &Player{ID: player.ID, Name: player.Name}
}

func mapToCorePlayer(player *Player) *core.Player {
	if player == nil {
		return nil
	}
	return &core.Player{ID: player.ID, Name: player.Name}
}
