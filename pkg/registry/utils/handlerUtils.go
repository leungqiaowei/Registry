package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

const defaultForwardTimeout = 30 * time.Second

// GetQueryParamCaseInsensitive 获取不区分大小写的查询参数
func GetQueryParamCaseInsensitive(params map[string][]string, paramName string) string {
	for key, values := range params {
		if strings.EqualFold(key, paramName) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

// GetFileParams 从请求中提取文件名和标签参数
func GetFileParams(r *http.Request, defaultTag string) (string, string, string, string, error) {
	param := r.URL.Query()

	fileName := GetQueryParamCaseInsensitive(param, "filename")
	if fileName == "" {
		return "", "", "", "", fmt.Errorf("filename is required")
	}
	fileName = filepath.Clean(fileName)
	if fileName == "." || fileName == ".." || strings.HasPrefix(fileName, "..") {
		return "", "", "", "", fmt.Errorf("invalid filename")
	}

	tag := GetQueryParamCaseInsensitive(param, "tag")
	if tag == "" {
		tag = defaultTag
	}

	owner := r.RemoteAddr

	fileType := strings.ToLower(r.Header.Get("FileType"))
	if fileType == "" {
		fileType = "file"
	}
	if fileType != "file" && fileType != "folder" && fileType != "completion" {
		return "", "", "", "", fmt.Errorf("invalid file type: %s", fileType)
	}

	return fileName, tag, owner, fileType, nil
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logs.Infof("failed to encode JSON response: %v", err)
	}
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, map[string]interface{}{
		"error":   http.StatusText(statusCode),
		"message": message,
	})
}

func CopyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// ForwardRequest 将 HTTP 请求转发给订阅者
func ForwardRequest(r *http.Request, newURL string, w http.ResponseWriter) error {
	logs.Infof("Forwarding to subscriber: %s", newURL)

	var bodyBytes []byte
	if r.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(r.Body)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "failed to read request body")
			return fmt.Errorf("failed to read request body: %w", err)
		}
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, newURL, bytes.NewReader(bodyBytes))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create forward request")
		logs.Infof("Failed to create HTTP request to %s: %v", newURL, err)
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header = r.Header.Clone()
	client := &http.Client{Timeout: defaultForwardTimeout}
	resp, err := client.Do(req)
	if err != nil {
		WriteError(w, http.StatusBadGateway, "failed to forward request to target")
		logs.Infof("Failed to forward data to %s: %v", newURL, err)
		return fmt.Errorf("failed to forward data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = fmt.Sprintf("target server responded with status %d", resp.StatusCode)
		}
		WriteError(w, http.StatusBadGateway, message)
		logs.Infof("Target server responded with status %d", resp.StatusCode)
		return fmt.Errorf("target server responded with status %d", resp.StatusCode)
	}

	CopyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		logs.Infof("Error writing response: %v", err)
		return fmt.Errorf("error writing response: %w", err)
	}

	logs.Infof("Data successfully forwarded to %s with status %d", newURL, resp.StatusCode)
	return nil
}

// TransformURL 将请求的 URL 转换为接收数据的 URL
func TransformURL(r *http.Request) (string, error) {
	clusterID := r.Header.Get("clusterID")
	if clusterID == "" {
		clusterID = r.Header.Get("ClusterID")
	}
	if clusterID == "" {
		return "", fmt.Errorf("missing clusterID header")
	}

	parsedURL, err := url.Parse(r.URL.String())
	if err != nil {
		logs.Infof("Failed to parse request URL: %v", err)
		return "", fmt.Errorf("failed to parse request URL")
	}

	_, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		port = "8081"
	}

	parsedURL.Path = strings.Replace(parsedURL.Path, "/forward", "/receive", 1)

	newURL := url.URL{
		Scheme:   "http",
		Host:     net.JoinHostPort(clusterID, port),
		Path:     parsedURL.Path,
		RawQuery: parsedURL.Query().Encode(),
	}

	return newURL.String(), nil
}

// TransformTargetURL 将请求中的 target 参数解析并转换为目标 URL
func TransformTargetURL(r *http.Request) (string, error) {
	parsedURL, err := url.Parse(r.URL.String())
	if err != nil {
		return "", fmt.Errorf("failed to parse request URL")
	}

	targetURLStr := parsedURL.Query().Get("target")
	if targetURLStr == "" {
		return "", fmt.Errorf("target URL is missing in the request")
	}

	targetURL, err := url.Parse(targetURLStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse target URL")
	}

	if targetURL.Scheme == "" || targetURL.Host == "" {
		return "", fmt.Errorf("target URL must include scheme and host")
	}

	return targetURL.String(), nil
}
