package httprouter

import (
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/health"
	"hex-microservice/http/rest"
	"hex-microservice/lookup"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
	org "github.com/julienschmidt/httprouter"
)

func paramFunc(r *http.Request, key string) string {
	return org.ParamsFromContext(r.Context()).ByName(key)
}

// newHttpRouter returns a http.Handler that adapts the service with the use of the httprouter router.
func New(log logr.Logger, mappedURL string, h health.Service, a adder.Service, l lookup.Service, d deleter.Service) http.Handler {
	r := org.New()
	r.HandleMethodNotAllowed = false

	s := rest.New(log, h, a, l, d, paramFunc)

	// this is bad: https://github.com/julienschmidt/httprouter/issues/183
	r.HandlerFunc(http.MethodGet, fmt.Sprintf("/:%s", rest.UrlParameterCode), func(rw http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/health") {
			s.Health(time.Now())(rw, r)
			return
		}

		s.RedirectGet(mappedURL)(rw, r)
	})

	r.Handler(http.MethodDelete, fmt.Sprintf("/:%s/:%s", rest.UrlParameterCode, rest.UrlParameterToken), s.RedirectDelete(mappedURL))
	r.Handler(http.MethodPost, "/", s.RedirectPost(mappedURL))

	return r
}
