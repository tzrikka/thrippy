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

var OAuthTemplate = links.NewTemplate(
	"Jira app using OAuth 2.0 (3LO)",
	[]string{
		"https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/",
		"https://developer.atlassian.com/console/myapps/",
	},
	links.OAuthCredFields,
	oauthModifier,
	atlassian.OAuthChecker,
)

// oauthModifier adjusts the given [oauth.Config]
// for Bitbucket Cloud OAuth 2.0 (3LO) apps, based on
// https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/.
func oauthModifier(o *oauth.Config) {
	if o.Config.Endpoint.AuthURL == "" {
		o.Config.Endpoint.AuthURL = "https://auth.atlassian.com/authorize"
	}

	if o.Config.Endpoint.TokenURL == "" {
		o.Config.Endpoint.TokenURL = "https://auth.atlassian.com/oauth/token"
	}

	o.Config.Scopes = append(o.Config.Scopes, "read:me")
}

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

	return links.EncodeMetadataAsJSON(atlassian.APITokenMetadata{
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
