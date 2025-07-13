package chatgpt

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tzrikka/thrippy/pkg/client"
)

const (
	modelsURL = "https://api.openai.com/v1/models"
)

// get is a ChatGPT-specific HTTP GET wrapper for [client.HTTPRequest].
// Based on https://platform.openai.com/docs/api-reference/authentication.
func get(ctx context.Context, url, apiKey string) (map[string]any, error) {
	resp, err := client.HTTPRequest(ctx, http.MethodGet, url, "Bearer "+apiKey)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	return m, json.Unmarshal(resp, &m)
}
