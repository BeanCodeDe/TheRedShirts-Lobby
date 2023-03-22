package core

import (
	"fmt"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/db"
	"github.com/google/uuid"
)

type (

	//Facade
	CoreFacade struct {
		db db.DB
	}

	Core interface {
		CreateLobby(lobby *Lobby) error
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

	Player struct {
		ID   uuid.UUID
		Name string
	}
)

func NewCore() (Core, error) {
	db, err := db.NewConnection()
	if err != nil {
		return nil, fmt.Errorf("error while initializing database: %v", err)
	}
	return &CoreFacade{db: db}, nil
}

func mapToDBLobby(lobby *Lobby) *db.Lobby {
	return &db.Lobby{ID: lobby.ID, Name: lobby.Name, Owner: lobby.Owner.ID, Password: lobby.Password, Difficulty: lobby.Difficulty}
}

func mapToDBPlayer(player *Player, lobbyId uuid.UUID) *db.Player {
	return &db.Player{ID: player.ID, Name: player.Name, LobbyId: lobbyId}
}

func mapToLobby(lobby *db.Lobby, owner *Player, players []*Player) *Lobby {
	return &Lobby{ID: lobby.ID, Name: lobby.Name, Owner: owner, Password: lobby.Password, Difficulty: lobby.Difficulty, Players: players}
}

func mapToPlayer(player *db.Player) *Player {
	return &Player{ID: player.ID, Name: player.Name}
}
