package handler

import (
	"bytes"
	"hit.edu/framework/pkg/registry/edge"
	"hit.edu/framework/pkg/registry/utils"
	"io"
	"net/http"
	"strings"
	"time"
)

// EdgeRequestHandler 将云端 HTTP 请求通过长连接代理到指定边侧节点。
// 方法 任意
// URL /edges/{edge_id}/{target_path}
type EdgeRequestHandler struct {
	Manager *edge.TunnelManager
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *EdgeRequestHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewEdgeRequestHandler(manager *edge.TunnelManager) *EdgeRequestHandler {
	dh := &EdgeRequestHandler{Manager: manager}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &EdgeRequestHandler{}

func (d *EdgeRequestHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		edgeID, targetPath, ok := parseEdgeProxyPath(r.URL.Path)
		if !ok {
			utils.WriteError(w, http.StatusBadRequest, "path must be /edges/{edge_id}/{target_path}")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, "failed to read request body")
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))

		resp, err := d.Manager.Request(r.Context(), edgeID, &edge.TunnelRequest{
			Method: r.Method,
			Path:   targetPath,
			Query:  r.URL.Query(),
			Header: r.Header.Clone(),
			Body:   body,
		}, 30*time.Second)
		if err != nil {
			utils.WriteError(w, http.StatusBadGateway, err.Error())
			return
		}

		utils.CopyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(resp.Body)
	}
}

func parseEdgeProxyPath(path string) (string, string, bool) {
	trimmed := strings.TrimPrefix(path, "/edges/")
	if trimmed == path || trimmed == "" {
		return "", "", false
	}
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], "/" + parts[1], true
}
