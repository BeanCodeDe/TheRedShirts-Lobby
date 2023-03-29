package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/db"
	"github.com/google/uuid"
)

func (core CoreFacade) CreateLobby(lobby *Lobby) error {
	dbLobby := mapToDBLobby(lobby)

	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	if err := tx.CreateLobby(dbLobby); err != nil {
		if !errors.Is(err, db.ErrLobbyAlreadyExists) {
			return fmt.Errorf("error while creating lobby: %v", err)
		}
		foundLobby, err := tx.GetLobbyById(lobby.ID)
		if err != nil {
			return fmt.Errorf("something went wrong while checking if lobby [%v] is already created: %v", lobby.ID, err)
		}

		if lobby.Name != foundLobby.Name || lobby.Password != foundLobby.Password {
			return fmt.Errorf("request of lobby [%v] doesn't match lobby from database [%v]", lobby, foundLobby)
		}

	}

	if err := core.joinLobby(tx, lobby.Owner.ID, lobby.Owner.Name, lobby.Owner.Team, lobby.ID, lobby.Password); err != nil {
		return err
	}

	return nil
}

func (core CoreFacade) UpdateLobby(lobby *Lobby) error {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)
	err = core.updateLobby(tx, lobby)
	return err
}

func (core CoreFacade) updateLobby(tx db.DBTx, lobby *Lobby) error {
	dbLobby, err := tx.GetLobbyById(lobby.ID)
	if err != nil {
		return fmt.Errorf("something went wrong while loading lobby [%v] from database: %v", lobby.ID, err)
	}

	dbLobby.Name = lobby.Name
	dbLobby.Difficulty = lobby.Difficulty
	dbLobby.Owner = lobby.Owner.ID
	dbLobby.Password = lobby.Password

	if err := tx.UpdateLobby(dbLobby); err != nil {
		if err != nil {
			return fmt.Errorf("something went wrong while updating lobby [%v]: %v", lobby.ID, err)
		}
	}
	return nil
}
func (core CoreFacade) DeleteLobby(lobbyId uuid.UUID) error {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	err = core.deleteLobby(tx, lobbyId)
	return err
}

func (core CoreFacade) deleteLobby(tx db.DBTx, lobbyId uuid.UUID) error {
	if err := tx.DeleteAllPlayerInLobby(lobbyId); err != nil {
		return fmt.Errorf("an error accourd while deleting all players from lobby [%v]: %v", lobbyId, err)
	}

	if err := tx.DeleteLobby(lobbyId); err != nil {
		return fmt.Errorf("an error accourd while deleting lobby [%v]: %v", lobbyId, err)
	}
	return nil
}

func (core CoreFacade) GetLobby(lobbyId uuid.UUID) (*Lobby, error) {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)

	lobby, err := core.getLobby(tx, lobbyId)
	return lobby, err
}

func (core CoreFacade) getLobby(tx db.DBTx, lobbyId uuid.UUID) (*Lobby, error) {
	lobby, err := tx.GetLobbyById(lobbyId)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while loading lobby [%v] from database: %v", lobbyId, err)
	}

	players, err := tx.GetAllPlayersInLobby(lobbyId)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while loading players of lobby [%v] from database: %v", lobby.ID, err)
	}

	owner, err := tx.GetPlayerById(lobby.Owner)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while loading owner [%v] of lobby [%v] from database: %v", lobby.Owner, lobby.ID, err)
	}

	return mapToLobby(lobby, mapToPlayer(owner), mapToPlayers(players)), nil
}

func (core CoreFacade) GetLobbies() ([]*Lobby, error) {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)

	lobbies, err := tx.GetAllLobbies()
	if err != nil {
		return nil, fmt.Errorf("something went wrong while loading all lobbies from database: %v", err)
	}

	coreLobbies := make([]*Lobby, len(lobbies))
	for index, lobby := range lobbies {

		players, err := tx.GetAllPlayersInLobby(lobby.ID)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while loading players of lobby [%v] from database: %v", lobby.ID, err)
		}

		owner, err := tx.GetPlayerById(lobby.Owner)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while loading owner [%v] of lobby [%v] from database: %v", lobby.Owner, lobby.ID, err)
		}
		coreLobbies[index] = mapToLobby(lobby, mapToPlayer(owner), mapToPlayers(players))
	}

	return coreLobbies, nil
}

func (core CoreFacade) JoinLobby(join *Join) error {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	err = core.joinLobby(tx, join.PlayerId, join.Name, join.Team, join.LobbyID, join.Password)
	return err
}

func (core CoreFacade) joinLobby(tx db.DBTx, playerId uuid.UUID, playerName string, teamName string, lobbyId uuid.UUID, password string) error {

	player, err := core.getPlayer(tx, playerId)
	if err != nil {
		return err
	}

	if player != nil {
		if lobbyId != player.LobbyId {
			if err := core.LeaveLobby(lobbyId, playerId); err != nil {
				return err
			}
		} else {
			player.LastRefresh = time.Now()
			player.Team = teamName
			if err := tx.UpdatePlayer(mapToDBPlayer(player, lobbyId)); err != nil {
				return fmt.Errorf("something went wrong while creating player %v from database: %v", playerId, err)
			}
			return nil
		}
	}

	lobby, err := tx.GetLobbyById(lobbyId)
	if err != nil {
		return fmt.Errorf("something went wrong while loading lobby %v from database: %v", lobbyId, err)
	}
	if lobby.Password != password {
		return ErrWrongLobbyPassword
	}
	if err := tx.CreatePlayer(&db.Player{ID: playerId, Name: playerName, Team: teamName, LobbyId: lobbyId, LastRefresh: time.Now()}); err != nil {
		return fmt.Errorf("something went wrong while creating player %v from database: %v", playerId, err)
	}

	return nil
}

func (core CoreFacade) LeaveLobby(lobbyId uuid.UUID, playerId uuid.UUID) error {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)

	err = core.leaveLobby(tx, lobbyId, playerId)
	return err
}

func (core CoreFacade) leaveLobby(tx db.DBTx, lobbyId uuid.UUID, playerId uuid.UUID) error {
	player, err := core.getPlayer(tx, playerId)
	if err != nil {
		return err
	}

	if player.LobbyId != lobbyId {
		return nil
	}

	lobby, err := core.getLobby(tx, lobbyId)
	if err != nil {
		return err
	}

	if lobby.Owner.ID == playerId {
		foundNewOwner := findPlayerNot(lobby.Players, playerId)
		if foundNewOwner == nil {
			if err := core.deleteLobby(tx, lobbyId); err != nil {
				return err
			}
		} else {
			lobby.Owner = foundNewOwner
			if err := core.updateLobby(tx, lobby); err != nil {
				return err
			}
		}
	}

	if err := tx.DeletePlayer(playerId); err != nil {
		return fmt.Errorf("something went wrong while deleting player %v from database: %v", playerId, err)
	}
	return nil
}

func mapToDBLobby(lobby *Lobby) *db.Lobby {
	return &db.Lobby{ID: lobby.ID, Name: lobby.Name, Owner: lobby.Owner.ID, Password: lobby.Password, Difficulty: lobby.Difficulty}
}

func mapToLobby(lobby *db.Lobby, owner *Player, players []*Player) *Lobby {
	return &Lobby{ID: lobby.ID, Name: lobby.Name, Owner: owner, Password: lobby.Password, Difficulty: lobby.Difficulty, Players: players}
}
