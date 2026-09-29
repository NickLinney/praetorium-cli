package coordinator

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	Endpoint string
	Timeout  time.Duration
}

func (client Client) Execute(ctx context.Context, body []byte) ([]byte, int, error) {
	endpoint, err := executeURL(client.Endpoint)
	if err != nil {
		return nil, 0, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, 0, fmt.Errorf("create coordinator request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	timeout := client.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	response, err := (&http.Client{Timeout: timeout}).Do(request)
	if err != nil {
		return nil, 0, fmt.Errorf("coordinator request failed: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	if err != nil {
		return nil, response.StatusCode, fmt.Errorf("read coordinator response: %w", err)
	}
	return responseBody, response.StatusCode, nil
}

func executeURL(endpoint string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid coordinator endpoint %q", endpoint)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("coordinator endpoint must use http or https, got %q", parsed.Scheme)
	}
	if strings.HasSuffix(strings.TrimRight(parsed.Path, "/"), "/execute") {
		return parsed.String(), nil
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/execute"
	return parsed.String(), nil
}
