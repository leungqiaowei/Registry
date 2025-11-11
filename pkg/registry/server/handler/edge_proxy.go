package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	edgeResponses    = make(map[string]chan map[string]interface{})
	edgeResponsesMux sync.RWMutex
)

// EdgeRequestHandler 对应边侧请求代理
// 方法 GET, POST, PUT, DELETE 等
// URL /edges/*
type EdgeRequestHandler struct {
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *EdgeRequestHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewEdgeRequestHandler() *EdgeRequestHandler {
	dh := &EdgeRequestHandler{}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &EdgeRequestHandler{}

func (d *EdgeRequestHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":     "client_id_required",
				"message":   "必须提供 client_id 参数",
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return
		}

		// 查找客户端
		client := findClientByID(clientID)
		if client == nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":     "client_not_found",
				"message":   fmt.Sprintf("客户端 %s 未连接", clientID),
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return
		}

		// 提取请求的路径（移除 /edges 前缀）
		requestPath := r.URL.Path
		if len(requestPath) > len("/edges") {
			requestPath = requestPath[len("/edges"):]
		}

		// 生成请求ID
		requestID := fmt.Sprintf("edge-req-%d", time.Now().UnixNano())

		// 创建响应通道
		responseChan := make(chan map[string]interface{}, 1)
		edgeResponsesMux.Lock()
		edgeResponses[requestID] = responseChan
		edgeResponsesMux.Unlock()

		// 清理函数
		defer func() {
			edgeResponsesMux.Lock()
			delete(edgeResponses, requestID)
			close(responseChan)
			edgeResponsesMux.Unlock()
		}()

		// 构建请求体（如果是POST/PUT等方法）
		var requestBody interface{}
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				requestBody = body
			}
		}

		// 构建边侧请求消息
		edgeRequestMsg := WSMessage{
			Type: "edge_request",
			Payload: map[string]interface{}{
				"request_id":   requestID,
				"method":       r.Method,
				"path":         requestPath,
				"query_params": r.URL.Query(),
				"headers":      r.Header,
				"body":         requestBody,
			},
		}

		log.Printf("发送边侧请求到 %s: %s %s", clientID, r.Method, requestPath)

		// 发送请求到边侧设备
		if err := sendToClient(client, edgeRequestMsg); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":     "send_request_failed",
				"message":   fmt.Sprintf("无法发送请求到边侧设备: %v", err),
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return
		}

		// 等待边侧响应（设置超时）
		select {
		case response := <-responseChan:
			// 收到边侧响应
			status, _ := response["status"].(string)
			data, _ := response["data"].(map[string]interface{})

			if status == "success" {
				statusCode, _ := data["status_code"].(float64)
				responseData, _ := data["data"]

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(int(statusCode))
				json.NewEncoder(w).Encode(responseData)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":     "edge_request_failed",
					"message":   data["error"],
					"timestamp": time.Now().Format(time.RFC3339),
				})
			}

		case <-time.After(30 * time.Second):
			// 超时
			w.WriteHeader(http.StatusGatewayTimeout)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":     "request_timeout",
				"message":   "边侧设备响应超时",
				"timestamp": time.Now().Format(time.RFC3339),
			})
		}
	}
}

// HandleEdgeResponse 处理来自边侧的响应（需要在main包中调用）
func HandleEdgeResponse(response map[string]interface{}) {
	payload, ok := response["payload"].(map[string]interface{})
	if !ok {
		log.Printf("边侧响应格式错误")
		return
	}

	requestID, ok := payload["request_id"].(string)
	if !ok {
		log.Printf("边侧响应缺少request_id")
		return
	}

	edgeResponsesMux.RLock()
	responseChan, exists := edgeResponses[requestID]
	edgeResponsesMux.RUnlock()

	if exists {
		responseChan <- payload
		log.Printf("边侧响应已传递: %s", requestID)
	} else {
		log.Printf("收到未知请求ID的边侧响应: %s", requestID)
	}
}
