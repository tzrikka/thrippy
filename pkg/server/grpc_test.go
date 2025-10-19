package server

import (
	"reflect"
	"testing"

	"github.com/lithammer/shortuuid/v4"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
	intlinks "github.com/tzrikka/thrippy/internal/links"
	"github.com/tzrikka/thrippy/pkg/links"
	"github.com/tzrikka/thrippy/pkg/secrets"
)

func TestCreateLink(t *testing.T) {
	cmd := &cli.Command{Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "grpc-addr",
			Value: "127.0.0.1:0",
		},
		&cli.BoolFlag{
			Name:  "dev",
			Value: true,
		},
	}}
	addr, err := startGRPCServer(t.Context(), cmd, secrets.NewTestManager())
	if err != nil {
		t.Fatal(err)
	}

	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(addr, creds)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := thrippypb.NewThrippyServiceClient(conn)

	tests := []struct {
		name    string
		req     *thrippypb.CreateLinkRequest
		wantErr bool
	}{
		{
			name: "invalid_template",
			req: thrippypb.CreateLinkRequest_builder{
				Template: proto.String("bad-template-id"),
			}.Build(),
			wantErr: true,
		},
		{
			name: "oauth_without_client_id",
			req: thrippypb.CreateLinkRequest_builder{
				Template: proto.String("generic-oauth"),
				OauthConfig: thrippypb.OAuthConfig_builder{
					AuthUrl:      proto.String("111"),
					ClientSecret: proto.String("222"),
				}.Build(),
			}.Build(),
			wantErr: true,
		},
		{
			name: "generic_oauth",
			req: thrippypb.CreateLinkRequest_builder{
				Template: proto.String("generic-oauth"),
				OauthConfig: thrippypb.OAuthConfig_builder{
					AuthUrl:      proto.String("111"),
					ClientId:     proto.String("222"),
					ClientSecret: proto.String("333"),
				}.Build(),
			}.Build(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.CreateLink(t.Context(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLink() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteLinkOAuth(t *testing.T) {
	cmd := &cli.Command{Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "grpc-addr",
			Value: "127.0.0.1:0",
		},
		&cli.BoolFlag{
			Name:  "dev",
			Value: true,
		},
	}}
	addr, err := startGRPCServer(t.Context(), cmd, secrets.NewTestManager())
	if err != nil {
		t.Fatal(err)
	}

	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(addr, creds)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := thrippypb.NewThrippyServiceClient(conn)
	resp1, err := client.CreateLink(t.Context(), thrippypb.CreateLinkRequest_builder{
		Template: proto.String("generic-oauth"),
		OauthConfig: thrippypb.OAuthConfig_builder{
			ClientId:     proto.String("111"),
			ClientSecret: proto.String("222"),
		}.Build(),
	}.Build())
	if err != nil {
		t.Fatalf("CreateLink() error = %v", err)
	}

	tests := []struct {
		name    string
		req     *thrippypb.DeleteLinkRequest
		wantErr bool
	}{
		{
			name:    "missing_id",
			req:     thrippypb.DeleteLinkRequest_builder{}.Build(),
			wantErr: true,
		},
		{
			name: "invalid_id",
			req: thrippypb.DeleteLinkRequest_builder{
				LinkId: proto.String("111"),
			}.Build(),
			wantErr: true,
		},
		{
			name: "link_not_found_is_error",
			req: thrippypb.DeleteLinkRequest_builder{
				LinkId: proto.String(shortuuid.New()),
			}.Build(),
			wantErr: true,
		},
		{
			name: "link_not_found_isnt_error",
			req: thrippypb.DeleteLinkRequest_builder{
				LinkId:       proto.String(shortuuid.New()),
				AllowMissing: proto.Bool(true),
			}.Build(),
		},
		{
			name: "happy_path",
			req: thrippypb.DeleteLinkRequest_builder{
				LinkId: proto.String(resp1.GetLinkId()),
			}.Build(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := client.DeleteLink(t.Context(), tt.req); (err != nil) != tt.wantErr {
				t.Fatalf("DeleteLink() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr || tt.req.GetAllowMissing() {
				return
			}

			req3 := thrippypb.GetLinkRequest_builder{LinkId: proto.String(resp1.GetLinkId())}.Build()
			if _, err = client.GetLink(t.Context(), req3); err == nil {
				t.Errorf("GetLink() error = %v, wantErr true", err)
			}
		})
	}
}

func TestGetLinkOAuth(t *testing.T) {
	cmd := &cli.Command{Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "grpc-addr",
			Value: "127.0.0.1:0",
		},
		&cli.BoolFlag{
			Name:  "dev",
			Value: true,
		},
	}}
	addr, err := startGRPCServer(t.Context(), cmd, secrets.NewTestManager())
	if err != nil {
		t.Fatal(err)
	}

	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(addr, creds)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := thrippypb.NewThrippyServiceClient(conn)
	resp1, err := client.CreateLink(t.Context(), thrippypb.CreateLinkRequest_builder{
		Template: proto.String("generic-oauth"),
		OauthConfig: thrippypb.OAuthConfig_builder{
			ClientId:     proto.String("111"),
			ClientSecret: proto.String("222"),
		}.Build(),
	}.Build())
	if err != nil {
		t.Fatalf("CreateLink() error = %v", err)
	}

	tests := []struct {
		name       string
		req        *thrippypb.GetLinkRequest
		wantID     string
		wantSecret string
		wantErr    bool
	}{
		{
			name:    "missing_id",
			req:     thrippypb.GetLinkRequest_builder{}.Build(),
			wantErr: true,
		},
		{
			name: "invalid_id",
			req: thrippypb.GetLinkRequest_builder{
				LinkId: proto.String("111"),
			}.Build(),
			wantErr: true,
		},
		{
			name: "link_not_found",
			req: thrippypb.GetLinkRequest_builder{
				LinkId: proto.String(shortuuid.New()),
			}.Build(),
			wantErr: true,
		},
		{
			name: "happy_path",
			req: thrippypb.GetLinkRequest_builder{
				LinkId: proto.String(resp1.GetLinkId()),
			}.Build(),
			wantID:     "111",
			wantSecret: "222",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp2, err := client.GetLink(t.Context(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLink() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := resp2.GetOauthConfig()
			if got.GetClientId() != tt.wantID {
				t.Errorf("GetLink()...GetClientId() = %q, want %q", got.GetClientId(), tt.wantID)
			}
			if got.GetClientSecret() != tt.wantSecret {
				t.Errorf("GetLink()...GetClientSecret() = %q, want %q", got.GetClientSecret(), tt.wantSecret)
			}
		})
	}
}

func TestGetLinkNonOAuth(t *testing.T) {
	cmd := &cli.Command{Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "grpc-addr",
			Value: "127.0.0.1:0",
		},
		&cli.BoolFlag{
			Name:  "dev",
			Value: true,
		},
	}}
	addr, err := startGRPCServer(t.Context(), cmd, secrets.NewTestManager())
	if err != nil {
		t.Fatal(err)
	}

	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(addr, creds)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := thrippypb.NewThrippyServiceClient(conn)
	resp, err := client.CreateLink(t.Context(), thrippypb.CreateLinkRequest_builder{
		Template: proto.String("slack-bot-token"),
	}.Build())
	if err != nil {
		t.Fatalf("CreateLink() error = %v", err)
	}

	got, err := client.GetLink(t.Context(), thrippypb.GetLinkRequest_builder{
		LinkId: proto.String(resp.GetLinkId()),
	}.Build())
	if err != nil {
		t.Errorf("GetLink() error = %v", err)
		return
	}

	wantTemplate := "slack-bot-token"
	if got.GetTemplate() != wantTemplate {
		t.Errorf("GetLink().GetTemplate() = %q, want %q", got.GetTemplate(), wantTemplate)
	}
	if got.GetOauthConfig() != nil {
		t.Errorf("GetLink().GetOauthConfig() = %v, want nil", got.GetOauthConfig())
	}
}

func TestSetAndGetCredentials(t *testing.T) {
	links.Templates["test"] = intlinks.NewTemplate("Generic link", nil, nil, nil, nil)

	cmd := &cli.Command{Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "grpc-addr",
			Value: "127.0.0.1:0",
		},
		&cli.BoolFlag{
			Name:  "dev",
			Value: true,
		},
	}}
	addr, err := startGRPCServer(t.Context(), cmd, secrets.NewTestManager())
	if err != nil {
		t.Fatal(err)
	}

	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(addr, creds)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	tests := []struct {
		name     string
		template string
		req      *thrippypb.SetCredentialsRequest
		want     map[string]string
	}{
		{
			name:     "generic_creds",
			template: "test",
			req: thrippypb.SetCredentialsRequest_builder{
				GenericCreds: map[string]string{
					"aaa": "111",
					"bbb": "222",
				},
			}.Build(),
			want: map[string]string{
				"aaa": "111",
				"bbb": "222",
			},
		},
		{
			name:     "token",
			template: "generic-oauth",
			req: thrippypb.SetCredentialsRequest_builder{
				Token: thrippypb.OAuthToken_builder{
					AccessToken:  proto.String("access_token"),
					Expiry:       proto.String("2025-05-17T10:11:12Z"),
					RefreshToken: proto.String("refresh_token"),
				}.Build(),
			}.Build(),
			want: map[string]string{
				"access_token":  "access_token",
				"expiry":        "2025-05-17T10:11:12Z",
				"refresh_token": "refresh_token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := thrippypb.NewThrippyServiceClient(conn)
			resp1, err := client.CreateLink(t.Context(), thrippypb.CreateLinkRequest_builder{
				Template: proto.String(tt.template),
				OauthConfig: thrippypb.OAuthConfig_builder{
					ClientId:     proto.String("111"),
					ClientSecret: proto.String("222"),
				}.Build(),
			}.Build())
			if err != nil {
				t.Fatalf("CreateLink() error = %v", err)
			}

			tt.req.SetLinkId(resp1.GetLinkId())
			_, err = client.SetCredentials(t.Context(), tt.req)
			if err != nil {
				t.Errorf("SetCredentials() error = %v", err)
				return
			}

			got, err := client.GetCredentials(t.Context(), thrippypb.GetCredentialsRequest_builder{
				LinkId: proto.String(resp1.GetLinkId()),
			}.Build())
			if err != nil {
				t.Errorf("GetCredentials() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got.GetCredentials(), tt.want) {
				t.Errorf("GetCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}
