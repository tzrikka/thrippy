package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"strconv"

	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"

	"github.com/tzrikka/thrippy/internal/logger"
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
	bi, _ := debug.ReadBuildInfo()
	path := configFile()

	cmd := &cli.Command{
		Name:    "thrippy",
		Usage:   "Manage third-party auth configs and tokens",
		Version: bi.Main.Version,
		Commands: []*cli.Command{
			serverCommand(path),
			linkTemplatesCommand,
			createLinkCommand,
			deleteLinkCommand,
			getLinkCommand,
			setCredsCommand,
			startOAuthCommand(path),
			getCredsCommand,
			getMetaCommand,
		},
		Flags:                 flags(path),
		EnableShellCompletion: true,
		Suggest:               true,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func flags(path altsrc.StringSourcer) []cli.Flag {
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
				toml.TOML("grpc.address", path),
			),
		},
	}

	flags = append(flags, client.GRPCFlags(path)...)
	flags = append(flags, server.GRPCFlags(path)...)
	flags = append(flags, secrets.ManagerFlags(path)...)
	flags = append(flags, secrets.AWSFlags(path)...)
	flags = append(flags, secrets.VaultFlags(path)...)
	return flags
}

// configFile returns the path to the app's configuration file.
// It also creates an empty file if it doesn't already exist.
func configFile() altsrc.StringSourcer {
	path, err := xdg.CreateFile(xdg.ConfigHome, ConfigDirName, ConfigFileName)
	if err != nil {
		logger.FatalError(context.Background(), "failed to create config file", err)
	}
	return altsrc.StringSourcer(path)
}
