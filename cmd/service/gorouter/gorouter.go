package gorouter

import (
	"context"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/health"
	"hex-microservice/http/rest"
	"hex-microservice/lookup"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
)

//
// See: https://benhoyt.com/writings/web-service-stdlib/
//

type router struct{}

type goRouter func(http.ResponseWriter, *http.Request)

func (f goRouter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	f(rw, r)
}

const (
	ErrorAlreadyExists    = "already-exists"
	ErrorDatabase         = "database"
	ErrorInternal         = "internal"
	ErrorMalformedJSON    = "malformed-json"
	ErrorMethodNotAllowed = "method-not-allowed"
	ErrorNotFound         = "not-found"
	ErrorValidation       = "validation"
)

const varsKey = "UrlParameter"

func match(r *http.Request, path string, vars ...string) *http.Request {
	matches := strings.Split(path, "/")
	lenMatches := len(matches)
	lenVars := len(vars)

	if lenMatches <= 0 || lenMatches != lenVars {
		return nil
	}

	parts := make(map[string]string, lenMatches)

	for i, m := range matches {
		parts[vars[i]] = m
	}

	ctx := context.WithValue(r.Context(), varsKey, parts)

	return r.WithContext(ctx)
}

func paramFunc(r *http.Request, key string) string {
	if rv := r.Context().Value(varsKey); rv != nil {
		if kv, ok := rv.(map[string]string); ok {
			return kv[key]
		}
	}

	return ""
}

func withoutTrailing(path string) string {
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}

	return path
}

func withoutPrefix(path, prefix string) string {
	return strings.TrimLeft(path, prefix)
}

// newGoRouter
func New(log logr.Logger, mappedURL string, h health.Service, a adder.Service, l lookup.Service, d deleter.Service) http.Handler {
	s := rest.New(log, h, a, l, d, paramFunc)

	return goRouter(func(rw http.ResponseWriter, r *http.Request) {
		path := withoutTrailing(r.URL.Path)

		log.Info("router", "method", r.Method, "path", path)

		switch path {
		case "/health":
			switch r.Method {
			case "GET":
				s.Health()(rw, r)
			}
		}

		const prefix = "/"
		if r := match(r, withoutPrefix(path, prefix), rest.UrlParameterCode); r != nil {
			switch r.Method {
			case "GET":
				s.RedirectGet(mappedURL)(rw, r)
			}
		}
	})
}
