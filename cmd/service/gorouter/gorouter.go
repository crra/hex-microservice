package gorouter

import (
	"context"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/health"
	"hex-microservice/http/rest"
	"hex-microservice/lookup"
	"net/http"
	"regexp"

	"github.com/go-logr/logr"
)

//
// See: https://benhoyt.com/writings/web-service-stdlib/
//

type router struct{}

type goRouter func(http.ResponseWriter, *http.Request)

func (f goRouter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// delegate to the anonymous function
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

func match(r *http.Request, pattern *regexp.Regexp, vars ...string) *http.Request {
	matches := pattern.FindStringSubmatch(r.URL.Path)
	lenMatches := len(matches)
	if lenMatches <= 0 {
		return nil
	}

	parts := make(map[string]string, lenMatches)

	for i, v := range vars {
		if i > lenMatches {
			break
		}
		parts[v] = matches[i]
	}

	ctx := context.WithValue(r.Context(), varsKey, parts)

	return r.WithContext(ctx)
}

// newGoRouter
func New(log logr.Logger, mappedURL string, h health.Service, a adder.Service, l lookup.Service, d deleter.Service) http.Handler {
	reCode := regexp.MustCompile(`^/([^/]+)$`)

	s := rest.New(log, h, a, l, d, func(r *http.Request, key string) string {
		if rv := r.Context().Value(varsKey); rv != nil {
			if kv, ok := rv.(map[string]string); ok {
				return kv[key]
			}
		}

		return ""
	})

	return goRouter(func(rw http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		log.Info("router", "method", r.Method, "path", path)

		switch path {
		case "/health":
			switch r.Method {
			case "GET":
				s.Health()(rw, r)
			}
		}

		if r := match(r, reCode, rest.UrlParameterCode); r != nil {
			switch r.Method {
			case "GET":
				s.RedirectGet(mappedURL)(rw, r)
			}
		}
	})
}
