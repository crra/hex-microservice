package main

import (
	"fmt"
	handler "hex-microservice/http"
	"hex-microservice/shortener"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/julienschmidt/httprouter"
)

// adapt takes a regular http.HandlerFunc and adapts it to use with httprouter.Handle.
func adapt(h http.HandlerFunc) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		h(rw, r)
	}
}

// newHttpRouter returns a http.Handler that adapts the service with the use of the httprouter router.
func newHttpRouter(log logr.Logger, service shortener.Service) http.Handler {
	router := httprouter.New()

	handler := handler.New(log, service, func(r *http.Request, key string) string {
		return httprouter.ParamsFromContext(r.Context()).ByName(key)
	})

	router.GET(fmt.Sprintf("/:%s", urlParameterCode), adapt(handler.Get))
	router.POST("/", adapt(handler.Post))

	return router
}
