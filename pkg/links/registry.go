// Package links defines the authentication details of well-known third-party
// services, as templates for link creation, with special logic to check the
// usability of user-provided credentials and return metadata about them.
package links

import (
	"github.com/tzrikka/thrippy/pkg/links/chatgpt"
	"github.com/tzrikka/thrippy/pkg/links/claude"
	"github.com/tzrikka/thrippy/pkg/links/gemini"
	"github.com/tzrikka/thrippy/pkg/links/github"
	"github.com/tzrikka/thrippy/pkg/links/google"
	"github.com/tzrikka/thrippy/pkg/links/slack"
	"github.com/tzrikka/thrippy/pkg/links/templates"
)

// Templates is a map of all the link templates that Thrippy recognizes and supports.
var Templates = map[string]templates.Template{
	"chatgpt": chatgpt.Template,
	"claude":  claude.Template,
	"gemini":  gemini.Template,
	"generic-oauth": templates.New(
		"Generic link", nil, nil, nil, nil,
	),
	"github-app-jwt":         github.AppJWTTemplate,
	"github-app-user":        github.AppUserTemplate,
	"github-user-pat":        github.UserPATTemplate,
	"github-webhook":         github.WebhookTemplate,
	"google-service-account": google.ServiceAccountTemplate,
	"google-user-oauth":      google.UserOAuthTemplate,
	"slack-bot-token":        slack.BotTokenTemplate,
	"slack-oauth":            slack.OAuthTemplate,
	"slack-oauth-gov":        slack.OAuthGovTemplate,
	"slack-socket-mode":      slack.SocketModeTemplate,
}
