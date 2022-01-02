package main

import (
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/lookup"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
)

func newGorillaMuxRouter(log logr.Logger, mappedURL string, a adder.Service, l lookup.Service, d deleter.Service) http.Handler {
	router := mux.NewRouter().StrictSlash(true)

	/*
			r.HandleFunc("/products", ProductsHandler).

		  Methods("GET").
	*/

	return router
}
