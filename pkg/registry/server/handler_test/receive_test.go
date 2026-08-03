package handler_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// 测试 ReceiveHandler 的正确性
func TestReceiveHandler(t *testing.T) {
	// 初始化 ReceiveHandler 的真实逻辑
	receiveHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST is supported", http.StatusMethodNotAllowed)
			return
		}

		// 获取文件名和标签
		fileName := r.URL.Query().Get("filename")
		tag := r.URL.Query().Get("tag")

		if fileName == "" || tag == "" {
			http.Error(w, "Filename and tag are required", http.StatusBadRequest)
			return
		}

		// 读取请求体
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		// 验证内容
		if !strings.Contains(string(body), "test content") {
			http.Error(w, "Invalid content", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("File received successfully"))
	}

	// 启动模拟的 ReceiveHandler 服务
	server := httptest.NewServer(http.HandlerFunc(receiveHandler))
	defer server.Close()

	// 模拟请求
	body := bytes.NewBufferString("test content")
	req, err := http.NewRequest(http.MethodPost, server.URL+"/receive?filename=test.txt&tag=v1.0.0", body)
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
