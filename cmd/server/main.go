package main

import (
	"PBGW/internal"
	"fmt"
	"log"
	"net/http"
)

func main() {
	var mux = http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World")
	})

	configurer := internal.GetConfigurer()
	var httpServer = &http.Server{
		Addr:           configurer.Address,
		Handler:        mux,
		ReadTimeout:    configurer.ReadTimeout,
		WriteTimeout:   configurer.WriteTimeout,
		MaxHeaderBytes: configurer.MaxHeaderBytes,
	}

	log.Fatal(httpServer.ListenAndServe())
}
