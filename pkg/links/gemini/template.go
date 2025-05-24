package gemini

import (
	"github.com/tzrikka/thrippy/pkg/links/templates"
)

var Template = templates.New(
	"Gemini using a static API key",
	[]string{
		"https://ai.google.dev/gemini-api/docs/api-key",
		"https://aistudio.google.com/app/apikey",
		"https://console.cloud.google.com/apis/credentials",
	},
	[]string{"api_key"},
	nil,
	nil,
)
