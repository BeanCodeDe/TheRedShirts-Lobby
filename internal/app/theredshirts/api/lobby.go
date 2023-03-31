package api

import (
	"fmt"
	"net/http"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/core"
	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	lobby_root_path          = "/lobby"
	lobby_update_status_path = "/status"
	lobby_id_param           = "lobbyId"
)

type (
	LobbyCreate struct {
		ID                  uuid.UUID              `param:"lobbyId" validate:"required"`
		Name                string                 `json:"name" validate:"required"`
		Owner               *Player                `json:"owner" validate:"required"`
		Password            string                 `json:"password"`
		Difficulty          int                    `json:"difficulty" validate:"required"`
		MissionLength       int                    `json:"mission_length" validate:"required"`
		NumberOfCrewMembers int                    `json:"number_of_crew_members" validate:"required"`
		MaxPlayers          int                    `json:"max_players" validate:"required"`
		ExpansionPacks      []string               `json:"expansion_packs"`
		PlayerPayload       map[string]interface{} `json:"player_payload"`
		Payload             map[string]interface{} `json:"payload"`
	}

	LobbyUpdate struct {
		ID                  uuid.UUID              `param:"lobbyId" validate:"required"`
		Name                string                 `json:"name" validate:"required"`
		Owner               uuid.UUID              `query:"owner" validate:"required"`
		Status              string                 `json:"status" validate:"required"`
		Password            string                 `json:"password"`
		Difficulty          int                    `json:"difficulty" validate:"required"`
		MissionLength       int                    `json:"mission_length" validate:"required"`
		NumberOfCrewMembers int                    `json:"number_of_crew_members" validate:"required"`
		MaxPlayers          int                    `json:"max_players" validate:"required"`
		ExpansionPacks      []string               `json:"expansion_packs"`
		Payload             map[string]interface{} `json:"payload"`
	}

	LobbyUpdateStatus struct {
		ID     uuid.UUID `param:"lobbyId" validate:"required"`
		Owner  uuid.UUID `query:"owner" validate:"required"`
		Status string    `json:"status"`
	}

	Lobby struct {
		ID                  uuid.UUID              `json:"id"`
		Name                string                 `json:"name"`
		Status              string                 `json:"status"`
		Owner               *Player                `json:"owner"`
		Difficulty          int                    `json:"difficulty"`
		MissionLength       int                    `json:"mission_length"`
		NumberOfCrewMembers int                    `json:"number_of_crew_members" `
		MaxPlayers          int                    `json:"max_players" `
		ExpansionPacks      []string               `json:"expansion_packs" `
		Players             []*Player              `json:"players"`
		Payload             map[string]interface{} `json:"payload"`
	}
)

func initLobbyInterface(group *echo.Group, api *EchoApi) {
	group.POST("", api.createLobbyId)
	group.PUT("/:"+lobby_id_param, api.createLobby)
	group.PATCH("/:"+lobby_id_param, api.updateLobby)
	group.PATCH("/:"+lobby_id_param+lobby_update_status_path, api.updateStatusLobby)
	group.DELETE("/:"+lobby_id_param, api.deleteLobby)
	group.GET("", api.getAllLobbies)
	group.GET("/:"+lobby_id_param, api.getLobby)
}

func (api *EchoApi) createLobbyId(context echo.Context) error {
	logger := context.Get(context_key).(*util.Context).Logger
	logger.Debug("Create lobby Id")
	return context.String(http.StatusCreated, uuid.NewString())
}

func (api *EchoApi) createLobby(context echo.Context) error {
	customContext := context.Get(context_key).(*util.Context)
	logger := customContext.Logger
	logger.Debug("Create lobby")

	lobby, err := bindLobbyCreationDTO(context)

	if err != nil {
		logger.Warnf("Error while binding lobby: %v", err)
		return echo.ErrBadRequest
	}
	coreLobby := mapLobbyCreateToCoreLobby(lobby)
	err = api.core.CreateLobby(customContext, coreLobby, lobby.PlayerPayload)

	if err != nil {
		logger.Warnf("Error while creating lobby: %v", err)
		return echo.ErrInternalServerError
	}

	return context.NoContent(http.StatusCreated)
}

func (api *EchoApi) updateLobby(context echo.Context) error {
	logger := context.Get(context_key).(*util.Context).Logger
	logger.Debug("Update lobby")

	lobby, err := bindLobbyUpdateDTO(context)

	if err != nil {
		logger.Warnf("Error while binding lobby: %v", err)
		return echo.ErrBadRequest
	}

	coreLobby := mapLobbyUpdateToCoreLobby(lobby)
	err = api.core.UpdateLobby(coreLobby)

	if err != nil {
		logger.Warnf("Error while creating lobby: %v", err)
		return echo.ErrInternalServerError
	}

	return context.NoContent(http.StatusOK)
}

func (api *EchoApi) updateStatusLobby(context echo.Context) error {
	logger := context.Get(context_key).(*util.Context).Logger
	logger.Debug("Update status of lobby")

	lobby, err := bindLobbyUpdateStatusDTO(context)

	if err != nil {
		logger.Warnf("Error while binding lobby: %v", err)
		return echo.ErrBadRequest
	}

	coreLobby := mapLobbyUpdateStatusToCoreLobby(lobby)
	err = api.core.UpdateLobby(coreLobby)

	if err != nil {
		logger.Warnf("Error while creating lobby: %v", err)
		return echo.ErrInternalServerError
	}

	return context.NoContent(http.StatusOK)
}

func (api *EchoApi) deleteLobby(context echo.Context) error {
	logger := context.Get(context_key).(*util.Context).Logger
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
	logger := context.Get(context_key).(*util.Context).Logger
	logger.Debug("Get all lobbies")

	lobbies, err := api.core.GetLobbies()
	if err != nil {
		logger.Warnf("Error while loading lobby: %v", err)
		return echo.ErrInternalServerError
	}
	return context.JSON(http.StatusOK, mapToLobbies(lobbies))
}

func (api *EchoApi) getLobby(context echo.Context) error {
	logger := context.Get(context_key).(*util.Context).Logger
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

func bindLobbyCreationDTO(context echo.Context) (*LobbyCreate, error) {
	var lobby = new(LobbyCreate)
	if err := context.Bind(lobby); err != nil {
		return nil, fmt.Errorf("could not bind lobby, %v", err)
	}
	if err := context.Validate(lobby); err != nil {
		return nil, fmt.Errorf("could not validate lobby, %v", err)
	}
	return lobby, nil
}

func bindLobbyUpdateDTO(context echo.Context) (*LobbyUpdate, error) {
	var lobby = new(LobbyUpdate)
	if err := context.Bind(lobby); err != nil {
		return nil, fmt.Errorf("could not bind lobby, %v", err)
	}
	if err := context.Validate(lobby); err != nil {
		return nil, fmt.Errorf("could not validate lobby, %v", err)
	}

	return lobby, nil
}

func bindLobbyUpdateStatusDTO(context echo.Context) (*LobbyUpdateStatus, error) {
	var lobby = new(LobbyUpdateStatus)
	if err := context.Bind(lobby); err != nil {
		return nil, fmt.Errorf("could not bind lobby, %v", err)
	}
	if err := context.Validate(lobby); err != nil {
		return nil, fmt.Errorf("could not validate lobby, %v", err)
	}

	return lobby, nil
}

func getLobbyId(context echo.Context) (uuid.UUID, error) {
	lobbyId, err := uuid.Parse(context.Param(lobby_id_param))
	if err != nil {
		return uuid.Nil, fmt.Errorf("error while binding lobbbyId: %v", err)
	}
	return lobbyId, nil
}

func mapLobbyCreateToCoreLobby(lobby *LobbyCreate) *core.Lobby {
	return &core.Lobby{ID: lobby.ID, Name: lobby.Name, Owner: mapToCorePlayer(lobby.Owner), Password: lobby.Password, Difficulty: lobby.Difficulty, MissionLength: lobby.MissionLength, NumberOfCrewMembers: lobby.NumberOfCrewMembers, MaxPlayers: lobby.MaxPlayers, ExpansionPacks: lobby.ExpansionPacks, Payload: lobby.Payload}
}

func mapLobbyUpdateToCoreLobby(lobby *LobbyUpdate) *core.Lobby {
	return &core.Lobby{ID: lobby.ID, Status: lobby.Status, Name: lobby.Name, Owner: &core.Player{ID: lobby.Owner}, Password: lobby.Password, Difficulty: lobby.Difficulty, MissionLength: lobby.MissionLength, NumberOfCrewMembers: lobby.NumberOfCrewMembers, MaxPlayers: lobby.MaxPlayers, ExpansionPacks: lobby.ExpansionPacks, Payload: lobby.Payload}
}

func mapLobbyUpdateStatusToCoreLobby(lobby *LobbyUpdateStatus) *core.Lobby {
	return &core.Lobby{ID: lobby.ID, Status: lobby.Status, Owner: &core.Player{ID: lobby.Owner}}
}

func mapToLobby(lobby *core.Lobby) *Lobby {
	return &Lobby{ID: lobby.ID, Status: lobby.Status, Name: lobby.Name, Owner: mapToPlayer(lobby.Owner), Difficulty: lobby.Difficulty, MissionLength: lobby.MissionLength, NumberOfCrewMembers: lobby.NumberOfCrewMembers, MaxPlayers: lobby.MaxPlayers, ExpansionPacks: lobby.ExpansionPacks, Players: mapToPlayers(lobby.Players), Payload: lobby.Payload}
}

func mapToLobbies(coreLobbies []*core.Lobby) []*Lobby {
	lobbies := make([]*Lobby, len(coreLobbies))
	for index, lobby := range coreLobbies {
		lobbies[index] = mapToLobby(lobby)
	}
	return lobbies
}
