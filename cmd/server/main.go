package main

import (
	"PBGW/internal"
	"PBGW/internal/http/handler/web"
	"PBGW/internal/middleware"
	"log"
	"net/http"
)

func main() {
	var mux = http.NewServeMux()

	// 핸들러 등록
	handlerWeb := web.HandlerWeb{}
	handlerWeb.RegisterHandlers(mux)

	// 미들웨어 적용
	finalMux := middleware.Maintenance(mux)

	configurer := internal.NewServerConfigure()
	var httpServer = &http.Server{
		Addr:           configurer.Address,
		Handler:        finalMux,
		ReadTimeout:    configurer.ReadTimeout,
		WriteTimeout:   configurer.WriteTimeout,
		MaxHeaderBytes: configurer.MaxHeaderBytes,
	}

	log.Fatal(httpServer.ListenAndServe())
}
