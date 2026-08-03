package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/server/handler"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadFolder(t *testing.T) {
	dataSpecList := data.NewDataSpecList() // 假设 NewFileMapping 方法初始化成功

	rootPath, err := createTestDirWithFiles("./")
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// 初始化 UploadHandler
	uploadHandler := handler.NewUploadHandler("./tmp", dataSpecList)

	// 创建 httptest 服务器
	server := httptest.NewServer(http.HandlerFunc(uploadHandler.GetHandler()))
	defer server.Close()

	url := fmt.Sprintf("%s/upload?filename=%s&tag=%s", server.URL, "test", "v1.0.0")

	// 调用 Traverse 方法模拟上传
	err = utils.Traverse(rootPath, url)
	if err != nil {
		t.Fatalf("Traverse failed: %v", err)
	}

	// 检查日志或其他成功信号
	t.Logf("Folder upload completed successfully for root path: %s", rootPath)
}

func createTestDirWithFiles(baseDir string) (string, error) {
	// 创建临时文件夹
	tempDir, err := os.MkdirTemp(baseDir, "testdir")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	// 创建子目录
	dir1 := filepath.Join(tempDir, "test1")
	dir2 := filepath.Join(tempDir, "test2")
	if err := os.Mkdir(dir1, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %v", dir1, err)
	}
	if err := os.Mkdir(dir2, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %v", dir2, err)
	}

	// 创建 1GB 的数据
	data := make([]byte, 1<<30)

	// 创建文件路径
	filePaths := []string{
		filepath.Join(dir1, "file1.txt"),
		filepath.Join(dir2, "file2.txt"),
	}

	// 写入文件内容
	for _, filePath := range filePaths {
		err := os.WriteFile(filePath, data, 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write file %s: %v", filePath, err)
		}
	}

	// 返回创建的临时目录路径
	return tempDir, nil
}
