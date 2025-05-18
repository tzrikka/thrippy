package github

import (
	"testing"
)

func Test_normalizeRFC3339(t *testing.T) {
	tests := []struct {
		name string
		t    string
		want string
	}{
		{
			name: "no_replacements",
			t:    "2025-12-21T20:19:18Z",
			want: "2025-12-21T20:19:18Z",
		},
		{
			name: "remove_millisecs",
			t:    "2025-12-21T20:19:18.123Z",
			want: "2025-12-21T20:19:18Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeRFC3339(tt.t); got != tt.want {
				t.Errorf("normalizeRFC3339() = %q, want %q", got, tt.want)
			}
		})
	}
}
