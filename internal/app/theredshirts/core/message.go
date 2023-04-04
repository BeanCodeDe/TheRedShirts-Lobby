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

type message struct {
	senderPlayerId uuid.UUID
	lobbyId        uuid.UUID
	topic          string
	payload        map[string]interface{}
}

func (core CoreFacade) createMessageId(context *util.Context, lobbyId uuid.UUID, senderPlayerId uuid.UUID) (string, error) {
	msgId, err := core.messageAdapter.CreateMessageId(context, lobbyId, senderPlayerId)
	if err != nil {
		return "", fmt.Errorf("error while creating message id: %v", err)
	}
	return msgId, nil
}

func (core CoreFacade) createMessage(context *util.Context, message *message) error {
	if message.senderPlayerId == uuid.Nil {
		message.senderPlayerId = core.lobbyPlayerId
	}

	msgId, err := core.createMessageId(context, message.lobbyId, message.senderPlayerId)
	if err != nil {
		return err
	}
	if err := core.messageAdapter.CreateMessage(context, &adapter.Message{Topic: message.topic, Message: message.payload}, message.lobbyId, msgId, message.senderPlayerId); err != nil {
		return fmt.Errorf("error while sending message with topic %s: %v", message.topic, err)
	}
	return nil
}
