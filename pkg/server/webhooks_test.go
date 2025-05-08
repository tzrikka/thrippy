package server

import (
	"testing"
)

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
