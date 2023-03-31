package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/adapter"
	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/db"
	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/google/uuid"
)

type (

	//Facade
	CoreFacade struct {
		db          db.DB
		chatAdapter adapter.ChatAdapter
	}

	Core interface {
		CreateLobby(context *util.Context, lobby *Lobby, playerPayload map[string]interface{}) error
		GetLobby(lobbyId uuid.UUID) (*Lobby, error)
		UpdateLobby(lobby *Lobby) error
		UpdateLobbyStatus(lobby *Lobby) error
		GetLobbies() ([]*Lobby, error)
		DeleteLobby(lobbyId uuid.UUID) error
		CreatePlayer(context *util.Context, join *Player, password string) error
		UpdatePlayer(context *util.Context, player *Player) error
		DeletePlayer(context *util.Context, lobbyId uuid.UUID, playerId uuid.UUID) error
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
	chatAdapter := adapter.NewChatAdapter()
	core := &CoreFacade{db: db, chatAdapter: *chatAdapter}
	core.startCleanUp()
	return core, nil
}
