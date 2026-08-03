package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
)

// DeleteHandler 对应文件删除请求
// 方法 DELETE
// URL /delete
// Param filename
type DeleteHandler struct {
	DataPath     string
	DataSpecList *data.DataSpecList
	Handler      func(w http.ResponseWriter, r *http.Request)
}

func (d *DeleteHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewDeleteHandler(dataPath string, dataSpecList *data.DataSpecList) *DeleteHandler {
	dh := &DeleteHandler{DataPath: dataPath, DataSpecList: dataSpecList}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &DeleteHandler{}

func (d *DeleteHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			utils.WriteError(w, http.StatusMethodNotAllowed, "only DELETE is supported")
			return
		}

		fileName, tag, _, _, err := utils.GetFileParams(r, "v1.0.0")
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("error getting file parameters: %v", err))
			return
		}

		if err := d.DataSpecList.DeleteFile(fileName, tag); err != nil {
			utils.WriteError(w, http.StatusNotFound, fmt.Sprintf("failed to delete file: %v", err))
			return
		}

		utils.WriteJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("file %s (tag: %s) deleted successfully", fileName, tag)})
	}
}
