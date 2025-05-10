// Package links defines the authentication details of well-known
// third-party services, as templates for link creation, and special
// logic per service to check the usability of private credentials.
package links

import (
	"slices"

	_ "golang.org/x/oauth2"

	"github.com/tzrikka/trippy/pkg/oauth"
)

// Template defines the authentication details of a well-known third-party service.
type Template struct {
	Description string
	OAuthFunc   func(*oauth.Config)
	CredsFields []string
}

// Templates is a map of all the link templates that Trippy recognizes and supports.
var Templates = map[string]Template{
	"generic": {
		Description: "Generic link",
		OAuthFunc:   noOAuth,
	},
	"slack-bot-token": {
		Description: "Slack with a static bot token (https://docs.slack.dev/authentication/tokens#bot)",
		OAuthFunc:   noOAuth,
		CredsFields: []string{"bot_token", "optional_app_token"},
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

// ModifyOAuthByTemplate fills in all the missing OAuth
// configuration details, based on the given link type ID.
// It also normalizes (i.e. sorts and compacts) OAuth scopes.
func ModifyOAuthByTemplate(o *oauth.Config, template string) {
	t, ok := Templates[template]
	if !ok {
		return
	}

	if o == nil {
		return
	}

	t.OAuthFunc(o)

	slices.Sort(o.Config.Scopes)
	o.Config.Scopes = slices.Compact(o.Config.Scopes)
}

func noOAuth(o *oauth.Config) {
	// Do nothing.
}

// oauthCredsFields is based on [oauth2.Token].
var oauthCredsFields = []string{"access_token", "expiry", "refresh_token", "token_type"}
