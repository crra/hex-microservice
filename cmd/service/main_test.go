package main

import (
	"encoding/json"
	"hex-microservice/health"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-logr/stdr"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	log := stdr.New(log.New(os.Stdout, "", log.Lshortfile))

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

			router := ri.new(log, "", healthService, nil, nil, nil)
			request := httptest.NewRequest("GET", "/health", nil)
			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, request)
			if assert.Equal(t, "application/json", responseRecorder.Header().Get("content-type")) {
				response := &healthResponse{}
				json.Unmarshal(responseRecorder.Body.Bytes(), response)
				assert.Equal(t, &healthResponse{Name: name, Version: version}, response)
			}
		})
	}
}
