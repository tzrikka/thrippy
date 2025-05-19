package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
			want: "id",
		},
		{
			name: "with_memo",
			memo: "memo",
			want: "id_memo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := constructStateParam("id", tt.memo); got != tt.want {
				t.Errorf("constructStateParam() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseStateParam(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		wantID   string
		wantMemo string
		wantErr  bool
	}{
		{
			name: "empty",
		},
		{
			name:     "no_id",
			state:    "_memo",
			wantMemo: "memo",
		},
		{
			name:   "valid_id_only",
			state:  "AQYywDkK3hiH9FEERA3aU5",
			wantID: "AQYywDkK3hiH9FEERA3aU5",
		},
		{
			name:     "valid_id_and_simple_memo",
			state:    "AQYywDkK3hiH9FEERA3aU5_memo",
			wantID:   "AQYywDkK3hiH9FEERA3aU5",
			wantMemo: "memo",
		},
		{
			name:     "valid_id_and_complex_memo",
			state:    "AQYywDkK3hiH9FEERA3aU5_foo_bar",
			wantID:   "AQYywDkK3hiH9FEERA3aU5",
			wantMemo: "foo_bar",
		},
		{
			name:    "invalid_id_only",
			state:   "111",
			wantID:  "111",
			wantErr: true,
		},
		{
			name:     "invalid_id_and_simple_memo",
			state:    "111_memo",
			wantID:   "111",
			wantMemo: "memo",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotMemo, err := parseStateParam(tt.state)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStateParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID != tt.wantID {
				t.Errorf("parseStateParam() got ID = %q, want %q", gotID, tt.wantID)
			}
			if gotMemo != tt.wantMemo {
				t.Errorf("parseStateParam() got memo = %q, want %q", gotMemo, tt.wantMemo)
			}
		})
	}
}
