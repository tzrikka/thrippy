// Package oauth is a collection of simple, stateless OAuth utility functions.
package oauth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	trippypb "github.com/tzrikka/trippy/proto/trippy/v1"
)

const (
	timeout = time.Second * 3
)

type Config struct {
	Config    *oauth2.Config
	AuthCodes map[string]string
}

// FromProto converts a wire-protocol [trippypb.OAuthConfig]
// message into a [Config] struct which is usable in Go.
// This function returns nil if the input is also nil.
func FromProto(c *trippypb.OAuthConfig) *Config {
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
	}
}

// ToString returns a human-readable string representation of a
// [trippypb.OAuthConfig] message, for pretty-printing in the CLI.
// This function returns an empty string if the input is nil.
func ToString(c *trippypb.OAuthConfig) string {
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
		lines = append(lines, fmt.Sprintf("Auth Codes: %v", acs))
	}

	return strings.Join(lines, "\n")
}

// ToProto converts this struct into a [trippypb.OAuthConfig]
// protocol-buffer message, for transmission over gRPC.
// This function returns nil if the receiver is nil.
func (c *Config) ToProto() *trippypb.OAuthConfig {
	if c == nil {
		return nil
	}

	return trippypb.OAuthConfig_builder{
		AuthUrl:   proto.String(c.Config.Endpoint.AuthURL),
		TokenUrl:  proto.String(c.Config.Endpoint.TokenURL),
		AuthStyle: proto.Int64(int64(c.Config.Endpoint.AuthStyle)),

		ClientId:     proto.String(c.Config.ClientID),
		ClientSecret: proto.String(c.Config.ClientSecret),

		Scopes:    c.Config.Scopes,
		AuthCodes: c.AuthCodes,
	}.Build()
}

// ToJSON converts this struct into a JSON representation of a
// [trippypb.OAuthConfig] protocol-buffer message, for storage in the
// secrets manager. This function returns "{}" if the receiver is nil.
func (c *Config) ToJSON() (string, error) {
	if c == nil {
		return "{}", nil
	}

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

// TokenToProto converts the given [oauth2.Token] into a [trippypb.OAuthConfig]
// protocol-buffer message, for transmission over gRPC and then storage.
func TokenToProto(t *oauth2.Token) *trippypb.OAuthToken {
	if t.Expiry.IsZero() && t.ExpiresIn > 0 { // If both are 0, the access token never expires.
		t.Expiry = time.Now().Add(time.Second * time.Duration(t.ExpiresIn))
	}

	t.Expiry = t.Expiry.UTC() // Whether or not it was already populated.

	o := trippypb.OAuthToken_builder{
		AccessToken: proto.String(t.AccessToken),
		Expiry:      proto.String(t.Expiry.Format(time.RFC3339)),
	}.Build()

	if t.RefreshToken != "" {
		o.SetRefreshToken(t.RefreshToken)
	}
	if t.TokenType != "" {
		o.SetTokenType(t.TokenType)
	}

	return o
}

// TokenFromProto converts a wire-protocol [trippypb.OAuthToken] message into an
// [oauth2.Token] struct. This function returns nil if the input is also nil.
func TokenFromProto(o *trippypb.OAuthToken) *oauth2.Token {
	if o == nil {
		return nil
	}

	t, _ := time.Parse(time.RFC3339, o.GetExpiry())
	return &oauth2.Token{
		AccessToken:  o.GetAccessToken(),
		Expiry:       t,
		RefreshToken: o.GetRefreshToken(),
		TokenType:    o.GetTokenType(),
	}
}
