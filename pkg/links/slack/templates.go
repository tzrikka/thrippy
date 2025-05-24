package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"golang.org/x/oauth2"

	"github.com/tzrikka/thrippy/pkg/links/templates"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

const (
	defaultBaseURL = "https://slack.com"
	govBaseURL     = "https://slack-gov.com" // https://docs.slack.dev/govslack
)

var (
	BotTokenTemplate = templates.New(
		"Slack app using a static bot token",
		[]string{
			"https://docs.slack.dev/authentication/tokens#bot",
			"https://api.slack.com/apps",
		},
		[]string{"bot_token", "app_token_optional"},
		nil,
		botTokenChecker,
	)

	OAuthTemplate = templates.New(
		"Slack app using OAuth v2",
		[]string{
			"https://docs.slack.dev/authentication/installing-with-oauth",
			"https://api.slack.com/apps",
		},
		templates.OAuthCredFields,
		oauthModifier(defaultBaseURL),
		oauthChecker,
	)

	OAuthGovTemplate = templates.New(
		"GovSlack app using OAuth v2",
		[]string{
			"https://docs.slack.dev/authentication/installing-with-oauth",
			"https://docs.slack.dev/govslack",
		},
		templates.OAuthCredFields,
		oauthModifier(govBaseURL),
		govOAuthChecker,
	)
)

// oauthModifier returns a function that adjusts an [oauth.Config] for Slack
// apps, based on the given base URL ([defaultBaseURL] or [govBaseURL]).
func oauthModifier(baseURL string) func(*oauth.Config) {
	return func(o *oauth.Config) {
		// https://docs.slack.dev/authentication/installing-with-oauth
		if o.Config.Endpoint.AuthURL == "" {
			o.Config.Endpoint.AuthURL = baseURL + "/oauth/v2/authorize"
		}

		// https://docs.slack.dev/reference/methods/oauth.v2.access
		if o.Config.Endpoint.TokenURL == "" {
			o.Config.Endpoint.TokenURL = baseURL + "/api/oauth.v2.access"
		}

		// https://docs.slack.dev/authentication/installing-with-oauth
		if o.Config.Endpoint.AuthStyle == oauth2.AuthStyleAutoDetect {
			o.Config.Endpoint.AuthStyle = oauth2.AuthStyleInHeader
		}

		// https://docs.slack.dev/reference/scopes/users.read
		// (required by https://docs.slack.dev/reference/methods/bots.info).
		o.Config.Scopes = append(o.Config.Scopes, "users:read")
	}
}

// botTokenChecker checks the given static bot token for
// Slack, and returns metadata about it in JSON format.
func botTokenChecker(ctx context.Context, m map[string]string, _ *oauth.Config, _ *oauth2.Token) (string, error) {
	return genericChecker(ctx, m["bot_token"], defaultBaseURL)
}

// oauthChecker checks the given static bot token for
// Slack, and returns metadata about it in JSON format.
func oauthChecker(ctx context.Context, _ map[string]string, _ *oauth.Config, t *oauth2.Token) (string, error) {
	return genericChecker(ctx, t.AccessToken, defaultBaseURL)
}

// govOAuthChecker checks the given static bot token for
// GovSlack, and returns metadata about it in JSON format.
func govOAuthChecker(ctx context.Context, _ map[string]string, _ *oauth.Config, t *oauth2.Token) (string, error) {
	return genericChecker(ctx, t.AccessToken, govBaseURL)
}

func genericChecker(ctx context.Context, botToken, baseURL string) (string, error) {
	if botToken == "" {
		return "", errors.New("missing bot token")
	}

	auth, err := authTest(ctx, baseURL, botToken)
	if err != nil {
		return "", fmt.Errorf("auth test error: %w", err)
	}

	bot, err := botsInfo(ctx, baseURL, botToken, auth)
	if err != nil {
		return "", fmt.Errorf("bot info error: %w", err)
	}

	j, err := json.Marshal(metadata{
		AppID:        bot.AppID,
		BotID:        bot.ID,
		BotName:      bot.Name,
		BotUpdated:   time.Unix(int64(bot.Updated), 0).UTC().Format(time.RFC3339),
		EnterpriseID: auth.EnterpriseID,
		TeamID:       auth.TeamID,
		TeamName:     auth.Team,
		URL:          auth.URL,
		UserID:       auth.UserID,
		UserName:     auth.User,
	})
	if err != nil {
		return "", err
	}
	return string(j), nil
}

type metadata struct {
	AppID        string `json:"app_id"`
	BotID        string `json:"bot_id"`
	BotName      string `json:"bot_name"`
	BotUpdated   string `json:"bot_updated"`
	EnterpriseID string `json:"enterprise_id,omitempty"`
	TeamID       string `json:"team_id"`
	TeamName     string `json:"team_name"`
	URL          string `json:"url"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
}
