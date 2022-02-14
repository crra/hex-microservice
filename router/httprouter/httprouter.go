package httprouter

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

	"github.com/go-logr/logr"
	org "github.com/julienschmidt/httprouter"
)

type httpRouter struct {
	log       logr.Logger
	mappedURL string
	router    *org.Router
}

func paramFunc(r *http.Request, key string) string {
	return org.ParamsFromContext(r.Context()).ByName(key)
}

// newHttpRouter returns a http.Handler that adapts the service with the use of the httprouter router.
func New(log logr.Logger, mappedURL string) router.Router {
	r := org.New()
	r.HandleMethodNotAllowed = false
	return &httpRouter{
		log:       log,
		mappedURL: mappedURL,
		router:    r,
	}
}

func param(name string) string {
	return ":" + name
}

func (hr *httpRouter) MountV1(v1Path string, healthPath string, h health.Service, servicePath string, a adder.Service, l lookup.Service, i invalidator.Service) {
	service := rest.NewV1(hr.log, h, a, l, i, paramFunc)

	hr.router.Handler(http.MethodGet, "/"+value.Join("/", v1Path, healthPath),
		service.Health(time.Now()))

	hr.router.Handler(http.MethodGet, "/"+value.Join("/", v1Path, servicePath, param(rest.UrlParameterCode)),
		service.RedirectGet(hr.mappedURL))

	hr.router.Handler(http.MethodPost, "/"+value.Join("/", v1Path, servicePath),
		service.RedirectPost(hr.mappedURL, servicePath))

	hr.router.Handler(http.MethodDelete, "/"+value.Join("/", v1Path, servicePath, param(rest.UrlParameterCode), param(rest.UrlParameterToken)),
		service.RedirectInvalidate(hr.mappedURL))
}

func (hr *httpRouter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	hr.router.ServeHTTP(rw, r)
}
