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

	"github.com/tzrikka/trippy/pkg/oauth"
	trippypb "github.com/tzrikka/trippy/proto/trippy/v1"
)

type server struct {
	trippypb.UnimplementedTrippyServiceServer
	resp *trippypb.GetLinkResponse
	err  error
}

func (s *server) GetLink(_ context.Context, _ *trippypb.GetLinkRequest) (*trippypb.GetLinkResponse, error) {
	return s.resp, s.err
}

func TestLinkOAuthConfig(t *testing.T) {
	tests := []struct {
		name    string
		resp    *trippypb.GetLinkResponse
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
			resp: trippypb.GetLinkResponse_builder{
				Template:    proto.String("template"),
				OauthConfig: trippypb.OAuthConfig_builder{}.Build(),
			}.Build(),
			wantErr: true,
		},
		{
			name:    "not_found",
			respErr: status.Error(codes.NotFound, "link not found"),
		},
		{
			name: "happy_path",
			resp: trippypb.GetLinkResponse_builder{
				Template: proto.String("template"),
				OauthConfig: trippypb.OAuthConfig_builder{
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
			lis, err := net.Listen("tcp", ":0")
			if err != nil {
				t.Fatal(err)
			}
			s := grpc.NewServer()
			trippypb.RegisterTrippyServiceServer(s, &server{resp: tt.resp, err: tt.respErr})
			go func() {
				s.Serve(lis)
			}()

			got, err := LinkOAuthConfig(t.Context(), lis.Addr().String(), InsecureCreds(), "link ID")
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
