package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/google/uuid"
)

type (
	ChatAdapter struct {
		ServerUrl string
	}

	PlayerCreate struct {
		Name string `json:"name"`
	}
)

const (
	chat_player_path  = "%s/chat/%s/player/%s"
	correlation_id    = "X-Correlation-ID"
	content_typ_value = "application/json; charset=utf-8"
	content_typ       = "Content-Type"
)

func NewChatAdapter() *ChatAdapter {
	serverUrl := util.GetEnvWithFallback("CHAT_SERVER_URL", "http://theredshirts-chat:1204")
	return &ChatAdapter{ServerUrl: serverUrl}
}

func (adapter *ChatAdapter) AddPlayerToChat(context *util.Context, lobbyId uuid.UUID, playerId uuid.UUID, playerCreate *PlayerCreate) error {
	/*response, err := adapter.sendAddPlayerRequest(context, lobbyId, playerId, playerCreate)
	if err != nil {
		return fmt.Errorf("error while adding player to chat: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("wrong status of response while adding player to chat: %v", response.StatusCode)
	}*/
	return nil
}

func (adapter *ChatAdapter) DeletePlayerFromChat(context *util.Context, lobbyId uuid.UUID, playerId uuid.UUID) error {
	/*response, err := adapter.sendDeletePlayerRequest(context, lobbyId, playerId)
	if err != nil {
		return fmt.Errorf("error while deleting player from chat: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("wrong status of response while deleting player from chat: %v", response.StatusCode)
	}*/
	return nil
}

func (adapter *ChatAdapter) sendAddPlayerRequest(context *util.Context, lobbyId uuid.UUID, playerId uuid.UUID, playerCreate *PlayerCreate) (*http.Response, error) {
	client := &http.Client{}
	profileCreateMarshalled, err := json.Marshal(playerCreate)
	if err != nil {
		return nil, fmt.Errorf("player could not be marshaled: %v", err)
	}

	path := fmt.Sprintf(chat_player_path, adapter.ServerUrl, lobbyId, playerId)
	req, err := http.NewRequest(http.MethodPut, path, bytes.NewBuffer(profileCreateMarshalled))
	if err != nil {
		return nil, fmt.Errorf("request for player to add to chat could not be build: %v", err)
	}

	req.Header.Set(correlation_id, context.CorrelationId)
	req.Header.Set(content_typ, content_typ_value)
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("request for player to add to chat not possible: %v", err)
	}
	return resp, nil
}

func (adapter *ChatAdapter) sendDeletePlayerRequest(context *util.Context, lobbyId uuid.UUID, playerId uuid.UUID) (*http.Response, error) {
	client := &http.Client{}

	path := fmt.Sprintf(chat_player_path, adapter.ServerUrl, lobbyId, playerId)
	req, err := http.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, fmt.Errorf("request to delete player from chat could not be build: %v", err)
	}

	req.Header.Set(correlation_id, context.CorrelationId)
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("request to delete player from chat not possible: %v", err)
	}
	return resp, nil
}
