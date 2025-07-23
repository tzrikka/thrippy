package client

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

type server struct {
	thrippypb.UnimplementedThrippyServiceServer
	resp *thrippypb.GetLinkResponse
	err  error
}

func (s *server) GetLink(_ context.Context, _ *thrippypb.GetLinkRequest) (*thrippypb.GetLinkResponse, error) {
	return s.resp, s.err
}

func TestLinkOAuthConfig(t *testing.T) {
	tests := []struct {
		name    string
		resp    *thrippypb.GetLinkResponse
		respErr error
		want    *oauth.Config
		wantErr bool
	}{
		{
			name: "nil",
		},
		{
			name:    "grpc_error",
			respErr: errors.New("error"),
			wantErr: true,
		},
		{
			name: "invalid_oauth_error",
			resp: thrippypb.GetLinkResponse_builder{
				Template:    proto.String("template"),
				OauthConfig: thrippypb.OAuthConfig_builder{}.Build(),
			}.Build(),
			wantErr: true,
		},
		{
			name:    "not_found",
			respErr: status.Error(codes.NotFound, "link not found"),
		},
		{
			name: "happy_path",
			resp: thrippypb.GetLinkResponse_builder{
				Template: proto.String("template"),
				OauthConfig: thrippypb.OAuthConfig_builder{
					AuthUrl:      proto.String("111"),
					ClientId:     proto.String("222"),
					ClientSecret: proto.String("333"),
				}.Build(),
			}.Build(),
			want: &oauth.Config{
				Config: &oauth2.Config{
					Endpoint: oauth2.Endpoint{
						AuthURL: "111",
					},
					ClientID:     "222",
					ClientSecret: "333",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := net.ListenConfig{}
			lis, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer lis.Close()

			s := grpc.NewServer()
			thrippypb.RegisterThrippyServiceServer(s, &server{resp: tt.resp, err: tt.respErr})
			go func() {
				_ = s.Serve(lis)
			}()

			got, err := LinkOAuthConfig(t.Context(), lis.Addr().String(), insecureCreds(), "link ID")
			if (err != nil) != tt.wantErr {
				t.Errorf("LinkOAuthConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LinkOAuthConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
