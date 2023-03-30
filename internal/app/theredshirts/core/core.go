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
		CreateLobby(context *util.Context, lobby *Lobby) error
		GetLobby(lobbyId uuid.UUID) (*Lobby, error)
		UpdateLobby(lobby *Lobby) error
		GetLobbies() ([]*Lobby, error)
		DeleteLobby(lobbyId uuid.UUID) error
		JoinLobby(context *util.Context, join *Join) error
		LeaveLobby(context *util.Context, lobbyId uuid.UUID, playerId uuid.UUID) error
	}

	//Objects
	Lobby struct {
		ID         uuid.UUID
		Name       string
		Owner      *Player
		Password   string
		Difficulty string
		Players    []*Player
	}

	Join struct {
		PlayerId uuid.UUID
		LobbyID  uuid.UUID
		Name     string
		Team     string
		Password string
	}

	Player struct {
		ID          uuid.UUID
		Name        string
		Team        string
		LastRefresh time.Time
		LobbyId     uuid.UUID
	}
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
