package edge

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

const (
	MessageTypeRegister  = "register"
	MessageTypeHeartbeat = "heartbeat"
	MessageTypeRequest   = "request"
	MessageTypeResponse  = "response"
	MessageTypeError     = "error"
)

type TunnelMessage struct {
	Type      string          `json:"type"`
	ID        string          `json:"id,omitempty"`
	EdgeID    string          `json:"edge_id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type TunnelRequest struct {
	Method string      `json:"method"`
	Path   string      `json:"path"`
	Query  url.Values  `json:"query,omitempty"`
	Header http.Header `json:"header,omitempty"`
	Body   []byte      `json:"body,omitempty"`
}

type TunnelResponse struct {
	StatusCode int         `json:"status_code"`
	Header     http.Header `json:"header,omitempty"`
	Body       []byte      `json:"body,omitempty"`
	Error      string      `json:"error,omitempty"`
}

type TunnelError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewMessage(messageType, id, edgeID string, payload interface{}) (*TunnelMessage, error) {
	var raw json.RawMessage
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		raw = data
	}
	return &TunnelMessage{
		Type:      messageType,
		ID:        id,
		EdgeID:    edgeID,
		Timestamp: time.Now(),
		Payload:   raw,
	}, nil
}
