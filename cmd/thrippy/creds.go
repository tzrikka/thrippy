package main

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"os"
	"slices"
	"sort"

	"github.com/pkg/browser"
	altsrc "github.com/urfave/cli-altsrc/v3"
	"github.com/urfave/cli-altsrc/v3/toml"
	"github.com/urfave/cli/v3"
	"google.golang.org/protobuf/proto"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
	"github.com/tzrikka/thrippy/pkg/client"
)

// startOAuthCommand is a function rather than a var because it
// depends on the runtime return value of [configDir] and [configFile].
func startOAuthCommand(configFilePath altsrc.StringSourcer) *cli.Command {
	return &cli.Command{
		Name:      "start-oauth",
		Usage:     "Starts a 3-legged OAuth 2.0 flow for a specific link",
		UsageText: "thrippy start-oauth [--base-url <http[s]://host:port>] <link ID>",
		Category:  "link credentials",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "base-url",
				Aliases: []string{"u"},
				Usage:   "Thrippy HTTP server's base URL",
				Value:   fmt.Sprintf("http://localhost:%d", DefaultHTTPPort),
				Sources: cli.NewValueSourceChain(
					toml.TOML("client.webhook_base_url", configFilePath),
				),
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if err := checkLinkIDArg(cmd); err != nil {
				return err
			}

			// Get the link's nonce.
			conn, err := client.Connection(cmd.String("grpc-addr"), client.GRPCCreds(ctx, cmd))
			if err != nil {
				return err
			}
			defer conn.Close()

			id := cmd.Args().First()
			c := thrippypb.NewThrippyServiceClient(conn)
			resp, err := c.GetLink(ctx, thrippypb.GetLinkRequest_builder{LinkId: proto.String(id)}.Build())
			if err != nil {
				return err
			}

			nonce := resp.GetOauthConfig().GetNonce()
			if len(nonce) == 0 {
				return fmt.Errorf("link %q does not have OAuth configured", id)
			}

			// Construct the OAuth start URL and open it in a browser.
			u, err := url.JoinPath(cmd.String("base-url"), "start")
			if err != nil {
				return err
			}

			u = fmt.Sprintf("%s?id=%s&nonce=%s", u, id, nonce)
			fmt.Println("Opening a browser with this URL:", u)

			if err := browser.OpenURL(u); err != nil {
				return err
			}
			return nil
		},
	}
}

var setCredsCommand = &cli.Command{
	Name:        "set-creds",
	Usage:       "Sets static credentials for a specific link",
	UsageText:   `thrippy set-creds [global options] <link ID> --kv "key=value" [--kv ...]`,
	Description: "Note that this command overwrites existing data, it does not append to it",
	Category:    "link credentials",
	Flags: []cli.Flag{
		&cli.StringMapFlag{
			Name:  "kv",
			Usage: `one or more "key=value" pairs`,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := checkLinkIDArg(cmd); err != nil {
			return err
		}

		kv, err := readFiles(cmd.StringMap("kv"))
		if err != nil {
			return err
		}

		conn, err := client.Connection(cmd.String("grpc-addr"), client.GRPCCreds(ctx, cmd))
		if err != nil {
			return err
		}
		defer conn.Close()

		c := thrippypb.NewThrippyServiceClient(conn)
		_, err = c.SetCredentials(ctx, thrippypb.SetCredentialsRequest_builder{
			LinkId:       proto.String(cmd.Args().First()),
			GenericCreds: kv,
		}.Build())
		if err != nil {
			return err
		}

		return nil
	},
}

var getCredsCommand = &cli.Command{
	Name:      "get-creds",
	Usage:     "Retrieves all saved credentials for a specific link",
	UsageText: "thrippy get-creds [global options] <link ID>",
	Category:  "link credentials",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := checkLinkIDArg(cmd); err != nil {
			return err
		}

		conn, err := client.Connection(cmd.String("grpc-addr"), client.GRPCCreds(ctx, cmd))
		if err != nil {
			return err
		}
		defer conn.Close()

		c := thrippypb.NewThrippyServiceClient(conn)
		resp, err := c.GetCredentials(ctx, thrippypb.GetCredentialsRequest_builder{
			LinkId: proto.String(cmd.Args().First()),
		}.Build())
		if err != nil {
			return err
		}

		// Maximum length of credential keys (for pretty-printing).
		l := 0
		kv := resp.GetCredentials()
		for k := range kv {
			if len(k) > l {
				l = len(k)
			}
		}

		// Sort credential keys before enumerating them.
		ks := slices.Collect(maps.Keys(kv))
		sort.Strings(ks)
		for _, k := range ks {
			fmt.Printf("- %-*s  %s\n", l, k, kv[k])
		}

		return nil
	},
}

// readFiles converts the values of keys that reference
// file paths ("@path") into the contents of these files.
func readFiles(m map[string]string) (map[string]string, error) {
	for k, v := range m {
		if len(v) == 0 || v[0] != '@' {
			continue
		}

		pem, err := os.ReadFile(v[1:])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", k, err)
		}

		m[k] = string(pem)
	}

	return m, nil
}
