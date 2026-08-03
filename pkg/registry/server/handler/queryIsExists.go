package handler

import (
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
)

// QueryIsExistsHandler 对应查询文件是否存在请求
// 方法 GET
// URL /query/exits
// Param filename
type QueryIsExistsHandler struct {
	DataPath     string
	DataSpecList *data.DataSpecList
	Handler      func(w http.ResponseWriter, r *http.Request)
}

func (d *QueryIsExistsHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewQueryIsExistsHandler(dataPath string, dataSpecList *data.DataSpecList) *QueryIsExistsHandler {
	dh := &QueryIsExistsHandler{DataPath: dataPath, DataSpecList: dataSpecList}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &QueryIsExistsHandler{}

func (d *QueryIsExistsHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "only GET is supported")
			return
		}

		fileName, tag, _, _, err := utils.GetFileParams(r, "v1.0.0")
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		file, err := d.DataSpecList.GetDataSpec(fileName, tag)
		if err != nil {
			utils.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		utils.WriteJSON(w, http.StatusOK, file)
	}
}
