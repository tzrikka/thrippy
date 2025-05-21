package client

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/rs/zerolog/log"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
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
			Name: "grpc-server-ca-cert", // Either TLS and mTLS.
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_GRPC_SERVER_CA_CERT"),
				toml.TOML("grpc.client.server_ca_cert", configFilePath),
			),
			Hidden:    true,
			TakesFile: true,
		},
		&cli.StringFlag{
			Name: "grpc-server-name-override", // Either TLS and mTLS, but only for testing.
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
func GRPCCreds(cmd *cli.Command) credentials.TransportCredentials {
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
		log.Fatal().Msg("missing server CA cert file for gRPC client with m/TLS")
	}

	// Using mTLS requires the client's X.509 PEM-encoded public cert
	// and private key. If one of them is missing it's an error.
	if certPath == "" && keyPath != "" {
		log.Fatal().Msg("missing client public cert file for gRPC client with mTLS")
	}
	if certPath != "" && keyPath == "" {
		log.Fatal().Msg("missing client private key file for gRPC client with mTLS")
	}

	// If both of them are missing, we use TLS.
	if certPath == "" && keyPath == "" {
		creds, err := credentials.NewClientTLSFromFile(caPath, nameOverride)
		if err != nil {
			log.Fatal().Err(err).Str("path", caPath).
				Msg("error in server CA cert for gRPC client with TLS")
		}
		return creds
	}

	// If all 3 are specified, we use mTLS.
	msg := "server CA cert file for gRPC client with mTLS"
	ca := x509.NewCertPool()
	pem, err := os.ReadFile(caPath) //gosec:disable G304 -- user-specified file by design
	if err != nil {
		log.Fatal().Err(err).Str("path", caPath).Msg("failed to read " + msg)
	}
	if ok := ca.AppendCertsFromPEM(pem); !ok {
		log.Fatal().Err(err).Str("path", caPath).Msg("failed to parse " + msg)
	}

	msg = "gRPC client with mTLS"
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Fatal().Err(err).Str("cert", certPath).Str("key", keyPath).
			Msg("failed to load client PEM key pair for " + msg)
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      ca,
		ServerName:   nameOverride,
		MinVersion:   tls.VersionTLS13,
	})
}
