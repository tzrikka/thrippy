package chatgpt

import (
	"github.com/tzrikka/thrippy/pkg/links/templates"
)

var Template = templates.New(
	"ChatGPT using a static API key",
	[]string{
		"https://platform.openai.com/docs/api-reference/authentication",
		"https://platform.openai.com/api-keys",
	},
	[]string{"api_key"},
	nil,
	nil,
)
