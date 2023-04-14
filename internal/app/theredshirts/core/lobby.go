package core

import (
	"errors"
	"fmt"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/db"
	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/google/uuid"
)

func (core CoreFacade) CreateLobby(context *util.Context, lobby *Lobby) error {
	context.Logger.Debugf("Creating lobby %+v:", *lobby)
	tx, err := core.startTransaction()
	if err != nil {
		return err
	}
	defer core.rollback(tx)
	lobby.Status = lobby_open
	if err := core.createLobby(tx, context, lobby); err != nil {
		return err
	}
	return core.commit(tx, context)
}

func (core CoreFacade) createLobby(tx *transaction, context *util.Context, lobby *Lobby) error {
	dbLobby := mapToDBLobby(lobby)

	if err := tx.dbTx.CreateLobby(dbLobby); err != nil {
		if !errors.Is(err, db.ErrLobbyAlreadyExists) {
			return fmt.Errorf("error while creating lobby: %v", err)
		}
		foundLobby, err := tx.dbTx.GetLobbyById(lobby.ID)
		if err != nil {
			return fmt.Errorf("something went wrong while checking if lobby [%v] is already created: %v", lobby.ID, err)
		}

		if lobby == nil {
			return fmt.Errorf("lobby not found")
		}

		if lobby.Name != foundLobby.Name || lobby.Password != foundLobby.Password {
			return fmt.Errorf("request of lobby [%v] doesn't match lobby from database [%v]", lobby, foundLobby)
		}

	}

	if err := core.createPlayer(context, tx, lobby.Owner.ID, lobby.Owner.Name, lobby.ID, lobby.Password, lobby.Owner.Spectator, lobby.Owner.Payload); err != nil {
		return err
	}

	return nil
}

func (core CoreFacade) UpdateLobby(context *util.Context, lobby *Lobby, playerId uuid.UUID) error {
	context.Logger.Debugf("Updating lobby %+v:", *lobby)
	tx, err := core.startTransaction()
	if err != nil {
		return err
	}
	defer core.rollback(tx)

	dbLobby, err := tx.dbTx.GetLobbyById(lobby.ID)
	if err != nil {
		return fmt.Errorf("something went wrong while loading lobby [%v] from database: %v", lobby.ID, err)
	}

	if dbLobby == nil {
		return fmt.Errorf("lobby not found")
	}

	if dbLobby.Owner != playerId {
		return fmt.Errorf("player [%v] is not owner [%v] of the lobby [%v]", lobby.Owner.ID, dbLobby.Owner, lobby.ID)
	}

	if err := core.updateLobby(context, tx, lobby, playerId); err != nil {
		return err
	}

	return core.commit(tx, context)
}

func (core CoreFacade) updateLobby(context *util.Context, tx *transaction, lobby *Lobby, playerId uuid.UUID) error {
	dbLobby, err := tx.dbTx.GetLobbyById(lobby.ID)
	if err != nil {
		return fmt.Errorf("something went wrong while loading lobby [%v] from database: %v", lobby.ID, err)
	}

	if dbLobby == nil {
		return fmt.Errorf("lobby not found")
	}

	dbLobby.Name = lobby.Name
	dbLobby.Status = lobby.Status
	dbLobby.Difficulty = lobby.Difficulty
	dbLobby.Owner = lobby.Owner.ID
	dbLobby.Password = lobby.Password
	dbLobby.MissionLength = lobby.MissionLength
	dbLobby.NumberOfCrewMembers = lobby.NumberOfCrewMembers
	dbLobby.MaxPlayers = lobby.MaxPlayers
	dbLobby.ExpansionPacks = lobby.ExpansionPacks
	dbLobby.Payload = lobby.Payload

	if err := tx.dbTx.UpdateLobby(dbLobby); err != nil {
		if err != nil {
			return fmt.Errorf("something went wrong while updating lobby [%v]: %v", lobby.ID, err)
		}
	}
	tx.messages = append(tx.messages, &message{senderPlayerId: playerId, lobbyId: lobby.ID, topic: PLAYER_UPDATES_LOBBY, payload: map[string]interface{}{}})

	return nil
}

func (core CoreFacade) UpdateLobbyStatus(context *util.Context, lobby *Lobby, playerId uuid.UUID) error {
	context.Logger.Debugf("Updating lobby status %+v:", *lobby)
	tx, err := core.startTransaction()
	if err != nil {
		return err
	}
	defer core.rollback(tx)
	if err = core.updateLobbyStatus(context, tx, lobby, playerId); err != nil {
		return err
	}
	return core.commit(tx, context)
}

func (core CoreFacade) updateLobbyStatus(context *util.Context, tx *transaction, lobby *Lobby, playerId uuid.UUID) error {
	dbLobby, err := tx.dbTx.GetLobbyById(lobby.ID)
	if err != nil {
		return fmt.Errorf("something went wrong while loading lobby [%v] from database: %v", lobby.ID, err)
	}

	if dbLobby == nil {
		return fmt.Errorf("lobby not found")
	}

	if dbLobby.Owner != playerId {
		return fmt.Errorf("player [%v] is not owner [%v] of the lobby [%v]", playerId, dbLobby.Owner, lobby.ID)
	}

	dbLobby.Status = lobby.Status

	if err := tx.dbTx.UpdateLobby(dbLobby); err != nil {
		if err != nil {
			return fmt.Errorf("something went wrong while updating state of lobby [%v]: %v", lobby.ID, err)
		}
	}

	tx.messages = append(tx.messages, &message{senderPlayerId: playerId, lobbyId: lobby.ID, topic: PLAYER_UPDATES_LOBBY, payload: map[string]interface{}{}})
	return nil
}

func (core CoreFacade) DeleteLobby(context *util.Context, lobbyId uuid.UUID, playerId uuid.UUID) error {
	context.Logger.Debugf("Deleting lobby: LobbyId [%v], OwnerId [%v]", lobbyId, playerId)
	tx, err := core.startTransaction()
	if err != nil {
		return err
	}
	defer core.rollback(tx)
	if err = core.deleteLobby(tx, context, lobbyId, playerId); err != nil {
		return nil
	}
	return core.commit(tx, context)
}

func (core CoreFacade) deleteLobby(tx *transaction, context *util.Context, lobbyId uuid.UUID, playerId uuid.UUID) error {
	lobby, err := tx.dbTx.GetLobbyById(lobbyId)
	if err != nil {
		return fmt.Errorf("an error accourd while loading lobby [%v]: %v", lobbyId, err)
	}

	if lobby == nil {
		return nil
	}

	if lobby.Owner != playerId {
		return fmt.Errorf("player [%v] is not owner [%v] of the lobby [%v]", playerId, lobby.Owner, lobbyId)
	}

	if err := tx.dbTx.DeleteLobby(lobbyId); err != nil {
		return fmt.Errorf("an error accourd while deleting lobby [%v]: %v", lobbyId, err)
	}
	return nil
}

func (core CoreFacade) GetLobby(context *util.Context, lobbyId uuid.UUID) (*Lobby, error) {
	context.Logger.Debugf("Get lobby: LobbyId [%v]", lobbyId)
	tx, err := core.startTransaction()
	if err != nil {
		return nil, err
	}
	defer core.rollback(tx)

	lobby, err := core.getLobby(tx, lobbyId)
	if err != nil {
		return nil, err
	}

	return lobby, core.commit(tx, context)
}

func (core CoreFacade) getLobby(tx *transaction, lobbyId uuid.UUID) (*Lobby, error) {
	lobby, err := tx.dbTx.GetLobbyById(lobbyId)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while loading lobby [%v] from database: %v", lobbyId, err)
	}

	if lobby == nil {
		return nil, fmt.Errorf("lobby not found")
	}

	players, err := tx.dbTx.GetAllPlayersInLobby(lobbyId)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while loading players of lobby [%v] from database: %v", lobby.ID, err)
	}

	owner, err := tx.dbTx.GetPlayerById(lobby.Owner)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while loading owner [%v] of lobby [%v] from database: %v", lobby.Owner, lobby.ID, err)
	}

	return mapToLobby(lobby, mapToPlayer(owner), mapToPlayers(players)), nil
}

func (core CoreFacade) GetLobbies(context *util.Context) ([]*Lobby, error) {
	tx, err := core.startTransaction()
	if err != nil {
		return nil, err
	}
	defer core.rollback(tx)

	lobbies, err := tx.dbTx.GetAllLobbies()
	if err != nil {
		return nil, fmt.Errorf("something went wrong while loading all lobbies from database: %v", err)
	}

	coreLobbies := make([]*Lobby, len(lobbies))
	for index, lobby := range lobbies {

		players, err := tx.dbTx.GetAllPlayersInLobby(lobby.ID)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while loading players of lobby [%v] from database: %v", lobby.ID, err)
		}

		owner, err := tx.dbTx.GetPlayerById(lobby.Owner)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while loading owner [%v] of lobby [%v] from database: %v", lobby.Owner, lobby.ID, err)
		}
		coreLobbies[index] = mapToLobby(lobby, mapToPlayer(owner), mapToPlayers(players))
	}

	return coreLobbies, core.commit(tx, context)
}

func mapToDBLobby(lobby *Lobby) *db.Lobby {
	return &db.Lobby{ID: lobby.ID, Status: lobby.Status, Name: lobby.Name, Owner: lobby.Owner.ID, Password: lobby.Password, Difficulty: lobby.Difficulty, MissionLength: lobby.MissionLength, NumberOfCrewMembers: lobby.NumberOfCrewMembers, MaxPlayers: lobby.MaxPlayers, ExpansionPacks: lobby.ExpansionPacks, Payload: lobby.Payload}
}

func mapToLobby(lobby *db.Lobby, owner *Player, players []*Player) *Lobby {
	return &Lobby{ID: lobby.ID, Status: lobby.Status, Name: lobby.Name, Owner: owner, Password: lobby.Password, Difficulty: lobby.Difficulty, MissionLength: lobby.MissionLength, NumberOfCrewMembers: lobby.NumberOfCrewMembers, MaxPlayers: lobby.MaxPlayers, ExpansionPacks: lobby.ExpansionPacks, Players: players, Payload: lobby.Payload}
}
