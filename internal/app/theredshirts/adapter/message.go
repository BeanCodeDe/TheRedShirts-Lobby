package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/util"
	"github.com/google/uuid"
)

type (
	MessageAdapter struct {
		ServerUrl string
	}
	Message struct {
		Topic   string                 `json:"topic"`
		Message map[string]interface{} `json:"message"`
	}
)

const (
	create_message_id_path = "%s/message/%s/msg"
	create_message_path    = "%s/message/%s/msg/%s"
	correlation_id         = "X-Correlation-ID"
	content_typ_value      = "application/json; charset=utf-8"
	content_typ            = "Content-Type"
	header_player_id       = "playerId"
)

func NewMessageAdapter() (*MessageAdapter, error) {
	serverUrl := util.GetEnvWithFallback("MESSAGE_SERVER_URL", "http://theredshirts-message:1203")

	return &MessageAdapter{ServerUrl: serverUrl}, nil
}

func (adapter *MessageAdapter) CreateMessageId(context *util.Context, lobbyId uuid.UUID, senderPlayerId uuid.UUID) (string, error) {
	response, err := adapter.sendCreateMessageId(context, lobbyId, senderPlayerId)
	if err != nil {
		return "", fmt.Errorf("error while creating message id: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("wrong status of response while creating message id: %v", response.StatusCode)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("could not parse response Body: %v", err)
	}

	msgId := string(bodyBytes)

	return msgId, nil
}

func (adapter *MessageAdapter) CreateMessage(context *util.Context, message *Message, lobbyId uuid.UUID, msgId string, senderPlayerId uuid.UUID) error {
	response, err := adapter.sendCreateMessage(context, message, lobbyId, senderPlayerId, msgId)
	if err != nil {
		return fmt.Errorf("error while creating message: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("wrong status of response while creating message: %v", response.StatusCode)
	}

	return nil
}

func (adapter *MessageAdapter) sendCreateMessageId(context *util.Context, lobbyId uuid.UUID, playerId uuid.UUID) (*http.Response, error) {
	client := &http.Client{}

	path := fmt.Sprintf(create_message_id_path, adapter.ServerUrl, lobbyId)
	req, err := http.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, fmt.Errorf("request to create message id could not be build: %v", err)
	}

	req.Header.Set(correlation_id, context.CorrelationId)
	req.Header.Set(header_player_id, playerId.String())
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("request to create message id not possible: %v", err)
	}
	return resp, nil
}

func (adapter *MessageAdapter) sendCreateMessage(context *util.Context, message *Message, lobbyId uuid.UUID, playerId uuid.UUID, msgId string) (*http.Response, error) {
	client := &http.Client{}
	jsonReq, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("error while marshal message: %v", err)
	}
	path := fmt.Sprintf(create_message_path, adapter.ServerUrl, lobbyId, msgId)
	req, err := http.NewRequest(http.MethodPut, path, bytes.NewBuffer(jsonReq))
	if err != nil {
		return nil, fmt.Errorf("request to create message could not be build: %v", err)
	}

	req.Header.Set(correlation_id, context.CorrelationId)
	req.Header.Set("uber-trace-id", context.CorrelationId)
	req.Header.Set(header_player_id, playerId.String())
	req.Header.Set(content_typ, content_typ_value)
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("request to create message not possible: %v", err)
	}
	return resp, nil
}
