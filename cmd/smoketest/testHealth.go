package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// testHealth performs the smoke test for the health endpoint.
func (a *app) testHealth() error {
	url := a.url + "/health"

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

	type healthResponse struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Uptime  string `json:"uptime"`
	}

	var hr healthResponse
	if err := json.Unmarshal(body, &hr); err != nil {
		return fmt.Errorf("can't parse response body: %v", err)
	}

	// just check if values are not empty
	if hr.Name == "" {
		return fmt.Errorf("'version' is empty")
	}

	if hr.Version == "" {
		return fmt.Errorf("'version' is empty")
	}

	if hr.Uptime == "" {
		return fmt.Errorf("'uptime' is empty")
	}

	fmt.Fprintf(a.status, "Success: 'health endpoint operational'\n")
	return nil
}
