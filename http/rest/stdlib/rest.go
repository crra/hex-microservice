package stdlib

import (
	"encoding/json"
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/http/url"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/vmihailenco/msgpack"
)

const (
	UrlParameterCode  = "code"
	UrlParameterToken = "token"

	headerFieldContentType = "content-type"

	contentTypeMessagePack = "application/x-msgpack"
	contentTypeJson        = "application/json"

	resourceName = "redirect"

	titleEmptyBody             = "Error processing request body, the content is empty"
	titleProcessingFieldFormat = "Error processing field: '%s'"
	missingParameterFormat     = "Error missing parameter: '%s'"
	customCodeAlreadyTaken     = "Error code already taken: '%s'"
)

type ParamFn func(r *http.Request, key string) string

type Handler interface {
	Health(now time.Time) http.HandlerFunc
	RedirectGet(mappingUrl string) http.HandlerFunc
	RedirectPost(mappingUrl string) http.HandlerFunc
	RedirectInvalidate(mappingUrl string) http.HandlerFunc
}

type converter struct {
	unmarshal func([]byte, any) error
	marshal   func(any) ([]byte, error)
}

type link struct {
	Href string `json:"href,omitempty"`
	Rel  string `json:"rel,omitempty"`
	T    string `json:"type,omitempty"`
}

// redirectResponse is the redirect that is returned to the client.
type redirectResponse struct {
	Code string `json:"code" msgpack:"code"`
	URL  string `json:"url" msgpack:"url"`

	Links []link `json:"_links,omitempty"`
}

// handler is the implementation of the REST service.
type handler struct {
	log     logr.Logger
	paramFn ParamFn
	// services
	adder       adder.Service
	lookup      lookup.Service
	invalidator invalidator.Service
	health      health.Service
	converters  map[string]converter
}

type ApiError struct {
	StatusCode int    `json:"status"`
	Title      string `json:"title"`
}

func urlForCode(mappedUrl, code string) string {
	return url.Join(mappedUrl, code)
}

func urlForCodeAndToken(mappedUrl, code, token string) string {
	return url.Join(mappedUrl, code, token)
}

func New(log logr.Logger, health health.Service, adder adder.Service, lookup lookup.Service, invalidator invalidator.Service, paramFn ParamFn) Handler {
	return &handler{
		log:     log,
		paramFn: paramFn,

		health:      health,
		adder:       adder,
		lookup:      lookup,
		invalidator: invalidator,
		// NOTE: not really sure if this is a good pattern with the lookup table,
		// but it was taken from the original example.
		converters: map[string]converter{
			contentTypeJson:        {json.Unmarshal, json.Marshal},
			contentTypeMessagePack: {msgpack.Unmarshal, msgpack.Marshal},
		},
	}
}

// writeResponse is a helper function that write the necessary data to the response.
func writeResponse(w http.ResponseWriter, contentType string, body []byte, statusCode int) error {
	w.Header().Set(headerFieldContentType, contentType)
	w.WriteHeader(statusCode)

	_, err := w.Write(body)

	return err
}

// writeApiError is a helper function that writes an ApiError to the response.
func writeApiError(w http.ResponseWriter, log logr.Logger, apiErr ApiError) {
	errReturn, err := json.Marshal(apiErr)
	if err != nil {
		log.Error(err, "error marshalling error")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerFieldContentType, contentTypeJson)
	http.Error(w, string(errReturn), apiErr.StatusCode)
	return
}

// getIP returns the requestor's (could be proxied or direct or faked) ip address (either V4 or V6).
func getIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}

	return r.RemoteAddr
}
