package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
)

type UploadHandler struct {
	DataPath     string
	DataSpecList *data.DataSpecList
	Handler      func(w http.ResponseWriter, r *http.Request)
}

func (d *UploadHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewUploadHandler(dataPath string, dataSpecList *data.DataSpecList) *UploadHandler {
	dh := &UploadHandler{DataPath: dataPath, DataSpecList: dataSpecList}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &UploadHandler{}

func (d *UploadHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.WriteError(w, http.StatusMethodNotAllowed, "only POST is supported")
			return
		}

		fileName, tag, owner, fileType, err := utils.GetFileParams(r, "v1.0.0")
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("error getting file parameters: %v", err))
			return
		}

		switch fileType {
		case "folder", "completion":
			utils.ReceiveDir(w, r, d.DataPath)
			if _, err := d.DataSpecList.SaveFolder(d.DataPath, fileName, tag, owner, true); err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save folder: %v", err))
				return
			}
		case "file":
			if _, err := d.DataSpecList.SaveFile(d.DataPath, fileName, tag, owner, true, r.Body); err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save file: %v", err))
				return
			}
		default:
			utils.WriteError(w, http.StatusBadRequest, "invalid file type")
			return
		}

		utils.WriteJSON(w, http.StatusCreated, map[string]string{"message": "file uploaded successfully"})
	}
}
