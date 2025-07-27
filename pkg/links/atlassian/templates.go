package atlassian

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/tzrikka/thrippy/internal/links"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

// APITokenMetadata is reused by multiple Atlassian checker functions.
type APITokenMetadata struct {
	AccountID   string `json:"account_id"`
	AccountType string `json:"account_type"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	TimeZone    string `json:"time_zone,omitempty"`
}

// OAuthMetadata is reused by multiple Atlassian checker functions.
type OAuthMetadata struct {
	CloudID string `json:"cloud_id"`
	Name    string `json:"name"`
	URL     string `json:"url"`
}

// OAuthChecker checks the given OAuth token, and returns metadata for
// API calls in the corresponding Atlassian Cloud workspace in JSON format.
func OAuthChecker(ctx context.Context, _ map[string]string, _ *oauth.Config, t *oauth2.Token) (string, error) {
	res, err := AccessibleResources(ctx, t.AccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to get Atlassian Cloud resource: %w", err)
	}

	return links.EncodeMetadataAsJSON(OAuthMetadata{
		CloudID: res.ID,
		Name:    res.Name,
		URL:     res.URL,
	})
}
