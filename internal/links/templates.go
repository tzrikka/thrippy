// Package links provides standard types and helper functions
// for all the templates that are implemented in [pkg/links].
//
// [pkg/links]: https://pkg.go.dev/github.com/tzrikka/thrippy/pkg/links
package links

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/protobuf/proto"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

type OAuthFunc func(*oauth.Config)

type CheckerFunc func(context.Context, map[string]string, *oauth.Config, *oauth2.Token) (string, error)

type Template struct {
	description string
	links       []string
	credFields  []string
	oauthFunc   OAuthFunc
	checkerFunc CheckerFunc
}

// NewTemplate defines the authentication details of a well-known third-party service.
func NewTemplate(description string, links, credFields []string, of OAuthFunc, cf CheckerFunc) Template {
	return Template{
		description: description,
		links:       links,
		credFields:  credFields,
		oauthFunc:   of,
		checkerFunc: cf,
	}
}

func (t Template) Description() string {
	return t.description
}

// CredFields returns a copy of all the expected field names
// in the link's credentials, based on the link's template.
func (t Template) CredFields() []*thrippypb.CredentialField {
	if len(t.credFields) == 0 {
		return nil
	}

	fields := make([]*thrippypb.CredentialField, len(t.credFields))
	for i, name := range t.credFields {
		fields[i] = thrippypb.CredentialField_builder{Name: proto.String(name)}.Build()
		if prefix, ok := strings.CutSuffix(name, "_optional"); ok {
			name = prefix
			fields[i].SetName(name)
			fields[i].SetOptional(true)
		}
		if prefix, ok := strings.CutSuffix(name, "_manual"); ok {
			name = prefix
			fields[i].SetName(name)
			fields[i].SetManual(true)
		}
	}
	return fields
}

// Check checks the usability of the provided credentials (either the map or
// the token), and returns JSON-serialized metadata about them for storage.
func (t Template) Check(ctx context.Context, m map[string]string, oc *oauth.Config, ot *oauth2.Token) (string, error) {
	if t.checkerFunc == nil {
		return "", nil
	}
	return t.checkerFunc(ctx, m, oc, ot)
}

// OAuthCredFields is a reusable standard based on [oauth2.Token].
var OAuthCredFields = []string{"access_token", "expiry", "refresh_token", "token_type"}

// ModifyOAuthByTemplate fills in all the missing OAuth
// configuration details, based on the given link template.
// It also normalizes (i.e. sorts and compacts) OAuth scopes.
func ModifyOAuthByTemplate(o *oauth.Config, t Template, found bool) {
	if !found || o == nil {
		return
	}

	if t.oauthFunc != nil {
		t.oauthFunc(o)
	}

	slices.Sort(o.Config.Scopes)
	o.Config.Scopes = slices.Compact(o.Config.Scopes)
}

// EncodeMetadataAsJSON converts the given struct into a JSON string.
func EncodeMetadataAsJSON(v any) (string, error) {
	sb := new(strings.Builder)
	if err := json.NewEncoder(sb).Encode(v); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// NormalizeBaseURL checks that the given URL is valid,
// and strips any suffixes after the host address.
func NormalizeBaseURL(baseURL string) (string, error) {
	if baseURL == "" {
		return "", fmt.Errorf("missing base URL")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	if u.Host == "" {
		return "", fmt.Errorf("invalid base URL: no host")
	}

	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}
