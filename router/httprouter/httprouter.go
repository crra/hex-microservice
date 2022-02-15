package httprouter

import (
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/http/rest/stdlib"
	"hex-microservice/http/url"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	org "github.com/julienschmidt/httprouter"
)

func paramFunc(r *http.Request, key string) string {
	return org.ParamsFromContext(r.Context()).ByName(key)
}

func param(name string) string {
	return ":" + name
}

// newHttpRouter returns a http.Handler that adapts the service with the use of the httprouter router.
func New(log logr.Logger, mappedURL string, mappedPath string, healthPath string, hs health.Service, servicePath string, as adder.Service, ls lookup.Service, is invalidator.Service) http.Handler {
	router := org.New()
	router.HandleMethodNotAllowed = false

	serviceMappedUrl := url.Join(mappedURL, mappedPath, servicePath)
	handler := stdlib.New(log, hs, as, ls, is, paramFunc)

	router.Handler(http.MethodGet, url.AbsPath(mappedPath, healthPath),
		handler.Health(time.Now()))

	router.Handler(http.MethodGet, url.AbsPath(mappedPath, servicePath, param(stdlib.UrlParameterCode)),
		handler.RedirectGet(serviceMappedUrl))

	router.Handler(http.MethodPost, url.AbsPath(mappedPath, servicePath),
		handler.RedirectPost(serviceMappedUrl))

	router.Handler(http.MethodDelete, url.AbsPath(mappedPath, servicePath, param(stdlib.UrlParameterCode), param(stdlib.UrlParameterToken)),
		handler.RedirectInvalidate(serviceMappedUrl))

	return router
}
