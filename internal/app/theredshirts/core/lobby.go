package core

import (
	"errors"
	"fmt"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/db"
)

func (core CoreFacade) CreateLobby(lobby *Lobby) error {
	dbLobby := mapToDBLobby(lobby)

	if err := core.db.CreateLobby(dbLobby); err != nil {
		if errors.Is(err, db.ErrLobbyAlreadyExists) {
			foundLobby, err := core.db.GetLobbyById(lobby.ID)
			if err != nil {
				return fmt.Errorf("something went wrong while checking if lobby [%v] is already created: %v", lobby.ID, err)
			}

			if lobby.Name != foundLobby.Name || lobby.Password != foundLobby.Password {
				return fmt.Errorf("request of lobby [%v] doesn't match lobby from database [%v]", lobby, foundLobby)
			}

			return nil
		}
		return fmt.Errorf("error while creating lobby: %v", err)
	}
	return nil
}
