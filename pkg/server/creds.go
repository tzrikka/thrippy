package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log/slog"
	"os"

	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/tzrikka/thrippy/internal/logger"
)

// GRPCFlags defines global (but hidden) CLI flags. The purpose of
// these CLI flags is to initialize the gRPC server in non-dev mode via
// environment variables and/or the application's configuration file.
// See also [client.GRPCFlags].
//
// [client.GRPCFlags]: https://pkg.go.dev/github.com/tzrikka/thrippy/pkg/client#GRPCFlags
func GRPCFlags(configFilePath altsrc.StringSourcer) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name: "grpc-client-ca-cert", // Only mTLS.
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_GRPC_CLIENT_CA_CERT"),
				toml.TOML("grpc.server.client_ca_cert", configFilePath),
			),
			Hidden:    true,
			TakesFile: true,
		},
		&cli.StringFlag{
			Name: "grpc-server-cert", // Both TLS and mTLS.
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_GRPC_SERVER_CERT"),
				toml.TOML("grpc.server.server_cert", configFilePath),
			),
			Hidden:    true,
			TakesFile: true,
		},
		&cli.StringFlag{
			Name: "grpc-server-key", // Both TLS and mTLS.
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_GRPC_SERVER_KEY"),
				toml.TOML("grpc.server.server_key", configFilePath),
			),
			Hidden:    true,
			TakesFile: true,
		},
	}
}

// GRPCCreds initializes gRPC server credentials, based on CLI flags.
// Errors here will abort the application with a log message.
// See also [client.GRPCCreds].
//
// [client.GRPCCreds]: https://pkg.go.dev/github.com/tzrikka/thrippy/pkg/client#GRPCCreds
func GRPCCreds(ctx context.Context, cmd *cli.Command) []grpc.ServerOption {
	if cmd.Bool("dev") {
		return nil
	}

	// Either TLS and mTLS.
	certPath := cmd.String("grpc-server-cert")
	keyPath := cmd.String("grpc-server-key")
	// Only mTLS.
	caPath := cmd.String("grpc-client-ca-cert")

	// The server's X.509 PEM-encoded public cert and private key are required
	// for both TLS and mTLS. If either of them is missing it's an error.
	if certPath == "" {
		logger.Fatal(ctx, "missing server public cert file for gRPC client with m/TLS")
	}
	if keyPath == "" {
		logger.Fatal(ctx, "missing server private key file for gRPC client with m/TLS")
	}

	// Using mTLS requires the client's CA cert (on many Linux systems,
	// "/etc/ssl/cert.pem" contains the system-wide set of root CAs).
	// If it's missing, we use TLS.
	if caPath == "" {
		msg := "gRPC server with TLS"
		creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
		if err != nil {
			logger.FatalError(ctx, "failed to create credentials for "+msg, err)
		}
		slog.Info("using "+msg, slog.String("cert", certPath), slog.String("key", keyPath))
		return []grpc.ServerOption{grpc.Creds(creds)}
	}

	// If all 3 are specified, we use mTLS.
	msg := "client CA cert file for gRPC server with mTLS"
	ca := x509.NewCertPool()
	pem, err := os.ReadFile(caPath) //gosec:disable G304 -- specified by admin by design
	if err != nil {
		logger.FatalError(ctx, "failed to read "+msg, err, slog.String("path", caPath))
	}
	if ok := ca.AppendCertsFromPEM(pem); !ok {
		logger.Fatal(ctx, "failed to parse "+msg, slog.String("path", caPath))
	}

	msg = "gRPC server with mTLS"
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		logger.FatalError(ctx, "failed to load server PEM key pair for "+msg,
			err, slog.String("cert", certPath), slog.String("key", keyPath))
	}

	slog.Info("using "+msg, slog.String("server_cert", certPath), slog.String("server_key", keyPath),
		slog.String("client_ca_cert", caPath))

	return []grpc.ServerOption{grpc.Creds(credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    ca,
		MinVersion:   tls.VersionTLS13,
	}))}
}
