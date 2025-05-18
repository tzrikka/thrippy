package slack

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type httpTestResponse struct {
	slackResponse
	Response string `json:"response,omitempty"`
}

func TestGet(t *testing.T) {
	tests := []struct {
		name        string
		startServer bool
		respBody    string
		wantErr     bool
		wantResp    string
	}{
		{
			name:        "happy_path",
			startServer: true,
			respBody:    `{"ok": true, "response": "response"}`,
			wantResp:    "response",
		},
		{
			name:        "bad_response",
			startServer: true,
			respBody:    "bad",
			wantErr:     true,
		},
		{
			name:        "slack_not_ok",
			startServer: true,
			respBody:    `{"ok": false}`,
		},
		{
			name:    "server_not_responding",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewUnstartedServer(handler(t, tt.respBody))
			if tt.startServer {
				s.Start()
			}
			defer s.Close()

			got := &httpTestResponse{}
			err := get(t.Context(), s.URL, "token", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantResp != "" && got.Response != tt.wantResp {
				t.Errorf("get() response = %v, want %q", got, tt.wantResp)
			}
		})
	}
}

func TestPost(t *testing.T) {
	tests := []struct {
		name        string
		startServer bool
		respBody    string
		wantErr     bool
		wantResp    string
	}{
		{
			name:        "happy_path",
			startServer: true,
			respBody:    `{"ok": true, "response": "response"}`,
			wantResp:    "response",
		},
		{
			name:        "bad_response",
			startServer: true,
			respBody:    "bad",
			wantErr:     true,
		},
		{
			name:        "slack_not_ok",
			startServer: true,
			respBody:    `{"ok": false}`,
		},
		{
			name:    "server_not_responding",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewUnstartedServer(handler(t, tt.respBody))
			if tt.startServer {
				s.Start()
			}
			defer s.Close()

			got := &httpTestResponse{}
			err := post(t.Context(), s.URL, "token", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("post() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantResp != "" && got.Response != tt.wantResp {
				t.Errorf("post() response = %v, want %q", got, tt.wantResp)
			}
		})
	}
}

func handler(t *testing.T, resp string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Authorization")
		want := "Bearer token"
		if got != want {
			t.Errorf("authorization header = %q, want %q", got, want)
		}

		n, err := fmt.Fprint(w, resp)
		if err != nil {
			t.Errorf("failed to write resp: %v", err)
		}
		if n != len(resp) {
			t.Errorf("wrote %d resp bytes, want %d", n, len(resp))
		}
	})
}
