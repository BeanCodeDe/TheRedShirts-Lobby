package core

import (
	"fmt"
	"time"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/adapter"
	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/db"
	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/google/uuid"
)

func (core CoreFacade) CreatePlayer(context *util.Context, player *Player, password string) error {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	err = core.createPlayer(context, tx, player.ID, player.Name, player.LobbyId, password, player.Payload)
	return err
}

func (core CoreFacade) createPlayer(context *util.Context, tx db.DBTx, playerId uuid.UUID, playerName string, lobbyId uuid.UUID, password string, payload map[string]interface{}) error {

	player, err := core.getPlayer(tx, playerId)
	if err != nil {
		return err
	}

	if player != nil {
		return nil
	}

	lobby, err := tx.GetLobbyById(lobbyId)
	if err != nil {
		return fmt.Errorf("something went wrong while loading lobby %v from database: %v", lobbyId, err)
	}
	if lobby.Password != password {
		return ErrWrongLobbyPassword
	}
	if err := tx.CreatePlayer(&db.Player{ID: playerId, Name: playerName, LobbyId: lobbyId, LastRefresh: time.Now(), Payload: payload}); err != nil {
		return fmt.Errorf("something went wrong while creating player %v from database: %v", playerId, err)
	}
	if err := core.chatAdapter.AddPlayerToChat(context, lobbyId, playerId, &adapter.PlayerCreate{Name: playerName}); err != nil {
		return fmt.Errorf("error while adding player to chat: %v", err)
	}
	return nil
}

func (core CoreFacade) UpdatePlayer(context *util.Context, player *Player) error {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	err = core.updatePlayer(context, tx, player)
	return err
}

func (core CoreFacade) updatePlayer(context *util.Context, tx db.DBTx, player *Player) error {

	foundPlayer, err := tx.GetPlayerById(player.ID)
	if err != nil {
		return err
	}

	if foundPlayer != nil {
		return nil
	}

	foundPlayer.LastRefresh = time.Now()
	foundPlayer.Name = player.Name
	foundPlayer.Payload = player.Payload
	if err := tx.UpdatePlayer(foundPlayer); err != nil {
		return fmt.Errorf("something went wrong while creating player %v from database: %v", player.ID, err)
	}
	return nil
}

func (core CoreFacade) DeletePlayer(context *util.Context, lobbyId uuid.UUID, playerId uuid.UUID) error {
	tx, err := core.db.StartTransaction()
	defer tx.HandleTransaction(err)

	err = core.deletePlayer(context, tx, lobbyId, playerId)
	return err
}

func (core CoreFacade) deletePlayer(context *util.Context, tx db.DBTx, lobbyId uuid.UUID, playerId uuid.UUID) error {
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
			if err := core.deleteLobby(tx, context, lobbyId, playerId); err != nil {
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
	return nil
}

func (core CoreFacade) getPlayer(tx db.DBTx, playerId uuid.UUID) (*Player, error) {
	player, err := tx.GetPlayerById(playerId)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while loading player [%v] from database: %v", playerId, err)
	}
	return mapToPlayer(player), nil
}

func mapToPlayer(player *db.Player) *Player {
	if player == nil {
		return nil
	}
	return &Player{ID: player.ID, Name: player.Name, LastRefresh: player.LastRefresh, LobbyId: player.LobbyId, Payload: player.Payload}
}

func mapToPlayers(dbPlayers []*db.Player) []*Player {
	players := make([]*Player, len(dbPlayers))
	for index, player := range dbPlayers {
		players[index] = mapToPlayer(player)
	}
	return players
}

func findPlayerNot(players []*Player, notPlayerId uuid.UUID) *Player {
	for _, player := range players {
		if player.ID != notPlayerId {
			return player
		}
	}
	return nil
}
