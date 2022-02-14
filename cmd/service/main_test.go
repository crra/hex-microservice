package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"hex-microservice/meta/value"
	"hex-microservice/repository"
	"hex-microservice/repository/memory"
	"hex-microservice/repository/sqlite"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/stdr"
	"github.com/stretchr/testify/assert"
)

var discardingLogger = stdr.New(log.New(io.Discard, "", log.Lshortfile))

var testRepositories = []struct {
	name   string
	config string
	new    newRepositoryFn
}{
	{name: "memory", new: memory.New},
	{name: "sqlite", config: "file::memory:?cache=shared&_journal_mode=WAL&_foreign_keys=true", new: sqlite.New},
}

const (
	contentTypeMessagePack = "application/x-msgpack"
	contentTypeJson        = "application/json"
)

const (
	mappedUrl = "https://service.arpa"

	healthTestName    = "name"
	healthTestVersion = "version"
	servicePath       = "_service_"
	healthPath        = "_health_"
)

var (
	v1HealthPrefix = value.Join("/", mappedUrl, v1Path, healthPath)
	v1ServicePath  = value.Join("/", mappedUrl, v1Path, servicePath)
)

var (
	healthTestNow         = time.Now()
	healthTestStartupTime = healthTestNow.Add(-1 * time.Minute)
)

type v1HealthResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

type link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
	T    string `json:"type"`
}

type v1CreateResponse struct {
	Code  string `json:"code"`
	URL   string `json:"url"`
	Links []link `json:"_links"`
}

func matrix(t *testing.T, f func(*testing.T, http.Handler, repository.RedirectRepository)) {
	for _, routerImp := range routerImplementations {
		routerImp := routerImp // pin

		for _, repoImp := range testRepositories {
			repoImp := repoImp // pin

			t.Run(fmt.Sprintf("router:%s, repository:%s", routerImp.name, repoImp.name), func(t *testing.T) {
				repository, close, err := repoImp.new(context.Background(), repoImp.config)
				if assert.NoError(t, err) {
					defer close()

					router := routerImp.new(discardingLogger, mappedUrl)
					router.MountV1(
						v1Path,
						healthPath,
						health.New(healthTestName, healthTestVersion, healthTestStartupTime),

						servicePath,
						adder.New(discardingLogger, repository),
						lookup.New(discardingLogger, repository),
						invalidator.New(discardingLogger, repository),
					)

					f(t, router, repository)
				}
			})
		}
	}
}

func v1UrlForCode(code string) string {
	return value.Join("/", mappedUrl, v1Path, servicePath, code)
}

func v1UrlForCodeAndToken(code, token string) string {
	return value.Join("/", mappedUrl, v1Path, servicePath, code, token)
}

func TestV1Health(t *testing.T) {
	matrix(t, func(t *testing.T, router http.Handler, _ repository.RedirectRepository) {
		request := httptest.NewRequest(http.MethodGet, v1HealthPrefix, nil)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)

		if assert.Equal(t, http.StatusOK, responseRecorder.Result().StatusCode) {
			if assert.Contains(t, responseRecorder.Header().Get("content-type"), "application/json") {
				response := &v1HealthResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), response)
				if assert.NoError(t, err) {
					assert.Equal(t, &v1HealthResponse{
						Name:    healthTestName,
						Version: healthTestVersion,
						Uptime:  healthTestNow.Sub(healthTestStartupTime).Round(time.Second).String(),
					}, response)
				}
			}
		}
	})
}

func TestV1RedirectGetRoot(t *testing.T) {
	t.Parallel()

	matrix(t, func(t *testing.T, router http.Handler, _ repository.RedirectRepository) {
		request := httptest.NewRequest(http.MethodGet, v1ServicePath, nil)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)

		assert.Equal(t, http.StatusNotFound, responseRecorder.Result().StatusCode)
	})
}

func TestV1RedirectGetExisting(t *testing.T) {
	const (
		code  = "code"
		token = "token"
		url   = "https://example.com/"
	)

	matrix(t, func(t *testing.T, router http.Handler, repository repository.RedirectRepository) {
		err := repository.Store(adder.RedirectStorage{
			Code:  code,
			Token: token,
			URL:   url,
		})
		if assert.NoError(t, err) {
			request := httptest.NewRequest(http.MethodGet, v1UrlForCode(code), nil)
			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, request)

			if assert.Equal(t, http.StatusTemporaryRedirect, responseRecorder.Result().StatusCode) {
				assert.Equal(t, url, responseRecorder.Result().Header.Get("location"))
			}
		}
	})
}

func TestV1RedirectAdd(t *testing.T) {
	const url = "https://example.com/"
	const customCode = "_code_"
	const payload = `{ "url": "` + url + `", "custom_code": "` + customCode + `" }`

	matrix(t, func(t *testing.T, router http.Handler, _ repository.RedirectRepository) {
		request := httptest.NewRequest("POST", v1ServicePath, strings.NewReader(payload))
		request.Header.Set("Content-Type", contentTypeJson)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)

		if assert.Equal(t, http.StatusCreated, responseRecorder.Result().StatusCode) {
			response := &v1CreateResponse{}
			err := json.Unmarshal(responseRecorder.Body.Bytes(), response)
			if assert.NoError(t, err) {
				assert.Equal(t, url, response.URL)
				assert.Equal(t, customCode, response.Code, "Custom code not applied")

				if assert.NotEmpty(t, response.Links) {
					deleteUrl := value.FirstValueFromSlice(response.Links, func(l link) bool {
						return l.T == http.MethodDelete
					})
					assert.NotNil(t, deleteUrl, "Delete URL missing")

					getUrl := value.FirstValueFromSlice(response.Links, func(l link) bool {
						return l.T == http.MethodGet
					})
					if assert.NotNil(t, getUrl, "Get URL missing") {
						if assert.Equal(t, v1UrlForCode(customCode), getUrl.Href) {
							request := httptest.NewRequest(http.MethodGet, getUrl.Href, nil)
							responseRecorder := httptest.NewRecorder()
							router.ServeHTTP(responseRecorder, request)

							if assert.Equal(t, http.StatusTemporaryRedirect, responseRecorder.Result().StatusCode) {
								assert.Equal(t, url, responseRecorder.Result().Header.Get("location"))
							}
						}
					}
				}
			}
		}
	})
}

func TestV1InvalidateNonExisting(t *testing.T) {
	const (
		code  = "code"
		token = "token"
	)
	matrix(t, func(t *testing.T, router http.Handler, repository repository.RedirectRepository) {
		request := httptest.NewRequest(http.MethodDelete, v1UrlForCodeAndToken(code, token), nil)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)

		assert.Equal(t, http.StatusNotFound, responseRecorder.Result().StatusCode)
	})
}

func TestV1InvalidateExisting(t *testing.T) {
	const (
		code  = "code"
		token = "token"
	)
	matrix(t, func(t *testing.T, router http.Handler, repository repository.RedirectRepository) {
		err := repository.Store(adder.RedirectStorage{
			Code:  code,
			Token: token,
		})

		if assert.NoError(t, err) {
			request := httptest.NewRequest(http.MethodDelete, v1UrlForCodeAndToken(code, token), nil)
			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, request)

			assert.Equal(t, http.StatusNoContent, responseRecorder.Result().StatusCode)
		}
	})
}
