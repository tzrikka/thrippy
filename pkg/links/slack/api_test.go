package slack

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type httpTestResponse struct {
	response

	Response string `json:"response,omitempty"`
}

func TestGetPost(t *testing.T) {
	tests := []struct {
		name        string
		funcName    string
		startServer bool
		respBody    string
		wantErr     bool
		wantResp    string
	}{
		// Get.
		{
			name:        "get_happy_path",
			funcName:    "get",
			startServer: true,
			respBody:    `{"ok": true, "response": "response"}`,
			wantResp:    "response",
		},
		{
			name:        "get_bad_response",
			funcName:    "get",
			startServer: true,
			respBody:    "bad",
			wantErr:     true,
		},
		{
			name:        "slack_not_ok",
			funcName:    "get",
			startServer: true,
			respBody:    `{"ok": false}`,
		},
		{
			name:     "get_server_not_responding",
			funcName: "get",
			wantErr:  true,
		},

		// Post.
		{
			name:        "post_happy_path",
			funcName:    "post",
			startServer: true,
			respBody:    `{"ok": true, "response": "response"}`,
			wantResp:    "response",
		},
		{
			name:        "post_bad_response",
			funcName:    "post",
			startServer: true,
			respBody:    "bad",
			wantErr:     true,
		},
		{
			name:        "post_slack_not_ok",
			funcName:    "post",
			startServer: true,
			respBody:    `{"ok": false}`,
		},
		{
			name:     "post_server_not_responding",
			funcName: "post",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewUnstartedServer(handler(t, tt.respBody))
			if tt.startServer {
				s.Start()
			}
			defer s.Close()

			fn := get
			if tt.funcName == "post" {
				fn = post
			}

			got := &httpTestResponse{}
			err := fn(t.Context(), s.URL, "token", got)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s() error = %v, wantErr %v", tt.funcName, err, tt.wantErr)
				return
			}
			if tt.wantResp != "" && got.Response != tt.wantResp {
				t.Errorf("%s() response = %v, want %q", tt.funcName, got, tt.wantResp)
			}
		})
	}
}

func handler(t *testing.T, resp string) http.HandlerFunc {
	t.Helper()

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
