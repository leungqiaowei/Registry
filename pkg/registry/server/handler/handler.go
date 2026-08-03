package handler

import "net/http"

// TODO: 泛用类
type Handler interface {
	//
	NewHandlerFunc() func(w http.ResponseWriter, r *http.Request)

	//
	GetHandler() func(w http.ResponseWriter, r *http.Request)
}
