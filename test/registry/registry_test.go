package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

const baseURL = "http://localhost:8081"

// 测试 GET 请求
func TestGetFile(t *testing.T) {
	// 发送 GET 请求来获取文件
	resp, err := http.Get(fmt.Sprintf("%s/download?filename=1_v2.txt", baseURL))
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	// 断言返回的状态码
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	// 读取响应内容并进行断言（假设你返回的是文件内容）
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// 根据返回的内容做一些断言
	if string(body) != "This is the test file content" {
		t.Fatalf("Expected file content, but got: %s", body)
	}
}

// 测试 POST 请求
func TestPostFile(t *testing.T) {
	// 创建一个包含文件数据的表单
	fileData := []byte("This is the content of the test file.")
	resp, err := http.Post(fmt.Sprintf("%s/upload?filename=fileData", baseURL), "application/octet-stream", bytes.NewReader(fileData))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	// 断言返回的状态码
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}

// 测试转发功能（假设你有一个转发文件的功能）
func TestForwardFile(t *testing.T) {
	// 创建一个包含文件数据的表单
	fileData := []byte("This is the content to forward.")
	resp, err := http.Post(fmt.Sprintf("%s/forward?target=http://localhost:8081&fileName=1txt", baseURL), "application/octet-stream", bytes.NewReader(fileData))
	if err != nil {
		t.Fatalf("Failed to make POST request for forwarding: %v", err)
	}
	defer resp.Body.Close()

	// 断言返回的状态码
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}

// 模拟一个延时响应的请求，以测试并发请求时的行为
func TestPostLargeFile(t *testing.T) {
	fileData := make([]byte, 1024*1024*1024*5) // 10 GB 的数据
	resp, err := http.Post(fmt.Sprintf("%s/upload?filename=fileData", baseURL), "application/octet-stream", bytes.NewReader(fileData))
	if err != nil {
		t.Fatalf("Failed to make POST request for large file: %v", err)
	}
	defer resp.Body.Close()

	// 断言返回的状态码
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}

// 测试并发请求的处理能力
func TestConcurrency(t *testing.T) {
	// 模拟并发请求
	for i := 0; i < 10; i++ {
		go func(i int) {
			fileData := []byte(fmt.Sprintf("This is request #%d", i))
			resp, err := http.Post(fmt.Sprintf("%s/upload", baseURL), "application/octet-stream", bytes.NewReader(fileData))
			if err != nil {
				t.Errorf("Failed to make POST request #%d: %v", i, err)
				return
			}
			defer resp.Body.Close()

			// 断言返回的状态码
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d, but got %d for request #%d", http.StatusOK, resp.StatusCode, i)
			}
		}(i)
	}

	// 等待一段时间以确保所有并发请求完成
	time.Sleep(3 * time.Second)
}
