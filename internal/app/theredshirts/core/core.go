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
		db             db.DB
		messageAdapter *adapter.MessageAdapter
		lobbyPlayerId  uuid.UUID
	}

	transaction struct {
		dbTx     db.DBTx
		messages []*message
	}

	Core interface {
		CreateLobby(context *util.Context, lobby *Lobby) error
		GetLobby(context *util.Context, lobbyId uuid.UUID) (*Lobby, error)
		UpdateLobby(context *util.Context, lobby *Lobby, playerId uuid.UUID) error
		UpdateLobbyStatus(context *util.Context, lobby *Lobby, playerId uuid.UUID) error
		GetLobbies(context *util.Context) ([]*Lobby, error)
		DeleteLobby(context *util.Context, lobbyId uuid.UUID, playerId uuid.UUID) error
		CreatePlayer(context *util.Context, join *Player, password string) error
		GetPlayer(context *util.Context, playerId uuid.UUID) (*Player, error)
		UpdatePlayer(context *util.Context, player *Player, playerId uuid.UUID) error
		UpdatePlayerLastRefresh(context *util.Context, playerId uuid.UUID) error
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
		Spectator   bool
		Payload     map[string]interface{}
	}
)

const (
	lobby_open    = "OPEN"
	lobby_playing = "PLAYING"
)

var (
	ErrWrongLobbyPassword = errors.New("wrong password")
	ErrLobbyFull          = errors.New("lobby is full")
)

func NewCore() (Core, error) {
	db, err := db.NewConnection()
	if err != nil {
		return nil, fmt.Errorf("error while initializing database: %v", err)
	}
	messageAdapter, err := adapter.NewMessageAdapter()
	if err != nil {
		return nil, fmt.Errorf("erro while initializing messageadapter: %v", err)
	}
	lobbyPlayerId, err := util.GetEnvUUID("LOBBY_USER")
	if err != nil {
		return nil, fmt.Errorf("error while loading lobby user from env: %v", err)
	}
	core := &CoreFacade{db: db, messageAdapter: messageAdapter, lobbyPlayerId: lobbyPlayerId}
	core.startCleanUp()
	return core, nil
}

func (core CoreFacade) startTransaction() (*transaction, error) {
	tx, err := core.db.StartTransaction()
	if err != nil {
		return nil, fmt.Errorf("error while starting transaction: %v", err)
	}
	return &transaction{dbTx: tx, messages: make([]*message, 0)}, nil
}

func (core CoreFacade) commit(tx *transaction, context *util.Context) error {
	if err := tx.dbTx.Commit(); err != nil {
		return fmt.Errorf("error while commiting transaction: %v", err)
	}
	for _, message := range tx.messages {
		err := core.createMessage(context, message)
		if err != nil {
			context.Logger.Warnf("Error while creating message after transaction: %v", err)
		}
	}
	return nil
}

func (core CoreFacade) rollback(tx *transaction) error {
	if err := tx.dbTx.Rollback(); err != nil {
		return fmt.Errorf("error while rollback transaction: %v", err)
	}
	return nil
}
