package templates

import (
	"reflect"
	"testing"
)

func TestTemplateCredFields(t *testing.T) {
	tests := []struct {
		name       string
		credFields []string
		want       []string
	}{
		{
			name: "nil",
		},
		{
			name:       "empty",
			credFields: []string{},
		},
		{
			name:       "one_element",
			credFields: []string{"one"},
			want:       []string{"one"},
		},
		{
			name:       "five_elements",
			credFields: []string{"1", "2", "3", "4", "5"},
			want:       []string{"1", "2", "3", "4", "5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := Template{credFields: tt.credFields}
			if got := tr.CredFields(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Template.CredFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeMetadataAsJSON(t *testing.T) {
	tests := []struct {
		name    string
		v       any
		want    string
		wantErr bool
	}{
		{
			name: "nil",
			v:    nil,
			want: "null\n",
		},
		{
			name: "empty",
			v:    struct{}{},
			want: "{}\n",
		},
		{
			name: "simple",
			v: struct {
				Visible string `json:"visible"`
				hidden  string
			}{
				Visible: "good",
				hidden:  "bad",
			},
			want: `{"visible":"good"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeMetadataAsJSON(tt.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeMetadataAsJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EncodeMetadataAsJSON() = %q, want %q", got, tt.want)
			}
		})
	}
}
