package handler

import (
	"hit.edu/framework/pkg/registry/data"
	"hit.edu/framework/pkg/registry/utils"
	"net/http"
)

type SubscribeListHandler struct {
	Subscribers *data.SubscriptionManager
	Handler     func(w http.ResponseWriter, r *http.Request)
}

func (d *SubscribeListHandler) GetHandler() func(w http.ResponseWriter, r *http.Request) {
	return d.Handler
}

func NewScribeListHandler(subscribers *data.SubscriptionManager) *SubscribeListHandler {
	dh := &SubscribeListHandler{Subscribers: subscribers}
	dh.Handler = dh.NewHandlerFunc()
	return dh
}

var _ Handler = &SubscribeListHandler{}

func (d *SubscribeListHandler) NewHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteError(w, http.StatusMethodNotAllowed, "only GET is supported")
			return
		}
		utils.WriteJSON(w, http.StatusOK, d.Subscribers.GetSubscriptionList())
	}
}
