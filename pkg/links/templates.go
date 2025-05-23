// Package links defines the authentication details of well-known third-party
// services, as templates for link creation, with special logic to check the
// usability of user-provided credentials and return metadata about them.
package links

import (
	"context"
	"slices"

	"golang.org/x/oauth2"

	"github.com/tzrikka/thrippy/pkg/links/github"
	"github.com/tzrikka/thrippy/pkg/links/google"
	"github.com/tzrikka/thrippy/pkg/links/slack"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

// Template defines the authentication details
// of a well-known third-party service.
type Template struct {
	description string
	links       []string
	credFields  []string
	oauthFunc   func(*oauth.Config)
	checkerFunc func(context.Context, map[string]string, *oauth.Config, *oauth2.Token) (string, error)
}

func (t Template) Description() string {
	return t.description
}

// CredsFields returns a copy of all the expected field names
// in the link's credentials, based on the link's template.
func (t Template) CredFields() []string {
	if len(t.credFields) == 0 {
		return nil
	}
	return slices.Clone(t.credFields)
}

// Check checks the usability of the provided credentials (either the map or
// the token), and returns JSON-serialized metadata about them for storage.
func (t Template) Check(ctx context.Context, m map[string]string, oc *oauth.Config, ot *oauth2.Token) (string, error) {
	if t.checkerFunc == nil {
		return "", nil
	}
	return t.checkerFunc(ctx, m, oc, ot)
}

// OAuthCredFields is a standard based on [oauth2.Token].
var OAuthCredFields = []string{"access_token", "expiry", "refresh_token", "token_type"}

// Templates is a map of all the link templates that Thrippy recognizes and supports.
var Templates = map[string]Template{
	"chatgpt": {
		description: "ChatGPT using a static API key",
		links: []string{
			"https://platform.openai.com/docs/api-reference/authentication",
			"https://platform.openai.com/api-keys",
		},
		credFields: []string{"api_key"},
	},
	"claude": {
		description: "Claude using a static API key",
		links: []string{
			"https://docs.anthropic.com/en/api/overview",
			"https://console.anthropic.com/settings/keys",
		},
		credFields: []string{"api_key"},
	},
	"gemini": {
		description: "Gemini with a static API key",
		links: []string{
			"https://ai.google.dev/gemini-api/docs/api-key",
			"https://aistudio.google.com/app/apikey",
			"https://console.cloud.google.com/apis/credentials",
		},
		credFields: []string{"api_key"},
	},
	"generic-oauth": {
		description: "Generic link",
	},
	"github-app-jwt": {
		description: "GitHub app installation using JWTs based on static credentials",
		links: []string{
			"https://docs.github.com/en/apps/using-github-apps/about-using-github-apps",
			"https://github.com/settings/apps",
		},
		credFields: []string{
			"client_id", "private_key", // Must be entered manually.
			"api_base_url", "install_id", // Added automatically by Thrippy.
		},
		oauthFunc:   github.AppInstallModifier,
		checkerFunc: github.JWTChecker,
	},
	"github-app-user": {
		description: "GitHub app authorization to act on behalf of a user",
		links: []string{
			"https://docs.github.com/en/apps/using-github-apps/authorizing-github-apps",
			"https://github.com/settings/apps",
		},
		credFields:  append([]string{"base_url_optional"}, OAuthCredFields...),
		oauthFunc:   github.AppAuthzModifier,
		checkerFunc: github.UserChecker,
	},
	"github-user-pat": {
		description: "GitHub with a user's static Personal Access Token (PAT)",
		links: []string{
			"https://docs.github.com/en/rest/authentication/authenticating-to-the-rest-api?apiVersion=2022-11-28#authenticating-with-a-personal-access-token",
			"https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens",
			"https://github.com/settings/personal-access-tokens",
			"https://github.com/settings/tokens",
		},
		credFields:  []string{"base_url_optional", "pat"},
		checkerFunc: github.UserChecker,
	},
	"google-user-oauth": {
		description: "Google APIs using OAuth 2.0 to act on behalf of a user",
		links: []string{
			"https://developers.google.com/workspace/guides/get-started",
			"https://console.cloud.google.com/auth/overview",
		},
		credFields:  OAuthCredFields,
		oauthFunc:   google.OAuthModifier,
		checkerFunc: google.UserTokenChecker,
	},
	"google-service-account": {
		description: "Google APIs using a static GCP service account key",
		links: []string{
			"https://cloud.google.com/iam/docs/service-account-overview",
			"https://developers.google.com/identity/protocols/oauth2/service-account",
			"https://console.cloud.google.com/iam-admin/serviceaccounts",
		},
		credFields:  []string{"key"},
		checkerFunc: google.ServiceKeyChecker,
	},
	"slack-bot-token": {
		description: "Slack app using a static bot token",
		links: []string{
			"https://docs.slack.dev/authentication/tokens#bot",
			"https://api.slack.com/apps",
		},
		credFields:  []string{"bot_token", "app_token_optional"},
		checkerFunc: slack.BotTokenChecker,
	},
	"slack-oauth": {
		description: "Slack app using OAuth v2",
		links: []string{
			"https://docs.slack.dev/authentication/installing-with-oauth",
			"https://api.slack.com/apps",
		},
		credFields:  OAuthCredFields,
		oauthFunc:   slack.OAuthModifier(slack.DefaultBaseURL),
		checkerFunc: slack.OAuthChecker,
	},
	"slack-oauth-gov": {
		description: "GovSlack app using OAuth v2",
		links: []string{
			"https://docs.slack.dev/authentication/installing-with-oauth",
			"https://docs.slack.dev/govslack",
		},
		credFields:  OAuthCredFields,
		oauthFunc:   slack.OAuthModifier(slack.GovBaseURL),
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
