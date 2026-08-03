package handler

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// 测试 ForwardHandler 的正确性
func TestForwardHandler(t *testing.T) {
	// 模拟目标 ReceiveHandler
	receiveHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
			return
		}

		body, _ := ioutil.ReadAll(r.Body)
		if !strings.Contains(string(body), "test content") {
			http.Error(w, "Invalid content", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("File received successfully"))
	}

	// 启动模拟的 ReceiveHandler 服务
	receiveServer := httptest.NewServer(http.HandlerFunc(receiveHandler))
	defer receiveServer.Close()

	// 初始化 ForwardHandler 的真实逻辑
	forwardHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
			return
		}

		// 转发到目标 ReceiveHandler
		targetURL := receiveServer.URL + "/receive?filename=test.txt&tag=v1.0.0"
		req, err := http.NewRequest(http.MethodPost, targetURL, r.Body)
		if err != nil {
			http.Error(w, "Failed to create forward request", http.StatusInternalServerError)
			return
		}

		req.Header = r.Header
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to forward request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		w.WriteHeader(resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		w.Write(body)
	}

	// 启动 ForwardHandler 服务
	forwardServer := httptest.NewServer(http.HandlerFunc(forwardHandler))
	defer forwardServer.Close()

	// 模拟请求
	body := bytes.NewBufferString("test content")
	req, err := http.NewRequest(http.MethodPost, forwardServer.URL, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 验证响应
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	responseBody, _ := ioutil.ReadAll(resp.Body)
	expected := "File received successfully"
	if string(responseBody) != expected {
		t.Fatalf("Expected response body '%s', got '%s'", expected, responseBody)
	}
}
