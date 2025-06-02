package claude

import (
	"context"

	"golang.org/x/oauth2"

	"github.com/tzrikka/thrippy/pkg/links/templates"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

var Template = templates.New(
	"Claude using a static API key",
	[]string{
		"https://docs.anthropic.com/en/api/overview",
		"https://console.anthropic.com/settings/keys",
	},
	[]string{"api_key_manual"},
	nil,
	apiKeyChecker,
)

func apiKeyChecker(ctx context.Context, m map[string]string, _ *oauth.Config, _ *oauth2.Token) (string, error) {
	if _, err := get(ctx, modelsURL, m["api_key"]); err != nil {
		return "", err
	}
	return "", nil
}
