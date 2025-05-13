package oauth

import (
	"reflect"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/protobuf/proto"

	trippypb "github.com/tzrikka/trippy/proto/trippy/v1"
)

func TestFromProto(t *testing.T) {
	tests := []struct {
		name string
		oac  *trippypb.OAuthConfig
		want *Config
	}{
		{
			name: "nil",
		},
		{
			name: "client_id_and_secret",
			oac: trippypb.OAuthConfig_builder{
				ClientId:     proto.String("id"),
				ClientSecret: proto.String("secret"),
			}.Build(),
			want: &Config{
				Config: &oauth2.Config{
					ClientID:     "id",
					ClientSecret: "secret",
				},
			},
		},
		{
			name: "auth_and _token_urls",
			oac: trippypb.OAuthConfig_builder{
				AuthUrl:  proto.String("auth"),
				TokenUrl: proto.String("token"),
			}.Build(),
			want: &Config{
				Config: &oauth2.Config{
					Endpoint: oauth2.Endpoint{
						AuthURL:  "auth",
						TokenURL: "token",
					},
				},
			},
		},
		{
			name: "scopes",
			oac: trippypb.OAuthConfig_builder{
				Scopes: []string{"111", "222"},
			}.Build(),
			want: &Config{
				Config: &oauth2.Config{
					Scopes: []string{"111", "222"},
				},
			},
		},
		{
			name: "opts",
			oac: trippypb.OAuthConfig_builder{
				AuthCodes: map[string]string{"aaa": "111", "bbb": "222"},
			}.Build(),
			want: &Config{
				Config:    &oauth2.Config{},
				AuthCodes: map[string]string{"aaa": "111", "bbb": "222"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromProto(tt.oac); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromProto() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name string
		oac  *trippypb.OAuthConfig
		want string
	}{
		{
			name: "nil",
			want: "",
		},
		{
			name: "empty",
			oac:  trippypb.OAuthConfig_builder{}.Build(),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.oac); got != tt.want {
				t.Errorf("ToString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TesConfigToProto(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want *trippypb.OAuthConfig
	}{
		{
			name: "nil",
		},
		{
			name: "auth_and_token_urls",
			cfg: &Config{
				Config: &oauth2.Config{
					Endpoint: oauth2.Endpoint{
						AuthURL:  "auth",
						TokenURL: "token",
					},
				},
			},
			want: trippypb.OAuthConfig_builder{
				AuthUrl:      proto.String("auth"),
				TokenUrl:     proto.String("token"),
				AuthStyle:    proto.Int64(0),
				ClientId:     proto.String(""),
				ClientSecret: proto.String(""),
			}.Build(),
		},
		{
			name: "client_id_and_secret",
			cfg: &Config{
				Config: &oauth2.Config{
					ClientID:     "id",
					ClientSecret: "secret",
				},
			},
			want: trippypb.OAuthConfig_builder{
				AuthUrl:      proto.String(""),
				TokenUrl:     proto.String(""),
				AuthStyle:    proto.Int64(0),
				ClientId:     proto.String("id"),
				ClientSecret: proto.String("secret"),
			}.Build(),
		},
		{
			name: "scopes",
			cfg: &Config{
				Config: &oauth2.Config{
					Scopes: []string{"111", "222"},
				},
			},
			want: trippypb.OAuthConfig_builder{
				AuthUrl:      proto.String(""),
				TokenUrl:     proto.String(""),
				AuthStyle:    proto.Int64(0),
				ClientId:     proto.String(""),
				ClientSecret: proto.String(""),
				Scopes:       []string{"111", "222"},
			}.Build(),
		},
		{
			name: "opts",
			cfg: &Config{
				Config:    &oauth2.Config{},
				AuthCodes: map[string]string{"aaa": "111", "bbb": "222"},
			},
			want: trippypb.OAuthConfig_builder{
				AuthUrl:      proto.String(""),
				TokenUrl:     proto.String(""),
				AuthStyle:    proto.Int64(0),
				ClientId:     proto.String(""),
				ClientSecret: proto.String(""),
				AuthCodes:    map[string]string{"aaa": "111", "bbb": "222"},
			}.Build(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.ToProto(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.ToProto() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigToJSON(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		want    string
		wantErr bool
	}{
		{
			name: "nil",
			want: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cfg.ToJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Config.ToJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConfigAuthCodes(t *testing.T) {
	tests := []struct {
		name string
		acs  map[string]string
		want []oauth2.AuthCodeOption
	}{
		{
			name: "nil",
		},
		{
			name: "offline",
			acs: map[string]string{
				"access_type": "offline",
			},
			want: []oauth2.AuthCodeOption{
				oauth2.AccessTypeOffline,
			},
		},
		{
			name: "consent",
			acs: map[string]string{
				"prompt": "consent",
			},
			want: []oauth2.AuthCodeOption{
				oauth2.ApprovalForce,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{AuthCodes: tt.acs}
			if got := c.authCodes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.authCodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenToProto(t *testing.T) {
	tests := []struct {
		name string
		t    *oauth2.Token
		want *trippypb.OAuthToken
	}{
		{
			name: "empty",
			t:    &oauth2.Token{},
			want: trippypb.OAuthToken_builder{
				AccessToken: proto.String(""),
				Expiry:      proto.String("0001-01-01T00:00:00Z"),
			}.Build(),
		},
		{
			name: "endless_access_token",
			t:    &oauth2.Token{AccessToken: "access"},
			want: trippypb.OAuthToken_builder{
				AccessToken: proto.String("access"),
				Expiry:      proto.String("0001-01-01T00:00:00Z"),
			}.Build(),
		},
		{
			name: "expiring_access_token",
			t: &oauth2.Token{
				AccessToken: "access",
				Expiry:      time.Unix(1500000005, 1234),
			},
			want: trippypb.OAuthToken_builder{
				AccessToken: proto.String("access"),
				Expiry:      proto.String("2017-07-14T02:40:05Z"),
			}.Build(),
		},
		{
			name: "all_fields",
			t: &oauth2.Token{
				AccessToken:  "access",
				TokenType:    "bearer",
				RefreshToken: "refresh",
				Expiry:       time.Unix(1500000005, 1234),
				ExpiresIn:    1234,
			},
			want: trippypb.OAuthToken_builder{
				AccessToken:  proto.String("access"),
				Expiry:       proto.String("2017-07-14T02:40:05Z"),
				RefreshToken: proto.String("refresh"),
				TokenType:    proto.String("bearer"),
			}.Build(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TokenToProto(tt.t)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToProto() = %v, want %v", got, tt.want)
			}
		})
	}
}
