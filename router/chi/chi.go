package chi

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

	org "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-logr/logr"
)

type chi struct {
	log       logr.Logger
	mappedURL string
	router    *org.Mux
}

// New returns a http.Handler that exposes the service with the chi router.
func New(log logr.Logger, mappedURL string) router.Router {
	r := org.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	r.NotFound(http.NotFound)
	r.MethodNotAllowed(http.NotFound)

	return &chi{
		log:       log,
		mappedURL: mappedURL,
		router:    r,
	}
}

func param(name string) string {
	return "{" + name + "}"
}

func (c *chi) MountV1(v1Path string, healthPath string, h health.Service, servicePath string, a adder.Service, l lookup.Service, i invalidator.Service) {
	service := rest.NewV1(c.log, h, a, l, i, org.URLParam)

	c.router.Get("/"+value.Join("/", v1Path, healthPath),
		service.Health(time.Now()))

	c.router.Get("/"+value.Join("/", v1Path, servicePath, param(rest.UrlParameterCode)),
		service.RedirectGet(c.mappedURL))

	c.router.Post("/"+value.Join("/", v1Path, servicePath),
		service.RedirectPost(c.mappedURL, servicePath))

	c.router.Delete("/"+value.Join("/", v1Path, servicePath, param(rest.UrlParameterCode), param(rest.UrlParameterToken)),
		service.RedirectInvalidate(c.mappedURL))
}

func (c *chi) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	c.router.ServeHTTP(rw, r)
}
