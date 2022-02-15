package ginimp

import (
	"encoding/json"
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/http/url"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/vmihailenco/msgpack"
)

const (
	UrlParameterCode  = "code"
	UrlParameterToken = "token"

	resourceName = "redirect"

	contentTypeMessagePack = "application/x-msgpack"
	contentTypeJson        = "application/json"
)

const (
	customCodeAlreadyTaken = "Error code already taken: '%s'"
)

type Handler interface {
	Health(now time.Time) gin.HandlerFunc
	RedirectGet(mappingUrl string) gin.HandlerFunc
	RedirectPost(mappingUrl string) gin.HandlerFunc
	RedirectInvalidate(mappingUrl string) gin.HandlerFunc
}

type converter struct {
	unmarshal func([]byte, any) error
	marshal   func(any) ([]byte, error)
}

type handler struct {
	log logr.Logger

	// services
	adder       adder.Service
	lookup      lookup.Service
	invalidator invalidator.Service
	health      health.Service
	converters  map[string]converter
}

func urlForCode(mappedUrl, code string) string {
	return url.Join(mappedUrl, code)
}

func urlForCodeAndToken(mappedUrl, code, token string) string {
	return url.Join(mappedUrl, code, token)
}

func New(log logr.Logger, health health.Service, adder adder.Service, lookup lookup.Service, invalidator invalidator.Service) Handler {
	return &handler{
		log: log,

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
