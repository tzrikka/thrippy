package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"

	"github.com/tzrikka/thrippy/pkg/client"
)

const (
	maxHealthzResponseSize = 1024 // 1 KiB.
)

func healthCheckCommand(configFilePath altsrc.StringSourcer) *cli.Command {
	return &cli.Command{
		Name:     "health-check",
		Usage:    "Sends a single GET request to http://localhost:port/healthz",
		Category: "server monitoring",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return sendHealthzRequest(ctx, cmd.Int("webhook-port"))
		},
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
		},
	}
}

func sendHealthzRequest(ctx context.Context, port int) error {
	url := fmt.Sprintf("http://localhost:%d/healthz", port)
	req, cancel, err := client.ConstructRequest(ctx, http.MethodGet, url, "", nil)
	if err != nil {
		return err
	}
	defer cancel()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxHealthzResponseSize))
	if err != nil {
		return fmt.Errorf("failed to read HTTP response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		if len(body) == 0 {
			return errors.New(resp.Status)
		}
		return fmt.Errorf("%s: %s", resp.Status, string(body))
	}

	return nil
}
