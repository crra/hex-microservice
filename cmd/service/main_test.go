package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/http/url"
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

	"github.com/gin-gonic/gin"
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
	headerFieldContentType = "content-type"
	contentTypeMessagePack = "application/x-msgpack"
	contentTypeJson        = "application/json"
)

const (
	mappedUrl   = "https://service.arpa"
	mappedPath  = "_path_"
	servicePath = "_service_"
	healthPath  = "_health_"

	healthTestName    = "name"
	healthTestVersion = "version"
)

var (
	healthURL  = url.Join(mappedUrl, mappedPath, healthPath)
	serviceURL = url.Join(mappedUrl, mappedPath, servicePath)
)

func urlForCode(code string) string {
	return url.Join(mappedUrl, mappedPath, servicePath, code)
}

func urlForCodeAndToken(code, token string) string {
	return url.Join(mappedUrl, mappedPath, servicePath, code, token)
}

var (
	healthTestNow         = time.Now()
	healthTestStartupTime = healthTestNow.Add(-1 * time.Minute)
)

type healthResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

type link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
	T    string `json:"type"`
}

type createResponse struct {
	Code  string `json:"code"`
	URL   string `json:"url"`
	Links []link `json:"_links"`
}

func matrix(t *testing.T, f func(*testing.T, http.Handler, repository.RedirectRepository)) {
	gin.SetMode(gin.TestMode)

	for _, routerImp := range routerImplementations {
		routerImp := routerImp // pin

		for _, repoImp := range testRepositories {
			repoImp := repoImp // pin

			t.Run(fmt.Sprintf("router:%s,repository:%s", routerImp.name, repoImp.name), func(t *testing.T) {
				repository, close, err := repoImp.new(context.Background(), repoImp.config)
				if assert.NoError(t, err) {
					defer close()

					router := routerImp.new(
						discardingLogger,
						mappedUrl,
						mappedPath,

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

func TestHealth(t *testing.T) {
	matrix(t, func(t *testing.T, router http.Handler, _ repository.RedirectRepository) {
		request := httptest.NewRequest(http.MethodGet, healthURL, nil)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)

		if assert.Equal(t, http.StatusOK, responseRecorder.Result().StatusCode) {
			if assert.Contains(t, responseRecorder.Header().Get(headerFieldContentType), contentTypeJson) {
				response := &healthResponse{}
				err := json.Unmarshal(responseRecorder.Body.Bytes(), response)
				if assert.NoError(t, err) {
					assert.Equal(t, &healthResponse{
						Name:    healthTestName,
						Version: healthTestVersion,
						Uptime:  healthTestNow.Sub(healthTestStartupTime).Round(time.Second).String(),
					}, response)
				}
			}
		}
	})
}

func TestRedirectGetRoot(t *testing.T) {
	matrix(t, func(t *testing.T, router http.Handler, _ repository.RedirectRepository) {
		request := httptest.NewRequest(http.MethodGet, serviceURL, nil)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)

		assert.Equal(t, http.StatusNotFound, responseRecorder.Result().StatusCode)
	})
}

func TestRedirectGetExisting(t *testing.T) {
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
			request := httptest.NewRequest(http.MethodGet, urlForCode(code), nil)
			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, request)

			if assert.Equal(t, http.StatusTemporaryRedirect, responseRecorder.Result().StatusCode) {
				assert.Equal(t, url, responseRecorder.Result().Header.Get("location"))
			}
		}
	})
}

func TestRedirectAdd(t *testing.T) {
	const url = "https://example.com/"
	const payload = `{ "url": "` + url + `" }`

	matrix(t, func(t *testing.T, router http.Handler, _ repository.RedirectRepository) {
		request := httptest.NewRequest(http.MethodPost, serviceURL, strings.NewReader(payload))
		request.Header.Set(headerFieldContentType, contentTypeJson)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)

		if assert.Equal(t, http.StatusCreated, responseRecorder.Result().StatusCode) {
			response := &createResponse{}
			err := json.Unmarshal(responseRecorder.Body.Bytes(), response)
			if assert.NoError(t, err) {
				assert.Equal(t, url, response.URL)

				if assert.NotEmpty(t, response.Links) {
					deleteUrl := value.FirstValueFromSlice(response.Links, func(l link) bool {
						return l.T == http.MethodDelete
					})
					assert.NotNil(t, deleteUrl, "Delete URL missing")

					getUrl := value.FirstValueFromSlice(response.Links, func(l link) bool {
						return l.T == http.MethodGet
					})
					if assert.NotNil(t, getUrl, "Get URL missing") {
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
	})
}

func TestRedirectAddAndRead(t *testing.T) {
	const url = "https://example.com/"
	const customCode = "_code_"
	const payload = `{ "url": "` + url + `", "custom_code": "` + customCode + `" }`

	matrix(t, func(t *testing.T, router http.Handler, _ repository.RedirectRepository) {
		request := httptest.NewRequest(http.MethodPost, serviceURL, strings.NewReader(payload))
		request.Header.Set(headerFieldContentType, contentTypeJson)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)

		if assert.Equal(t, http.StatusCreated, responseRecorder.Result().StatusCode) {
			response := &createResponse{}
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
						if assert.Equal(t, urlForCode(customCode), getUrl.Href) {
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

func TestInvalidateNonExisting(t *testing.T) {
	const (
		code  = "code"
		token = "token"
	)
	matrix(t, func(t *testing.T, router http.Handler, repository repository.RedirectRepository) {
		request := httptest.NewRequest(http.MethodDelete, urlForCodeAndToken(code, token), nil)
		responseRecorder := httptest.NewRecorder()
		router.ServeHTTP(responseRecorder, request)

		assert.Equal(t, http.StatusNotFound, responseRecorder.Result().StatusCode)
	})
}

func TestInvalidateExisting(t *testing.T) {
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
			request := httptest.NewRequest(http.MethodDelete, urlForCodeAndToken(code, token), nil)
			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, request)

			assert.Equal(t, http.StatusNoContent, responseRecorder.Result().StatusCode)
		}
	})
}
