package main

import (
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	httpservice "hex-microservice/http"
	"hex-microservice/lookup"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/julienschmidt/httprouter"
)

// adapt takes a regular http.HandlerFunc and adapts it to use with httprouter.Handle.
func adapt(h http.HandlerFunc) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		h(rw, r)
	}
}

// newHttpRouter returns a http.Handler that adapts the service with the use of the httprouter router.
func newHttpRouter(log logr.Logger, mappedURL string, a adder.Service, l lookup.Service, d deleter.Service) http.Handler {
	r := httprouter.New()

	s := httpservice.New(log, a, l, d, func(r *http.Request, key string) string {
		return httprouter.ParamsFromContext(r.Context()).ByName(key)
	})

	r.GET(fmt.Sprintf("/:%s", httpservice.UrlParameterCode), adapt(s.RedirectGet(mappedURL)))
	r.POST("/", adapt(s.RedirectPost(mappedURL)))
	r.GET(fmt.Sprintf("/:%s/:%s", httpservice.UrlParameterCode, httpservice.UrlParameterToken), adapt(s.RedirectDelete(mappedURL)))

	return r
}
