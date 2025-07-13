package main

import (
	"context"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestCheckLinkIDArg(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no_args",
			args:    []string{"test", "cmd"},
			wantErr: true,
		},
		{
			name:    "multiple_args",
			args:    []string{"test", "cmd", "one", "two"},
			wantErr: true,
		},
		{
			name:    "single_invalid_id",
			args:    []string{"test", "cmd", "111"},
			wantErr: true,
		},
		{
			name: "single_valid_id",
			args: []string{"test", "cmd", "KrRnsTfSR3Pvo6KxUSUV47"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &cli.Command{Name: "test", Commands: []*cli.Command{
				{
					Name: "cmd",
					Action: func(_ context.Context, cmd *cli.Command) error {
						return checkLinkIDArg(cmd)
					},
				},
			}}
			if err := app.Run(t.Context(), tt.args); (err != nil) != tt.wantErr {
				t.Errorf("cli.Command.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
