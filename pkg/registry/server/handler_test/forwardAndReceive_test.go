package handler

import (
	"bytes"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/server/handler"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 测试真实的 ForwardHandler 和 ReceiveHandler 的交互
func TestForwardAndReceiveHandlersReal(t *testing.T) {
	// 初始化数据
	//fileMapping := _else.NewFileMapping() // 假设 NewFileMapping 方法初始化成功
	dataSpecList := data.NewDataSpecList()
	subscribers := data.NewSubscriptionManager()

	// 实例化真实的 ReceiveHandler
	receiveHandler := handler.NewReceiveHandler("/tmp/data", dataSpecList, subscribers)

	// 实例化真实的 ForwardHandler
	forwardHandler := handler.NewForwardHandler("/tmp/data", dataSpecList)

	// 启动 ReceiveHandler 的 HTTP 服务器
	receiveServer := httptest.NewServer(http.HandlerFunc(receiveHandler.GetHandler()))
	defer receiveServer.Close()

	// 修改 ForwardHandler 的目标 URL
	originalHandler := forwardHandler.GetHandler()
	forwardHandler.Handler = func(w http.ResponseWriter, r *http.Request) {
		// 注入目标 URL 到请求中
		query := r.URL.Query()
		query.Set("target", receiveServer.URL)
		r.URL.RawQuery = query.Encode()

		// 调用原始的 ForwardHandler
		originalHandler(w, r)
	}

	// 启动 ForwardHandler 的 HTTP 服务器
	forwardServer := httptest.NewServer(http.HandlerFunc(forwardHandler.GetHandler()))
	defer forwardServer.Close()

	// 模拟向 ForwardHandler 发送请求
	requestBody := bytes.NewBufferString("test content")
	req, err := http.NewRequest(http.MethodPost, forwardServer.URL+"/forward?filename=test.txt&tag=v1.0.0", requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("FileType", "file")

	// 使用 HTTP 客户端发送请求
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
