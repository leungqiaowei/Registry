package handler

import (
	"hit.edu/framework/pkg/registry/utils"
	"testing"
)

func TestDirectTransfer(t *testing.T) {
	err := utils.Traverse("D:\\Programming\\GolandProjects\\Registry\\tmp\\data\\downloads", "http://localhost:8080/receive?filename=downloads")
	if err != nil {
		t.Fatalf("传输失败: %v", err)
	}
	t.Log("传输成功")
}
