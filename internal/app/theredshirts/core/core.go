package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/db"
	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/google/uuid"
)

type (

	//Facade
	CoreFacade struct {
		db db.DB
	}

	Core interface {
		CreateLobby(context *util.Context, lobby *Lobby) error
		GetLobby(lobbyId uuid.UUID) (*Lobby, error)
		UpdateLobby(lobby *Lobby) error
		UpdateLobbyStatus(lobby *Lobby) error
		GetLobbies() ([]*Lobby, error)
		DeleteLobby(context *util.Context, lobbyId uuid.UUID, ownerId uuid.UUID) error
		CreatePlayer(context *util.Context, join *Player, password string) error
		GetPlayer(playerId uuid.UUID) (*Player, error)
		UpdatePlayer(context *util.Context, player *Player) error
		UpdatePlayerLastRefresh(playerId uuid.UUID) error
		DeletePlayer(context *util.Context, playerId uuid.UUID) error
	}

	//Objects
	Lobby struct {
		ID                  uuid.UUID
		Status              string
		Name                string
		Owner               *Player
		Password            string
		Difficulty          int
		MissionLength       int
		NumberOfCrewMembers int
		MaxPlayers          int
		ExpansionPacks      []string
		Players             []*Player
		Payload             map[string]interface{}
	}

	Player struct {
		ID          uuid.UUID
		Name        string
		LastRefresh time.Time
		LobbyId     uuid.UUID
		Payload     map[string]interface{}
	}
)

const (
	lobby_open    = "OPEN"
	lobby_playing = "PLAYING"
)

var (
	ErrWrongLobbyPassword = errors.New("wrong password")
)

func NewCore() (Core, error) {
	db, err := db.NewConnection()
	if err != nil {
		return nil, fmt.Errorf("error while initializing database: %v", err)
	}
	core := &CoreFacade{db: db}
	return core, nil
}
