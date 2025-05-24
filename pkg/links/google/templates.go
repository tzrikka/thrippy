package google

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleoauth2 "google.golang.org/api/oauth2/v2"

	"github.com/tzrikka/thrippy/pkg/links/templates"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

var (
	ServiceAccountTemplate = templates.New(
		"Google APIs using a static GCP service account key",
		[]string{
			"https://cloud.google.com/iam/docs/service-account-overview",
			"https://developers.google.com/identity/protocols/oauth2/service-account",
			"https://console.cloud.google.com/iam-admin/serviceaccounts",
		},
		[]string{"key"},
		nil,
		serviceKeyChecker,
	)

	UserOAuthTemplate = templates.New(
		"Google APIs using OAuth 2.0 to act on behalf of a user",
		[]string{
			"https://developers.google.com/workspace/guides/get-started",
			"https://console.cloud.google.com/auth/overview",
		},
		templates.OAuthCredFields,
		oauthModifier,
		userTokenChecker,
	)
)

// oauthModifier adjusts the given [oauth.Config] for Google
// OAuth 2.0 authorizations, to act on behalf of a user.
func oauthModifier(o *oauth.Config) {
	if o.Config.Endpoint.AuthURL == "" {
		o.Config.Endpoint.AuthURL = google.Endpoint.AuthURL
	}

	if o.Config.Endpoint.TokenURL == "" {
		o.Config.Endpoint.TokenURL = google.Endpoint.TokenURL
	}

	if o.Config.Endpoint.AuthStyle == oauth2.AuthStyleAutoDetect {
		o.Config.Endpoint.AuthStyle = google.Endpoint.AuthStyle
	}

	// https://developers.google.com/identity/protocols/oauth2/scopes#oauth2
	o.Config.Scopes = append(o.Config.Scopes,
		googleoauth2.UserinfoEmailScope,
		googleoauth2.UserinfoProfileScope,
		googleoauth2.OpenIDScope,
	)

	if o.AuthCodes == nil {
		o.AuthCodes = map[string]string{}
	}
	if _, ok := o.AuthCodes["access_type"]; !ok {
		o.AuthCodes["access_type"] = "offline" // [oauth2.AccessTypeOffline].
	}
}

// userTokenChecker checks the given OAuth token,
// and returns metadata about its owner in JSON format.
func userTokenChecker(ctx context.Context, _ map[string]string, o *oauth.Config, t *oauth2.Token) (string, error) {
	user, token, err := oauthUserInfo(ctx, o, t)
	if err != nil {
		return "", err
	}

	j, err := json.Marshal(oauthMetadata{
		Email:         user.Email,
		ID:            user.Id,
		FamilyName:    user.FamilyName,
		GivenName:     user.GivenName,
		Name:          user.Name,
		Picture:       user.Picture,
		Scopes:        token.Scope,
		VerifiedEmail: strconv.FormatBool(*user.VerifiedEmail),
	})
	if err != nil {
		return "", err
	}

	return string(j), nil
}

// serviceKeyChecker checks the given Google Cloud service
// account key, and returns metadata about it in JSON format.
func serviceKeyChecker(ctx context.Context, m map[string]string, _ *oauth.Config, _ *oauth2.Token) (string, error) {
	email, id, err := serviceAccountInfo(ctx, m["key"])
	if err != nil {
		return "", err
	}

	matches := regexp.MustCompile(`"project_id":\s*"(.*?)"`).FindStringSubmatch(m["key"])
	if len(matches) < 2 {
		return "", errors.New("project ID not found in service account key")
	}

	j, err := json.Marshal(oauthMetadata{
		Email:   email,
		ID:      id,
		Project: matches[1],
	})
	if err != nil {
		return "", err
	}

	return string(j), nil
}

type oauthMetadata struct {
	Email         string `json:"email"`
	ID            string `json:"id"`
	FamilyName    string `json:"family_name,omitempty"`
	GivenName     string `json:"given_name,omitempty"`
	Name          string `json:"name,omitempty"`
	Picture       string `json:"picture,omitempty"`
	Scopes        string `json:"scopes,omitempty"`
	VerifiedEmail string `json:"verified_email,omitempty"`
	Project       string `json:"project,omitempty"`
}
