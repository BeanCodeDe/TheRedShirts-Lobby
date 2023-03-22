package db

import (
	"errors"
	"strings"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/google/uuid"
)

type (
	Lobby struct {
		ID         uuid.UUID `db:"id"`
		Name       string    `db:"name"`
		Owner      uuid.UUID `db:"owner"`
		Password   string    `db:"password"`
		Difficulty string    `db:"difficulty"`
	}

	Player struct {
		ID      uuid.UUID `db:"id"`
		Name    string    `db:"name"`
		LobbyId uuid.UUID `db:"lobby_id"`
	}

	DB interface {
		Close()
		//Lobby
		CreateLobby(lobby *Lobby) error
		UpdateLobby(lobby *Lobby) error
		DeleteLobby(id uuid.UUID) error
		GetLobbyById(id uuid.UUID) (*Lobby, error)
		GetAllLobbies() ([]*Lobby, error)
		//Player
		CreatePlayer(player *Player) error
		DeletePlayer(id uuid.UUID) error
		GetPlayerById(id uuid.UUID) (*Player, error)
		GetAllPlayersInLobby(lobbyId uuid.UUID) ([]*Player, error)
	}
)

const (
	schema_name = "theredshirts_lobby"
)

func NewConnection() (DB, error) {
	switch db := strings.ToLower(util.GetEnvWithFallback("DATABASE", "postgresql")); db {
	case "postgresql":
		return newPostgresConnection()
	default:
		return nil, errors.New("no configuration for %s found")
	}
}
