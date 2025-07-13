package bitbucket

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/tzrikka/thrippy/internal/links"
	"github.com/tzrikka/thrippy/pkg/links/atlassian"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

var APITokenTemplate = links.NewTemplate(
	"BitBucket with a user's static API token",
	[]string{
		"https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/",
		"https://support.atlassian.com/bitbucket-cloud/docs/api-tokens/",
		"https://developer.atlassian.com/cloud/bitbucket/rest/intro/#api-tokens",
		"https://id.atlassian.com/manage-profile/security/api-tokens",
	},
	[]string{"email_manual", "api_token_manual"},
	nil,
	apiTokenChecker,
)

// apiTokenChecker checks the given static API token for
// BitBucket Cloud, and returns metadata about it in JSON format.
func apiTokenChecker(ctx context.Context, m map[string]string, _ *oauth.Config, _ *oauth2.Token) (string, error) {
	// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/#api-user-get
	url := "https://api.bitbucket.org/2.0/user"
	user := &User{}
	if err := atlassian.CurrentUser(ctx, url, m["email"], m["api_token"], user); err != nil {
		return "", fmt.Errorf("error in getting current Confluence Cloud user: %w", err)
	}

	return links.EncodeMetadataAsJSON(user)
}

// https://developer.atlassian.com/cloud/bitbucket/rest/api-group-users/#api-user-get
type User struct {
	AccountID   string `json:"account_id"`
	Type        string `json:"type"`
	CreatedOn   string `json:"created_on"`
	DisplayName string `json:"display_name"`
	Nickname    string `json:"nickname,omitempty"`
	Username    string `json:"username"`
	UUID        string `json:"uuid"`
}
