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
	},
	"slack-oauth": {
		Description: "Slack with OAuth v2 (https://docs.slack.dev/authentication/installing-with-oauth)",
		OAuthFunc:   SlackOAuth,
	},
	"slack-oauth-gov": {
		Description: "GovSlack with OAuth v2 (https://docs.slack.dev/govslack)",
		OAuthFunc:   GovSlackOAuth,
	},
}

func noOAuth(o *oauth.Config) {
	// Do nothing.
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
