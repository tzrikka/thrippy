package github

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/tzrikka/thrippy/pkg/client"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

const (
	DefaultBaseURL = "https://github.com"
	badBaseURL     = "https://bad-base-url"
)

// AuthBaseURL returns the base URL for GitHub: either [DefaultBaseURL]
// or a link-specific URL for GitHub Enterprise Server (GHES).
func AuthBaseURL(o *oauth.Config) string {
	baseURL, ok := o.Params["base_url"] // Link creation.
	if !ok {
		baseURL = o.Config.Endpoint.AuthURL // Anytime afterwards.
	}

	if baseURL == "" {
		return DefaultBaseURL
	}

	// Custom base URL for GitHub Enterprise Server (GHES): normalize it.
	if strings.HasPrefix(baseURL, "http://") {
		baseURL = strings.Replace(baseURL, "http://", "https://", 1)
	}
	if !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return badBaseURL
	}
	if u.Host == "" {
		return badBaseURL
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""

	return u.String()
}

// APIBaseURL transforms the given GitHub base URL
// into an API endpoint URL, based on [this].
//
// [this]: https://docs.github.com/en/enterprise-server/apps/sharing-github-apps/making-your-github-app-available-for-github-enterprise-server#the-app-code-must-use-the-correct-urls
func APIBaseURL(baseURL string) string {
	if baseURL == DefaultBaseURL {
		return "https://api.github.com"
	}
	return baseURL + "/api/v3"
}

// generateJWT generates a JSON Web Token (JWT) for a GitHub app. Based on:
// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-json-web-token-jwt-for-a-github-app
func generateJWT(clientID, privateKey string) (string, error) {
	// Input sanity checks.
	if clientID == "" {
		return "", errors.New("missing credential: client_id")
	}
	if privateKey == "" {
		return "", errors.New("missing credential: private_key")
	}

	// Parse the private key.
	privateKey = strings.ReplaceAll(privateKey, "\\n", "\n")
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return "", errors.New("failed to decode PEM private key")
	}

	pk, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Generate and sign the JWT.
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(time.Minute * 10).Unix(),
		"iss": clientID,
	})

	signedToken, err := token.SignedString(pk)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return signedToken, nil
}

const (
	mimeType = "application/vnd.github+json"
)

// get is a GitHub-specific HTTP GET wrapper for [client.HTTPRequest].
func get(ctx context.Context, url, token string) (map[string]any, error) {
	resp, err := client.HTTPRequest(ctx, http.MethodGet, url, mimeType, token)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	return m, json.Unmarshal(resp, &m)
}
