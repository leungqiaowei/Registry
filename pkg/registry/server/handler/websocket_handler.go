package handler

import (
	"hit.edu/framework/pkg/registry/edge"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
)

// WebSocketHandler 接收边侧主动发起的云边隧道长连接。
// 方法 GET
// URL /edge/ws?edge_id=<edge_id>
type WebSocketHandler struct {
	Manager *edge.TunnelManager
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *WebSocketHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewWebSocketHandler(manager *edge.TunnelManager) *WebSocketHandler {
	dh := &WebSocketHandler{Manager: manager}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &WebSocketHandler{}

func (d *WebSocketHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "only GET is supported")
			return
		}
		if err := d.Manager.AcceptEdge(w, r); err != nil {
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
}

// ClientsListHandler 查询当前在线边侧节点。
// 方法 GET
// URL /edge/clients
type ClientsListHandler struct {
	Manager *edge.TunnelManager
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *ClientsListHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewClientsListHandler(manager *edge.TunnelManager) *ClientsListHandler {
	dh := &ClientsListHandler{Manager: manager}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &ClientsListHandler{}

func (d *ClientsListHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "only GET is supported")
			return
		}
		utils.WriteJSON(w, http.StatusOK, d.Manager.ListEdges())
	}
}

// HealthCheckHandler 通过云边隧道检查边侧节点健康。
// 方法 GET
// URL /edge/health?edge_id=<edge_id>
type HealthCheckHandler struct {
	Manager *edge.TunnelManager
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (d *HealthCheckHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewHealthCheckHandler(manager *edge.TunnelManager) *HealthCheckHandler {
	dh := &HealthCheckHandler{Manager: manager}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &HealthCheckHandler{}

func (d *HealthCheckHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "only GET is supported")
			return
		}

		edgeID := r.URL.Query().Get("edge_id")
		if edgeID == "" {
			utils.WriteError(w, http.StatusBadRequest, "edge_id is required")
			return
		}

		resp, err := d.Manager.Request(r.Context(), edgeID, &edge.TunnelRequest{
			Method: http.MethodGet,
			Path:   "/health",
		}, 10_000_000_000)
		if err != nil {
			utils.WriteError(w, http.StatusBadGateway, err.Error())
			return
		}
		utils.CopyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(resp.Body)
	}
}
