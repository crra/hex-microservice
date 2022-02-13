package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const testHappyName = "happy path"

// testHealth performs the smoke test for the health endpoint.
func (a *app) testHappy() error {
	url := a.newUrl("")

	const testUrl = "https://example.com"
	customCode := strings.Replace(uuid.New().String(), "-", "", -1)[:25]

	body, err := a.happyCreateRedirect(url, testUrl, customCode)
	if err != nil {
		return err
	}

	cleanupUrl, err := a.testHappyCreated(body, testUrl, customCode)
	if cleanupUrl != "" {
		// TODO: defer delete once
	}
	if err != nil {
		return err
	}

	return nil
}

func (a *app) testHappyCreated(body []byte, testUrl, customCode string) (string, error) {
	var cleanupUrl string
	type link struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
		T    string `json:"type"`
	}

	var createResponse struct {
		Code  string `json:"code"`
		URL   string `json:"url"`
		Links []link `json:"_links"`
	}

	if err := json.Unmarshal(body, &createResponse); err != nil {
		return cleanupUrl, fmt.Errorf("can't parse response body: %v", err)
	}

	if len(createResponse.Links) == 0 {
		return cleanupUrl, fmt.Errorf("'links' not provided")
	}

	delete := sliceGetFirstEqual(createResponse.Links, func(l link) bool {
		return l.T == http.MethodDelete
	})
	if delete == nil {
		return cleanupUrl, fmt.Errorf("'delete link' missing")
	}

	// from now on we can cleanup the urls created
	cleanupUrl = delete.Href

	if createResponse.URL != testUrl {
		return cleanupUrl, fmt.Errorf("'url' is not equal")
	}

	if createResponse.Code != customCode {
		return cleanupUrl, fmt.Errorf("'custom code' is not applied")
	}

	get := sliceGetFirstEqual(createResponse.Links, func(l link) bool {
		return l.T == http.MethodGet
	})
	if get == nil {
		return cleanupUrl, fmt.Errorf("'get link' missing")
	}

	if get.Href != a.newUrl(customCode) {
		return cleanupUrl, fmt.Errorf("'get link' not as expected")
	}

	return cleanupUrl, nil
}

// TODO: create library
func sliceGetFirstEqual[E any](elements []E, fn func(E) bool) *E {
	for _, e := range elements {
		if fn(e) {
			return &e
		}
	}

	return nil
}

func (a *app) happyCreateRedirect(url, redirectUrl, customCode string) ([]byte, error) {
	var body []byte

	payload := fmt.Sprintf(`{"url": "%s", "custom_code": "%s" }`, redirectUrl, customCode)

	request, err := http.NewRequestWithContext(a.parent, http.MethodPost, url,
		bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return body, err
	}

	request.Header.Set("Content-Type", contentTypeJson)

	response, err := mustSuccess(a.client, request)
	if err != nil {
		return body, err
	}

	body, err = io.ReadAll(response.Body)
	if err != nil {
		return body, fmt.Errorf("can't read response body: %v", err)
	}

	return body, nil
}
