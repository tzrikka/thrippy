package chatgpt

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tzrikka/thrippy/pkg/client"
)

const (
	modelsURL = "https://api.openai.com/v1/models"
	mimeType  = "application/json"
)

// get is a ChatGPT-specific HTTP GET wrapper for [client.HTTPRequest].
func get(ctx context.Context, url, apiKey string) (map[string]any, error) {
	// https://platform.openai.com/docs/api-reference/authentication
	resp, err := client.HTTPRequest(ctx, http.MethodGet, url, mimeType, apiKey)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	return m, json.Unmarshal(resp, &m)
}
