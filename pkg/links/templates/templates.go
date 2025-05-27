package templates

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"golang.org/x/oauth2"

	"github.com/tzrikka/thrippy/pkg/oauth"
)

type (
	OAuthFunc   func(*oauth.Config)
	CheckerFunc func(context.Context, map[string]string, *oauth.Config, *oauth2.Token) (string, error)

	Template struct {
		description string
		links       []string
		credFields  []string
		oauthFunc   OAuthFunc
		checkerFunc CheckerFunc
	}
)

// New defines the authentication details of a well-known third-party service.
func New(description string, links, credFields []string, of OAuthFunc, cf CheckerFunc) Template {
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

// CredsFields returns a copy of all the expected field names
// in the link's credentials, based on the link's template.
func (t Template) CredFields() []string {
	if len(t.credFields) == 0 {
		return nil
	}
	return slices.Clone(t.credFields)
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
	sb := strings.Builder{}
	if err := json.NewEncoder(&sb).Encode(v); err != nil {
		return "", err
	}
	return sb.String(), nil
}
