// Package server implements Thrippy's gRPC service,
// and an HTTP server for OAuth webhooks.
package server

import (
	"context"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/tzrikka/thrippy/pkg/secrets"
)

// Start initializes Thrippy's gRPC and HTTP servers, and the default logger.
func Start(ctx context.Context, cmd *cli.Command) error {
	var handler slog.Handler
	if cmd.Bool("dev") {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
		})
	}

	slog.SetDefault(slog.New(handler))
	if cmd.Bool("dev") {
		slog.Warn("********** DEV MODE - UNSAFE IN PRODUCTION! **********")
	}

	sm, err := secrets.NewManager(ctx, cmd)
	if err != nil {
		return err
	}

	if _, err := startGRPCServer(ctx, cmd, sm); err != nil {
		return err
	}

	return newHTTPServer(ctx, cmd).run()
}
