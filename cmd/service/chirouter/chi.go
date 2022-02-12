package chirouter

import (
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/health"
	"hex-microservice/http/rest"
	"hex-microservice/lookup"
	"net/http"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-logr/logr"
)

// New returns a http.Handler that exposes the service with the chi router.
func New(log logr.Logger, mappedURL string, h health.Service, a adder.Service, l lookup.Service, d deleter.Service) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	r.NotFound(http.NotFound)
	r.MethodNotAllowed(http.NotFound)

	s := rest.New(log, h, a, l, d, chi.URLParam)

	r.Get("/health", s.Health(time.Now()))

	r.Get(fmt.Sprintf("/{%s}", rest.UrlParameterCode), s.RedirectGet(mappedURL))
	r.Post("/", s.RedirectPost(mappedURL))
	r.Delete(fmt.Sprintf("/{%s}/{%s}", rest.UrlParameterCode, rest.UrlParameterToken), s.RedirectDelete(mappedURL))

	return r
}
