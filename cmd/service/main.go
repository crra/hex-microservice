package main

// Encoding: https://www.google.com -> 98sj1-293
// Decoding: http://localhost:8000/98sj1-293 -> https://www.google.com

// repo <- service -> serializer -> http

import (
	"context"
	"errors"
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/cmd/service/chirouter"
	"hex-microservice/cmd/service/ginrouter"
	"hex-microservice/cmd/service/gorouter"
	"hex-microservice/cmd/service/httprouter"
	"hex-microservice/cmd/service/muxrouter"
	"hex-microservice/customcontext"
	"hex-microservice/deleter"
	"hex-microservice/health"
	"hex-microservice/lookup"
	"hex-microservice/meta/value"
	"hex-microservice/repository"
	"hex-microservice/repository/memory"
	"hex-microservice/repository/mongo"
	"hex-microservice/repository/redis"
	"hex-microservice/repository/sqlite"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"

	"github.com/spf13/viper"
)

// externally set by the build system
var (
	version = "dev-build"
	name    = "shortener"
)

// default server values
const (
	defaultBind           = "localhost:8000"
	defaultMapped         = "http://" + defaultBind
	defaultRepositoryArgs = ""

	// considder: https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	defaultServerIdleTimeout    = 120 * time.Second
	defaultServerReadTimeout    = 5 * time.Second
	defaultServerWriteTimeout   = 5 * time.Second
	ServerShutdownGraceDuration = 5 * time.Second
)

// used configuration keys
const (
	configKeyBind       = "bind"
	configKeyMapped     = "mapped"
	configKeyRouter     = "router"
	configKeyRepository = "repository"
)

var (
	// used default repository implementation
	defaultRepository = repositoryImplementations[0]
	// used default router implementation
	defaultRouter = routerImplementations[0]
)

// repositoryImpl represents a repository implementation that can be instantiated.
type repositoryImpl struct {
	name string
	new  func(context.Context, string) (repository.RedirectRepository, repository.Close, error)
}

// String returns the string representation of the routerImpl.
func (r routerImpl) String() string { return r.name }

// repositoryImpl represents a router implementation that can be instantiated.
type routerImpl struct {
	name string
	new  func(logr.Logger, string, health.Service, adder.Service, lookup.Service, deleter.Service) http.Handler
}

// String returns the string representation of the repositoryImpl.
func (r repositoryImpl) String() string { return r.name }

// available router implementations
var routerImplementations = []routerImpl{
	{"go", gorouter.New},
	{"chi", chirouter.New},
	{"gorilla", muxrouter.New},
	{"httprouter", httprouter.New},
	{"gin", ginrouter.New},
}

// available repository implementations
var repositoryImplementations = []repositoryImpl{
	{"memory", memory.New},
	{"redis", redis.New},
	{"mongodb", mongo.New},
	{"sqlite", sqlite.New},
}

// configuration describes the user defined configuration options.
type configuration struct {
	Bind           string
	Mapped         string
	Router         routerImpl
	Repository     repositoryImpl
	RepositoryArgs string
}

// getConfiguration retrieves the configuration of the service.
func getConfiguration(log logr.Logger) (*configuration, error) {
	v := viper.New()

	v.SetConfigName(name)
	v.SetConfigType("env")
	v.AddConfigPath(fmt.Sprintf("$HOME/.%s", name))
	v.AddConfigPath(".")
	v.AutomaticEnv()

	v.SetDefault(configKeyBind, defaultBind)
	v.SetDefault(configKeyMapped, defaultMapped)
	v.SetDefault(configKeyRepository, defaultRepository.String())
	v.SetDefault(configKeyRouter, defaultRouter.String())

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("fatal error reading config file: %w", err)
		}
	}

	// check if provided value is supported and fall back to default otherwise
	router, ok := value.FirstByString(routerImplementations, strings.ToLower, v.GetString(configKeyRouter))
	if !ok {
		router = defaultRouter
		log.Info("default configuration value due to unsupported value", "key", configKeyRouter, "provided", v.GetString(configKeyRouter), "using", defaultRouter)
	}

	repositoryType := defaultRouter.String()
	repositoryArgs := defaultRepositoryArgs

	if parts, err := url.Parse(v.GetString(configKeyRepository)); err == nil {
		repositoryType = parts.Scheme
		repositoryArgs = parts.String()
	}

	repository, ok := value.FirstByString(repositoryImplementations, strings.ToLower, repositoryType)
	if !ok {
		repository = defaultRepository
		repositoryArgs = defaultRepositoryArgs
		log.Info("default configuration value due to unsupported value", "key", configKeyRepository, "provided", v.GetString(configKeyRepository), "using", repository)
	}

	return &configuration{
		Bind:           v.GetString(configKeyBind),
		Mapped:         v.GetString(configKeyMapped),
		Router:         router,
		Repository:     repository,
		RepositoryArgs: repositoryArgs,
	}, nil
}

// run encloses the program in a function that can take dependencies (parameters) and can return an error.
func run(parent context.Context, log logr.Logger) error {
	// process user input
	c, err := getConfiguration(log)
	if err != nil {
		return fmt.Errorf("error processing configuration: %w", err)
	}

	// initialize the configured repository
	// use a factory function (new) of the supported type
	repository, close, err := c.Repository.new(parent, c.RepositoryArgs)
	if err != nil {
		return fmt.Errorf("error creating repository: %w", err)
	}

	defer close()

	// the service is the domain core
	lookupService := lookup.New(log, repository)
	adderService := adder.New(log, repository)
	deleteService := deleter.New(log, repository)
	healthService := health.New(name, version, time.Now())

	// initialize the configured router
	// use a factory function (new) of the supported type
	router := c.Router.new(log, c.Mapped, healthService, adderService, lookupService, deleteService)

	// use the built-in http server
	server := &http.Server{
		Addr:         c.Bind,
		Handler:      router,
		IdleTimeout:  defaultServerIdleTimeout,
		ReadTimeout:  defaultServerReadTimeout,
		WriteTimeout: defaultServerWriteTimeout,
	}

	// a server context that handles an error during startup
	serverCtx, serverCtxCancel := customcontext.WithErrorCanceller(parent)
	defer serverCtxCancel(nil)

	// start the server in an own go routine
	// and cancel the server context on errors
	go func(cancel func(error)) {
		log.Info("Server started")
		cancel(server.ListenAndServe())
	}(serverCtxCancel)

	log.Info("Waiting for shutdown")
	<-serverCtx.Done()
	log.Info("Shutdown requested")

	// propagate application errors (e.g. during startup)
	if err := serverCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	// perform a graceful shutdown but cancel after a timeout
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), ServerShutdownGraceDuration)
	defer timeoutCancel()

	if err := server.Shutdown(timeoutCtx); err != nil {
		log.Info("Shutdown with error")
		return err
	}

	log.Info("Regular shutdown performed")
	return nil
}

// main is the entrypoint of the program.
// main is the only place where external dependencies (e.g. output stream, logger, filesystem)
// are resolved and where final errors are handled (e.g. writing to the console).
func main() {
	// use the built in logger
	log := stdr.New(log.New(os.Stdout, "", log.Lshortfile))

	// create a parent context that listens on os signals (e.g. CTRL-C)
	context, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// cancel the parent context and all children if an os signal arrives
	go func() {
		<-context.Done()
		cancel()
	}()

	// run the program and clean up
	if err := run(context, log); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
}
