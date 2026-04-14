package handler

import "net/http"

type httpHandler interface {
	RegisterHandlers(mux *http.ServeMux)
}
