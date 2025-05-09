// Package links defines the configuration details of
// well-known third-party services, as optional templates.
package links

import (
	"slices"

	"github.com/tzrikka/trippy/pkg/oauth"
)

// Type represents the details of a supported link type.
type Type struct {
	Description string
	OAuthFunc   func(*oauth.Config)
	CredsFields []string
}

// Types is a map of all supported link types.
var Types = map[string]Type{
	"generic-oauth": {
		Description: "Generic OAuth 2.0",
		OAuthFunc:   noOAuth,
	},
	"slack-bot-token": {
		Description: "Slack with a static bot token (https://docs.slack.dev/authentication/tokens#bot)",
		OAuthFunc:   noOAuth,
		CredsFields: []string{"bot_token_req", "app_token_opt"},
	},
	"slack-oauth": {
		Description: "Slack with OAuth v2 (https://docs.slack.dev/authentication/installing-with-oauth)",
		OAuthFunc:   slackOAuth,
		CredsFields: oauthCredsFields,
	},
	"slack-oauth-gov": {
		Description: "GovSlack with OAuth v2 (https://docs.slack.dev/govslack)",
		OAuthFunc:   govSlackOAuth,
		CredsFields: oauthCredsFields,
	},
}

// ModifyOAuthConfigByType fills in all the missing OAuth
// configuration details, based on the given link type ID.
// It also normalizes (i.e. sorts and compacts) OAuth scopes.
func ModifyOAuthByType(o *oauth.Config, linkType string) {
	t, ok := Types[linkType]
	if !ok {
		return
	}

	t.OAuthFunc(o)

	slices.Sort(o.Config.Scopes)
	o.Config.Scopes = slices.Compact(o.Config.Scopes)
}

func noOAuth(o *oauth.Config) {
	// Do nothing.
}

// oauthCredsFields is based on: https://pkg.go.dev/golang.org/x/oauth2#Token.
var oauthCredsFields = []string{"access_token", "expiry", "refresh_token", "token_type"}
