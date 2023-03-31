package core

import (
	"fmt"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/db"
	"github.com/google/uuid"
)

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

func mapToDBPlayer(player *Player, lobbyId uuid.UUID) *db.Player {
	return &db.Player{ID: player.ID, Name: player.Name, LobbyId: lobbyId, Payload: player.Payload}
}

func findPlayerNot(players []*Player, notPlayerId uuid.UUID) *Player {
	for _, player := range players {
		if player.ID != notPlayerId {
			return player
		}
	}
	return nil
}
