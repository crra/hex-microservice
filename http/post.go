package service

import (
	"errors"
	"fmt"
	"hex-microservice/adder"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	validate "gopkg.in/dealancer/validate.v2"
)

// RedirectPost implements the "post" verb of the REST context that creates a new redirect.
func (h *handler) RedirectPost(mappingUrl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			h.log.Error(err, "reading document body")
			return
		}

		// Converter for different content types
		contentType := r.Header.Get("Content-Type")
		converter, ok := h.converters[contentType]
		if !ok {
			h.log.Error(nil, "unsupported content type", "contentType", contentType)
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}

		// extract body
		if len(requestBody) == 0 {
			h.log.Error(err, "empty body")
			writeApiError(w, h.log, ApiError{
				StatusCode: http.StatusBadRequest,
				Title:      titleEmptyBody,
			})
			return
		}

		red := redirectRequest{}
		if err := converter.unmarshal(requestBody, &red); err != nil {
			h.log.Error(err, "unable to unmarshal the request", "contentType", contentType)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// validate
		if err := validate.Validate(red); err != nil {
			var errValidation validate.ErrorValidation
			if errors.As(err, &errValidation) {
				fieldName := errValidation.FieldName()

				h.log.Error(err, "error validating request",
					"fieldValue", reflect.ValueOf(&red).Elem().FieldByName(fieldName),
					"request", red,
				)
				writeApiError(w, h.log, ApiError{
					StatusCode: http.StatusBadRequest,
					Title:      fmt.Sprintf(titleProcessingFieldFormat, fieldName),
				})
				return
			}

			h.log.Error(err, "error validating request", "request", red)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// store
		results, err := h.adder.Add(
			adder.RedirectCommand{
				URL:        red.URL,
				ClientInfo: getIP(r),
			})
		if err != nil {
			h.log.Error(err, "error adding request", "request", red)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// response to client
		asResponse := redirectResponse{
			Code: results[0].Code,
			URL:  red.URL,
			Links: []link{
				{
					Href: strings.Join([]string{mappingUrl, results[0].Code}, "/"),
					Rel:  resourceName,
					T:    "GET",
				},
				{
					Href: strings.Join([]string{mappingUrl, results[0].Code, results[0].Token}, "/"),
					Rel:  resourceName,
					T:    "DELETE",
				},
			},
		}

		responseBody, err := converter.marshal(asResponse)
		if err != nil {
			h.log.Error(err, "marshalling response", "contentType", contentType, "response", asResponse)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if err := writeResponse(w, contentType, responseBody, http.StatusCreated); err != nil {
			h.log.Error(err, "error writing the response to the response object")
			return
		}
	}
}
