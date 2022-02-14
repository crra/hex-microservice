package gorouter

import (
	"context"
	"encoding/json"
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/http/rest"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"hex-microservice/router"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

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

	if lenMatches == 1 && matches[0] == "" || lenMatches != lenVars {
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

var withoutPrefix = strings.TrimPrefix

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	b, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		http.Error(w, `{"error":"`+ErrorInternal+`"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	w.Write(b)
}

func jsonError(w http.ResponseWriter, status int, error string, data map[string]interface{}) {
	response := struct {
		Status int                    `json:"status"`
		Error  string                 `json:"error"`
		Data   map[string]interface{} `json:"data,omitempty"`
	}{
		Status: status,
		Error:  error,
		Data:   data,
	}

	writeJSON(w, status, response)
}

type goRouter struct {
	log       logr.Logger
	mappedURL string

	v1            rest.HandlerV1
	v1Path        string
	v1ServicePath string
	v1HealthPath  string
}

// New creates a new router inspired by: https://benhoyt.com/writings/web-service-stdlib/.
func New(log logr.Logger, mappedURL string) router.Router {
	return &goRouter{
		log:       log,
		mappedURL: mappedURL,
	}
}

func (gr *goRouter) MountV1(v1Path string, healthPath string, h health.Service, sericePath string, a adder.Service, l lookup.Service, i invalidator.Service) {
	gr.v1Path = v1Path
	gr.v1HealthPath = healthPath
	gr.v1ServicePath = sericePath

	gr.v1 = rest.NewV1(gr.log, h, a, l, i, paramFunc)
}

func (gr *goRouter) handleV1(path string, rw http.ResponseWriter, r *http.Request) {
	gr.log.Info("router", "method", r.Method, "path", "/"+gr.v1Path+path)

	// e.g. "/health"
	if path == gr.v1HealthPath {
		switch r.Method {
		case http.MethodGet:
			gr.v1.Health(time.Now())(rw, r)
			return
		}
	}

	if strings.HasPrefix(path, gr.v1ServicePath) {
		// e.g "/service"
		if path == gr.v1ServicePath {
			switch r.Method {
			case http.MethodPost:
				gr.v1.RedirectPost(gr.mappedURL, gr.v1ServicePath)(rw, r)
				return
			}
		}

		if r := match(r, withoutPrefix(path, gr.v1ServicePath+"/"), rest.UrlParameterCode); r != nil {
			switch r.Method {
			case http.MethodGet:
				gr.v1.RedirectGet(gr.mappedURL)(rw, r)
				return
			}
		}

		if r := match(r, withoutPrefix(path, gr.v1ServicePath+"/"), rest.UrlParameterCode, rest.UrlParameterToken); r != nil {
			switch r.Method {
			case http.MethodDelete:
				gr.v1.RedirectInvalidate(gr.mappedURL)(rw, r)
				return
			}
		}
	}

	jsonError(rw, http.StatusNotFound, ErrorNotFound, nil)
	return
}

func (gr *goRouter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	path := withoutTrailing(r.URL.Path)

	v1Path := "/" + gr.v1Path + "/"
	if strings.HasPrefix(path, v1Path) && gr.v1 != nil {
		gr.handleV1(path[len(v1Path):], rw, r)
		return
	}

	jsonError(rw, http.StatusNotFound, ErrorNotFound, nil)
}
