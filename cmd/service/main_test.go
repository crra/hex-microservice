package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/health"
	"hex-microservice/lookup"
	"hex-microservice/repository/memory"
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

const (
	contentTypeMessagePack = "application/x-msgpack"
	contentTypeJson        = "application/json"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	const (
		name    = "name"
		version = "version"
	)
	start := time.Now()

	healthService := health.New(name, version, start)

	type healthResponse struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	for _, ri := range routerImplementations {
		ri := ri // pin

		t.Run(ri.name, func(t *testing.T) {
			t.Parallel()

			router := ri.new(discardingLogger, "", healthService, nil, nil, nil)
			request := httptest.NewRequest(http.MethodGet, "/health", nil)
			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, request)

			if assert.Equal(t, http.StatusOK, responseRecorder.Result().StatusCode) {
				if assert.Contains(t, responseRecorder.Header().Get("content-type"), "application/json") {
					response := &healthResponse{}
					json.Unmarshal(responseRecorder.Body.Bytes(), response)
					assert.Equal(t, &healthResponse{Name: name, Version: version}, response)
				}
			}
		})
	}
}

func TestRedirectGetRoot(t *testing.T) {
	t.Parallel()

	repository, close, err := memory.New(context.Background(), "")
	if assert.NoError(t, err) {
		defer close()

		lookupService := lookup.New(discardingLogger, repository)

		for _, ri := range routerImplementations {
			ri := ri // pin

			t.Run(ri.name, func(t *testing.T) {
				t.Parallel()

				router := ri.new(discardingLogger, "", nil, nil, lookupService, nil)
				request := httptest.NewRequest(http.MethodGet, "/", nil)
				responseRecorder := httptest.NewRecorder()
				router.ServeHTTP(responseRecorder, request)

				assert.Equal(t, http.StatusNotFound, responseRecorder.Result().StatusCode)
			})
		}

	}
}

func TestRedirectGetExisting(t *testing.T) {
	t.Parallel()

	const (
		code  = "code"
		token = "token"
		url   = "https://example.com/"
	)

	repository, close, err := memory.New(context.Background(), "")
	if assert.NoError(t, err) {
		defer close()

		err = repository.Store(adder.RedirectStorage{
			Code:  code,
			Token: token,
			URL:   url,
		})

		if assert.NoError(t, err) {
			lookupService := lookup.New(discardingLogger, repository)

			for _, ri := range routerImplementations {
				ri := ri // pin

				t.Run(ri.name, func(t *testing.T) {
					t.Parallel()

					router := ri.new(discardingLogger, "", nil, nil, lookupService, nil)
					request := httptest.NewRequest(http.MethodGet, "/"+code, nil)
					responseRecorder := httptest.NewRecorder()
					router.ServeHTTP(responseRecorder, request)

					if assert.Equal(t, http.StatusTemporaryRedirect, responseRecorder.Result().StatusCode) {
						assert.Equal(t, url, responseRecorder.Result().Header.Get("location"))
					}
				})
			}
		}
	}
}

func TestRedirectAdd(t *testing.T) {
	t.Parallel()

	const url = "https://example.com/"

	payload := fmt.Sprintf(`{ "url": "%s" }`, url)

	for _, ri := range routerImplementations {
		ri := ri // pin

		t.Run(ri.name, func(t *testing.T) {
			t.Parallel()

			repository, close, err := memory.New(context.Background(), "")
			if assert.NoError(t, err) {
				defer close()

				adderService := adder.New(discardingLogger, repository)

				router := ri.new(discardingLogger, "", nil, adderService, nil, nil)

				request := httptest.NewRequest("POST", "/", strings.NewReader(payload))
				request.Header.Set("Content-Type", contentTypeJson)

				responseRecorder := httptest.NewRecorder()
				router.ServeHTTP(responseRecorder, request)

				assert.Equal(t, http.StatusCreated, responseRecorder.Result().StatusCode)
			}
		})
	}
}

func TestRedirectDeleteNonExisting(t *testing.T) {
	t.Parallel()

	const (
		code  = "code"
		token = "token"
	)

	repository, close, err := memory.New(context.Background(), "")
	if assert.NoError(t, err) {
		defer close()

		deleterService := deleter.New(discardingLogger, repository)

		for _, ri := range routerImplementations {
			ri := ri // pin

			t.Run(ri.name, func(t *testing.T) {
				t.Parallel()

				router := ri.new(discardingLogger, "", nil, nil, nil, deleterService)

				request := httptest.NewRequest("DELETE", "/"+code+"/"+token, nil)

				responseRecorder := httptest.NewRecorder()
				router.ServeHTTP(responseRecorder, request)

				assert.Equal(t, http.StatusNotFound, responseRecorder.Result().StatusCode)
			})
		}
	}
}

func TestRedirectDeleteExisting(t *testing.T) {
	t.Parallel()

	const (
		code  = "code"
		token = "token"
	)

	for _, ri := range routerImplementations {
		ri := ri // pin

		t.Run(ri.name, func(t *testing.T) {
			t.Parallel()

			repository, close, err := memory.New(context.Background(), "")
			if assert.NoError(t, err) {
				defer close()

				err = repository.Store(adder.RedirectStorage{
					Code:  code,
					Token: token,
				})

				if assert.NoError(t, err) {
					deleterService := deleter.New(discardingLogger, repository)

					router := ri.new(discardingLogger, "", nil, nil, nil, deleterService)

					request := httptest.NewRequest("DELETE", "/"+code+"/"+token, nil)

					responseRecorder := httptest.NewRecorder()
					router.ServeHTTP(responseRecorder, request)

					assert.Equal(t, http.StatusNoContent, responseRecorder.Result().StatusCode)
				}
			}
		})
	}
}
