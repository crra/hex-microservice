package gorillamux

import (
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/http/rest/stdlib"
	"hex-microservice/http/url"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-logr/logr"
	org "github.com/gorilla/mux"
)

func paramFunc(r *http.Request, key string) string {
	return org.Vars(r)[key]
}

func param(name string) string {
	return "{" + name + "}"
}

func New(log logr.Logger, mappedURL string, mappedPath string, healthPath string, hs health.Service, servicePath string, as adder.Service, ls lookup.Service, is invalidator.Service) http.Handler {
	router := org.NewRouter()
	router.StrictSlash(true)
	router.NotFoundHandler = http.NotFoundHandler()
	router.MethodNotAllowedHandler = http.NotFoundHandler()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.StripSlashes)

	serviceMappedUrl := url.Join(mappedURL, mappedPath, servicePath)
	handler := stdlib.New(log, hs, as, ls, is, paramFunc)

	router.HandleFunc(url.AbsPath(mappedPath, healthPath),
		handler.Health(time.Now())).
		Methods(http.MethodGet)

	router.HandleFunc(url.AbsPath(mappedPath, servicePath, param(stdlib.UrlParameterCode)),
		handler.RedirectGet(serviceMappedUrl)).
		Methods(http.MethodGet)

	router.HandleFunc(url.AbsPath(mappedPath, servicePath),
		handler.RedirectPost(serviceMappedUrl)).
		Methods(http.MethodPost)

	router.HandleFunc(url.AbsPath(mappedPath, servicePath, param(stdlib.UrlParameterCode), param(stdlib.UrlParameterToken)),
		handler.RedirectInvalidate(serviceMappedUrl)).
		Methods(http.MethodDelete)

	return router
}
