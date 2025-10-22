// Package server implements Thrippy's gRPC service,
// and an HTTP server for OAuth webhooks.
package server

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/urfave/cli/v3"

	"github.com/tzrikka/thrippy/pkg/secrets"
)

// Start initializes Thrippy's gRPC and HTTP servers, and logging.
func Start(ctx context.Context, cmd *cli.Command) error {
	initLog(cmd.Bool("dev"))

	sm, err := secrets.NewManager(ctx, cmd)
	if err != nil {
		return err
	}

	if _, err := startGRPCServer(ctx, cmd, sm); err != nil {
		return err
	}

	return newHTTPServer(cmd).run()
}

// initLog initializes the logger for the Thrippy server,
// based on whether it's running in development mode or not.
func initLog(devMode bool) {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	if !devMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Caller().Logger()
		return
	}

	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05.000",
	}).With().Caller().Logger()

	log.Warn().Msg("********** DEV MODE - UNSAFE IN PRODUCTION! **********")
}
