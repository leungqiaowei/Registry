package handler

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"io"
	"net/http"
)

// DownloadHandler 对应文件下载请求
// 方法 GET
// URL /download
// Param filename
// TODO:
type DownloadHandler struct {
	DataPath     string
	DataSpecList *data.DataSpecList
	Handler      func(w http.ResponseWriter, r *http.Request)
}

func (d *DownloadHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewDownloadHandler(dataPath string, dataSpecList *data.DataSpecList) *DownloadHandler {
	dh := &DownloadHandler{DataPath: dataPath, DataSpecList: dataSpecList}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &DownloadHandler{}

func (d *DownloadHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "only GET is supported")
			return
		}

		fileName, tag, _, fileType, err := utils.GetFileParams(r, "v1.0.0")
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("error getting file parameters: %v", err))
			return
		}

		spec, err := d.DataSpecList.GetDataSpec(fileName, tag)
		if err != nil {
			utils.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		switch fileType {
		case "folder", "completion":
			folderPath := d.DataSpecList.GetFilePath(fileName, tag)
			if folderPath == "" {
				utils.WriteError(w, http.StatusNotFound, "folder not found")
				return
			}
			if err := utils.Traverse(folderPath, r.URL.Query().Get("target")); err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("error traversing directory: %v", err))
				return
			}
			utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "folder dispatched successfully"})
		case "file":
			file, err := d.DataSpecList.LoadFile(fileName, tag)
			if err != nil {
				utils.WriteError(w, http.StatusNotFound, fmt.Sprintf("error loading file: %v", err))
				return
			}
			defer file.Close()
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			if _, err = io.Copy(w, file); err != nil {
				logs.Infof("Failed to send file %s (tag: %s): %v", fileName, tag, err)
				return
			}
			logs.Infof("File %s (tag: %s) downloaded successfully", fileName, tag)
		default:
			utils.WriteError(w, http.StatusBadRequest, "invalid file type")
			return
		}

		if !spec.IsPermanent {
			if err := d.DataSpecList.DeleteFile(fileName, tag); err != nil {
				logs.Infof("Failed to delete non-permanent file %s (tag: %s): %v", fileName, tag, err)
			}
		}
	}
}
