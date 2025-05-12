package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"golang.org/x/oauth2"

	"github.com/tzrikka/trippy/pkg/oauth"
)

// OAuthModifier adjusts the given [oauth.Config] for Slack apps.
func OAuthModifier(o *oauth.Config) {
	// https://docs.slack.dev/authentication/installing-with-oauth
	if o.Config.Endpoint.AuthURL == "" {
		o.Config.Endpoint.AuthURL = "https://slack.com/oauth/v2/authorize"
	}
	// https://docs.slack.dev/reference/methods/oauth.v2.access
	if o.Config.Endpoint.TokenURL == "" {
		o.Config.Endpoint.TokenURL = "https://slack.com/api/oauth.v2.access"
	}
	// https://docs.slack.dev/authentication/installing-with-oauth
	if o.Config.Endpoint.AuthStyle == oauth2.AuthStyleAutoDetect {
		o.Config.Endpoint.AuthStyle = oauth2.AuthStyleInHeader
	}

	// https://docs.slack.dev/reference/scopes/users.read
	// (required by https://docs.slack.dev/reference/methods/bots.info).
	o.Config.Scopes = append(o.Config.Scopes, "users:read")
}

// GovOAuthModifier adjusts the given [oauth.Config] for [GovSlack] apps.
//
// [GovSlack]: https://docs.slack.dev/govslack
func GovOAuthModifier(o *oauth.Config) {
	// https://docs.slack.dev/authentication/installing-with-oauth
	if o.Config.Endpoint.AuthURL == "" {
		o.Config.Endpoint.AuthURL = "https://slack-gov.com/oauth/v2/authorize"
	}
	// https://docs.slack.dev/reference/methods/oauth.v2.access
	if o.Config.Endpoint.TokenURL == "" {
		o.Config.Endpoint.TokenURL = "https://slack-gov.com/api/oauth.v2.access"
	}
	// https://docs.slack.dev/authentication/installing-with-oauth
	if o.Config.Endpoint.AuthStyle == oauth2.AuthStyleAutoDetect {
		o.Config.Endpoint.AuthStyle = oauth2.AuthStyleInHeader
	}

	// https://docs.slack.dev/reference/scopes/users.read
	// (required by https://docs.slack.dev/reference/methods/bots.info).
	o.Config.Scopes = append(o.Config.Scopes, "users:read")
}

// BotTokenChecker checks the given static bot token for Slack.
// Based on https://docs.slack.dev/reference/methods/auth.test
// and https://docs.slack.dev/reference/methods/bots.info
func BotTokenChecker(ctx context.Context, m map[string]string, _ *oauth2.Token) (string, error) {
	return genericChecker(ctx, m["bot_token"], "https://slack.com")
}

// BotTokenChecker checks the given OAuth token for Slack.
// Based on https://docs.slack.dev/reference/methods/auth.test
// and https://docs.slack.dev/reference/methods/bots.info
func OAuthChecker(ctx context.Context, _ map[string]string, t *oauth2.Token) (string, error) {
	return genericChecker(ctx, t.AccessToken, "https://slack.com")
}

// BotTokenChecker checks the given OAuth token for GovSlack.
// Based on https://docs.slack.dev/reference/methods/auth.test
// and https://docs.slack.dev/reference/methods/bots.info
func GovOAuthChecker(ctx context.Context, _ map[string]string, t *oauth2.Token) (string, error) {
	return genericChecker(ctx, t.AccessToken, "https://slack-gov.com")
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
