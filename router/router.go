package router

import (
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"net/http"
)

type Router interface {
	MountV1(v1Prefix string, healthpath string, hs health.Service, servicePath string, as adder.Service, ls lookup.Service, is invalidator.Service)

	ServeHTTP(http.ResponseWriter, *http.Request)
}
