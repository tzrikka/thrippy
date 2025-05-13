package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	maxSize = 1 << 20 // 2^20 bytes = 1 MiB
)

func HTTPRequest(ctx context.Context, httpMethod, url, mimeType, token string) ([]byte, error) {
	// Construct the request.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, httpMethod, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to construct HTTP request: %w", err)
	}

	req.Header.Set("Accept", mimeType)
	req.Header.Set("Authorization", "Bearer "+token)

	// Send the request to the server.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read and return the response body.
	body, err := io.ReadAll(http.MaxBytesReader(nil, resp.Body, maxSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response's body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		s := fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		if len(body) > 0 {
			s = fmt.Sprintf("%s: %s", s, string(body))
		}
		return nil, errors.New(s)
	}

	return body, nil
}
