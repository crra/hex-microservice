package httprouter

import (
	"context"
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/health"
	"hex-microservice/http/rest"
	"hex-microservice/lookup"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	org "github.com/julienschmidt/httprouter"
)

// adapt takes a regular http.HandlerFunc and adapts it to use with httprouter.Handle.
func adapt(h http.HandlerFunc) org.Handle {
	return func(rw http.ResponseWriter, r *http.Request, _ org.Params) {
		h(rw, r)
	}
}

func paramFunc(r *http.Request, key string) string {
	return org.ParamsFromContext(r.Context()).ByName(key)
}

// newHttpRouter returns a http.Handler that adapts the service with the use of the httprouter router.
func New(log logr.Logger, mappedURL string, h health.Service, a adder.Service, l lookup.Service, d deleter.Service) http.Handler {
	r := org.New()

	s := rest.New(log, h, a, l, d, paramFunc)

	// this is bad: https://github.com/julienschmidt/httprouter/issues/183
	r.GET(fmt.Sprintf("/:%s", rest.UrlParameterCode), func(rw http.ResponseWriter, r *http.Request, p org.Params) {
		r = r.WithContext(context.WithValue(r.Context(), org.ParamsKey, p))

		if strings.HasPrefix(r.URL.Path, "/health") {
			s.Health()(rw, r)
			return
		}

		s.RedirectGet(mappedURL)(rw, r)
	})

	r.DELETE(fmt.Sprintf("/:%s/:%s", rest.UrlParameterCode, rest.UrlParameterToken), adapt(s.RedirectDelete(mappedURL)))
	r.POST("/", adapt(s.RedirectPost(mappedURL)))

	return r
}
