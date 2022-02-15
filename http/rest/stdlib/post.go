package stdlib

import (
	"errors"
	"fmt"
	"hex-microservice/adder"
	"io/ioutil"
	"net/http"
	"reflect"

	validate "gopkg.in/dealancer/validate.v2"
)

// redirectRequest is the redirect that is requested by the client.
type redirectRequest struct {
	// mandatory
	URL string `json:"url" msgpack:"url"  validate:"empty=false & format=url"`
	// optional
	CustomCode string `json:"custom_code" msgpack:"custom_code" validate:"empty=true | gte=5 & lte=25"`
}

// RedirectPost implements the "post" verb of the REST context that creates a new redirect.
func (h *handler) RedirectPost(mappingUrl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			h.log.Error(err, "reading document body")
			return
		}

		// Converter for different content types
		contentType := r.Header.Get(headerFieldContentType)
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
				CustomCode: red.CustomCode,
				ClientInfo: getIP(r),
			})
		if err != nil {
			if red.CustomCode != "" && errors.Is(err, adder.ErrDuplicate) {
				writeApiError(w, h.log, ApiError{
					StatusCode: http.StatusConflict,
					Title:      fmt.Sprintf(customCodeAlreadyTaken, red.CustomCode),
				})
				return
			}

			h.log.Error(err, "error adding request", "request", red)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// response to client}
		result := results[0]
		asResponse := redirectResponse{
			Code: result.Code,
			URL:  red.URL,
			Links: []link{
				{
					Href: urlForCode(mappingUrl, result.Code),
					Rel:  resourceName,
					T:    http.MethodGet,
				},
				{
					Href: urlForCodeAndToken(mappingUrl, result.Code, result.Token),
					Rel:  resourceName,
					T:    http.MethodDelete,
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
