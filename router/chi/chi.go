package chi

import (
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/http/rest/stdlib"
	"hex-microservice/http/url"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"net/http"
	"time"

	org "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-logr/logr"
)

func param(name string) string {
	return "{" + name + "}"
}

// New returns a http.Handler that exposes the service with the chi router.
func New(log logr.Logger, mappedURL string, mappedPath string, healthPath string, hs health.Service, servicePath string, as adder.Service, ls lookup.Service, is invalidator.Service) http.Handler {
	router := org.NewRouter()
	router.NotFound(http.NotFound)
	router.MethodNotAllowed(http.NotFound)

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.StripSlashes)

	serviceMappedUrl := url.Join(mappedURL, mappedPath, servicePath)
	handler := stdlib.New(log, hs, as, ls, is, org.URLParam)

	router.Get(url.AbsPath(mappedPath, healthPath),
		handler.Health(time.Now()))

	router.Get(url.AbsPath(mappedPath, servicePath, param(stdlib.UrlParameterCode)),
		handler.RedirectGet(serviceMappedUrl))

	router.Post(url.AbsPath(mappedPath, servicePath),
		handler.RedirectPost(serviceMappedUrl))

	router.Delete(url.AbsPath(mappedPath, servicePath, param(stdlib.UrlParameterCode), param(stdlib.UrlParameterToken)),
		handler.RedirectInvalidate(serviceMappedUrl))

	return router
}
