package gin

import (
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/http/rest/ginimp"
	"hex-microservice/http/url"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"net/http"
	"time"

	org "github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
)

func param(name string) string {
	return ":" + name
}

// newHttpRouter returns a http.Handler that adapts the service with the use of the httprouter router.
func New(log logr.Logger, mappedURL string, mappedPath string, healthPath string, hs health.Service, servicePath string, as adder.Service, ls lookup.Service, is invalidator.Service) http.Handler {
	router := org.Default()
	router.HandleMethodNotAllowed = false
	router.Use(org.Logger())
	router.Use(org.Recovery())

	serviceMappedUrl := url.Join(mappedURL, mappedPath, servicePath)
	handler := ginimp.New(log, hs, as, ls, is)

	router.GET(url.AbsPath(mappedPath, healthPath),
		handler.Health(time.Now()))

	router.GET(url.AbsPath(mappedPath, servicePath, param(ginimp.UrlParameterCode)),
		handler.RedirectGet(serviceMappedUrl))

	router.POST(url.AbsPath(mappedPath, servicePath),
		handler.RedirectPost(serviceMappedUrl))

	router.DELETE(url.AbsPath(mappedPath, servicePath, param(ginimp.UrlParameterCode), param(ginimp.UrlParameterToken)),
		handler.RedirectInvalidate(serviceMappedUrl))

	return router
}
