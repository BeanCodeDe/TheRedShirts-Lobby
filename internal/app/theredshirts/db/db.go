package db

import (
	"errors"
	"strings"
	"time"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/google/uuid"
)

type (
	Lobby struct {
		ID                  uuid.UUID              `db:"id"`
		Status              string                 `db:"status"`
		Name                string                 `db:"name"`
		Owner               uuid.UUID              `db:"owner"`
		Password            string                 `db:"password"`
		Difficulty          int                    `db:"difficulty"`
		MissionLength       int                    `db:"mission_length"`
		NumberOfCrewMembers int                    `db:"number_of_crew_members"`
		MaxPlayers          int                    `db:"max_players"`
		ExpansionPacks      []string               `db:"expansion_packs"`
		Payload             map[string]interface{} `db:"payload"`
	}

	Player struct {
		ID          uuid.UUID              `db:"id"`
		Name        string                 `db:"name"`
		LobbyId     uuid.UUID              `db:"lobby_id"`
		LastRefresh time.Time              `db:"last_refresh"`
		Payload     map[string]interface{} `db:"payload"`
	}

	DB interface {
		Close()
		StartTransaction() (DBTx, error)
	}

	DBTx interface {
		HandleTransaction(err error)
		//Lobby
		CreateLobby(lobby *Lobby) error
		UpdateLobby(lobby *Lobby) error
		DeleteLobby(id uuid.UUID) error
		DeleteEmptyLobbies() error
		GetLobbyById(id uuid.UUID) (*Lobby, error)
		GetAllLobbies() ([]*Lobby, error)
		//Player
		CreatePlayer(player *Player) error
		DeletePlayer(id uuid.UUID) error
		DeleteAllPlayerInLobby(lobbyId uuid.UUID) error
		DeletePlayerOlderRefreshDate(time time.Time) error
		UpdatePlayer(player *Player) error
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
