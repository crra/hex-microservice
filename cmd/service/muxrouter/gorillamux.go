package muxrouter

import (
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/health"
	"hex-microservice/http/rest"
	"hex-microservice/lookup"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
)

func New(log logr.Logger, mappedURL string, h health.Service, a adder.Service, l lookup.Service, d deleter.Service) http.Handler {
	r := mux.NewRouter().StrictSlash(true)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	s := rest.New(log, h, a, l, d, func(r *http.Request, key string) string {
		return mux.Vars(r)[key]
	})

	r.HandleFunc("/health", s.Health()).Methods("GET")

	r.HandleFunc("/", s.RedirectGet(mappedURL)).Methods("GET")
	/*
			r.HandleFunc("/products", ProductsHandler).

		  Methods("GET").
	*/

	return r
}
