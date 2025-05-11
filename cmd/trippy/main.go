// Trippy manages authentication configurations
// and tokens for third-party (3P) services.
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"

	"github.com/rs/zerolog/log"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"

	"github.com/tzrikka/trippy/pkg/secrets"
	"github.com/tzrikka/xdg"
)

const (
	configDirName  = "trippy"
	configFileName = "config.toml"

	configDirPerm  = 0o700
	configFilePerm = 0o600

	defaultGRPCPort = 14460
	defaultHTTPPort = 14470
)

func main() {
	buildInfo, _ := debug.ReadBuildInfo()
	configFilePath := configFile(configDir())

	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:  "dev",
			Usage: "simple setup, but unsafe for production",
		},
		&cli.StringFlag{
			Name:    "grpc-addr",
			Aliases: []string{"a"},
			Usage:   "gRPC server address and port",
			Value:   net.JoinHostPort("", strconv.Itoa(defaultGRPCPort)),
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("TRIPPY_GRPC_ADDRESS"),
				toml.TOML("grpc.address", configFilePath),
			),
		},
	}
	// TODO: Client & server gRPC transport credentials.
	flags = append(flags, secrets.ManagerFlags(configFilePath)...)
	flags = append(flags, secrets.VaultFlags(configFilePath)...)

	cmd := &cli.Command{
		Name:    "trippy",
		Usage:   "Manage third-party auth configs and tokens",
		Version: buildInfo.Main.Version,
		Commands: []*cli.Command{
			serverCommand(configFilePath),
			linkTemplatesCommand,
			createLinkCommand,
			getLinkCommand,
			startOAuthCommand(configFilePath),
			setCredsCommand,
			getCredsCommand,
			getMetaCommand,
		},
		Flags:                 flags,
		EnableShellCompletion: true,
		Suggest:               true,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// configDir returns the path to the app's configuration directory.
// It also creates the directory if it doesn't already exist.
func configDir() string {
	path, err := xdg.ConfigHome()
	if err != nil {
		log.Fatal().Err(err).Caller().Send()
	}

	path = filepath.Join(path, configDirName)
	if err := os.MkdirAll(path, configDirPerm); err != nil {
		log.Fatal().Err(err).Caller().Send()
	}

	return path
}

// configFile returns the path to the app's configuration file.
// It also creates an empty file if it doesn't already exist.
func configFile(path string) altsrc.StringSourcer {
	path = filepath.Join(path, configFileName)

	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, configFilePerm) //gosec:disable G304 -- constructed and cleaned by us
	if err != nil {
		log.Fatal().Err(err).Caller().Send()
	}
	_ = f.Close()

	return altsrc.StringSourcer(path)
}
