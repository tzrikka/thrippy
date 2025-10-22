package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectURL(t *testing.T) {
	tests := []struct {
		name        string
		webhookAddr string
		want        string
	}{
		{
			name: "empty",
		},
		{
			name:        "foo",
			webhookAddr: "foo",
			want:        "https://foo/callback",
		},
		{
			name:        "foo.bar",
			webhookAddr: "foo.bar",
			want:        "https://foo.bar/callback",
		},
		{
			name:        "full_http_url",
			webhookAddr: "http://example.com/blah",
			want:        "https://example.com/callback",
		},
		{
			name:        "full_https_url",
			webhookAddr: "https://example.com/blah",
			want:        "https://example.com/callback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := redirectURL(tt.webhookAddr); got != tt.want {
				t.Errorf("redirectURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTMLResponse(t *testing.T) {
	tests := []struct {
		name   string
		status int
		msg    string
	}{
		{
			name:   "ok",
			status: http.StatusOK,
		},
		{
			name:   "bad_request",
			status: http.StatusBadRequest,
		},
		{
			name:   "internal_server_error",
			status: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				htmlResponse(w, tt.status, tt.msg)
			}))
			defer s.Close()

			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, s.URL, http.NoBody)
			if err != nil {
				t.Errorf("failed to create HTTP request: %v", err)
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("HTTP request error = %v", err)
				return
			}

			if resp.StatusCode != tt.status {
				t.Errorf("htmlResponse() status = %d, want %d", resp.StatusCode, tt.status)
			}

			_ = resp.Body.Close()
		})
	}
}

func TestConstructStateParam(t *testing.T) {
	tests := []struct {
		name string
		memo string
		want string
	}{
		{
			name: "without_memo",
			want: "id_nonce",
		},
		{
			name: "with_memo",
			memo: "memo",
			want: "id_nonce_memo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := constructStateParam("id", "nonce", tt.memo); got != tt.want {
				t.Errorf("constructStateParam() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseStateParam(t *testing.T) {
	tests := []struct {
		name      string
		state     string
		wantID    string
		wantNonce string
		wantMemo  string
		wantErr   bool
	}{
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name:     "memo_only",
			state:    "__memo",
			wantMemo: "memo",
			wantErr:  true,
		},
		{
			name:      "no_id",
			state:     "_nonce_memo",
			wantNonce: "nonce",
			wantMemo:  "memo",
			wantErr:   true,
		},
		{
			name:     "no_nonce",
			state:    "id__memo",
			wantID:   "id",
			wantMemo: "memo",
			wantErr:  true,
		},
		{
			name:    "valid_id_missing_nonce",
			state:   "AQYywDkK3hiH9FEERA3aU5",
			wantID:  "AQYywDkK3hiH9FEERA3aU5",
			wantErr: true,
		},
		{
			name:     "valid_id_and_simple_memo_missing_nonce",
			state:    "AQYywDkK3hiH9FEERA3aU5__memo",
			wantID:   "AQYywDkK3hiH9FEERA3aU5",
			wantMemo: "memo",
			wantErr:  true,
		},
		{
			name:     "valid_id_and_complex_memo_missing_nonce",
			state:    "AQYywDkK3hiH9FEERA3aU5__foo_bar",
			wantID:   "AQYywDkK3hiH9FEERA3aU5",
			wantMemo: "foo_bar",
			wantErr:  true,
		},
		{
			name:    "invalid_id_only",
			state:   "111",
			wantID:  "111",
			wantErr: true,
		},
		{
			name:     "invalid_id_and_simple_memo",
			state:    "111__memo",
			wantID:   "111",
			wantMemo: "memo",
			wantErr:  true,
		},
		{
			name:      "all_valid",
			state:     "AQYywDkK3hiH9FEERA3aU5_X8cbAvTF2M2crW9YrfVMoB_nonce_memo",
			wantID:    "AQYywDkK3hiH9FEERA3aU5",
			wantNonce: "X8cbAvTF2M2crW9YrfVMoB",
			wantMemo:  "nonce_memo",
		},
		{
			name:      "all_invalid",
			state:     "111_222_memo",
			wantID:    "111",
			wantNonce: "222",
			wantMemo:  "memo",
			wantErr:   true,
		},
		{
			name:      "valid_nonce_and_simple_memo",
			state:     "_X8cbAvTF2M2crW9YrfVMoB_memo",
			wantNonce: "X8cbAvTF2M2crW9YrfVMoB",
			wantMemo:  "memo",
			wantErr:   true,
		},
		{
			name:      "valid_nonce_and_complex_memo",
			state:     "_X8cbAvTF2M2crW9YrfVMoB_foo_bar",
			wantNonce: "X8cbAvTF2M2crW9YrfVMoB",
			wantMemo:  "foo_bar",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotNonce, gotMemo, err := parseStateParam(tt.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStateParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("parseStateParam() got ID = %q, want %q", gotID, tt.wantID)
			}
			if gotNonce != tt.wantNonce {
				t.Errorf("parseStateParam() got nonce = %q, want %q", gotNonce, tt.wantNonce)
			}
			if gotMemo != tt.wantMemo {
				t.Errorf("parseStateParam() got memo = %q, want %q", gotMemo, tt.wantMemo)
			}
		})
	}
}
