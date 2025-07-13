package jira

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/tzrikka/thrippy/internal/links"
	"github.com/tzrikka/thrippy/pkg/links/atlassian"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

var APITokenTemplate = links.NewTemplate(
	"Jira with a user's static API token",
	[]string{
		"https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/",
		"https://id.atlassian.com/manage-profile/security/api-tokens",
	},
	[]string{"base_url_manual", "email_manual", "api_token_manual"},
	nil,
	apiTokenChecker,
)

// apiTokenChecker checks the given static API token for
// Jira Cloud, and returns metadata about it in JSON format.
func apiTokenChecker(ctx context.Context, m map[string]string, _ *oauth.Config, _ *oauth2.Token) (string, error) {
	baseURL, err := links.NormalizeURL(m["base_url"])
	if err != nil {
		return "", err
	}

	// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-myself/#api-rest-api-3-myself-get
	url := baseURL + "/rest/api/3/myself"
	user := &User{}
	if err := atlassian.CurrentUser(ctx, url, m["email"], m["api_token"], user); err != nil {
		return "", fmt.Errorf("error in getting current Jira Cloud user: %w", err)
	}

	return links.EncodeMetadataAsJSON(atlassian.Metadata{
		AccountID:   user.AccountID,
		AccountType: user.AccountType,
		Email:       user.EmailAddress,
		Name:        user.DisplayName,
		TimeZone:    user.TimeZone,
	})
}

// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-myself/#api-rest-api-3-myself-get
type User struct {
	AccountID    string `json:"accountId"`
	AccountType  string `json:"accountType"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	TimeZone     string `json:"timeZone"`
}
