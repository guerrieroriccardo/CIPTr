package apiclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/guerrieroriccardo/CIPTr/cli/internal/version"
)

// ErrUnauthorized is returned when the API responds with 401.
var ErrUnauthorized = errors.New("unauthorized")

// ErrForbidden is returned when the API responds with 403.
var ErrForbidden = errors.New("forbidden")

// Client wraps net/http to call the CIPTr REST API.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// New creates an API client pointing at the given base URL.
// Example: New("http://localhost:8080/api/v1")
func New(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

// envelope is the standard JSON response shape from the backend.
type envelope struct {
	Data  json.RawMessage `json:"data"`
	Error *string         `json:"error"`
}

// Get performs a GET request and decodes the response into result.
func (c *Client) Get(path string, result any) error {
	return c.do(http.MethodGet, path, nil, result)
}

// Post performs a POST request with a JSON body and decodes the response.
func (c *Client) Post(path string, body any, result any) error {
	return c.do(http.MethodPost, path, body, result)
}

// Put performs a PUT request with a JSON body and decodes the response.
func (c *Client) Put(path string, body any, result any) error {
	return c.do(http.MethodPut, path, body, result)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) error {
	return c.do(http.MethodDelete, path, nil, nil)
}

// do is the shared HTTP helper that handles envelope parsing.
func (c *Client) do(method, path string, body any, result any) error {
	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("X-CLI-Version", version.Version)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	if resp.StatusCode == http.StatusForbidden {
		return ErrForbidden
	}

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("decode response (status %d): %w", resp.StatusCode, err)
	}

	if env.Error != nil {
		return fmt.Errorf("API error: %s", *env.Error)
	}

	if result != nil && env.Data != nil {
		if err := json.Unmarshal(env.Data, result); err != nil {
			return fmt.Errorf("decode data: %w", err)
		}
	}

	return nil
}

// GetRaw performs a GET request and returns the raw response bytes.
// Useful for binary responses (e.g. PDF) that are not wrapped in the JSON envelope.
func (c *Client) GetRaw(path string) ([]byte, error) {
	url := c.BaseURL + path

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("X-CLI-Version", version.Version)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, ErrForbidden
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var env envelope
		if json.Unmarshal(raw, &env) == nil && env.Error != nil {
			return nil, fmt.Errorf("API error: %s", *env.Error)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, path)
	}

	return raw, nil
}

// Login authenticates and returns the JWT token.
func (c *Client) Login(username, password string) (string, error) {
	var result struct {
		Token string `json:"token"`
	}
	err := c.do(http.MethodPost, "/login", map[string]string{
		"username": username,
		"password": password,
	}, &result)
	if err != nil {
		return "", err
	}
	return result.Token, nil
}

// GuestLogin authenticates a guest user (no password) and returns the JWT token.
func (c *Client) GuestLogin(username string) (string, error) {
	var result struct {
		Token string `json:"token"`
	}
	err := c.do(http.MethodPost, "/guest-login", map[string]string{
		"username": username,
	}, &result)
	if err != nil {
		return "", err
	}
	return result.Token, nil
}
