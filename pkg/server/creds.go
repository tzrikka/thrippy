package server

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/rs/zerolog/log"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
func GRPCCreds(cmd *cli.Command) []grpc.ServerOption {
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
		log.Fatal().Msg("missing server public cert file for gRPC client with m/TLS")
	}
	if keyPath == "" {
		log.Fatal().Msg("missing server private key file for gRPC client with m/TLS")
	}

	// Using mTLS requires the client's CA cert (on many Linux systems,
	// "/etc/ssl/cert.pem" contains the system-wide set of root CAs).
	// If it's missing, we use TLS.
	if caPath == "" {
		msg := "gRPC server with TLS"
		creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create credentials for " + msg)
		}
		log.Info().Str("cert", certPath).Str("key", keyPath).Msg("using " + msg)
		return []grpc.ServerOption{grpc.Creds(creds)}
	}

	// If all 3 are specified, we use mTLS.
	msg := "client CA cert file for gRPC server with mTLS"
	ca := x509.NewCertPool()
	pem, err := os.ReadFile(caPath) //gosec:disable G304 -- user-specified file by design
	if err != nil {
		log.Fatal().Err(err).Str("path", caPath).Msg("failed to read " + msg)
	}
	if ok := ca.AppendCertsFromPEM(pem); !ok {
		log.Fatal().Err(err).Str("path", caPath).Msg("failed to parse " + msg)
	}

	msg = "gRPC server with mTLS"
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Fatal().Err(err).Str("cert", certPath).Str("key", keyPath).
			Msg("failed to load server PEM key pair for " + msg)
	}

	log.Info().Str("server_cert", certPath).Str("server_key", keyPath).
		Str("client_ca_cert", caPath).Msg("using " + msg)

	return []grpc.ServerOption{grpc.Creds(credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		RootCAs:      ca,
		MinVersion:   tls.VersionTLS13,
	}))}
}
