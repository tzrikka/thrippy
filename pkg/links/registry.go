// Package links defines the authentication details of well-known third-party
// services, as templates for link creation, with special logic to check the
// usability of user-provided credentials and return metadata about them.
package links

import (
	"github.com/tzrikka/thrippy/internal/links"
	"github.com/tzrikka/thrippy/pkg/links/atlassian/bitbucket"
	"github.com/tzrikka/thrippy/pkg/links/atlassian/confluence"
	"github.com/tzrikka/thrippy/pkg/links/atlassian/jira"
	"github.com/tzrikka/thrippy/pkg/links/chatgpt"
	"github.com/tzrikka/thrippy/pkg/links/claude"
	"github.com/tzrikka/thrippy/pkg/links/gemini"
	"github.com/tzrikka/thrippy/pkg/links/github"
	"github.com/tzrikka/thrippy/pkg/links/google"
	"github.com/tzrikka/thrippy/pkg/links/slack"
)

// Templates is a map of all the link templates that Thrippy recognizes and supports.
var Templates = map[string]links.Template{
	"bitbucket-app-oauth":   bitbucket.OAuthTemplate,
	"bitbucket-user-token":  bitbucket.APITokenTemplate,
	"chatgpt":               chatgpt.Template,
	"claude":                claude.Template,
	"confluence-app-oauth":  confluence.OAuthTemplate,
	"confluence-user-token": confluence.APITokenTemplate,
	"gemini":                gemini.Template,
	"generic-oauth": links.NewTemplate(
		"Generic link", nil, nil, nil, nil,
	),
	"github-app-jwt":         github.AppJWTTemplate,
	"github-app-user":        github.AppUserTemplate,
	"github-user-pat":        github.UserPATTemplate,
	"github-webhook":         github.WebhookTemplate,
	"google-service-account": google.ServiceAccountTemplate,
	"google-user-oauth":      google.UserOAuthTemplate,
	"jira-app-oauth":         jira.OAuthTemplate,
	"jira-user-token":        jira.APITokenTemplate,
	"slack-bot-token":        slack.BotTokenTemplate,
	"slack-oauth":            slack.OAuthTemplate,
	"slack-oauth-gov":        slack.OAuthGovTemplate,
	"slack-socket-mode":      slack.SocketModeTemplate,
}
