package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/tzrikka/thrippy/pkg/client"
)

// https://docs.slack.dev/reference/methods/auth.test/
type authTestResponse struct {
	response

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

// https://docs.slack.dev/reference/methods/bots.info/
type botsInfoResponse struct {
	response

	Bot *bot `json:"bot,omitempty"`
}

// https://docs.slack.dev/reference/methods/bots.info/
type bot struct {
	ID      string `json:"id"`
	TeamID  string `json:"team_id"`
	Name    string `json:"name"`
	AppID   string `json:"app_id"`
	UserID  string `json:"user_id"`
	Deleted bool   `json:"deleted"`
	Updated int64  `json:"updated"`
}

type response struct {
	OK               bool              `json:"ok"`
	Error            string            `json:"error,omitempty"`
	Needed           string            `json:"needed,omitempty"`   // Scope errors (undocumented).
	Provided         string            `json:"provided,omitempty"` // Scope errors (undocumented).
	Warning          string            `json:"warning,omitempty"`
	ResponseMetadata *responseMetadata `json:"response_metadata,omitempty"`
}

type responseMetadata struct {
	Messages   []string `json:"messages,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
	NextCursor string   `json:"next_cursor,omitempty"`
}

// authTest checks the caller's authentication & identity.
// Based on https://docs.slack.dev/reference/methods/auth.test/ (no scopes required).
func authTest(ctx context.Context, baseURL, botToken string) (*authTestResponse, error) {
	url := baseURL + "/api/auth.test"

	resp := new(authTestResponse)
	if err := post(ctx, url, botToken, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp, nil
}

// botsInfo gets information about a bot user.
// Based on https://docs.slack.dev/reference/methods/bots.info/
// (required scope: https://docs.slack.dev/reference/scopes/users.read).
func botsInfo(ctx context.Context, baseURL, botToken string, authTest *authTestResponse) (*bot, error) {
	url := fmt.Sprintf("%s/api/bots.info?bot=%s", baseURL, authTest.BotID)
	if authTest.TeamID != "" {
		url = fmt.Sprintf("%s&team_id=%s", url, authTest.TeamID)
	}
	if authTest.IsEnterpriseInstall {
		url = fmt.Sprintf("%s&enterprise_id=%s", url, authTest.EnterpriseID)
	}

	resp := new(botsInfoResponse)
	if err := get(ctx, url, botToken, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	if resp.Bot.AppID == "" {
		return nil, errors.New("empty response")
	}

	return resp.Bot, nil
}

// WebSocketURL generates a temporary Socket Mode WebSocket URL that your app
// can connect to in order to receive events and interactive payloads over.
// Based on https://docs.slack.dev/reference/methods/apps.connections.open/
// (required scope: https://docs.slack.dev/reference/scopes/connections.write).
func webSocketURL(ctx context.Context, baseURL, appLevelToken string) error {
	url := baseURL + "/api/apps.connections.open"

	resp := new(response)
	if err := post(ctx, url, appLevelToken, resp); err != nil {
		return err
	}
	if !resp.OK {
		return errors.New(resp.Error)
	}
	return nil
}

// get is a Slack-specific HTTP GET wrapper for [client.HTTPRequest].
func get(ctx context.Context, url, botToken string, jsonResp any) error {
	resp, err := client.HTTPRequest(ctx, http.MethodGet, url, "Bearer "+botToken)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp, jsonResp)
}

// post is a Slack-specific HTTP POST wrapper for [client.HTTPRequest].
func post(ctx context.Context, url, botToken string, jsonResp any) error {
	resp, err := client.HTTPRequest(ctx, http.MethodPost, url, "Bearer "+botToken)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp, jsonResp)
}
