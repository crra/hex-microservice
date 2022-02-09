package gorouter

import (
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

func match(path string, pattern *regexp.Regexp, vars ...string) bool {
	matches := pattern.FindStringSubmatch(path)
	if len(matches) <= 0 {
		return false
	}

	return true
}

// newGoRouter
func New(log logr.Logger, mappedURL string, h health.Service, a adder.Service, l lookup.Service, d deleter.Service) http.Handler {
	reCode := regexp.MustCompile(`^/([^/]+)$`)

	s := rest.New(log, h, a, l, d, func(r *http.Request, key string) string {
		// return httprouter.ParamsFromContext(r.Context()).ByName(key)
		return key
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

		switch {
		case match(path, reCode, rest.UrlParameterCode):
			switch r.Method {
			case "GET":
				// s.RedirectGet(mappedURL)(rw, r)
			}
		}

		/*

			var id string

			switch {
			case path == "/albums":
				switch r.Method {
				case "GET":
					s.getAlbums(w, r)
				case "POST":
					s.addAlbum(w, r)
				default:
					w.Header().Set("Allow", "GET, POST")
					s.jsonError(w, http.StatusMethodNotAllowed, ErrorMethodNotAllowed, nil)
				}

			case match(path, reAlbumsID, &id):
				switch r.Method {
				case "GET":
					s.getAlbumByID(w, r, id)
				default:
					w.Header().Set("Allow", "GET")
					s.jsonError(w, http.StatusMethodNotAllowed, ErrorMethodNotAllowed, nil)
				}

			default:
				s.jsonError(w, http.StatusNotFound, ErrorNotFound, nil)
			}
		*/
	})
}
