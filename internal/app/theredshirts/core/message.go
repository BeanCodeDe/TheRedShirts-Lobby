package core

import (
	"fmt"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/adapter"
	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/google/uuid"
)

const (
	//Topics
	PLAYER_JOINS_LOBBY   = "PLAYER_JOINS_LOBBY"
	PLAYER_LEAVES_LOBBY  = "PLAYER_LEAVES_LOBBY"
	PLAYER_UPDATES_LOBBY = "PLAYER_UPDATES_LOBBY"
	PLAYER_UPDATED       = "PLAYER_UPDATED"
	PLAYER_LAGGING       = "PLAYER_LAGGING"
)

func (core CoreFacade) createPlayerJoinsLobbyMessage(context util.Context, lobbyId uuid.UUID) error {
	return core.createMessage(&context, &adapter.Message{Topic: PLAYER_JOINS_LOBBY}, lobbyId)
}

func (core CoreFacade) createPlayerLeavesLobbyMessage(context util.Context, lobbyId uuid.UUID) error {
	return core.createMessage(&context, &adapter.Message{Topic: PLAYER_LEAVES_LOBBY}, lobbyId)
}

func (core CoreFacade) createPlayerUpdatesLobbyMessage(context util.Context, lobbyId uuid.UUID) error {
	return core.createMessage(&context, &adapter.Message{Topic: PLAYER_UPDATES_LOBBY}, lobbyId)
}

func (core CoreFacade) createPlayerUpdatedMessage(context util.Context, lobbyId uuid.UUID) error {
	return core.createMessage(&context, &adapter.Message{Topic: PLAYER_UPDATED}, lobbyId)
}

func (core CoreFacade) createPlayerLaggingMessage(context util.Context, lobbyId uuid.UUID) error {
	return core.createMessage(&context, &adapter.Message{Topic: PLAYER_LAGGING}, lobbyId)
}

func (core CoreFacade) createMessageId(context *util.Context, lobbyId uuid.UUID) (string, error) {
	msgId, err := core.messageAdapter.CreateMessageId(context, lobbyId)
	if err != nil {
		return "", fmt.Errorf("error while creating message id: %v", err)
	}
	return msgId, nil
}

func (core CoreFacade) createMessage(context *util.Context, message *adapter.Message, lobbyId uuid.UUID) error {
	msgId, err := core.createMessageId(context, lobbyId)
	if err != nil {
		return err
	}
	if err := core.messageAdapter.CreateMessage(context, nil, lobbyId, msgId); err != nil {
		return fmt.Errorf("error while sending message with topic %s: %v", message.Topic, err)
	}
	return nil
}
