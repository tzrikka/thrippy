package main

import (
	"errors"

	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"

	"github.com/tzrikka/thrippy/pkg/server"
)

func serverCommand(configFilePath altsrc.StringSourcer) *cli.Command {
	return &cli.Command{
		Name:      "server",
		Usage:     "Starts a local Thrippy server",
		UsageText: "thrippy server [global options] [command options]",
		Action:    server.Start,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "webhook-port",
				Aliases: []string{"p"},
				Usage:   "local port number for HTTP webhooks",
				Value:   DefaultHTTPPort,
				Sources: cli.NewValueSourceChain(
					cli.EnvVar("THRIPPY_WEBHOOK_PORT"),
					toml.TOML("server.webhook_port", configFilePath),
				),
				Validator: validatePort,
			},
			&cli.StringFlag{
				Name:    "webhook-addr",
				Aliases: []string{"w"},
				Usage:   "public address for HTTP webhooks",
				Sources: cli.NewValueSourceChain(
					cli.EnvVar("THRIPPY_WEBHOOK_ADDRESS"),
					toml.TOML("server.webhook_address", configFilePath),
				),
			},
			&cli.StringFlag{
				Name:    "fallback-url",
				Aliases: []string{"u"},
				Usage:   "optional destination for OAuth callbacks without a state",
				Sources: cli.NewValueSourceChain(
					cli.EnvVar("THRIPPY_FALLBACK_URL"),
					toml.TOML("server.fallback_url", configFilePath),
				),
			},
		},
	}
}

func validatePort(p int) error {
	if p < 0 || p > 65535 {
		return errors.New("out of range [0-65535]")
	}
	return nil
}
