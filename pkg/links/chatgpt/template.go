package chatgpt

import (
	"context"

	"golang.org/x/oauth2"

	"github.com/tzrikka/thrippy/internal/links"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

var Template = links.NewTemplate(
	"ChatGPT using a static API key",
	[]string{
		"https://platform.openai.com/docs/api-reference/authentication",
		"https://platform.openai.com/api-keys",
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
