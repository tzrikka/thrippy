package confluence

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"

	"github.com/tzrikka/thrippy/internal/links"
	"github.com/tzrikka/thrippy/pkg/links/atlassian"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

var APITokenTemplate = links.NewTemplate(
	"Confluence with a user's static API token",
	[]string{
		"https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/",
		"https://id.atlassian.com/manage-profile/security/api-tokens",
	},
	[]string{"base_url_manual", "email_manual", "api_token_manual"},
	nil,
	apiTokenChecker,
)

// apiTokenChecker checks the given static API token for
// Confluence Cloud, and returns metadata about it in JSON format.
func apiTokenChecker(ctx context.Context, m map[string]string, _ *oauth.Config, _ *oauth2.Token) (string, error) {
	baseURL, err := links.NormalizeURL(m["base_url"])
	if err != nil {
		return "", err
	}

	// https://developer.atlassian.com/cloud/confluence/rest/v1/api-group-users/#api-wiki-rest-api-user-current-get
	url := baseURL + "/wiki/rest/api/user/current"
	user := &User{}
	if err := atlassian.CurrentUser(ctx, url, m["email"], m["api_token"], user); err != nil {
		return "", fmt.Errorf("error in getting current Confluence Cloud user: %w", err)
	}

	return links.EncodeMetadataAsJSON(atlassian.Metadata{
		AccountID:   user.AccountID,
		AccountType: user.AccountType,
		Email:       user.Email,
		Name:        user.PublicName,
		TimeZone:    user.TimeZone,
	})
}

// https://developer.atlassian.com/cloud/confluence/rest/v1/api-group-users/#api-wiki-rest-api-user-current-get
type User struct {
	AccountID   string `json:"accountId"`
	AccountType string `json:"accountType"`
	Email       string `json:"email"`
	PublicName  string `json:"publicName"`
	TimeZone    string `json:"timeZone"`
}
