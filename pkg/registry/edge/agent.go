package edge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type AgentConfig struct {
	EdgeID            string
	CloudURL          string
	LocalBaseURL      string
	Token             string
	ReconnectInterval time.Duration
	HeartbeatInterval time.Duration
	RequestTimeout    time.Duration
}

type Agent struct {
	cfg AgentConfig
}

func NewAgent(cfg AgentConfig) *Agent {
	if cfg.EdgeID == "" {
		cfg.EdgeID = defaultEdgeID()
	}
	if cfg.LocalBaseURL == "" {
		cfg.LocalBaseURL = "http://127.0.0.1:8119"
	}
	if cfg.ReconnectInterval <= 0 {
		cfg.ReconnectInterval = 5 * time.Second
	}
	if cfg.HeartbeatInterval <= 0 {
		cfg.HeartbeatInterval = 30 * time.Second
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 25 * time.Second
	}
	return &Agent{cfg: cfg}
}

func (a *Agent) Run(ctx context.Context) error {
	if a.cfg.CloudURL == "" {
		return fmt.Errorf("cloud url is required")
	}
	for {
		if err := a.connectOnce(ctx); err != nil && ctx.Err() != nil {
			return ctx.Err()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(a.cfg.ReconnectInterval):
		}
	}
}

func (a *Agent) connectOnce(ctx context.Context) error {
	cloudURL, err := url.Parse(a.cfg.CloudURL)
	if err != nil {
		return err
	}
	query := cloudURL.Query()
	query.Set("edge_id", a.cfg.EdgeID)
	cloudURL.RawQuery = query.Encode()

	header := http.Header{}
	header.Set("X-Edge-ID", a.cfg.EdgeID)
	if a.cfg.Token != "" {
		header.Set("Authorization", "Bearer "+a.cfg.Token)
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, cloudURL.String(), header)
	if err != nil {
		return err
	}
	defer conn.Close()

	errCh := make(chan error, 2)
	go a.heartbeatLoop(ctx, conn, errCh)
	go a.readLoop(ctx, conn, errCh)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (a *Agent) heartbeatLoop(ctx context.Context, conn *websocket.Conn, errCh chan<- error) {
	ticker := time.NewTicker(a.cfg.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		case <-ticker.C:
			msg, err := NewMessage(MessageTypeHeartbeat, "", a.cfg.EdgeID, map[string]string{"status": "alive"})
			if err != nil {
				errCh <- err
				return
			}
			if err := conn.WriteJSON(msg); err != nil {
				errCh <- err
				return
			}
		}
	}
}

func (a *Agent) readLoop(ctx context.Context, conn *websocket.Conn, errCh chan<- error) {
	for {
		var msg TunnelMessage
		if err := conn.ReadJSON(&msg); err != nil {
			errCh <- err
			return
		}
		if msg.Type != MessageTypeRequest {
			continue
		}
		go a.handleRequest(ctx, conn, &msg)
	}
}

func (a *Agent) handleRequest(ctx context.Context, conn *websocket.Conn, msg *TunnelMessage) {
	var tunnelReq TunnelRequest
	if err := json.Unmarshal(msg.Payload, &tunnelReq); err != nil {
		a.writeResponse(conn, msg.ID, &TunnelResponse{StatusCode: http.StatusBadRequest, Error: err.Error()})
		return
	}

	localURL, err := url.Parse(strings.TrimRight(a.cfg.LocalBaseURL, "/") + tunnelReq.Path)
	if err != nil {
		a.writeResponse(conn, msg.ID, &TunnelResponse{StatusCode: http.StatusBadRequest, Error: err.Error()})
		return
	}
	localURL.RawQuery = tunnelReq.Query.Encode()

	reqCtx, cancel := context.WithTimeout(ctx, a.cfg.RequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, tunnelReq.Method, localURL.String(), bytes.NewReader(tunnelReq.Body))
	if err != nil {
		a.writeResponse(conn, msg.ID, &TunnelResponse{StatusCode: http.StatusBadRequest, Error: err.Error()})
		return
	}
	req.Header = tunnelReq.Header.Clone()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		a.writeResponse(conn, msg.ID, &TunnelResponse{StatusCode: http.StatusBadGateway, Error: err.Error()})
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		a.writeResponse(conn, msg.ID, &TunnelResponse{StatusCode: http.StatusBadGateway, Error: err.Error()})
		return
	}

	a.writeResponse(conn, msg.ID, &TunnelResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       body,
	})
}

func (a *Agent) writeResponse(conn *websocket.Conn, requestID string, resp *TunnelResponse) {
	msg, err := NewMessage(MessageTypeResponse, requestID, a.cfg.EdgeID, resp)
	if err != nil {
		return
	}
	_ = conn.WriteJSON(msg)
}

func defaultEdgeID() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return fmt.Sprintf("edge-%d", time.Now().UnixNano())
	}
	return hostname
}
