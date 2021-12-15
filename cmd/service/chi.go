package main

import (
	"fmt"
	http2 "hex-microservice/http"
	"hex-microservice/shortener"
	"net/http"

	chi "github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-logr/logr"
)

// newChiRouter returns a http.Handler that adapts the service with the use of the chi router.
func newChiRouter(log logr.Logger, service shortener.Service) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	s := http2.New(log, service, chi.URLParam)

	router.Get(fmt.Sprintf("/{%s}", urlParameterCode), s.Get)
	router.Post("/", s.Post)

	return router
}
