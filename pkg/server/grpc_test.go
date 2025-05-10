package server

import (
	"testing"

	"github.com/lithammer/shortuuid/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	"github.com/tzrikka/trippy/pkg/secrets"
	trippypb "github.com/tzrikka/trippy/proto/trippy/v1"
)

func TestCreateLink(t *testing.T) {
	addr, err := startGRPCServer(secrets.NewTestManager(), "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(addr, creds)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := trippypb.NewTrippyServiceClient(conn)

	tests := []struct {
		name    string
		req     *trippypb.CreateLinkRequest
		wantErr bool
	}{
		{
			name: "invalid_template",
			req: trippypb.CreateLinkRequest_builder{
				Template: proto.String("bad-template-id"),
			}.Build(),
			wantErr: true,
		},
		{
			name: "oauth_without_client_id",
			req: trippypb.CreateLinkRequest_builder{
				Template: proto.String("generic"),
				OauthConfig: trippypb.OAuthConfig_builder{
					AuthUrl:      proto.String("111"),
					ClientSecret: proto.String("222"),
				}.Build(),
			}.Build(),
			wantErr: true,
		},
		{
			name: "oauth_without_client_secret",
			req: trippypb.CreateLinkRequest_builder{
				Template: proto.String("generic"),
				OauthConfig: trippypb.OAuthConfig_builder{
					AuthUrl:  proto.String("111"),
					ClientId: proto.String("222"),
				}.Build(),
			}.Build(),
			wantErr: true,
		},
		{
			name: "generic_oauth",
			req: trippypb.CreateLinkRequest_builder{
				Template: proto.String("generic"),
				OauthConfig: trippypb.OAuthConfig_builder{
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

func TestGetLinkOAuth(t *testing.T) {
	addr, err := startGRPCServer(secrets.NewTestManager(), "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(addr, creds)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := trippypb.NewTrippyServiceClient(conn)
	resp1, err := client.CreateLink(t.Context(), trippypb.CreateLinkRequest_builder{
		Template: proto.String("generic"),
		OauthConfig: trippypb.OAuthConfig_builder{
			ClientId:     proto.String("111"),
			ClientSecret: proto.String("222"),
		}.Build(),
	}.Build())
	if err != nil {
		t.Fatalf("CreateLink() error = %v", err)
	}

	tests := []struct {
		name       string
		req        *trippypb.GetLinkRequest
		wantID     string
		wantSecret string
		wantErr    bool
	}{
		{
			name:    "missing_id",
			req:     trippypb.GetLinkRequest_builder{}.Build(),
			wantErr: true,
		},
		{
			name: "invalid_id",
			req: trippypb.GetLinkRequest_builder{
				LinkId: proto.String("111"),
			}.Build(),
			wantErr: true,
		},
		{
			name: "link_not_found",
			req: trippypb.GetLinkRequest_builder{
				LinkId: proto.String(shortuuid.New()),
			}.Build(),
			wantErr: true,
		},
		{
			name: "happy_path",
			req: trippypb.GetLinkRequest_builder{
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
	addr, err := startGRPCServer(secrets.NewTestManager(), "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(addr, creds)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := trippypb.NewTrippyServiceClient(conn)
	resp1, err := client.CreateLink(t.Context(), trippypb.CreateLinkRequest_builder{
		Template: proto.String("slack-bot-token"),
	}.Build())
	if err != nil {
		t.Fatalf("CreateLink() error = %v", err)
	}

	got, err := client.GetLink(t.Context(), trippypb.GetLinkRequest_builder{
		LinkId: proto.String(resp1.GetLinkId()),
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
