package gin

import (
	"hex-microservice/adder"
	"hex-microservice/health"
	"hex-microservice/http/rest"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"hex-microservice/meta/value"
	"hex-microservice/router"
	"net/http"
	"time"

	org "github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
)

type ginRouter struct {
	log       logr.Logger
	mappedURL string
	router    *org.Engine
}

// newHttpRouter returns a http.Handler that adapts the service with the use of the httprouter router.
func New(log logr.Logger, mappedURL string) router.Router {
	r := org.Default()
	r.HandleMethodNotAllowed = false

	return &ginRouter{
		log:       log,
		mappedURL: mappedURL,
		router:    r,
	}
}

func (gr *ginRouter) MountV1(v1Path string, healthPath string, h health.Service, servicePath string, a adder.Service, l lookup.Service, i invalidator.Service) {
	service := rest.NewGinV1(gr.log, h, a, l, i, func(r *http.Request, key string) string {
		panic("use 'gin.ShouldBind' instead")
	})

	gr.router.GET("/"+value.Join("/", v1Path, healthPath), service.Health(time.Now()))
}

func (gr *ginRouter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	gr.router.ServeHTTP(rw, r)
}
