// Package links defines the authentication details of well-known
// third-party services, as templates for link creation, and special
// logic per service to check the usability of private credentials,
// and return serialized metadata about them for storage.
package links

import (
	"context"
	"slices"

	"golang.org/x/oauth2"

	"github.com/tzrikka/trippy/pkg/links/slack"
	"github.com/tzrikka/trippy/pkg/oauth"
)

// Template defines the authentication details of a well-known third-party service.
type Template struct {
	description string
	credFields  []string
	oauthFunc   func(*oauth.Config)
	checkerFunc func(context.Context, map[string]string, *oauth2.Token) (string, error)
}

func (t Template) Description() string {
	return t.description
}

func (t Template) CredFields() []string {
	if len(t.credFields) == 0 {
		return nil
	}
	return slices.Clone(t.credFields)
}

// Check checks the usability of the provided credentials (either the map or
// the token), and returns JSON-serialized metadata about them for storage.
func (t Template) Check(ctx context.Context, m map[string]string, ot *oauth2.Token) (string, error) {
	if t.checkerFunc == nil {
		return "", nil
	}
	return t.checkerFunc(ctx, m, ot)
}

// oauthCredFields is based on [oauth2.Token].
var oauthCredFields = []string{"access_token", "expiry", "refresh_token", "token_type"}

// Templates is a map of all the link templates that Trippy recognizes and supports.
var Templates = map[string]Template{
	"generic": {
		description: "Generic link",
	},
	"slack-bot-token": {
		description: "Slack with a static bot token (https://docs.slack.dev/authentication/tokens#bot)",
		credFields:  []string{"bot_token", "app_token_optional"},
		checkerFunc: slack.BotTokenChecker,
	},
	"slack-oauth": {
		description: "Slack with OAuth v2 (https://docs.slack.dev/authentication/installing-with-oauth)",
		credFields:  oauthCredFields,
		oauthFunc:   slack.OAuthModifier,
		checkerFunc: slack.OAuthChecker,
	},
	"slack-oauth-gov": {
		description: "GovSlack with OAuth v2 (https://docs.slack.dev/govslack)",
		credFields:  oauthCredFields,
		oauthFunc:   slack.GovOAuthModifier,
		checkerFunc: slack.GovOAuthChecker,
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

	if t.oauthFunc != nil {
		t.oauthFunc(o)
	}

	slices.Sort(o.Config.Scopes)
	o.Config.Scopes = slices.Compact(o.Config.Scopes)
}
