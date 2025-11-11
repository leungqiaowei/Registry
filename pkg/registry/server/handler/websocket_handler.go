package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 生产环境应该验证来源
		},
	}

	clients    = make(map[*websocket.Conn]*ClientInfo)
	clientsMux sync.RWMutex
)

type ClientInfo struct {
	IP        string          `json:"ip"`
	Conn      *websocket.Conn `json:"-"`
	Connected time.Time       `json:"connected"`
	ClientID  string          `json:"client_id"`
}

type WSMessage struct {
	Type      string      `json:"type"` // heartbeat, health_request, health_response
	Payload   interface{} `json:"payload"`
	Timestamp string      `json:"timestamp"`
}

// WebSocketHandler 对应 WebSocket 连接请求
// 方法 GET
// URL /ws
type WebSocketHandler struct {
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *WebSocketHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewWebSocketHandler() *WebSocketHandler {
	dh := &WebSocketHandler{}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &WebSocketHandler{}

func (d *WebSocketHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is supported", http.StatusMethodNotAllowed)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket升级失败: %v", err)
			return
		}
		defer conn.Close()

		clientIP := r.RemoteAddr
		clientID := r.Header.Get("X-Client-ID")
		if clientID == "" {
			clientID = fmt.Sprintf("client-%d", time.Now().UnixNano())
		}

		// 注册客户端
		client := &ClientInfo{
			IP:        clientIP,
			Conn:      conn,
			Connected: time.Now(),
			ClientID:  clientID,
		}

		registerClient(client)
		defer unregisterClient(client)

		log.Printf("WebSocket客户端连接: %s (%s)", clientID, clientIP)

		// 处理消息
		for {
			var msg WSMessage
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Printf("读取WebSocket消息失败 %s: %v", clientID, err)
				break
			}

			handleClientMessage(client, msg)
		}
	}
}

// ClientsListHandler 对应客户端列表查询请求
// 方法 GET
// URL /edge/clients
type ClientsListHandler struct {
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *ClientsListHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewClientsListHandler() *ClientsListHandler {
	dh := &ClientsListHandler{}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &ClientsListHandler{}

func (d *ClientsListHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is supported", http.StatusMethodNotAllowed)
			return
		}

		clientsMux.RLock()
		defer clientsMux.RUnlock()

		clientList := make([]*ClientInfo, 0, len(clients))
		for _, client := range clients {
			clientList = append(clientList, client)
		}

		// 将客户端列表序列化为 JSON
		response, err := json.Marshal(clientList)
		if err != nil {
			http.Error(w, "Failed to serialize client list", http.StatusInternalServerError)
			return
		}

		// 设置响应头并返回结果
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	}
}

// HealthCheckHandler 对应健康检查请求
// 方法 GET
// URL /edge/health
type HealthCheckHandler struct {
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *HealthCheckHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewHealthCheckHandler() *HealthCheckHandler {
	dh := &HealthCheckHandler{}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &HealthCheckHandler{}

func (d *HealthCheckHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is supported", http.StatusMethodNotAllowed)
			return
		}

		clientID := r.URL.Query().Get("client_id")

		var results []map[string]interface{}

		if clientID == "" {
			// 检查所有客户端
			results = requestAllClientsHealth()
		} else {
			// 检查特定客户端
			result := requestClientHealth(clientID)
			results = []map[string]interface{}{result}
		}

		response, err := json.Marshal(map[string]interface{}{
			"timestamp":       time.Now().Format(time.RFC3339),
			"checked_clients": len(results),
			"results":         results,
		})
		if err != nil {
			http.Error(w, "Failed to serialize health check results", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	}
}

// 以下辅助函数保持不变
// 在云侧的 handleClientMessage 函数中修改
func handleClientMessage(client *ClientInfo, msg WSMessage) {
	switch msg.Type {
	case "heartbeat":
		log.Printf("收到心跳来自 %s", client.ClientID)
		// 更新最后活跃时间
		client.Connected = time.Now()

	case "health_response":
		log.Printf("收到健康响应来自 %s", client.ClientID)
		// 可以存储或转发健康响应
		if payload, ok := msg.Payload.(map[string]interface{}); ok {
			log.Printf("   健康状态: %v", payload)
		}

	case "edge_response":
		log.Printf("收到边侧响应来自 %s", client.ClientID)
		// 处理边侧响应
		HandleEdgeResponse(map[string]interface{}{
			"payload": msg.Payload,
		})

	default:
		log.Printf("收到未知消息类型来自 %s: %s", client.ClientID, msg.Type)
	}
}

// 向特定客户端发送消息
func sendToClient(client *ClientInfo, msg WSMessage) error {
	msg.Timestamp = time.Now().Format(time.RFC3339)
	return client.Conn.WriteJSON(msg)
}

// 向所有客户端广播消息
func broadcastToClients(msg WSMessage) {
	clientsMux.RLock()
	defer clientsMux.RUnlock()

	for _, client := range clients {
		if err := sendToClient(client, msg); err != nil {
			log.Printf("向客户端 %s 发送消息失败: %v", client.ClientID, err)
		}
	}
}

func requestAllClientsHealth() []map[string]interface{} {
	clientsMux.RLock()
	defer clientsMux.RUnlock()

	var results []map[string]interface{}
	healthMsg := WSMessage{
		Type: "health_request",
		Payload: map[string]string{
			"request_id": fmt.Sprintf("req-%d", time.Now().UnixNano()),
		},
	}

	for _, client := range clients {
		result := map[string]interface{}{
			"client_id": client.ClientID,
			"ip":        client.IP,
			"connected": client.Connected.Format(time.RFC3339),
		}

		if err := sendToClient(client, healthMsg); err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		} else {
			result["status"] = "health_request_sent"
			result["sent_at"] = time.Now().Format(time.RFC3339)
		}

		results = append(results, result)
	}

	return results
}

func requestClientHealth(clientID string) map[string]interface{} {
	client := findClientByID(clientID)
	if client == nil {
		return map[string]interface{}{
			"client_id": clientID,
			"status":    "error",
			"error":     "client not found",
		}
	}

	healthMsg := WSMessage{
		Type: "health_request",
		Payload: map[string]string{
			"request_id": fmt.Sprintf("req-%d", time.Now().UnixNano()),
		},
	}

	if err := sendToClient(client, healthMsg); err != nil {
		return map[string]interface{}{
			"client_id": clientID,
			"status":    "error",
			"error":     err.Error(),
		}
	}

	return map[string]interface{}{
		"client_id": clientID,
		"status":    "health_request_sent",
		"sent_at":   time.Now().Format(time.RFC3339),
	}
}

// 客户端管理
func registerClient(client *ClientInfo) {
	clientsMux.Lock()
	defer clientsMux.Unlock()
	clients[client.Conn] = client
}

func unregisterClient(client *ClientInfo) {
	clientsMux.Lock()
	defer clientsMux.Unlock()
	delete(clients, client.Conn)
	log.Printf("WebSocket客户端断开: %s", client.ClientID)
}

func findClientByID(clientID string) *ClientInfo {
	clientsMux.RLock()
	defer clientsMux.RUnlock()

	for _, client := range clients {
		if client.ClientID == clientID {
			return client
		}
	}
	return nil
}

// 请求客户端健康检查
func handleHealthCheckRequest(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")

	var results []map[string]interface{}

	if clientID == "" {
		// 检查所有客户端
		results = requestAllClientsHealth()
	} else {
		// 检查特定客户端
		result := requestClientHealth(clientID)
		results = []map[string]interface{}{result}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"timestamp":       time.Now().Format(time.RFC3339),
		"checked_clients": len(results),
		"results":         results,
	})
}
