package handler

import (
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
)

// QueryListHandler 对应查询文件列表请求
// 方法 GET
// URL /query/list
type QueryListHandler struct {
	DataPath     string
	DataSpecList *data.DataSpecList
	Handler      func(w http.ResponseWriter, r *http.Request)
}

func (d *QueryListHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewQueryListHandler(dataPath string, dataSpecList *data.DataSpecList) *QueryListHandler {
	dh := &QueryListHandler{DataPath: dataPath, DataSpecList: dataSpecList}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &QueryListHandler{}

func (d *QueryListHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "only GET is supported")
			return
		}
		utils.WriteJSON(w, http.StatusOK, d.DataSpecList.GetList())
	}
}
