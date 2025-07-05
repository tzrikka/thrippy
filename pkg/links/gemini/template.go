package gemini

import (
	"github.com/tzrikka/thrippy/internal/links"
)

var Template = links.NewTemplate(
	"Gemini using a static API key",
	[]string{
		"https://ai.google.dev/gemini-api/docs/api-key",
		"https://aistudio.google.com/app/apikey",
		"https://console.cloud.google.com/apis/credentials",
	},
	[]string{"api_key_manual"},
	nil,
	nil,
)
