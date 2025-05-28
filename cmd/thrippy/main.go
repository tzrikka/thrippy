package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"strconv"

	"github.com/rs/zerolog/log"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"

	"github.com/tzrikka/thrippy/pkg/client"
	"github.com/tzrikka/thrippy/pkg/secrets"
	"github.com/tzrikka/thrippy/pkg/server"
	"github.com/tzrikka/xdg"
)

const (
	ConfigDirName  = "thrippy"
	ConfigFileName = "config.toml"

	DefaultGRPCPort = 14460
	DefaultHTTPPort = 14470
)

func main() {
	buildInfo, _ := debug.ReadBuildInfo()
	configFilePath := configFile()

	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:  "dev",
			Usage: "simple setup, but unsafe for production",
		},
		&cli.StringFlag{
			Name:    "grpc-addr",
			Aliases: []string{"a"},
			Usage:   "gRPC server address and port",
			Value:   net.JoinHostPort("", strconv.Itoa(DefaultGRPCPort)),
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("THRIPPY_GRPC_ADDRESS"),
				toml.TOML("grpc.address", configFilePath),
			),
		},
	}
	flags = append(flags, client.GRPCFlags(configFilePath)...)
	flags = append(flags, server.GRPCFlags(configFilePath)...)
	flags = append(flags, secrets.ManagerFlags(configFilePath)...)
	flags = append(flags, secrets.VaultFlags(configFilePath)...)

	cmd := &cli.Command{
		Name:    "thrippy",
		Usage:   "Manage third-party auth configs and tokens",
		Version: buildInfo.Main.Version,
		Commands: []*cli.Command{
			serverCommand(configFilePath),
			linkTemplatesCommand,
			createLinkCommand,
			getLinkCommand,
			setCredsCommand,
			startOAuthCommand(configFilePath),
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

// configFile returns the path to the app's configuration file.
// It also creates an empty file if it doesn't already exist.
func configFile() altsrc.StringSourcer {
	path, err := xdg.CreateFile(xdg.ConfigHome, ConfigDirName, ConfigFileName)
	if err != nil {
		log.Fatal().Err(err).Caller().Send()
	}
	return altsrc.StringSourcer(path)
}
