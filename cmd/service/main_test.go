package main

import (
	"hex-microservice/health"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-logr/stdr"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	// t.Parallel()

	log := stdr.New(log.New(os.Stdout, "", log.Lshortfile))

	name := "name"
	version := "version"

	healthService := health.New(name, version)

	for _, ri := range routerImplementations {
		t.Run(ri.name, func(t *testing.T) {
			// t.Parallel()

			router := ri.new(log, "", healthService, nil, nil, nil)
			request := httptest.NewRequest("GET", "/health", nil)
			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, request)
			assert.Equal(t, "application/json", responseRecorder.Header().Get("content-type"))
		})
	}
}
