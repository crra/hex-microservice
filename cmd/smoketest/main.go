package main

import (
	"context"
	"errors"
	"fmt"
	"hex-microservice/customcontext"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"
)

const flagUrl = "url"

const defaultUrl = "http://localhost:8000"

type HTTPClientDo interface {
	Do(req *http.Request) (*http.Response, error)
}

const defaultHTTPTimeout = 5 * time.Second

var defaultHTTPClientConfiguration = http.Client{
	Timeout:   defaultHTTPTimeout,
	Transport: http.DefaultTransport,
}

var errNonOkStatusCode = errors.New("non-ok http status code")

// app holds the settings for the smoke test
type app struct {
	parent context.Context
	url    string
	status io.Writer
	client HTTPClientDo
}

// mustSuccess performs an HTTP call and ensures that the call returns with a success response.
func mustSuccess(client HTTPClientDo, r *http.Request) (*http.Response, error) {
	response, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return nil, fmt.Errorf("http-reqest for '%s': %w (%d)", r.URL, errNonOkStatusCode, response.StatusCode)
	}

	return response, nil
}

type testfunc func() error

func (a *app) testRunner(cmd *cobra.Command, args []string) error {
	testCtx, testCtxCancel := customcontext.WithErrorCanceller(a.parent)
	defer testCtxCancel(nil)

	// inspired by: https://gist.github.com/jesselucas/179e70a684b6df18189fdaaa24f852cf
	wg := sync.WaitGroup{}
	waitCh := make(chan struct{})

	go func() {
		for _, t := range []struct {
			name string
			f    testfunc
		}{{name: testHealthName, f: a.testHealth}} {
			t := t // pin
			wg.Add(1)
			go func(cancel func(error), wg *sync.WaitGroup) {
				defer wg.Done()
				if err := t.f(); err != nil {
					cancel(fmt.Errorf("%s: %w", t.name, err))
					return
				}
				fmt.Fprintf(a.status, "Success: %s\n", t.name)
			}(testCtxCancel, &wg)
		}

		wg.Wait()
		close(waitCh)
	}()

	select {
	case <-testCtx.Done():
		if err := testCtx.Err(); err != nil {
			return err
		}
	case <-waitCh:
		fmt.Fprintf(a.status, "Success: all tests completed\n")
	}

	return nil
}

func run(parent context.Context, log logr.Logger, status io.Writer, client HTTPClientDo) error {
	if client == nil {
		client = func() HTTPClientDo {
			c := http.Client(defaultHTTPClientConfiguration)
			return &c
		}()
	}

	app := &app{
		parent: parent,
		url:    defaultUrl,
		status: status,
		client: client,
	}
	cmd := &cobra.Command{
		SilenceErrors: true,
		SilenceUsage:  true,

		Args: func(cmd *cobra.Command, args []string) error {
			if _, err := url.ParseRequestURI(app.url); err != nil {
				return err
			}

			return nil
		},
		RunE: app.testRunner,
	}

	f := cmd.Flags()
	f.SortFlags = false // prefer the order defined by the code

	f.StringVar(&app.url, flagUrl, app.url, "url of the endpoint")

	return cmd.Execute()
}

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
	if err := run(context, log, os.Stdout, nil); err != nil {
		log.Error(err, "smoke test failed")
		os.Exit(1)
	}
}
