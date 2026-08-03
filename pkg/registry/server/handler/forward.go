package handler

import (
	"net/http"

	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
)

// ForwardHandler 处理 POST 请求并实时转发数据
// TODO:

type ForwardHandler struct {
	DataPath     string
	DataSpecList *data.DataSpecList
	Handler      func(w http.ResponseWriter, r *http.Request)
}

func (d *ForwardHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

// NewForwardHandler 创建一个新的 ForwardHandler 实例
func NewForwardHandler(dataPath string, dataSpecList *data.DataSpecList) *ForwardHandler {
	dh := &ForwardHandler{DataPath: dataPath, DataSpecList: dataSpecList}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &ForwardHandler{}

func (d *ForwardHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		clusterID := r.Header.Get("ClusterID")
		if clusterID == "" {
			clusterID = r.Header.Get("clusterID")
		}
		if clusterID == "" {
			utils.WriteError(w, http.StatusBadRequest, "missing ClusterID in request header")
			logs.Infof("Missing ClusterID in request header")
			return
		}

		newURL, err := utils.TransformURL(r)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		logs.Infof("Forwarding to URL: %s", newURL)
		if err := utils.ForwardRequest(r, newURL, w); err != nil {
			logs.Infof("Failed to forward request: %v", err)
			return
		}
	}
}
