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

	if err := core.joinLobby(context, tx, lobby.Owner.ID, lobby.Owner.Name, lobby.Owner.Team, lobby.ID, lobby.Password); err != nil {
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

	if dbLobby.Status == lobby_playing {
		return fmt.Errorf("can not change lobby when status is playing")
	}

	dbLobby.Name = lobby.Name
	dbLobby.Difficulty = lobby.Difficulty
	dbLobby.Owner = lobby.Owner.ID
	dbLobby.Password = lobby.Password
	dbLobby.MissionLength = lobby.MissionLength
	dbLobby.CrewMembers = lobby.CrewMembers
	dbLobby.MaxPlayers = lobby.MaxPlayers
	dbLobby.ExpansionPacks = lobby.ExpansionPacks

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

func (core CoreFacade) JoinLobby(context *util.Context, join *Join) error {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	err = core.joinLobby(context, tx, join.PlayerId, join.Name, join.Team, join.LobbyID, join.Password)
	return err
}

func (core CoreFacade) joinLobby(context *util.Context, tx db.DBTx, playerId uuid.UUID, playerName string, teamName string, lobbyId uuid.UUID, password string) error {

	player, err := core.getPlayer(tx, playerId)
	if err != nil {
		return err
	}

	if player != nil {
		if lobbyId != player.LobbyId {
			if err := core.leaveLobby(context, tx, lobbyId, playerId); err != nil {
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
	if err := core.chatAdapter.AddPlayerToChat(context, lobbyId, playerId, &adapter.PlayerCreate{Name: playerName, Team: teamName}); err != nil {
		return fmt.Errorf("error while adding player to chat: %v", err)
	}
	return nil
}

func (core CoreFacade) LeaveLobby(context *util.Context, lobbyId uuid.UUID, playerId uuid.UUID) error {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)

	err = core.leaveLobby(context, tx, lobbyId, playerId)
	return err
}

func (core CoreFacade) leaveLobby(context *util.Context, tx db.DBTx, lobbyId uuid.UUID, playerId uuid.UUID) error {
	context.Logger.Debugf("Player [%s] leaves lobby [%s]", playerId, lobbyId)
	player, err := core.getPlayer(tx, playerId)
	if err != nil {
		return err
	}

	if player == nil {
		context.Logger.Debugf("No player for id [%s] found", playerId)
		return nil
	}

	if player.LobbyId != lobbyId {
		context.Logger.Debugf("Lobby id [%s] from player doesent match id [%s] from request", player.LobbyId, lobbyId)
		return nil
	}

	lobby, err := core.getLobby(tx, lobbyId)
	if err != nil {
		return err
	}

	if lobby.Owner.ID == playerId {
		context.Logger.Debugf("Player who is leaving is also owner of lobby [%s]", lobbyId)
		foundNewOwner := findPlayerNot(lobby.Players, playerId)
		if foundNewOwner == nil {
			context.Logger.Debugf("No new owner found. Deleting lobby [%s]", lobbyId)
			if err := core.deleteLobby(tx, lobbyId); err != nil {
				return err
			}
		} else {
			context.Logger.Debugf("Player [%s] found to be the new owner of lobby [%s]", foundNewOwner.ID, lobbyId)
			lobby.Owner = foundNewOwner
			if err := core.updateLobby(tx, lobby); err != nil {
				return err
			}
		}
	}

	context.Logger.Debugf("Delete player [%s]", playerId)
	if err := tx.DeletePlayer(playerId); err != nil {
		return fmt.Errorf("something went wrong while deleting player %v from database: %v", playerId, err)
	}

	context.Logger.Debugf("Remove player [%s] from lobby chat [%s]", playerId, lobbyId)
	if err := core.chatAdapter.DeletePlayerFromChat(context, lobbyId, playerId); err != nil {
		return fmt.Errorf("error while deleting player from chat: %v", err)
	}
	return nil
}

func mapToDBLobby(lobby *Lobby) *db.Lobby {
	return &db.Lobby{ID: lobby.ID, Status: lobby.Status, Name: lobby.Name, Owner: lobby.Owner.ID, Password: lobby.Password, Difficulty: lobby.Difficulty, MissionLength: lobby.MissionLength, CrewMembers: lobby.CrewMembers, MaxPlayers: lobby.MaxPlayers, ExpansionPacks: lobby.ExpansionPacks}
}

func mapToLobby(lobby *db.Lobby, owner *Player, players []*Player) *Lobby {
	return &Lobby{ID: lobby.ID, Status: lobby.Status, Name: lobby.Name, Owner: owner, Password: lobby.Password, Difficulty: lobby.Difficulty, MissionLength: lobby.MissionLength, CrewMembers: lobby.CrewMembers, MaxPlayers: lobby.MaxPlayers, ExpansionPacks: lobby.ExpansionPacks, Players: players}
}
