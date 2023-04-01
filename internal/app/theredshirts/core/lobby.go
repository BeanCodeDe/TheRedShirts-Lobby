package core

import (
	"errors"
	"fmt"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/db"
	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/google/uuid"
)

func (core CoreFacade) CreateLobby(context *util.Context, lobby *Lobby) error {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)
	lobby.Status = lobby_open
	err = core.createLobby(tx, context, lobby)
	return err
}

func (core CoreFacade) createLobby(tx db.DBTx, context *util.Context, lobby *Lobby) error {
	dbLobby := mapToDBLobby(lobby)

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

	if err := core.createPlayer(context, tx, lobby.Owner.ID, lobby.Owner.Name, lobby.ID, lobby.Password, lobby.Owner.Payload); err != nil {
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

	if dbLobby.Owner == lobby.Owner.ID {
		return fmt.Errorf("only owner of the lobby can update the lobby")
	}

	dbLobby.Name = lobby.Name
	dbLobby.Difficulty = lobby.Difficulty
	dbLobby.Owner = lobby.Owner.ID
	dbLobby.Password = lobby.Password
	dbLobby.MissionLength = lobby.MissionLength
	dbLobby.NumberOfCrewMembers = lobby.NumberOfCrewMembers
	dbLobby.MaxPlayers = lobby.MaxPlayers
	dbLobby.ExpansionPacks = lobby.ExpansionPacks
	dbLobby.Payload = lobby.Payload

	if err := tx.UpdateLobby(dbLobby); err != nil {
		if err != nil {
			return fmt.Errorf("something went wrong while updating lobby [%v]: %v", lobby.ID, err)
		}
	}
	return nil
}

func (core CoreFacade) UpdateLobbyStatus(lobby *Lobby) error {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)
	err = core.updateLobbyStatus(tx, lobby)
	return err
}

func (core CoreFacade) updateLobbyStatus(tx db.DBTx, lobby *Lobby) error {
	dbLobby, err := tx.GetLobbyById(lobby.ID)
	if err != nil {
		return fmt.Errorf("something went wrong while loading lobby [%v] from database: %v", lobby.ID, err)
	}

	if dbLobby.Owner == lobby.Owner.ID {
		return fmt.Errorf("only owner of the lobby can change the state")
	}

	dbLobby.Status = lobby.Status

	if err := tx.UpdateLobby(dbLobby); err != nil {
		if err != nil {
			return fmt.Errorf("something went wrong while updating state of lobby [%v]: %v", lobby.ID, err)
		}
	}
	return nil
}

func (core CoreFacade) DeleteLobby(context *util.Context, lobbyId uuid.UUID, ownerId uuid.UUID) error {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	err = core.deleteLobby(tx, context, lobbyId, ownerId)
	return err
}

func (core CoreFacade) deleteLobby(tx db.DBTx, context *util.Context, lobbyId uuid.UUID, ownerId uuid.UUID) error {
	lobby, err := tx.GetLobbyById(lobbyId)
	if err != nil {
		return fmt.Errorf("an error accourd while deleting all players from lobby [%v]: %v", lobbyId, err)
	}

	if lobby.Owner != ownerId {
		return nil
	}

	players, err := tx.GetAllPlayersInLobby(lobbyId)
	if err != nil {
		return fmt.Errorf("error while loeading players to remove from lobby [%v]", lobbyId)
	}

	for _, player := range players {
		if err := core.chatAdapter.DeletePlayerFromChat(context, lobbyId, player.ID); err != nil {
			return fmt.Errorf("error while removing player [%v] from lobby chat [%v]: %v", ownerId, lobbyId, err)
		}
	}

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

	if lobby == nil {
		return nil, nil
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

func mapToDBLobby(lobby *Lobby) *db.Lobby {
	return &db.Lobby{ID: lobby.ID, Status: lobby.Status, Name: lobby.Name, Owner: lobby.Owner.ID, Password: lobby.Password, Difficulty: lobby.Difficulty, MissionLength: lobby.MissionLength, NumberOfCrewMembers: lobby.NumberOfCrewMembers, MaxPlayers: lobby.MaxPlayers, ExpansionPacks: lobby.ExpansionPacks, Payload: lobby.Payload}
}

func mapToLobby(lobby *db.Lobby, owner *Player, players []*Player) *Lobby {
	return &Lobby{ID: lobby.ID, Status: lobby.Status, Name: lobby.Name, Owner: owner, Password: lobby.Password, Difficulty: lobby.Difficulty, MissionLength: lobby.MissionLength, NumberOfCrewMembers: lobby.NumberOfCrewMembers, MaxPlayers: lobby.MaxPlayers, ExpansionPacks: lobby.ExpansionPacks, Players: players, Payload: lobby.Payload}
}
