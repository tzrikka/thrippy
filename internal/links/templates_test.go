package links

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
)

func TestTemplateCredFields(t *testing.T) {
	tests := []struct {
		name       string
		credFields []string
		want       []*thrippypb.CredentialField
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
			want: []*thrippypb.CredentialField{
				thrippypb.CredentialField_builder{Name: proto.String("one")}.Build(),
			},
		},
		{
			name:       "five_simple_elements",
			credFields: []string{"1", "2", "3", "4", "5"},
			want: []*thrippypb.CredentialField{
				thrippypb.CredentialField_builder{Name: proto.String("1")}.Build(),
				thrippypb.CredentialField_builder{Name: proto.String("2")}.Build(),
				thrippypb.CredentialField_builder{Name: proto.String("3")}.Build(),
				thrippypb.CredentialField_builder{Name: proto.String("4")}.Build(),
				thrippypb.CredentialField_builder{Name: proto.String("5")}.Build(),
			},
		},
		{
			name:       "manual_field",
			credFields: []string{"name_manual"},
			want: []*thrippypb.CredentialField{
				thrippypb.CredentialField_builder{
					Name:   proto.String("name"),
					Manual: proto.Bool(true),
				}.Build(),
			},
		},
		{
			name:       "optional_field",
			credFields: []string{"name_optional"},
			want: []*thrippypb.CredentialField{
				thrippypb.CredentialField_builder{
					Name:     proto.String("name"),
					Optional: proto.Bool(true),
				}.Build(),
			},
		},
		{
			name:       "manual_and_optional_field",
			credFields: []string{"name_manual_optional"},
			want: []*thrippypb.CredentialField{
				thrippypb.CredentialField_builder{
					Name:     proto.String("name"),
					Manual:   proto.Bool(true),
					Optional: proto.Bool(true),
				}.Build(),
			},
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

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
		wantErr bool
	}{
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name:    "arbitrary_string",
			baseURL: "string",
			wantErr: true,
		},
		{
			name:    "scheme_only",
			baseURL: "http://",
			wantErr: true,
		},
		{
			name:    "no_host",
			baseURL: "https:///foo",
			wantErr: true,
		},
		{
			name:    "valid_base_url",
			baseURL: "https://foo/",
			want:    "https://foo",
		},
		{
			name:    "valid_url_with_suffixes",
			baseURL: "https://foo.bar.com/dir/subdir?key=value#frag",
			want:    "https://foo.bar.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeBaseURL(tt.baseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeBaseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NormalizeBaseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
