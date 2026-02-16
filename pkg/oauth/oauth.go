// Package oauth is a collection of simple, stateless OAuth utility functions.
package oauth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lithammer/shortuuid/v4"
	"golang.org/x/oauth2"
	"google.golang.org/protobuf/encoding/protojson"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
)

const (
	timeout = 3 * time.Second
)

// Config contains the complete OAuth 2.0 configutation of a link:
// primarily the [oauth2.Config], but also optional [oauth2.AuthCodeOption]
// key-value pairs, and optional [oauth2.Endpoint] URL parameters
// that some links recognize.
//
// If specified and recognized, parameter values are injected into
// [oauth2.Endpoint] URLs in the [oauth2.Config] by the function
// [links.ModifyOAuthByTemplate], when the gRPC server is creating
// a new link. Either way, they are discarded when storing OAuth
// configurations in the secrets manager.
//
// [links.ModifyOAuthByTemplate]: https://pkg.go.dev/github.com/tzrikka/thrippy/pkg/links#ModifyOAuthByTemplate
type Config struct {
	Config    *oauth2.Config
	AuthCodes map[string]string
	Params    map[string]string
	Nonce     string
}

// FromProto converts a wire-protocol [OAuthConfig] protocol-buffer
// message into a [Config] struct which is a usable receiver in Go.
// This function returns nil if the input is also nil.
//
// [OAuthConfig]: https://github.com/tzrikka/thrippy/blob/main/proto/thrippy/v1/oauth.proto
func FromProto(c *thrippypb.OAuthConfig) *Config {
	if c == nil {
		return nil
	}

	return &Config{
		Config: &oauth2.Config{
			ClientID:     c.GetClientId(),
			ClientSecret: c.GetClientSecret(),

			Endpoint: oauth2.Endpoint{
				AuthURL:   c.GetAuthUrl(),
				TokenURL:  c.GetTokenUrl(),
				AuthStyle: oauth2.AuthStyle(c.GetAuthStyle()),
			},
			Scopes: c.GetScopes(),
		},
		AuthCodes: c.GetAuthCodes(),
		Params:    c.GetParams(),
		Nonce:     c.GetNonce(),
	}
}

// ToString returns a human-readable string representation of an [OAuthConfig]
// protocol-buffer message, for pretty-printing in the CLI application.
// This function returns an empty string if the input is nil.
//
// [OAuthConfig]: https://github.com/tzrikka/thrippy/blob/main/proto/thrippy/v1/oauth.proto
func ToString(c *thrippypb.OAuthConfig) string {
	if c.GetAuthUrl() == "" {
		return ""
	}

	lines := []string{
		"Auth URL:   " + c.GetAuthUrl(),
		"Token URL:  " + c.GetTokenUrl(),
		"Client ID:  " + c.GetClientId(),
		"Cli Secret: " + c.GetClientSecret(),
	}

	scopes := c.GetScopes()
	if len(scopes) > 0 {
		lines = append(lines, fmt.Sprintf("Scopes:     %v", scopes))
	}

	acs := c.GetAuthCodes()
	if len(acs) > 0 {
		line := fmt.Sprintf("Auth Codes: %v", acs)
		lines = append(lines, strings.Replace(line, "map", "", 1))
	}

	lines = append(lines, "", "Nonce: "+c.GetNonce())

	return strings.Join(lines, "\n")
}

// IsUsable checks whether this struct has any usable
// field values, or whether it's completely empty.
func (c *Config) IsUsable() bool {
	if c == nil {
		return false
	}

	s := fmt.Sprintf("%s%s%s%s",
		c.Config.Endpoint.AuthURL,
		c.Config.Endpoint.TokenURL,
		c.Config.ClientID,
		c.Config.ClientSecret)

	return len(s) > 0
}

// ToProto converts this struct into an [OAuthConfig] protocol-buffer message,
// for transmission over gRPC. This function returns nil if the receiver is nil.
//
// [OAuthConfig]: https://github.com/tzrikka/thrippy/blob/main/proto/thrippy/v1/oauth.proto
func (c *Config) ToProto() *thrippypb.OAuthConfig {
	if c == nil {
		return nil
	}

	return thrippypb.OAuthConfig_builder{
		AuthUrl:   new(c.Config.Endpoint.AuthURL),
		TokenUrl:  new(c.Config.Endpoint.TokenURL),
		AuthStyle: new(int64(c.Config.Endpoint.AuthStyle)),

		ClientId:     new(c.Config.ClientID),
		ClientSecret: new(c.Config.ClientSecret),

		Scopes:    c.Config.Scopes,
		AuthCodes: c.AuthCodes,

		// Params were already injected into the URLs, so no need to store them as a map.

		Nonce: new(c.Nonce),
	}.Build()
}

// ToJSON converts this struct into a JSON representation of an [OAuthConfig]
// protocol-buffer message, for storage in the secrets manager.
// This function returns "{}" if the receiver is nil.
//
// This function is not safe for concurrent use: it modifies the receiver's [Nonce] field.
// However, this is not a problem in practice because it's triggered by a single live user.
//
// [OAuthConfig]: https://github.com/tzrikka/thrippy/blob/main/proto/thrippy/v1/oauth.proto
func (c *Config) ToJSON() (string, error) {
	if c == nil {
		return "{}", nil
	}

	c.Nonce = shortuuid.New()

	j, err := protojson.Marshal(c.ToProto())
	if err != nil {
		return "", err
	}

	return string(j), nil
}

// AuthCodeURL returns a URL to an OAuth 2.0 provider's consent page
// that asks for permissions for the required scopes explicitly.
//
// State is an opaque value used by us to maintain state between the request
// (to this URL) and the subsequent callback redirect. The authorization
// server includes this value when redirecting the user back to us.
func (c *Config) AuthCodeURL(state string) string {
	return c.Config.AuthCodeURL(state, c.authCodes()...)
}

// Exchange converts a temporary authorization code into an access token.
//
// It is used after a resource provider redirects the user back
// to the callback URL (the URL obtained from [AuthCodeURL]).
//
// The code will be in the *http.Request.FormValue("code").
// Before calling Exchange, be sure to validate FormValue("state")
// if you are using it to protect against CSRF attacks.
func (c *Config) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	client := &http.Client{Timeout: timeout}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
	return c.Config.Exchange(ctx, code, c.authCodes()...)
}

func (c *Config) authCodes() []oauth2.AuthCodeOption {
	var acs []oauth2.AuthCodeOption
	for k, v := range c.AuthCodes {
		acs = append(acs, oauth2.SetAuthURLParam(k, v))
	}
	return acs
}

// RefreshToken returns a refreshed version of the given [oauth2.Token],
// as a map for storage in the secrets manager and transmission to the user,
// assuming that we already checked that it's no longer [oauth2.Token.Valid].
func (c *Config) RefreshToken(ctx context.Context, t *oauth2.Token, force bool) (map[string]any, error) {
	if force {
		t.AccessToken = ""
	}

	t, err := c.Config.TokenSource(ctx, t).Token()
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"access_token":  t.AccessToken,
		"expiry":        t.Expiry.UTC().Format(time.RFC3339),
		"refresh_token": t.RefreshToken,
		"token_type":    t.TokenType,
	}, nil
}

// TokenToProto converts the given [oauth2.Token] into an [OAuthConfig]
// protocol-buffer message, for transmission over gRPC and then storage.
//
// [OAuthConfig]: https://github.com/tzrikka/thrippy/blob/main/proto/thrippy/v1/oauth.proto
func TokenToProto(t *oauth2.Token) *thrippypb.OAuthToken {
	if t.Expiry.IsZero() && t.ExpiresIn > 0 { // If both are 0, the access token never expires.
		t.Expiry = time.Now().Add(time.Second * time.Duration(t.ExpiresIn))
	}

	t.Expiry = t.Expiry.UTC() // Whether or not it was already populated.
	o := thrippypb.OAuthToken_builder{AccessToken: new(t.AccessToken), Expiry: new(t.Expiry.Format(time.RFC3339))}.Build()
	if t.RefreshToken != "" {
		o.SetRefreshToken(t.RefreshToken)
	}
	if t.TokenType != "" {
		o.SetTokenType(t.TokenType)
	}

	return o
}

// TokenFromMap converts a map from the secrets manager into an [oauth2.Token]
// struct. This function returns nil if the input is also nil.
func TokenFromMap(m map[string]any) (*oauth2.Token, bool) {
	if m == nil {
		return nil, false
	}

	es, ok := m["expiry"].(string)
	if !ok {
		es = ""
	}
	et, _ := time.Parse(time.RFC3339, es)

	at, ok := m["access_token"].(string)
	if !ok {
		at = ""
	}

	rt, ok := m["refresh_token"].(string)
	if !ok {
		rt = ""
	}

	tt, ok := m["token_type"].(string)
	if !ok {
		tt = ""
	}

	t := &oauth2.Token{AccessToken: at, Expiry: et, RefreshToken: rt, TokenType: tt}

	if r, ok := m["raw"].(map[string]any); ok {
		t = t.WithExtra(r)
	}

	return t, t.AccessToken != ""
}

// TokenFromProto converts a wire-protocol [OAuthToken] message into an
// [oauth2.Token] struct. This function returns nil if the input is also nil.
//
// [OAuthConfig]: https://github.com/tzrikka/thrippy/blob/main/proto/thrippy/v1/oauth.proto
func TokenFromProto(o *thrippypb.OAuthToken) *oauth2.Token {
	if o == nil {
		return nil
	}

	e, _ := time.Parse(time.RFC3339, o.GetExpiry())
	t := &oauth2.Token{
		AccessToken:  o.GetAccessToken(),
		Expiry:       e,
		RefreshToken: o.GetRefreshToken(),
		TokenType:    o.GetTokenType(),
	}

	if r := o.GetRaw(); len(r) > 0 {
		m := make(map[string]any, len(r))
		for k, v := range r {
			m[k] = v
		}
		t = t.WithExtra(m)
	}

	return t
}
