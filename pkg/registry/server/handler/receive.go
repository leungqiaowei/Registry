package handler

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
)

// ReceiveHandler 文件/文件夹接收
// 方法 Post
// URL /receive

type ReceiveHandler struct {
	DataPath     string
	DataSpecList *data.DataSpecList
	Subscribers  *data.SubscriptionManager
	Handler      func(w http.ResponseWriter, r *http.Request)
}

func (d *ReceiveHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

// NewReceiveHandler 创建一个 ReceiveHandler 实例
func NewReceiveHandler(dataPath string, dataSpecList *data.DataSpecList, subscribers *data.SubscriptionManager) *ReceiveHandler {
	dh := &ReceiveHandler{DataPath: dataPath, DataSpecList: dataSpecList, Subscribers: subscribers}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &ReceiveHandler{}

func (d *ReceiveHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.WriteError(w, http.StatusMethodNotAllowed, "only POST is supported")
			return
		}

		flowType := r.Header.Get("FlowType")
		if flowType == "etcd" {
			targetURL, err := utils.TransformTargetURL(r)
			if err != nil {
				utils.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
			logs.Infof("Forwarding to URL: %s", targetURL)
			if err := utils.ForwardRequest(r, targetURL, w); err != nil {
				logs.Infof("Failed to forward request: %v", err)
			}
			return
		}

		fileName, tag, owner, fileType, err := utils.GetFileParams(r, "v1.0.0")
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("error getting file parameters: %v", err))
			return
		}

		if d.Subscribers.IsSubscribed(fileName, tag) {
			subscribers := d.Subscribers.GetSubscribers(fileName, tag)
			for _, subscriber := range subscribers {
				if err := utils.ForwardRequest(r, subscriber, w); err != nil {
					logs.Infof("Failed to forward request to subscriber %s: %v", subscriber, err)
					return
				}
			}
			return
		}

		switch fileType {
		case "folder", "completion":
			utils.ReceiveDir(w, r, d.DataPath)
			if _, err := d.DataSpecList.SaveFolder(d.DataPath, fileName, tag, owner, false); err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save folder: %v", err))
				return
			}
		case "file":
			if _, err := d.DataSpecList.SaveFile(d.DataPath, fileName, tag, owner, false, r.Body); err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("failed to save file: %v", err))
				return
			}
		default:
			utils.WriteError(w, http.StatusBadRequest, "invalid file type")
			return
		}

		utils.WriteJSON(w, http.StatusCreated, map[string]string{"message": "file received successfully"})
	}
}
