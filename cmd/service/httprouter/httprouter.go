package httprouter

import (
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/health"
	"hex-microservice/http/rest"
	"hex-microservice/lookup"
	"net/http"

	"github.com/go-logr/logr"
	org "github.com/julienschmidt/httprouter"
)

// adapt takes a regular http.HandlerFunc and adapts it to use with httprouter.Handle.
func adapt(h http.HandlerFunc) org.Handle {
	return func(rw http.ResponseWriter, r *http.Request, _ org.Params) {
		h(rw, r)
	}
}

// newHttpRouter returns a http.Handler that adapts the service with the use of the httprouter router.
func New(log logr.Logger, mappedURL string, h health.Service, a adder.Service, l lookup.Service, d deleter.Service) http.Handler {
	r := org.New()

	s := rest.New(log, h, a, l, d, func(r *http.Request, key string) string {
		return org.ParamsFromContext(r.Context()).ByName(key)
	})

	// this is bad: https://github.com/julienschmidt/httprouter/issues/183

	r.GET(fmt.Sprintf("/redirect/:%s", rest.UrlParameterCode), adapt(s.RedirectGet(mappedURL)))
	r.POST("/redirect/", adapt(s.RedirectPost(mappedURL)))
	r.GET(fmt.Sprintf("/redirect/:%s/:%s", rest.UrlParameterCode, rest.UrlParameterToken), adapt(s.RedirectDelete(mappedURL)))

	r.GET("/health", adapt(s.Health()))

	return r
}
