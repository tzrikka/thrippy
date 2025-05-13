package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	timeout = time.Second * 3
	maxSize = 1 << 20 // 2^20 bytes = 1 MiB
)

func get(ctx context.Context, url, botToken string, jsonResp any) error {
	resp, err := httpRequest(ctx, http.MethodGet, url, botToken)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp, jsonResp)
}

func post(ctx context.Context, url, botToken string, jsonResp any) error {
	resp, err := httpRequest(ctx, http.MethodPost, url, botToken)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp, jsonResp)
}

func httpRequest(ctx context.Context, method, url, token string) ([]byte, error) {
	// Construct the request.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to construct HTTP request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send the request.
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
