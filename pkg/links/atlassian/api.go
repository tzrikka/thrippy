package atlassian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/tzrikka/thrippy/pkg/client"
)

// CurrentUser is reused by multiple Atlassian checker functions.
// It is an Atlassian-specific HTTP GET wrapper for [client.HTTPRequest],
// based on:
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

// Metadata is reused by multiple Atlassian checker functions.
type Metadata struct {
	AccountID   string `json:"account_id"`
	AccountType string `json:"account_type"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	TimeZone    string `json:"time_zone,omitempty"`
}
