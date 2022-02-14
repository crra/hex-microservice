package gorillamux

import (
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/http/rest"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"hex-microservice/meta/value"
	"hex-microservice/router"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-logr/logr"
	org "github.com/gorilla/mux"
)

type mux struct {
	log       logr.Logger
	mappedURL string
	router    *org.Router
}

func New(log logr.Logger, mappedURL string) router.Router {
	r := org.NewRouter()
	r.StrictSlash(true)
	r.NotFoundHandler = http.NotFoundHandler()
	r.MethodNotAllowedHandler = http.NotFoundHandler()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	return &mux{
		log:       log,
		mappedURL: mappedURL,
		router:    r,
	}
}

func param(name string) string {
	return "{" + name + "}"
}

func (m *mux) MountV1(v1Path string, healthPath string, h health.Service, servicePath string, a adder.Service, l lookup.Service, i invalidator.Service) {
	service := rest.NewV1(m.log, h, a, l, i, func(r *http.Request, key string) string {
		return org.Vars(r)[key]
	})

	m.router.HandleFunc("/"+value.Join("/", v1Path, healthPath),
		service.Health(time.Now())).Methods("GET")

	m.router.HandleFunc("/"+value.Join("/", v1Path, servicePath, param(rest.UrlParameterCode)),
		service.RedirectGet(m.mappedURL)).Methods(http.MethodGet)

	m.router.HandleFunc("/"+value.Join("/", v1Path, servicePath),
		service.RedirectPost(m.mappedURL, servicePath)).Methods(http.MethodPost)

	m.router.HandleFunc("/"+value.Join("/", v1Path, servicePath, param(rest.UrlParameterCode), param(rest.UrlParameterToken)),
		service.RedirectInvalidate(m.mappedURL))
}

func (m *mux) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	m.router.ServeHTTP(rw, r)
}
