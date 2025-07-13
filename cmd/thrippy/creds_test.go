package main

import (
	"reflect"
	"testing"
)

func TestReadFiles(t *testing.T) {
	tests := []struct {
		name    string
		m       map[string]string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "nil",
		},
		{
			name: "empty",
			m:    map[string]string{},
			want: map[string]string{},
		},
		{
			name: "not_a_path",
			m:    map[string]string{"key": "val"},
			want: map[string]string{"key": "val"},
		},
		{
			name:    "bad_file_path",
			m:       map[string]string{"key": "@val"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readFiles(tt.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
