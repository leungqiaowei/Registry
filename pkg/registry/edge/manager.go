package edge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type TunnelManager struct {
	clients  map[string]*TunnelClient
	pending  map[string]chan *TunnelResponse
	mutex    sync.RWMutex
	upgrader websocket.Upgrader
}

type EdgeInfo struct {
	EdgeID      string    `json:"edge_id"`
	RemoteAddr  string    `json:"remote_addr"`
	ConnectedAt time.Time `json:"connected_at"`
	LastSeen    time.Time `json:"last_seen"`
	Status      string    `json:"status"`
}

func NewTunnelManager() *TunnelManager {
	return &TunnelManager{
		clients: make(map[string]*TunnelClient),
		pending: make(map[string]chan *TunnelResponse),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (m *TunnelManager) AcceptEdge(w http.ResponseWriter, r *http.Request) error {
	edgeID := r.URL.Query().Get("edge_id")
	if edgeID == "" {
		edgeID = r.Header.Get("X-Edge-ID")
	}
	if edgeID == "" {
		return fmt.Errorf("edge_id is required")
	}

	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	client := newTunnelClient(edgeID, r.RemoteAddr, conn, m)
	m.register(client)
	client.run()
	return nil
}

func (m *TunnelManager) register(client *TunnelClient) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if old := m.clients[client.EdgeID]; old != nil {
		old.Close()
	}
	m.clients[client.EdgeID] = client
}

func (m *TunnelManager) unregister(client *TunnelClient) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if current := m.clients[client.EdgeID]; current == client {
		delete(m.clients, client.EdgeID)
	}
}

func (m *TunnelManager) ListEdges() []EdgeInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	list := make([]EdgeInfo, 0, len(m.clients))
	for _, client := range m.clients {
		list = append(list, client.Info())
	}
	return list
}

func (m *TunnelManager) GetEdge(edgeID string) (*TunnelClient, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	client, ok := m.clients[edgeID]
	return client, ok
}

func (m *TunnelManager) Request(ctx context.Context, edgeID string, req *TunnelRequest, timeout time.Duration) (*TunnelResponse, error) {
	client, ok := m.GetEdge(edgeID)
	if !ok {
		return nil, fmt.Errorf("edge %s is not online", edgeID)
	}

	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	responseCh := make(chan *TunnelResponse, 1)
	m.mutex.Lock()
	m.pending[requestID] = responseCh
	m.mutex.Unlock()
	defer func() {
		m.mutex.Lock()
		delete(m.pending, requestID)
		m.mutex.Unlock()
	}()

	msg, err := NewMessage(MessageTypeRequest, requestID, edgeID, req)
	if err != nil {
		return nil, err
	}
	if err := client.Send(msg); err != nil {
		return nil, err
	}

	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, fmt.Errorf("request to edge %s timed out", edgeID)
	case resp := <-responseCh:
		if resp.Error != "" {
			return resp, fmt.Errorf(resp.Error)
		}
		return resp, nil
	}
}

func (m *TunnelManager) handleMessage(client *TunnelClient, msg *TunnelMessage) {
	client.UpdateLastSeen()
	switch msg.Type {
	case MessageTypeHeartbeat, MessageTypeRegister:
		return
	case MessageTypeResponse:
		var resp TunnelResponse
		if err := json.Unmarshal(msg.Payload, &resp); err != nil {
			return
		}
		m.mutex.RLock()
		ch := m.pending[msg.ID]
		m.mutex.RUnlock()
		if ch != nil {
			ch <- &resp
		}
	}
}
