package main

import (
	"PBGW/internal"
	"PBGW/internal/http/handler/web"
	"log"
	"net/http"
)

func main() {
	var mux = http.NewServeMux()

	// 핸들러 등록
	handlerWeb := web.HandlerWeb{}
	handlerWeb.RegisterHandlers(mux)

	configurer := internal.GetConfigure()
	var httpServer = &http.Server{
		Addr:           configurer.Address,
		Handler:        mux,
		ReadTimeout:    configurer.ReadTimeout,
		WriteTimeout:   configurer.WriteTimeout,
		MaxHeaderBytes: configurer.MaxHeaderBytes,
	}

	log.Fatal(httpServer.ListenAndServe())
}
