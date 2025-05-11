package links

import (
	"reflect"
	"testing"
)

func TestCredFields(t *testing.T) {
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
