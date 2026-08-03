package handler

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GetHandler 对应文件获取请求
// 方法 GET
// URL /get?file=/path/to/file
type GetFileHandler struct {
	// 基础根目录，所有文件必须在此目录下
	RootDir string
	// 处理器函数
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (g *GetFileHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return g.Handler
}

func NewGetFileHandler(rootDir string) *GetFileHandler {
	gh := &GetFileHandler{
		RootDir: filepath.Clean(rootDir),
	}
	gh.Handler = gh.NewHandlerFunc()
	return gh
}

var _ Handler = &GetFileHandler{}

func (g *GetFileHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET is supported", http.StatusMethodNotAllowed)
			return
		}

		// 获取文件路径参数
		filePath := r.URL.Query().Get("file")
		if filePath == "" {
			http.Error(w, "File parameter is required", http.StatusBadRequest)
			return
		}

		// 清理路径并确保它在根目录下
		cleanPath := filepath.Clean(filePath)
		if !strings.HasPrefix(cleanPath, "/") {
			cleanPath = "/" + cleanPath
		}

		absPath := filepath.Join(g.RootDir, cleanPath)

		// 安全检查：确保路径在根目录下
		relPath, err := filepath.Rel(g.RootDir, absPath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			http.Error(w, "Invalid file path", http.StatusBadRequest)
			return
		}

		// 打开文件
		file, err := os.Open(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "File not found", http.StatusNotFound)
			} else {
				http.Error(w, fmt.Sprintf("Error opening file: %v", err), http.StatusInternalServerError)
			}
			return
		}
		defer file.Close()

		// 获取文件信息
		fileInfo, err := file.Stat()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting file info: %v", err), http.StatusInternalServerError)
			return
		}

		// 如果是目录，返回错误（或者可以实现目录打包下载）
		if fileInfo.IsDir() {
			http.Error(w, "Directory download not supported", http.StatusBadRequest)
			return
		}

		// 设置响应头
		//filename := filepath.Base(cleanPath)
		w.Header().Set("Content-Type", "application/octet-stream")
		//w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
		w.WriteHeader(http.StatusOK)

		// 发送文件内容
		_, err = io.Copy(w, file)
		if err != nil {
			logs.Errorf("Failed to send file %s: %v", cleanPath, err)
			return
		}

		logs.Infof("File %s downloaded successfully", cleanPath)
	}
}
