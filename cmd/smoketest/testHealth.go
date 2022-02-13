package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const testHealthName = "health endpoint"

// testHealth performs the smoke test for the health endpoint.
func (a *app) testHealth() error {
	url := a.newUrl("health")

	request, err := http.NewRequestWithContext(a.parent, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	response, err := mustSuccess(a.client, request)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("can't read response body: %v", err)
	}

	var healthResponse struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Uptime  string `json:"uptime"`
	}

	if err := json.Unmarshal(body, &healthResponse); err != nil {
		return fmt.Errorf("can't parse response body: %v", err)
	}

	// just check if values are not empty
	if healthResponse.Name == "" {
		return fmt.Errorf("'version' is empty")
	}

	if healthResponse.Version == "" {
		return fmt.Errorf("'version' is empty")
	}

	if healthResponse.Uptime == "" {
		return fmt.Errorf("'uptime' is empty")
	}

	return nil
}
