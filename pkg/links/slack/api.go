package slack

import (
	"context"
	"errors"
	"fmt"
)

// https://docs.slack.dev/reference/methods/auth.test
type authTestResponse struct {
	slackResponse

	URL                 string `json:"url"`
	Team                string `json:"team"`
	User                string `json:"user"`
	TeamID              string `json:"team_id"`
	UserID              string `json:"user_id"`
	BotID               string `json:"bot_id"`
	ExpiresIn           int    `json:"expires_in"`
	EnterpriseID        string `json:"enterprise_id,omitempty"`
	IsEnterpriseInstall bool   `json:"is_enterprise_install"`
}

// https://docs.slack.dev/reference/methods/bots.info
type botsInfoResponse struct {
	slackResponse

	Bot bot `json:"bot"`
}

// https://docs.slack.dev/reference/methods/bots.info
type bot struct {
	ID      string            `json:"id"`
	Deleted bool              `json:"deleted"`
	Name    string            `json:"name"`
	Updated int               `json:"updated"`
	AppID   string            `json:"app_id"`
	UserID  string            `json:"user_id"`
	Icons   map[string]string `json:"icons"`
}

type slackResponse struct {
	OK               bool              `json:"ok"`
	Warning          string            `json:"warning,omitempty"`
	Error            string            `json:"error,omitempty"`
	ResponseMetadata *responseMetadata `json:"response_metadata,omitempty"`
}

type responseMetadata struct {
	Messages   []string `json:"messages,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
	NextCursor string   `json:"next_cursor,omitempty"`
}

// authTest checks the caller's authentication & identity.
// Based on https://docs.slack.dev/reference/methods/auth.test (no scopes required).
func authTest(ctx context.Context, baseURL, botToken string) (*authTestResponse, error) {
	url := baseURL + "/api/auth.test"

	resp := &authTestResponse{}
	if err := post(ctx, url, botToken, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp, nil
}

// botsInfo gets information about a bot user.
// Based on https://docs.slack.dev/reference/methods/bots.info
// (required scope: https://docs.slack.dev/reference/scopes/users.read).
func botsInfo(ctx context.Context, baseURL, botToken string, authTest *authTestResponse) (*bot, error) {
	url := fmt.Sprintf("%s/api/bots.info?bot=%s", baseURL, authTest.BotID)
	if authTest.TeamID != "" {
		url = fmt.Sprintf("%s&team_id=%s", url, authTest.TeamID)
	}
	if authTest.IsEnterpriseInstall {
		url = fmt.Sprintf("%s&enterprise_id=%s", url, authTest.EnterpriseID)
	}

	resp := &botsInfoResponse{}
	if err := get(ctx, url, botToken, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	if resp.Bot.AppID == "" {
		return nil, errors.New("empty response")
	}

	return &resp.Bot, nil
}
