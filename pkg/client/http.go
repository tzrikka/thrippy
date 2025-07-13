package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	maxSize = 10 << 20 // 10 MiB.
)

func HTTPRequest(ctx context.Context, httpMethod, url, authToken string) ([]byte, error) {
	return HTTPRequestWithHeaders(ctx, httpMethod, url, authToken, map[string]string{
		"Accept": "application/json",
	})
}

func HTTPRequestWithHeaders(ctx context.Context, httpMethod, url, authToken string, headers map[string]string) ([]byte, error) {
	req, cancel, err := constructRequest(ctx, httpMethod, url, authToken, headers)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		msg := resp.Status
		if len(body) > 0 {
			msg = fmt.Sprintf("%s: %s", msg, string(body))
		}
		return nil, errors.New(msg)
	}

	return body, nil
}

func constructRequest(ctx context.Context, method, url, token string, headers map[string]string) (*http.Request, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	req, err := http.NewRequestWithContext(ctx, method, url, http.NoBody)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("failed to construct HTTP request: %w", err)
	}

	if strings.HasPrefix(token, "Basic ") {
		token = strings.TrimPrefix(token, "Basic ")
		if user, pass, found := strings.Cut(token, ":"); found {
			req.SetBasicAuth(user, pass)
		}
	} else if token != "" {
		req.Header.Set("Authorization", token)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, cancel, nil
}
