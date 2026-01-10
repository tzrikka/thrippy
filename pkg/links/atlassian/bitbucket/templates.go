package bitbucket

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/tzrikka/thrippy/internal/links"
	"github.com/tzrikka/thrippy/pkg/client"
	"github.com/tzrikka/thrippy/pkg/links/atlassian"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

var APITokenTemplate = links.NewTemplate(
	"Bitbucket with a user's static API token",
	[]string{
		"https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/",
		"https://support.atlassian.com/bitbucket-cloud/docs/api-tokens/",
		"https://developer.atlassian.com/cloud/bitbucket/rest/intro/#api-tokens",
		"https://id.atlassian.com/manage-profile/security/api-tokens",
	},
	[]string{"email_manual", "api_token_manual"},
	nil,
	apiTokenChecker,
)

var OAuthTemplate = links.NewTemplate(
	"Bitbucket app using OAuth 2.0 (3LO)",
	[]string{
		"https://developer.atlassian.com/cloud/bitbucket/oauth-2/",
		"https://support.atlassian.com/bitbucket-cloud/docs/use-oauth-on-bitbucket-cloud/",
		"https://developer.atlassian.com/cloud/bitbucket/rest/intro/#bitbucket-oauth-2-0-scopes",
	},
	append(links.OAuthCredFields, "webhook_secret_manual_optional"),
	oauthModifier,
	oauthChecker,
)

var WebhookTemplate = links.NewTemplate(
	"Bitbucket webhook for all types of incoming events",
	[]string{
		"https://support.atlassian.com/bitbucket-cloud/docs/manage-webhooks/",
		"https://support.atlassian.com/bitbucket-cloud/docs/event-payloads/",
	},
	[]string{"webhook_secret_manual"},
	nil,
	nil,
)

// apiTokenChecker checks the given static API token for
// Bitbucket Cloud, and returns metadata about it in JSON format.
func apiTokenChecker(ctx context.Context, m map[string]string, _ *oauth.Config, _ *oauth2.Token) (string, error) {
	// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/#api-user-get
	url := "https://api.bitbucket.org/2.0/user"
	user := new(User)
	if err := atlassian.CurrentUser(ctx, url, m["email"], m["api_token"], user); err != nil {
		return "", fmt.Errorf("failed to get current Bitbucket Cloud user: %w", err)
	}

	return links.EncodeMetadataAsJSON(user)
}

// oauthModifier adjusts the given [oauth.Config]
// for Bitbucket Cloud OAuth 2.0 (3LO) apps, based on
// https://developer.atlassian.com/cloud/bitbucket/oauth-2/.
func oauthModifier(o *oauth.Config) {
	if o.Config.Endpoint.AuthURL == "" {
		o.Config.Endpoint.AuthURL = "https://bitbucket.org/site/oauth2/authorize"
	}

	if o.Config.Endpoint.TokenURL == "" {
		o.Config.Endpoint.TokenURL = "https://bitbucket.org/site/oauth2/access_token"
	}

	o.Config.Scopes = append(o.Config.Scopes, "account")
}

// User is based on: https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/#api-user-get
type User struct {
	AccountID   string `json:"account_id"`
	Type        string `json:"type"`
	CreatedOn   string `json:"created_on"`
	DisplayName string `json:"display_name"`
	Nickname    string `json:"nickname,omitempty"`
	Username    string `json:"username"`
	UUID        string `json:"uuid"`
}

// oauthChecker checks the user associated with the given OAuth token.
func oauthChecker(ctx context.Context, _ map[string]string, _ *oauth.Config, t *oauth2.Token) (string, error) {
	// Scenario 1: user manually sets a webhook secret (before the OAuth flow).
	if t.AccessToken == "" {
		ws, ok := t.Extra("webhook_secret").(string)
		if !ok || ws == "" {
			return "", errors.New("missing both OAuth token and 'webhook_secret'")
		}
		return "", nil
	}

	// Scenario 2: OAuth flow completed, we (should) have an access token.
	// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/#api-user-get
	url := "https://api.bitbucket.org/2.0/user"
	resp, err := client.HTTPRequest(ctx, http.MethodGet, url, "Bearer "+t.AccessToken)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}
