package github

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/oauth2"

	"github.com/tzrikka/thrippy/pkg/oauth"
)

// AppAuthzModifier adjusts the given [oauth.Config] for
// GitHub app authorizations, to act on behalf of a user.
func AppAuthzModifier(o *oauth.Config) {
	baseURL := AuthBaseURL(o)

	// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-user-access-token-for-a-github-app#using-the-web-application-flow-to-generate-a-user-access-token
	if o.Config.Endpoint.AuthURL == "" {
		o.Config.Endpoint.AuthURL = baseURL + "/login/oauth/authorize"
	}

	// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-user-access-token-for-a-github-app#using-the-web-application-flow-to-generate-a-user-access-token
	if o.Config.Endpoint.TokenURL == "" {
		o.Config.Endpoint.TokenURL = baseURL + "/login/oauth/access_token"
	}

	// https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#3-use-the-access-token-to-access-the-api
	if o.Config.Endpoint.AuthStyle == oauth2.AuthStyleAutoDetect {
		o.Config.Endpoint.AuthStyle = oauth2.AuthStyleInHeader
	}
}

// AppInstallModifier adjusts the given [oauth.Config] for GitHub app
// installations, and using them with JWTs based on static credentials.
func AppInstallModifier(o *oauth.Config) {
	baseURL := AuthBaseURL(o)

	appsDir := "apps" // In github.com.
	if baseURL != DefaultBaseURL {
		appsDir = "github-apps" // GitHub Enterprise Server (GHES).
	}

	appName := o.Params["app_name"]
	if appName == "" {
		appName = "unknown-app-name"
	}

	// https://docs.github.com/en/apps/using-github-apps/installing-a-github-app-from-a-third-party#installing-a-github-app
	if o.Config.Endpoint.AuthURL == "" {
		o.Config.Endpoint.AuthURL = fmt.Sprintf("%s/%s/%s/installations/new", baseURL, appsDir, appName)
	}

	// Use a JWT; creating app or installation tokens is out-of-scope,
	// because it's done automatically by GitHub SDKs.
	o.Config.Endpoint.TokenURL = ""

	// https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#3-use-the-access-token-to-access-the-api
	if o.Config.Endpoint.AuthStyle == oauth2.AuthStyleAutoDetect {
		o.Config.Endpoint.AuthStyle = oauth2.AuthStyleInHeader
	}
}

// JWTChecker generates and checks a JWT based on the given static credentials
// for a GitHub app, and returns metadata in JSON format about the GitHub app.
func JWTChecker(ctx context.Context, m map[string]string, o *oauth.Config, _ *oauth2.Token) (string, error) {
	jwt, err := generateJWT(m["client_id"], m["private_key"])
	if err != nil {
		return "", err
	}

	// https://docs.github.com/en/rest/apps/apps?apiVersion=2022-11-28#get-the-authenticated-app
	u := APIBaseURL(AuthBaseURL(o)) + "/app"
	resp, err := get(ctx, u, jwt)
	if err != nil {
		return "", fmt.Errorf("app details: %w", err)
	}

	owner := resp["owner"].(map[string]any)
	meta := appMetadata{
		Name:         resp["name"].(string),
		Slug:         resp["slug"].(string),
		OwnerLogin:   owner["login"].(string),
		OwnerType:    strings.ToLower(owner["type"].(string)),
		AppUpdatedAt: normalizeRFC3339(resp["updated_at"].(string)),
	}

	// The above must be specified manually by the user, but the following
	// is optional - until Thrippy adds the installation ID automatically.
	installID := m["install_id"]
	if installID == "" {
		j, err := json.Marshal(meta)
		if err != nil {
			return "", err
		}
		return string(j), nil
	}

	// https://docs.github.com/en/rest/apps/apps#get-an-installation-for-the-authenticated-app
	u = fmt.Sprintf("%s/installations/%s", u, installID)
	resp, err = get(ctx, u, jwt)
	if err != nil {
		return "", fmt.Errorf("app installation details: %w", err)
	}

	acct := resp["account"].(map[string]any)
	perms := fmt.Sprintf("%v", resp["permissions"])

	meta.Events = fmt.Sprintf("%v", resp["events"])
	meta.Permissions = strings.Replace(perms, "map", "", 1)
	meta.TargetLogin = acct["login"].(string)
	meta.TargetType = strings.ToLower(acct["type"].(string))
	meta.InstallUpdatedAt = normalizeRFC3339(resp["updated_at"].(string))
	meta.InstallURL = resp["html_url"].(string)

	j, err := json.Marshal(meta)
	if err != nil {
		return "", err
	}
	return string(j), nil
}

type appMetadata struct {
	// Before installation.
	Name         string `json:"app_name"`
	Slug         string `json:"app_slug"`
	OwnerLogin   string `json:"app_owner_login"`
	OwnerType    string `json:"app_owner_type"`
	AppUpdatedAt string `json:"app_updated_at"`

	// After installation.
	Events           string `json:"install_events"`
	Permissions      string `json:"install_permissions"`
	TargetLogin      string `json:"install_target_login"`
	TargetType       string `json:"install_target_type"`
	InstallUpdatedAt string `json:"install_updated_at"`
	InstallURL       string `json:"install_url"`
}

// normalizeRFC3339 strips sub-seconds from RFC-3339 timestamp strings.
func normalizeRFC3339(t string) string {
	return regexp.MustCompile(`\.\d+Z`).ReplaceAllString(t, "Z")
}

// UserChecker checks the given OAuth token,
// or static Personal Access Token (PAT) for GitHub. Based on:
// https://docs.github.com/en/rest/users/users#get-the-authenticated-user
func UserChecker(ctx context.Context, m map[string]string, o *oauth.Config, t *oauth2.Token) (string, error) {
	if o == nil {
		o = &oauth.Config{
			Config: &oauth2.Config{
				Endpoint: oauth2.Endpoint{},
			},
		}
	}
	if o.Config.Endpoint.AuthURL == "" && m["base_url_optional"] != "" {
		o.Config.Endpoint.AuthURL = m["base_url_optional"]
	}

	u := APIBaseURL(AuthBaseURL(o)) + "/user"
	token, ok := m["pat"]
	if !ok && t != nil {
		token = t.AccessToken
	}
	resp, err := get(ctx, u, token)
	if err != nil {
		return "", fmt.Errorf("user details: %w", err)
	}

	company, ok := resp["company"].(string)
	if !ok {
		company = ""
	}

	location, ok := resp["location"].(string)
	if !ok {
		location = ""
	}

	j, err := json.Marshal(userMetadata{
		Company:  company,
		Email:    resp["email"].(string),
		Location: location,
		Login:    resp["login"].(string),
		Name:     resp["name"].(string),
		URL:      resp["html_url"].(string),
		UserID:   strconv.FormatInt(int64(resp["id"].(float64)), 10),
	})
	if err != nil {
		return "", err
	}

	return string(j), nil
}

type userMetadata struct {
	Company  string `json:"company,omitempty"`
	Email    string `json:"email"`
	Location string `json:"location,omitempty"`
	Login    string `json:"login"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	UserID   string `json:"user_id"`
}
