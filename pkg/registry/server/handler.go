package server

import (
	"net/http"
	"time"

	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/server/handler"
)

const (
	defaultKeepAlivePeriod = 3 * time.Minute
)

// RegistryHandler 处理不同的HTTP请求
type RegistryHandler struct {
	// 处理文件上传
	UploadHandler handler.Handler
	// 处理文件下载
	DownloadHandler handler.Handler
	// 处理文件转发
	ForwardHandler handler.Handler
	// 处理文件接收
	ReceiveHandler handler.Handler
	// 处理文件删除
	// TODO:
	DeleteHandler handler.Handler
	// 处理文件查询
	// TODO:
	QueryIsExistsHandler handler.Handler
	// 处理文件列表查询
	QueryListHandler handler.Handler
	// 处理文件订阅
	SubscribeHandler handler.Handler
	// 获取文件订阅列表
	SubscribeListHandler handler.Handler
	// 处理文件目录上传
	CatalogueUploadHandler handler.Handler
	// 处理文件目录下载
	CatalogueDownloadHandler handler.Handler

	// 处理长连接ip保持
	WebSocketHandler   handler.Handler
	ClientsHandler     handler.Handler
	HealthHandler      handler.Handler
	EdgeRequestProxy   handler.Handler
	EdgeConnectHandler handler.Handler
}

func NewRegistryHandler(dataPath string, dataSpecList *data.DataSpecList, subscribers *data.SubscriptionManager) *RegistryHandler {
	// TODO: Download等改成Handler, 实现ServeHTTP等函数
	rh := &RegistryHandler{
		UploadHandler:        handler.NewUploadHandler(dataPath, dataSpecList),
		DownloadHandler:      handler.NewDownloadHandler(dataPath, dataSpecList),
		ForwardHandler:       handler.NewForwardHandler(dataPath, dataSpecList),
		ReceiveHandler:       handler.NewReceiveHandler(dataPath, dataSpecList, subscribers),
		DeleteHandler:        handler.NewDeleteHandler(dataPath, dataSpecList),
		QueryIsExistsHandler: handler.NewQueryIsExistsHandler(dataPath, dataSpecList),
		QueryListHandler:     handler.NewQueryListHandler(dataPath, dataSpecList),
		SubscribeHandler:     handler.NewSubscribeHandler(subscribers),
		SubscribeListHandler: handler.NewScribeListHandler(subscribers),
		//CatalogueUploadHandler:   handler.NewCatalogueUploadHandler(dataPath, fileMapping),
		//CatalogueDownloadHandler: handler.NewCatalogueDownloadHandler(dataPath, fileMapping),
		WebSocketHandler:   handler.NewWebSocketHandler(),
		ClientsHandler:     handler.NewClientsListHandler(),
		HealthHandler:      handler.NewHealthCheckHandler(),
		EdgeRequestProxy:   handler.NewEdgeRequestHandler(),
		EdgeConnectHandler: handler.NewConnectHandler(dataPath),
	}

	// TODO: 临时用法,注册路由
	http.HandleFunc("/download", rh.DownloadHandler.GetHandler())
	http.HandleFunc("/upload", rh.UploadHandler.GetHandler())
	http.HandleFunc("/forward", rh.ForwardHandler.GetHandler())
	http.HandleFunc("/receive", rh.ReceiveHandler.GetHandler())
	http.HandleFunc("/delete", rh.DeleteHandler.GetHandler())
	http.HandleFunc("/query/exits", rh.QueryIsExistsHandler.GetHandler())
	http.HandleFunc("/query/list", rh.QueryListHandler.GetHandler())
	http.HandleFunc("/subscribe", rh.SubscribeHandler.GetHandler())
	http.HandleFunc("/subscribe/list", rh.SubscribeListHandler.GetHandler())
	http.HandleFunc("/websocket", rh.WebSocketHandler.GetHandler())
	http.HandleFunc("/edge/clients", rh.ClientsHandler.GetHandler())
	http.HandleFunc("/edge/health", rh.HealthHandler.GetHandler())
	http.HandleFunc("/edge/connect", rh.EdgeConnectHandler.GetHandler())
	http.HandleFunc("/edges/", rh.EdgeRequestProxy.GetHandler())
	return rh
}

// CORSMiddleware 包装 handler 添加 CORS 支持
func CORSMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置 CORS 头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Allow-Credentials", "false")

		// 处理 OPTIONS 预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		// 继续处理请求
		h.ServeHTTP(w, r)
	}
}
