package claude

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tzrikka/thrippy/pkg/client"
)

const (
	modelsURL = "https://api.anthropic.com/v1/models"
	mimeType  = "application/json"
	version   = "anthropic-version: 2023-06-01"
	// X-api-key.
)

// get is an Anthropic-specific HTTP GET wrapper for [client.HTTPRequest].
func get(ctx context.Context, url, apiKey string) (map[string]any, error) {
	// https://docs.anthropic.com/en/api/overview
	// https://docs.anthropic.com/en/api/versioning
	// https://docs.anthropic.com/en/api/models-list
	headers := map[string]string{
		"anthropic-version": "2023-06-01",
		"x-api-key":         apiKey,
	}
	resp, err := client.HTTPRequest(ctx, http.MethodGet, url, mimeType, "", headers)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	return m, json.Unmarshal(resp, &m)
}
