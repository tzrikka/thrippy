package links

import (
	"golang.org/x/oauth2"

	"github.com/tzrikka/trippy/pkg/oauth"
)

// SlackOAuth adjusts the given [oauth.Config] for Slack (https://slack.com/).
// Based on https://docs.slack.dev/authentication/installing-with-oauth
// and https://docs.slack.dev/reference/methods/oauth.v2.access
func SlackOAuth(o *oauth.Config) {
	if o.Config.Endpoint.AuthURL == "" {
		o.Config.Endpoint.AuthURL = "https://slack.com/oauth/v2/authorize"
	}
	if o.Config.Endpoint.TokenURL == "" {
		o.Config.Endpoint.TokenURL = "https://slack.com/api/oauth.v2.access"
	}
	if o.Config.Endpoint.AuthStyle == oauth2.AuthStyleAutoDetect {
		o.Config.Endpoint.AuthStyle = oauth2.AuthStyleInHeader
	}

	// https://docs.slack.dev/reference/scopes/users.read
	o.Config.Scopes = append(o.Config.Scopes, "users:read")
}

// GovSlackOAuth adjusts the given [oauth.Config] for GovSlack (https://slack-gov.com/).
// Based on https://docs.slack.dev/authentication/installing-with-oauth
// and https://docs.slack.dev/reference/methods/oauth.v2.access
// and https://docs.slack.dev/govslack
func GovSlackOAuth(o *oauth.Config) {
	if o.Config.Endpoint.AuthURL == "" {
		o.Config.Endpoint.AuthURL = "https://slack-gov.com/oauth/v2/authorize"
	}
	if o.Config.Endpoint.TokenURL == "" {
		o.Config.Endpoint.TokenURL = "https://slack-gov.com/api/oauth.v2.access"
	}
	if o.Config.Endpoint.AuthStyle == oauth2.AuthStyleAutoDetect {
		o.Config.Endpoint.AuthStyle = oauth2.AuthStyleInHeader
	}

	// https://docs.slack.dev/reference/scopes/users.read
	o.Config.Scopes = append(o.Config.Scopes, "users:read")
}
