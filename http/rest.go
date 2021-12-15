package service

import (
	"errors"
	"hex-microservice/shortener"
	"io/ioutil"
	"log"
	"net/http"

	js "hex-microservice/serializer/json"
	ms "hex-microservice/serializer/msgpack"

	"github.com/go-logr/logr"
)

type ParamFn func(r *http.Request, key string) string

type RedirectHandler interface {
	Get(http.ResponseWriter, *http.Request)
	Post(http.ResponseWriter, *http.Request)
}

type handler struct {
	log             logr.Logger
	paramFn         ParamFn
	redirectService shortener.Service
}

func New(log logr.Logger, redirectService shortener.Service, paramFn ParamFn) RedirectHandler {
	return &handler{
		log:             log,
		paramFn:         paramFn,
		redirectService: redirectService,
	}
}

func writeResponse(w http.ResponseWriter, contentType string, body []byte, statusCode int) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	if _, err := w.Write(body); err != nil {
		log.Println(err)
	}
}

func (h *handler) serializer(contentType string) shortener.RedirectSerializer {
	if contentType == "application/x-msgpack" {
		return &ms.Redirect{}
	}

	return &js.Redirect{}
}

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	code := h.paramFn(r, "code")
	redirect, err := h.redirectService.Find(code)
	if err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, shortener.ErrRedirectNotFound) {
			status = http.StatusNotFound
		}

		http.Error(w, http.StatusText(status), status)
		return
	}

	http.Redirect(w, r, redirect.URL, http.StatusMovedPermanently)
}

func (h *handler) Post(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	contentType := r.Header.Get("Content-Type")
	redirect, err := h.serializer(contentType).Decode(requestBody)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err = h.redirectService.Store(redirect); err != nil {
		status := http.StatusInternalServerError

		if errors.Is(err, shortener.ErrRedirectInvalid) {
			status = http.StatusBadRequest
		}

		http.Error(w, http.StatusText(status), status)
		return
	}

	responseBody, err := h.serializer(contentType).Encode(redirect)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeResponse(w, contentType, responseBody, http.StatusCreated)
}
