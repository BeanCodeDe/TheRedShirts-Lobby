package core

import (
	"errors"
	"fmt"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/db"
	"github.com/google/uuid"
)

func (core CoreFacade) CreateLobby(lobby *Lobby) error {
	dbLobby := mapToDBLobby(lobby)

	tx, err := core.db.StartTransaction()
	defer core.db.HandleTransaction(tx, err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	if err := core.db.CreateLobby(dbLobby); err != nil {
		if !errors.Is(err, db.ErrLobbyAlreadyExists) {
			return fmt.Errorf("error while creating lobby: %v", err)
		}
		foundLobby, err := core.db.GetLobbyById(lobby.ID)
		if err != nil {
			return fmt.Errorf("something went wrong while checking if lobby [%v] is already created: %v", lobby.ID, err)
		}

		if lobby.Name != foundLobby.Name || lobby.Password != foundLobby.Password {
			return fmt.Errorf("request of lobby [%v] doesn't match lobby from database [%v]", lobby, foundLobby)
		}

	}

	if err := core.joinLobby(lobby.Owner.ID, lobby.Name, lobby.ID, lobby.Password); err != nil {
		return err
	}

	return nil
}

func (core CoreFacade) UpdateLobby(lobby *Lobby) error {
	tx, err := core.db.StartTransaction()
	defer core.db.HandleTransaction(tx, err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	dbLobby, err := core.db.GetLobbyById(lobby.ID)
	if err != nil {
		return fmt.Errorf("something went wrong while loading lobby [%v] from database: %v", lobby.ID, err)
	}

	dbLobby.Name = lobby.Name
	dbLobby.Difficulty = lobby.Difficulty
	dbLobby.Password = lobby.Password

	if err := core.db.UpdateLobby(dbLobby); err != nil {
		if err != nil {
			return fmt.Errorf("something went wrong while updating lobby [%v]: %v", lobby.ID, err)
		}
	}
	return nil
}

func (core CoreFacade) DeleteLobby(lobbyId uuid.UUID) error {
	tx, err := core.db.StartTransaction()
	defer core.db.HandleTransaction(tx, err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	if err := core.db.DeleteAllPlayerInLobby(lobbyId); err != nil {
		return fmt.Errorf("an error accourd while deleting all players from lobby [%v]: %v", lobbyId, err)
	}

	if err := core.db.DeleteLobby(lobbyId); err != nil {
		return fmt.Errorf("an error accourd while deleting lobby [%v]: %v", lobbyId, err)
	}
	return nil
}

func (core CoreFacade) GetLobby(lobbyId uuid.UUID) (*Lobby, error) {
	lobby, err := core.db.GetLobbyById(lobbyId)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while loading lobby [%v] from database: %v", lobby.ID, err)
	}

	players, err := core.db.GetAllPlayersInLobby(lobbyId)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while loading players of lobby [%v] from database: %v", lobby.ID, err)
	}

	owner, err := core.db.GetPlayerById(lobby.Owner)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while loading owner [%v] of lobby [%v] from database: %v", lobby.Owner, lobby.ID, err)
	}

	return mapToLobby(lobby, mapToPlayer(owner), mapToPlayers(players)), nil
}

func (core CoreFacade) GetLobbies() ([]*Lobby, error) {
	lobbies, err := core.db.GetAllLobbies()
	if err != nil {
		return nil, fmt.Errorf("something went wrong while loading all lobbies from database: %v", err)
	}

	coreLobbies := make([]*Lobby, len(lobbies))
	for index, lobby := range lobbies {

		players, err := core.db.GetAllPlayersInLobby(lobby.ID)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while loading players of lobby [%v] from database: %v", lobby.ID, err)
		}

		owner, err := core.db.GetPlayerById(lobby.Owner)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while loading owner [%v] of lobby [%v] from database: %v", lobby.Owner, lobby.ID, err)
		}
		coreLobbies[index] = mapToLobby(lobby, mapToPlayer(owner), mapToPlayers(players))
	}

	return coreLobbies, nil
}

func (core CoreFacade) JoinLobby(join *Join) error {
	tx, err := core.db.StartTransaction()
	defer core.db.HandleTransaction(tx, err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	err = core.joinLobby(join.PlayerId, join.Name, join.LobbyID, join.Password)
	return err
}

func (core CoreFacade) joinLobby(playerId uuid.UUID, playerName string, lobbyId uuid.UUID, password string) error {
	if err := core.LeaveLobby(playerId); err != nil {
		return err
	}
	lobby, err := core.db.GetLobbyById(lobbyId)
	if err != nil {
		return fmt.Errorf("something went wrong while loading lobby %v from database: %v", lobbyId, err)
	}
	if lobby.Password != password {
		return ErrWrongLobbyPassword
	}
	if err := core.db.CreatePlayer(&db.Player{ID: playerId, Name: playerName, LobbyId: lobbyId}); err != nil {
		return fmt.Errorf("something went wrong while creating player %v from database: %v", playerId, err)
	}

	return nil
}

func (core CoreFacade) LeaveLobby(playerId uuid.UUID) error {
	if err := core.db.DeletePlayer(playerId); err != nil {
		return fmt.Errorf("something went wrong while deleting player %v from database: %v", playerId, err)
	}
	return nil
}
