package google

import (
	"context"

	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"github.com/tzrikka/thrippy/pkg/oauth"
)

func oauthUserInfo(ctx context.Context, o *oauth.Config, t *oauth2.Token) (*googleoauth2.Userinfo, *googleoauth2.Tokeninfo, error) {
	// https://github.com/googleapis/google-api-go-client
	opt := option.WithTokenSource(o.Config.TokenSource(ctx, t))
	svc, err := googleoauth2.NewService(ctx, opt)
	if err != nil {
		return nil, nil, err
	}

	// https://developers.google.com/identity/openid-connect/openid-connect#obtainuserinfo
	ui, err := svc.Userinfo.V2.Me.Get().Do()
	if err != nil {
		return nil, nil, err
	}

	ti, err := svc.Tokeninfo().Do()
	if err != nil {
		return nil, nil, err
	}

	return ui, ti, nil
}

func serviceAccountInfo(ctx context.Context, jsonKey string) (string, string, error) {
	// https://cloud.google.com/docs/authentication/client-libraries#external-credentials
	opt := option.WithCredentialsJSON([]byte(jsonKey))
	svc, err := googleoauth2.NewService(ctx, opt)
	if err != nil {
		return "", "", err
	}

	// https://developers.google.com/identity/openid-connect/openid-connect#obtainuserinfo
	ui, err := svc.Userinfo.V2.Me.Get().Do()
	if err != nil {
		return "", "", err
	}

	return ui.Email, ui.Id, nil
}
