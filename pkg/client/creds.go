package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log/slog"
	"os"

	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/tzrikka/thrippy/internal/logger"
)

// insecureCreds should be used only in dev mode and unit tests.
func insecureCreds() credentials.TransportCredentials {
	return insecure.NewCredentials()
}

// GRPCFlags defines global (but hidden) CLI flags. The purpose of
// these CLI flags is to initialize gRPC clients in non-dev mode via
// environment variables and/or the application's configuration file.
// See also [server.GRPCFlags].
//
// [server.GRPCFlags]: https://pkg.go.dev/github.com/tzrikka/thrippy/pkg/server#GRPCFlags
func GRPCFlags(configFilePath altsrc.StringSourcer) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name: "grpc-client-cert", // Only mTLS.
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_GRPC_CLIENT_CERT"),
				toml.TOML("grpc.client.client_cert", configFilePath),
			),
			Hidden:    true,
			TakesFile: true,
		},
		&cli.StringFlag{
			Name: "grpc-client-key", // Only mTLS.
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_GRPC_CLIENT_KEY"),
				toml.TOML("grpc.client.client_key", configFilePath),
			),
			Hidden:    true,
			TakesFile: true,
		},
		&cli.StringFlag{
			Name: "grpc-server-ca-cert", // Both TLS and mTLS.
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_GRPC_SERVER_CA_CERT"),
				toml.TOML("grpc.client.server_ca_cert", configFilePath),
			),
			Hidden:    true,
			TakesFile: true,
		},
		&cli.StringFlag{
			Name: "grpc-server-name-override", // Both TLS and mTLS, but only for testing.
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_GRPC_SERVER_NAME_OVERRIDE"),
				toml.TOML("grpc.client.server_name_override", configFilePath),
			),
			Hidden: true,
		},
	}
}

// GRPCCreds initializes gRPC client credentials, based on CLI flags.
// Errors here will abort the application with a log message.
// See also [server.GRPCCreds].
//
// [server.GRPCCreds]: https://pkg.go.dev/github.com/tzrikka/thrippy/pkg/server#GRPCCreds
func GRPCCreds(ctx context.Context, cmd *cli.Command) credentials.TransportCredentials {
	if cmd.Bool("dev") {
		return insecureCreds()
	}

	// Either TLS and mTLS.
	caPath := cmd.String("grpc-server-ca-cert")
	nameOverride := cmd.String("grpc-server-name-override")
	// Only mTLS.
	certPath := cmd.String("grpc-client-cert")
	keyPath := cmd.String("grpc-client-key")

	// The server's CA cert is required either way (on many Linux systems,
	// "/etc/ssl/cert.pem" contains the system-wide set of root CAs).
	if caPath == "" {
		logger.Fatal(ctx, "missing server CA cert file for gRPC client with m/TLS")
	}

	// Using mTLS requires the client's X.509 PEM-encoded public cert
	// and private key. If one of them is missing it's an error.
	if certPath == "" && keyPath != "" {
		logger.Fatal(ctx, "missing client public cert file for gRPC client with mTLS")
	}
	if certPath != "" && keyPath == "" {
		logger.Fatal(ctx, "missing client private key file for gRPC client with mTLS")
	}

	// If both of them are missing, we use TLS.
	if certPath == "" && keyPath == "" {
		return newClientTLSFromFile(ctx, caPath, nameOverride, nil)
	}

	// If all 3 are specified, we use mTLS.
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		logger.FatalError(ctx, "failed to load client PEM key pair for gRPC client with mTLS", err,
			slog.String("cert", certPath), slog.String("key", keyPath))
	}

	return newClientTLSFromFile(ctx, caPath, nameOverride, []tls.Certificate{cert})
}

// newClientTLSFromFile constructs TLS credentials from the provided root
// certificate authority certificate file(s) to validate server connections.
//
// This function is based on [credentials.NewClientTLSFromFile], but uses
// TLS 1.3 as the minimum version (instead of 1.2), and support mTLS too.
func newClientTLSFromFile(ctx context.Context, caPath, serverNameOverride string, certs []tls.Certificate) credentials.TransportCredentials {
	b, err := os.ReadFile(caPath) //gosec:disable G304 // Specified by admin by design.
	if err != nil {
		logger.FatalError(ctx, "failed to read server CA cert file for gRPC client", err, slog.String("path", caPath))
	}

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		logger.Fatal(ctx, "failed to parse server CA cert file for gRPC client", slog.String("path", caPath))
	}

	cfg := &tls.Config{
		RootCAs:    cp,
		ServerName: serverNameOverride,
		MinVersion: tls.VersionTLS13,
	}
	if len(certs) > 0 {
		cfg.Certificates = certs
	}

	return credentials.NewTLS(cfg)
}
