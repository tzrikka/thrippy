package claude

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tzrikka/thrippy/pkg/client"
)

const (
	modelsURL = "https://api.anthropic.com/v1/models"
	version   = "anthropic-version: 2023-06-01"
	// X-api-key.
)

// get is an Anthropic-specific HTTP GET wrapper for [client.HTTPRequest].
// Based on https://docs.anthropic.com/en/api/overview and
// https://docs.anthropic.com/en/api/versioning and
// https://docs.anthropic.com/en/api/models-list.
func get(ctx context.Context, url, apiKey string) (map[string]any, error) {
	resp, err := client.HTTPRequestWithHeaders(ctx, http.MethodGet, url, "", map[string]string{
		"anthropic-version": "2023-06-01",
		"x-api-key":         apiKey,
	})
	if err != nil {
		return nil, err
	}

	var m map[string]any
	return m, json.Unmarshal(resp, &m)
}
