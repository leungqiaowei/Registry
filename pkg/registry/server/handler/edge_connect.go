package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	conn      *websocket.Conn
	clientID  string
	serverURL string
}

// ConnectHandler 对应 WebSocket 客户端连接请求
// 方法 GET
// URL /edge/connect
type ConnectHandler struct {
	ServerURL string
	Handler   func(w http.ResponseWriter, r *http.Request)
}

func (d *ConnectHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewConnectHandler(serverURL string) *ConnectHandler {
	serverURL = "ws://223.166.61.57:11006/websocket"
	dh := &ConnectHandler{
		ServerURL: serverURL,
	}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &ConnectHandler{}

func (d *ConnectHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is supported", http.StatusMethodNotAllowed)
			return
		}

		clientID := getClientID()

		log.Printf("开始WebSocket连接到: %s", d.ServerURL)
		log.Printf("客户端ID: %s", clientID)

		// 创建WebSocket客户端
		wsClient := &WSClient{
			clientID:  clientID,
			serverURL: d.ServerURL,
		}

		// 在goroutine中启动连接（避免阻塞HTTP请求）
		go wsClient.connectAndServe()

		// 返回连接启动响应
		response := map[string]interface{}{
			"status":     "connecting",
			"client_id":  clientID,
			"server_url": d.ServerURL,
			"timestamp":  time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(response)
	}
}

// HealthHandler 对应本地健康检查请求
// 方法 GET
// URL /health
type HealthHandler struct {
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *HealthHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewHealthHandler() *HealthHandler {
	dh := &HealthHandler{}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &HealthHandler{}

func (d *HealthHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is supported", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"status":    "OK",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "edge-client",
			"version":   "1.0",
			"metrics": map[string]interface{}{
				"uptime":     time.Since(startTime).Seconds(),
				"goroutines": runtime.NumGoroutine(),
			},
		}
		json.NewEncoder(w).Encode(response)
	}
}

// 以下为原有的 WSClient 方法，保持不变
func (w *WSClient) connectAndServe() {
	for {
		headers := http.Header{}
		headers.Set("X-Client-ID", w.clientID)

		conn, _, err := websocket.DefaultDialer.Dial(w.serverURL, headers)
		if err != nil {
			log.Printf("WebSocket连接失败: %v, 5秒后重试...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		w.conn = conn
		log.Printf("WebSocket连接成功")

		// 启动心跳
		go w.startHeartbeat()

		// 处理服务器消息
		w.handleMessages()

		// 连接断开后重试
		log.Printf("WebSocket连接断开，重新连接...")
		time.Sleep(3 * time.Second)
	}
}

func (w *WSClient) startHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if w.conn == nil {
			return
		}

		msg := map[string]interface{}{
			"type": "heartbeat",
			"payload": map[string]string{
				"client_id": w.clientID,
				"status":    "alive",
			},
			"timestamp": time.Now().Format(time.RFC3339),
		}

		if err := w.conn.WriteJSON(msg); err != nil {
			log.Printf("发送心跳失败: %v", err)
			return
		}
	}
}

// 在边侧设备的 handleMessages 函数中添加
func (w *WSClient) handleMessages() {
	for {
		var msg map[string]interface{}
		err := w.conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("读取WebSocket消息失败: %v", err)
			return
		}

		msgType, _ := msg["type"].(string)

		switch msgType {
		case "health_request":
			log.Printf("收到健康检查请求")
			w.sendHealthResponse()

		case "edge_request":
			log.Printf("收到边侧请求")
			w.handleEdgeRequest(msg)

		default:
			log.Printf("收到服务器消息: %s", msgType)
		}
	}
}

// handleEdgeRequest 处理来自云侧的边侧请求
// 在边侧设备中
func (w *WSClient) handleEdgeRequest(msg map[string]interface{}) {
	payload, ok := msg["payload"].(map[string]interface{})
	if !ok {
		log.Printf("边侧请求格式错误")
		return
	}

	requestID, _ := payload["request_id"].(string)
	method, _ := payload["method"].(string)
	path, _ := payload["path"].(string)
	queryParams, _ := payload["query_params"].(map[string]interface{})

	log.Printf("处理边侧请求: %s %s", method, path)

	// 构建请求URL
	url := fmt.Sprintf("http://localhost:8919%s", path)
	println(url)
	// 添加查询参数
	if len(queryParams) > 0 {
		url += "?"
		for key, value := range queryParams {
			if values, ok := value.([]interface{}); ok && len(values) > 0 {
				if strValue, ok := values[0].(string); ok {
					url += fmt.Sprintf("%s=%s&", key, strValue)
				}
			}
		}
		url = strings.TrimSuffix(url, "&")
	}

	// 创建HTTP请求
	var req *http.Request
	var err error

	if method == http.MethodPost || method == http.MethodPut {
		// 处理有请求体的方法
		if body, exists := payload["body"]; exists && body != nil {
			jsonBody, _ := json.Marshal(body)
			req, err = http.NewRequest(method, url, bytes.NewReader(jsonBody))
			if err == nil {
				req.Header.Set("Content-Type", "application/json")
			}
		} else {
			req, err = http.NewRequest(method, url, nil)
		}
	} else {
		// GET、DELETE等方法
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		log.Printf("创建请求失败: %v", err)
		w.sendEdgeResponse(requestID, "error", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// 设置请求头
	if headers, ok := payload["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			if strValues, ok := value.([]interface{}); ok && len(strValues) > 0 {
				if strValue, ok := strValues[0].(string); ok {
					req.Header.Set(key, strValue)
				}
			}
		}
	}

	// 发送请求
	client := &http.Client{Timeout: 25 * time.Second} // 比云侧超时短
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("请求边侧服务失败: %v", err)
		w.sendEdgeResponse(requestID, "error", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	var responseData interface{}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		responseData = map[string]interface{}{
			"status_code": resp.StatusCode,
			"status":      resp.Status,
			"error":       "读取响应体失败",
		}
	} else {
		// 尝试解析为JSON，如果不是JSON则保持原始数据
		if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
			responseData = string(bodyBytes)
		}
	}

	// 发送响应回云侧
	w.sendEdgeResponse(requestID, "success", map[string]interface{}{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"data":        responseData,
	})
}

// sendEdgeResponse 发送边侧请求响应回云侧
func (w *WSClient) sendEdgeResponse(requestID string, status string, data map[string]interface{}) {
	response := map[string]interface{}{
		"type": "edge_response",
		"payload": map[string]interface{}{
			"request_id": requestID,
			"status":     status,
			"data":       data,
			"timestamp":  time.Now().Format(time.RFC3339),
		},
	}

	if err := w.conn.WriteJSON(response); err != nil {
		log.Printf("发送边侧响应失败: %v", err)
	} else {
		log.Printf("边侧响应已发送: %s", requestID)
	}
}

func (w *WSClient) sendHealthResponse() {
	// 检查本地健康状态
	healthStatus := "healthy"
	var healthData map[string]interface{}

	resp, err := http.Get("http://localhost:8919/health")
	if err != nil {
		healthStatus = "unhealthy"
		healthData = map[string]interface{}{
			"error": err.Error(),
		}
	} else {
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&healthData)
	}

	response := map[string]interface{}{
		"type": "health_response",
		"payload": map[string]interface{}{
			"client_id": w.clientID,
			"status":    healthStatus,
			"data":      healthData,
			"timestamp": time.Now().Format(time.RFC3339),
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err := w.conn.WriteJSON(response); err != nil {
		log.Printf("发送健康响应失败: %v", err)
	} else {
		log.Printf("健康响应已发送")
	}
}

func getClientID() string {
	// 尝试获取稳定的客户端ID
	if hostname, err := os.Hostname(); err == nil {
		return fmt.Sprintf("client-%s", hostname)
	}
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}

func startHealthServer() {
	healthHandler := NewHealthHandler()

	http.HandleFunc("/health", healthHandler.GetHandler())

	go func() {
		log.Printf("本地健康检查服务启动在 :8919")
		if err := http.ListenAndServe(":8919", nil); err != nil {
			log.Fatalf("健康检查服务启动失败: %v", err)
		}
	}()
}

var startTime = time.Now()
