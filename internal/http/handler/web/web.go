package web

import "net/http"

// 사용자용 웹 서비스 핸들러 구현부

type HandlerWeb struct {
}

func (h HandlerWeb) RegisterHandlers(mux *http.ServeMux) {
	fs := http.FileServer(http.Dir("./web/dist"))
	mux.Handle("/web/", http.StripPrefix("/web/", fs))
}
