package github

import (
	"testing"

	"golang.org/x/oauth2"

	"github.com/tzrikka/thrippy/pkg/oauth"
)

func TestAuthBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		urlParam string
		authURL  string
		want     string
	}{
		{
			name: "default",
			want: DefaultBaseURL,
		},
		{
			name:     "param_without_scheme",
			urlParam: "base_url",
			want:     "https://base_url",
		},
		{
			name:     "param_with_http_scheme",
			urlParam: "http://base_url",
			want:     "https://base_url",
		},
		{
			name:     "param_with_https_scheme",
			urlParam: "https://base_url",
			want:     "https://base_url",
		},
		{
			name:     "param_with_path",
			urlParam: "https://base_url/foo/bar",
			want:     "https://base_url",
		},
		{
			name:    "auth_url",
			authURL: "https://base_url",
			want:    "https://base_url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &oauth.Config{
				Config: &oauth2.Config{Endpoint: oauth2.Endpoint{AuthURL: tt.authURL}},
			}
			if tt.urlParam != "" {
				o.Params = map[string]string{"base_url": tt.urlParam}
			}
			if got := AuthBaseURL(o); got != tt.want {
				t.Errorf("AuthBaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAPIBaseURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "default",
			url:  DefaultBaseURL,
			want: "https://api.github.com",
		},
		{
			name: "ghes",
			url:  "https://base_url",
			want: "https://base_url/api/v3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := APIBaseURL(tt.url); got != tt.want {
				t.Errorf("APIBaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
