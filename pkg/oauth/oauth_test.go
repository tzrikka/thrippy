package oauth

import (
	"reflect"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/protobuf/proto"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
)

func TestFromProto(t *testing.T) {
	tests := []struct {
		name string
		oac  *thrippypb.OAuthConfig
		want *Config
	}{
		{
			name: "nil",
		},
		{
			name: "client_id_and_secret",
			oac: thrippypb.OAuthConfig_builder{
				ClientId:     new("id"),
				ClientSecret: new("secret"),
			}.Build(),
			want: &Config{
				Config: &oauth2.Config{
					ClientID:     "id",
					ClientSecret: "secret",
				},
			},
		},
		{
			name: "auth_and_token_urls",
			oac: thrippypb.OAuthConfig_builder{
				AuthUrl:  new("auth"),
				TokenUrl: new("token"),
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
			oac: thrippypb.OAuthConfig_builder{
				Scopes: []string{"111", "222"},
			}.Build(),
			want: &Config{
				Config: &oauth2.Config{
					Scopes: []string{"111", "222"},
				},
			},
		},
		{
			name: "auth_codes",
			oac: thrippypb.OAuthConfig_builder{
				AuthCodes: map[string]string{"aaa": "111", "bbb": "222"},
			}.Build(),
			want: &Config{
				Config:    &oauth2.Config{},
				AuthCodes: map[string]string{"aaa": "111", "bbb": "222"},
			},
		},
		{
			name: "params",
			oac: thrippypb.OAuthConfig_builder{
				Params: map[string]string{"aaa": "111", "bbb": "222"},
			}.Build(),
			want: &Config{
				Config: &oauth2.Config{},
				Params: map[string]string{"aaa": "111", "bbb": "222"},
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
		oac  *thrippypb.OAuthConfig
		want string
	}{
		{
			name: "nil",
			want: "",
		},
		{
			name: "empty",
			oac:  thrippypb.OAuthConfig_builder{}.Build(),
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

func TestConfigIsUsable(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want bool
	}{
		{
			name: "nil",
		},
		{
			name: "empty",
			cfg:  &Config{Config: &oauth2.Config{}},
		},
		{
			name: "auth_url",
			cfg: &Config{Config: &oauth2.Config{
				Endpoint: oauth2.Endpoint{AuthURL: "auth_url"},
			}},
			want: true,
		},
		{
			name: "client_id",
			cfg: &Config{Config: &oauth2.Config{
				ClientID: "client_id",
			}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.IsUsable(); got != tt.want {
				t.Errorf("IsUsable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigToProto(t *testing.T) {
	tests := []struct {
		name string
		cfg  *Config
		want *thrippypb.OAuthConfig
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
			want: thrippypb.OAuthConfig_builder{
				AuthUrl:      new("auth"),
				TokenUrl:     new("token"),
				AuthStyle:    proto.Int64(0),
				ClientId:     new(""),
				ClientSecret: new(""),
				Nonce:        new(""),
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
			want: thrippypb.OAuthConfig_builder{
				AuthUrl:      new(""),
				TokenUrl:     new(""),
				AuthStyle:    proto.Int64(0),
				ClientId:     new("id"),
				ClientSecret: new("secret"),
				Nonce:        new(""),
			}.Build(),
		},
		{
			name: "scopes",
			cfg: &Config{
				Config: &oauth2.Config{
					Scopes: []string{"111", "222"},
				},
			},
			want: thrippypb.OAuthConfig_builder{
				AuthUrl:      new(""),
				TokenUrl:     new(""),
				AuthStyle:    proto.Int64(0),
				ClientId:     new(""),
				ClientSecret: new(""),
				Scopes:       []string{"111", "222"},
				Nonce:        new(""),
			}.Build(),
		},
		{
			name: "auth_codes",
			cfg: &Config{
				Config:    &oauth2.Config{},
				AuthCodes: map[string]string{"aaa": "111", "bbb": "222"},
			},
			want: thrippypb.OAuthConfig_builder{
				AuthUrl:      new(""),
				TokenUrl:     new(""),
				AuthStyle:    proto.Int64(0),
				ClientId:     new(""),
				ClientSecret: new(""),
				AuthCodes:    map[string]string{"aaa": "111", "bbb": "222"},
				Nonce:        new(""),
			}.Build(),
		},
		{
			name: "nonce",
			cfg: &Config{
				Config: &oauth2.Config{},
				Nonce:  "nonce",
			},
			want: thrippypb.OAuthConfig_builder{
				AuthUrl:      new(""),
				TokenUrl:     new(""),
				AuthStyle:    proto.Int64(0),
				ClientId:     new(""),
				ClientSecret: new(""),
				Nonce:        new("nonce"),
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
		want *thrippypb.OAuthToken
	}{
		{
			name: "empty",
			t:    &oauth2.Token{},
			want: thrippypb.OAuthToken_builder{
				AccessToken: new(""),
				Expiry:      new("0001-01-01T00:00:00Z"),
			}.Build(),
		},
		{
			name: "endless_access_token",
			t:    &oauth2.Token{AccessToken: "access"},
			want: thrippypb.OAuthToken_builder{
				AccessToken: new("access"),
				Expiry:      new("0001-01-01T00:00:00Z"),
			}.Build(),
		},
		{
			name: "expiring_access_token",
			t: &oauth2.Token{
				AccessToken: "access",
				Expiry:      time.Unix(1500000005, 1234),
			},
			want: thrippypb.OAuthToken_builder{
				AccessToken: new("access"),
				Expiry:      new("2017-07-14T02:40:05Z"),
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
			want: thrippypb.OAuthToken_builder{
				AccessToken:  new("access"),
				Expiry:       new("2017-07-14T02:40:05Z"),
				RefreshToken: new("refresh"),
				TokenType:    new("bearer"),
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

func TestTokenFromMap(t *testing.T) {
	tests := []struct {
		name      string
		m         map[string]any
		wantToken *oauth2.Token
		wantOK    bool
	}{
		{
			name: "nil",
		},
		{
			name: "access_token_without_expiry",
			m: map[string]any{
				"access_token": "access_token",
			},
			wantToken: &oauth2.Token{
				AccessToken: "access_token",
			},
			wantOK: true,
		},
		{
			name: "access_token_with_zero_expiry",
			m: map[string]any{
				"access_token": "access_token",
				"expiry":       "0001-01-01T00:00:00Z",
			},
			wantToken: &oauth2.Token{
				AccessToken: "access_token",
			},
			wantOK: true,
		},
		{
			name: "access_and_refresh_tokens",
			m: map[string]any{
				"access_token":  "access_token",
				"expiry":        "2024-12-06T03:02:01Z",
				"refresh_token": "refresh_token",
			},
			wantToken: &oauth2.Token{
				AccessToken:  "access_token",
				Expiry:       time.Unix(1733454121, 0).UTC(),
				RefreshToken: "refresh_token",
			},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, gotOK := TokenFromMap(tt.m)
			if !reflect.DeepEqual(gotToken, tt.wantToken) {
				t.Errorf("TokenFromMap() got = %#v, want %#v", gotToken, tt.wantToken)
			}
			if gotOK != tt.wantOK {
				t.Errorf("TokenFromMap() ok = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}
