package core

import (
	"fmt"
	"time"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func (core CoreFacade) startCleanUp() {
	log.Info("Start auto cleanup of lobbies")
	s := gocron.NewScheduler(time.UTC)

	s.Every(10).Seconds().Do(func() {
		correlationId := uuid.NewString()
		logger := log.WithFields(log.Fields{
			"Scavenger": correlationId,
		})
		context := &util.Context{CorrelationId: correlationId, Logger: logger}

		tx, err := core.startTransaction()
		if err != nil {
			logger.Warnf("something went wrong while creating transaction: %v", err)
			return
		}
		defer core.rollback(tx)

		if err := core.cleanUpAfkPlayers(context, tx); err != nil {
			log.Warn("Error while scheduling: %v", err)
			return
		}
		if err := core.commit(tx, context); err != nil {
			log.Warn("Error while committing changes: %v", err)
		}
	})

	s.StartAsync()
}

func (core CoreFacade) cleanUpAfkPlayers(context *util.Context, tx *transaction) error {
	warningTime := time.Now().Add(time.Second * -5)
	deleteTime := warningTime.Add(time.Second * -10)
	players, err := tx.dbTx.GetPlayersLastRefresh(warningTime)
	if err != nil {
		return fmt.Errorf("error while cleaning up afk players: %v", err)
	}
	for _, player := range players {
		if player.LastRefresh.Before(deleteTime) {
			err := core.deletePlayer(context, tx, player.ID)
			if err != nil {
				return err
			}
		} else {
			tx.messages = append(tx.messages, &message{senderPlayerId: uuid.Nil, lobbyId: player.LobbyId, topic: PLAYER_LAGGING, payload: map[string]interface{}{"player_id": player.ID}})
		}
	}
	return nil
}
