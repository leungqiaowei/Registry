package edge

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type TunnelClient struct {
	EdgeID      string
	RemoteAddr  string
	ConnectedAt time.Time
	lastSeen    time.Time
	conn        *websocket.Conn
	manager     *TunnelManager
	send        chan *TunnelMessage
	closeOnce   sync.Once
	mutex       sync.RWMutex
}

func newTunnelClient(edgeID, remoteAddr string, conn *websocket.Conn, manager *TunnelManager) *TunnelClient {
	now := time.Now()
	return &TunnelClient{
		EdgeID:      edgeID,
		RemoteAddr:  remoteAddr,
		ConnectedAt: now,
		lastSeen:    now,
		conn:        conn,
		manager:     manager,
		send:        make(chan *TunnelMessage, 32),
	}
}

func (c *TunnelClient) run() {
	go c.writeLoop()
	c.readLoop()
}

func (c *TunnelClient) Send(msg *TunnelMessage) error {
	select {
	case c.send <- msg:
		return nil
	default:
		return websocket.ErrCloseSent
	}
}

func (c *TunnelClient) Info() EdgeInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return EdgeInfo{
		EdgeID:      c.EdgeID,
		RemoteAddr:  c.RemoteAddr,
		ConnectedAt: c.ConnectedAt,
		LastSeen:    c.lastSeen,
		Status:      "online",
	}
}

func (c *TunnelClient) UpdateLastSeen() {
	c.mutex.Lock()
	c.lastSeen = time.Now()
	c.mutex.Unlock()
}

func (c *TunnelClient) Close() {
	c.closeOnce.Do(func() {
		close(c.send)
		_ = c.conn.Close()
	})
}

func (c *TunnelClient) readLoop() {
	defer func() {
		c.manager.unregister(c)
		c.Close()
	}()
	for {
		var msg TunnelMessage
		if err := c.conn.ReadJSON(&msg); err != nil {
			return
		}
		c.manager.handleMessage(c, &msg)
	}
}

func (c *TunnelClient) writeLoop() {
	for msg := range c.send {
		if err := c.conn.WriteJSON(msg); err != nil {
			c.Close()
			return
		}
	}
}
