package main

import (
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	httpservice "hex-microservice/http"
	"hex-microservice/lookup"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-logr/logr"
)

// newChiRouter returns a http.Handler that exposes the service with the chi router.
func newChiRouter(log logr.Logger, mappedURL string, a adder.Service, l lookup.Service, d deleter.Service) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	s := httpservice.New(log, a, l, d, chi.URLParam)

	r.Get(fmt.Sprintf("/{%s}", httpservice.UrlParameterCode), s.RedirectGet(mappedURL))
	r.Post("/", s.RedirectPost(mappedURL))
	r.Delete(fmt.Sprintf("/{%s}/{%s}", httpservice.UrlParameterCode, httpservice.UrlParameterToken), s.RedirectDelete(mappedURL))

	return r
}
