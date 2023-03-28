package core

import (
	"fmt"
	"time"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/db"
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
)

func (core CoreFacade) startCleanUp() {
	log.Info("Start auto cleanup of lobbies")
	s := gocron.NewScheduler(time.UTC)

	s.Every(10).Seconds().Do(func() {
		tx, err := core.db.StartTransaction()
		defer tx.HandleTransaction(err)
		if err := core.cleanUpAfkPlayers(tx); err != nil {
			log.Warn("Error while scheduling: %v", err)
			return
		}
		if err := core.cleanUpEmptyLobbies(tx); err != nil {
			log.Warn("Error while scheduling: %v", err)
			return
		}
	})

	//s.StartAsync()
}

func (core CoreFacade) cleanUpAfkPlayers(tx db.DBTx) error {
	if err := tx.DeletePlayerOlderRefreshDate(time.Now().Add(time.Hour * -5)); err != nil {
		return fmt.Errorf("error while cleaning up afk players: %v", err)
	}
	return nil
}

func (core CoreFacade) cleanUpEmptyLobbies(tx db.DBTx) error {
	if err := tx.DeletePlayerOlderRefreshDate(time.Now().Add(time.Second * -5)); err != nil {
		return fmt.Errorf("error while cleaning empty lobbies: %v", err)
	}
	return nil
}
