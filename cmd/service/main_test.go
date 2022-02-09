package main

import (
	"context"
	"encoding/json"
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/lookup"
	"hex-microservice/repository/memory"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-logr/stdr"
	"github.com/stretchr/testify/assert"
)

var discardingLogger = stdr.New(log.New(io.Discard, "", log.Lshortfile))

func TestHealth(t *testing.T) {
	t.Parallel()

	name := "name"
	version := "version"

	healthService := health.New(name, version)

	type healthResponse struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	for _, ri := range routerImplementations {
		t.Run(ri.name, func(t *testing.T) {
			t.Parallel()

			router := ri.new(discardingLogger, "", healthService, nil, nil, nil)
			request := httptest.NewRequest("GET", "/health", nil)
			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, request)

			if assert.Equal(t, http.StatusOK, responseRecorder.Result().StatusCode) {
				if assert.Equal(t, "application/json", responseRecorder.Header().Get("content-type")) {
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

	repository, err := memory.New(context.Background(), "")
	if assert.NoError(t, err) {
		lookupService := lookup.New(discardingLogger, repository)

		for _, ri := range routerImplementations {
			t.Run(ri.name, func(t *testing.T) {
				t.Parallel()

				router := ri.new(discardingLogger, "", nil, nil, lookupService, nil)
				request := httptest.NewRequest("GET", "/", nil)
				responseRecorder := httptest.NewRecorder()
				router.ServeHTTP(responseRecorder, request)

				assert.Equal(t, http.StatusNotFound, responseRecorder.Result().StatusCode)
			})
		}

	}
}

func TestRedirectGetExisting(t *testing.T) {
	t.Parallel()

	code := "code"
	token := "token"
	url := "https://example.com/"
	clientInfo := "local test client"
	created := time.Now()

	repository, err := memory.New(context.Background(), "")
	if assert.NoError(t, err) {
		err = repository.Store(adder.RedirectStorage{
			Code:       code,
			Token:      token,
			URL:        url,
			ClientInfo: clientInfo,
			CreatedAt:  created,
		})

		if assert.NoError(t, err) {
			lookupService := lookup.New(discardingLogger, repository)

			for _, ri := range routerImplementations {
				t.Run(ri.name, func(t *testing.T) {
					t.Parallel()

					router := ri.new(discardingLogger, "", nil, nil, lookupService, nil)
					request := httptest.NewRequest("GET", "/"+code, nil)
					responseRecorder := httptest.NewRecorder()
					router.ServeHTTP(responseRecorder, request)

					assert.Equal(t, http.StatusTemporaryRedirect, responseRecorder.Result().StatusCode)
				})
			}
		}
	}
}
