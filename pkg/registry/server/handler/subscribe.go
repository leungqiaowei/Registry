package handler

import (
	"fmt"
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
	"path/filepath"
)

// SubscribeHandler 文件数据订阅
// 方法 POST
// URL /subscribe
// Param filename tag client_ip
type SubscribeHandler struct {
	Subscribers *data.SubscriptionManager
	Handler     func(w http.ResponseWriter, r *http.Request)
}

func (d *SubscribeHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewSubscribeHandler(subscribers *data.SubscriptionManager) *SubscribeHandler {
	dh := &SubscribeHandler{Subscribers: subscribers}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &SubscribeHandler{}

func (d *SubscribeHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.WriteError(w, http.StatusMethodNotAllowed, "only POST is supported")
			return
		}

		param := r.URL.Query()
		fileName := utils.GetQueryParamCaseInsensitive(param, "filename")
		if fileName == "" {
			utils.WriteError(w, http.StatusBadRequest, "filename is required")
			return
		}
		fileName = filepath.Clean(fileName)

		tag := utils.GetQueryParamCaseInsensitive(param, "tag")
		if tag == "" {
			tag = "v1.0.0"
		}

		clientIP := utils.GetQueryParamCaseInsensitive(param, "client_ip")
		if clientIP == "" {
			utils.WriteError(w, http.StatusBadRequest, "client_ip is required")
			return
		}

		clientAddr := fmt.Sprintf("http://%s:8080/receive?filename=%s&tag=%s", clientIP, fileName, tag)
		d.Subscribers.Subscribe(fileName, tag, clientAddr)
		utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "subscription successful"})
	}
}
