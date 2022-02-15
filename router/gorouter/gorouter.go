package gorouter

import (
	"context"
	"encoding/json"
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/http/rest/stdlib"
	"hex-microservice/http/url"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

const (
	ErrorInternal              = "internal"
	ErrorNotFound              = "not-found"
	headerFieldContentType     = "content-type"
	contentTypeJsonWithCharset = "application/json; charset=utf-8"
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
	w.Header().Set(headerFieldContentType, contentTypeJsonWithCharset)
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
	log logr.Logger

	serviceMappedUrl string

	healthPath  string
	servicePath string

	handler stdlib.Handler
}

// New creates a new router inspired by: https://benhoyt.com/writings/web-service-stdlib/.
func New(log logr.Logger, mappedURL string, mappedPath string, healthPath string, hs health.Service, servicePath string, as adder.Service, ls lookup.Service, is invalidator.Service) http.Handler {
	return &goRouter{
		log:              log,
		serviceMappedUrl: url.Join(mappedURL, mappedPath, servicePath),

		healthPath:  url.AbsPath(mappedPath, healthPath),
		servicePath: url.AbsPath(mappedPath, servicePath),

		handler: stdlib.New(log, hs, as, ls, is, paramFunc),
	}
}

func (gr *goRouter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	path := withoutTrailing(r.URL.Path)

	gr.log.Info("router", "method", r.Method, "path", path)

	// e.g. "/health"
	if path == gr.healthPath {
		switch r.Method {
		case http.MethodGet:
			gr.handler.Health(time.Now())(rw, r)
			return
		}
	}

	if strings.HasPrefix(path, gr.servicePath) {
		// e.g "/service"
		if path == gr.servicePath {
			switch r.Method {
			case http.MethodPost:
				gr.handler.RedirectPost(gr.serviceMappedUrl)(rw, r)
				return
			}
		}

		if r := match(r, withoutPrefix(path, gr.servicePath+"/"), stdlib.UrlParameterCode); r != nil {
			switch r.Method {
			case http.MethodGet:
				gr.handler.RedirectGet(gr.serviceMappedUrl)(rw, r)
				return
			}
		}

		if r := match(r, withoutPrefix(path, gr.servicePath+"/"), stdlib.UrlParameterCode, stdlib.UrlParameterToken); r != nil {
			switch r.Method {
			case http.MethodDelete:
				gr.handler.RedirectInvalidate(gr.serviceMappedUrl)(rw, r)
				return
			}
		}
	}

	jsonError(rw, http.StatusNotFound, ErrorNotFound, nil)
	return
}
