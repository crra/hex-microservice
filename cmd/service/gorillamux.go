package main

import (
	"hex-microservice/shortener"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
)

func newGorillaMuxRouter(log logr.Logger, service shortener.Service) http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	/*
			r.HandleFunc("/products", ProductsHandler).

		  Methods("GET").
	*/

	return router
}
