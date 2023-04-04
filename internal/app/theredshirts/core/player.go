package core

import (
	"fmt"
	"time"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/db"
	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/google/uuid"
)

func (core CoreFacade) CreatePlayer(context *util.Context, player *Player, password string) error {
	context.Logger.Debugf("Creating Player: %+v", *player)
	tx, err := core.startTransaction()
	defer core.handleTransaction(tx, context, err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	err = core.createPlayer(context, tx, player.ID, player.Name, player.LobbyId, password, player.Payload)
	return err
}

func (core CoreFacade) createPlayer(context *util.Context, tx *transaction, playerId uuid.UUID, playerName string, lobbyId uuid.UUID, password string, payload map[string]interface{}) error {

	player, err := core.getPlayer(tx, playerId)
	if err != nil {
		return err
	}

	if player != nil {
		return nil
	}

	lobby, err := tx.dbTx.GetLobbyById(lobbyId)
	if err != nil {
		return fmt.Errorf("something went wrong while loading lobby %v from database: %v", lobbyId, err)
	}
	if lobby.Password != password {
		return ErrWrongLobbyPassword
	}
	if err := tx.dbTx.CreatePlayer(&db.Player{ID: playerId, Name: playerName, LobbyId: lobbyId, LastRefresh: time.Now(), Payload: payload}); err != nil {
		return fmt.Errorf("something went wrong while creating player %v from database: %v", playerId, err)
	}

	tx.messages = append(tx.messages, &message{senderPlayerId: playerId, lobbyId: lobbyId, topic: PLAYER_JOINS_LOBBY, payload: map[string]interface{}{"player_id": playerId}})

	return nil
}

func (core CoreFacade) UpdatePlayer(context *util.Context, player *Player) error {
	context.Logger.Debugf("Updating Player: %+v", *player)
	tx, err := core.startTransaction()
	defer core.handleTransaction(tx, context, err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	err = core.updatePlayer(context, tx, player)
	return err
}

func (core CoreFacade) updatePlayer(context *util.Context, tx *transaction, player *Player) error {

	foundPlayer, err := tx.dbTx.GetPlayerById(player.ID)
	if err != nil {
		return err
	}

	if foundPlayer == nil {
		return fmt.Errorf("player [%v] not found", player.ID)
	}

	foundPlayer.LastRefresh = time.Now()
	foundPlayer.Name = player.Name
	foundPlayer.Payload = player.Payload
	if err := tx.dbTx.UpdatePlayer(foundPlayer); err != nil {
		return fmt.Errorf("something went wrong while updating player [%v]: %v", player.ID, err)
	}
	tx.messages = append(tx.messages, &message{senderPlayerId: foundPlayer.ID, lobbyId: foundPlayer.LobbyId, topic: PLAYER_UPDATED, payload: map[string]interface{}{"player_id": foundPlayer.ID}})
	return nil
}

func (core CoreFacade) UpdatePlayerLastRefresh(context *util.Context, playerId uuid.UUID) error {
	context.Logger.Debugf("Updating Player last refresh [%v]", playerId)
	tx, err := core.startTransaction()
	defer core.handleTransaction(tx, context, err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	err = core.updatePlayerLastRefresh(tx, playerId)
	return err
}

func (core CoreFacade) updatePlayerLastRefresh(tx *transaction, playerId uuid.UUID) error {

	if err := tx.dbTx.UpdatePlayerLastRefresh(playerId, time.Now()); err != nil {
		return fmt.Errorf("something went wrong while updating last refresh of player [%v]: %v", playerId, err)
	}
	return nil
}

func (core CoreFacade) DeletePlayer(context *util.Context, playerId uuid.UUID) error {
	context.Logger.Debugf("Deleting Player [%v]", playerId)
	tx, err := core.startTransaction()
	defer core.handleTransaction(tx, context, err)
	if err != nil {
		return fmt.Errorf("something went wrong while creating transaction: %v", err)
	}
	err = core.deletePlayer(context, tx, playerId)
	return err
}

func (core CoreFacade) deletePlayer(context *util.Context, tx *transaction, playerId uuid.UUID) error {
	context.Logger.Debugf("Delete player [%s]", playerId)
	player, err := core.getPlayer(tx, playerId)
	if err != nil {
		return err
	}

	if player == nil {
		context.Logger.Debugf("No player for id [%s] found", playerId)
		return nil
	}

	lobby, err := core.getLobby(tx, player.LobbyId)
	if err != nil {
		return err
	}

	if lobby.Owner.ID == playerId {
		context.Logger.Debugf("Player who is leaving the lobby is also owner of lobby [%s]", player.LobbyId)
		foundNewOwner := findPlayerNot(lobby.Players, playerId)
		if foundNewOwner == nil {
			context.Logger.Debugf("No new owner found. Deleting lobby [%s]", player.LobbyId)
			if err := tx.dbTx.DeletePlayer(playerId); err != nil {
				return fmt.Errorf("error while deleting player [%v] from database: %v", playerId, err)
			}

			tx.messages = append(tx.messages, &message{senderPlayerId: uuid.Nil, lobbyId: player.LobbyId, topic: PLAYER_LEAVES_LOBBY, payload: map[string]interface{}{"player_id": playerId}})

			if err := core.deleteLobby(tx, context, player.LobbyId, playerId); err != nil {
				return err
			}
			return nil
		} else {
			context.Logger.Debugf("Player [%s] found to be the new owner of lobby [%s]", foundNewOwner.ID, player.LobbyId)
			lobby.Owner = foundNewOwner
			if err := core.updateLobby(context, tx, lobby); err != nil {
				return err
			}
		}
	}
	if err := tx.dbTx.DeletePlayer(playerId); err != nil {
		return fmt.Errorf("error while deleting player [%v] from database: %v", playerId, err)
	}

	tx.messages = append(tx.messages, &message{senderPlayerId: uuid.Nil, lobbyId: player.LobbyId, topic: PLAYER_LEAVES_LOBBY, payload: map[string]interface{}{"player_id": playerId}})

	return nil
}
func (core CoreFacade) GetPlayer(context *util.Context, playerId uuid.UUID) (*Player, error) {
	context.Logger.Debugf("Getting Player [%v]", playerId)
	tx, err := core.startTransaction()
	defer core.handleTransaction(tx, context, err)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while creating transaction: %v", err)
	}

	player, err := core.getPlayer(tx, playerId)
	return player, err
}

func (core CoreFacade) getPlayer(tx *transaction, playerId uuid.UUID) (*Player, error) {
	player, err := tx.dbTx.GetPlayerById(playerId)
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
