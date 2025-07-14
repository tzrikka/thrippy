package atlassian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/tzrikka/thrippy/pkg/client"
)

// CloudResource represents the details or a Confluence or Jira workspace
// for the purpose of checking OAuth tokens and making API calls.
type CloudResource struct {
	ID     string   `json:"id"`
	URL    string   `json:"url"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

// AccessibleResources is reused by multiple Atlassian checker functions
// for OAuth-based links. It is an Atlassian-specific HTTP GET wrapper
// for [client.HTTPRequest].
//
// It is based on:
//   - https://developer.atlassian.com/cloud/confluence/oauth-2-3lo-apps/#3--make-calls-to-the-api-using-the-access-token
//   - https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/#3--make-calls-to-the-api-using-the-access-token
func AccessibleResources(ctx context.Context, accessToken string) (*CloudResource, error) {
	url := "https://api.atlassian.com/oauth/token/accessible-resources"
	resp, err := client.HTTPRequest(ctx, http.MethodGet, url, "Bearer "+accessToken)
	if err != nil {
		return nil, err
	}

	jsonResp := []CloudResource{}
	if err := json.Unmarshal(resp, &jsonResp); err != nil {
		return nil, err
	}

	switch len(jsonResp) {
	case 0:
		return nil, errors.New("valid OAuth token with no Atlassian accessible resources")
	case 1:
		return &jsonResp[0], nil
	default:
		return nil, errors.New("multiple Atlassian accessible resources found")
	}
}

// CurrentUser is reused by multiple Atlassian checker functions
// for token-based links. It is an Atlassian-specific HTTP GET
// wrapper for [client.HTTPRequest].
//
// It is based on:
//   - https://developer.atlassian.com/cloud/bitbucket/rest/intro/#api-tokens
//   - https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/#api-user-get
//   - https://developer.atlassian.com/cloud/confluence/basic-auth-for-rest-apis/
//   - https://developer.atlassian.com/cloud/confluence/rest/v1/api-group-users/#api-wiki-rest-api-user-current-get
//   - https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/#ad-hoc-api-calls
//   - https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-myself/#api-rest-api-3-myself-get
func CurrentUser(ctx context.Context, url, email, apiToken string, jsonResp any) error {
	if email == "" {
		return errors.New("missing email address")
	}
	if apiToken == "" {
		return errors.New("missing API token")
	}

	auth := fmt.Sprintf("Basic %s:%s", email, apiToken)
	resp, err := client.HTTPRequest(ctx, http.MethodGet, url, auth)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp, jsonResp)
}
