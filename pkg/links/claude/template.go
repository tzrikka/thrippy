package claude

import (
	"github.com/tzrikka/thrippy/pkg/links/templates"
)

var Template = templates.New(
	"Claude using a static API key",
	[]string{
		"https://docs.anthropic.com/en/api/overview",
		"https://console.anthropic.com/settings/keys",
	},
	[]string{"api_key"},
	nil,
	nil,
)
