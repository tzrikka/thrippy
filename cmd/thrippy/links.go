package main

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"

	"github.com/lithammer/shortuuid/v4"
	"github.com/urfave/cli/v3"
	"google.golang.org/protobuf/proto"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
	"github.com/tzrikka/thrippy/pkg/client"
	"github.com/tzrikka/thrippy/pkg/links"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

var linkTemplatesCommand = &cli.Command{
	Name:      "link-templates",
	Usage:     "Lists all available templates for link creation",
	UsageText: "thrippy link-templates",
	Category:  "link",
	Action: func(_ context.Context, _ *cli.Command) error {
		// Maximum length of template ID (for pretty-printing).
		l := 0
		for id := range links.Templates {
			if len(id) > l {
				l = len(id)
			}
		}

		// Sort templates by ID before enumerating them.
		ids := slices.Collect(maps.Keys(links.Templates))
		sort.Strings(ids)
		for _, id := range ids {
			fmt.Printf("- %-*s  %s\n", l, id, links.Templates[id].Description())
		}

		return nil
	},
}

var createLinkCommand = &cli.Command{
	Name:      "create-link",
	Usage:     "Creates a new link configuration",
	UsageText: "thrippy create-link [global options] --template <...> [oauth options]",
	Category:  "link",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "template",
			Aliases:  []string{"t"},
			Usage:    `Link configuration template (see "link-templates" command)`,
			Required: true,
			Validator: func(v string) error {
				if _, ok := links.Templates[v]; !ok {
					return errors.New("invalid template")
				}
				return nil
			},
		},
		&cli.StringFlag{
			Name:  "auth-url",
			Usage: "optional OAuth 2.0 auth URL",
		},
		&cli.StringFlag{
			Name:  "token-url",
			Usage: "optional OAuth 2.0 token URL",
		},
		&cli.StringFlag{
			Name:  "client-id",
			Usage: "optional OAuth 2.0 client ID",
		},
		&cli.StringFlag{
			Name:  "client-secret",
			Usage: "optional OAuth 2.0 client secret",
		},
		&cli.StringSliceFlag{
			Name:  "scopes",
			Usage: "optional OAuth 2.0 scopes (comma delimited / multiple flags)",
		},
		&cli.StringMapFlag{
			Name:  "param",
			Usage: "optional OAuth 2.0 URL parameters",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		conn, err := client.Connection(cmd.String("grpc-addr"), client.GRPCCreds(ctx, cmd))
		if err != nil {
			return err
		}
		defer conn.Close()

		// Protocol buffers differentiate between empty and unset field values.
		hasOAuth := false
		o := &thrippypb.OAuthConfig{}
		if v := cmd.String("auth-url"); v != "" {
			o.SetAuthUrl(v)
			hasOAuth = true
		}
		if v := cmd.String("token-url"); v != "" {
			o.SetTokenUrl(v)
			hasOAuth = true
		}
		if v := cmd.String("client-id"); v != "" {
			o.SetClientId(v)
			hasOAuth = true
		}
		if v := cmd.String("client-secret"); v != "" {
			o.SetClientSecret(v)
			hasOAuth = true
		}
		if s := cmd.StringSlice("scopes"); len(s) > 0 {
			o.SetScopes(s)
			hasOAuth = true
		}
		if m := cmd.StringMap("param"); len(m) > 0 {
			o.SetParams(m)
			hasOAuth = true
		}
		if !hasOAuth {
			o = nil
		}

		c := thrippypb.NewThrippyServiceClient(conn)
		resp, err := c.CreateLink(ctx, thrippypb.CreateLinkRequest_builder{
			Template:    proto.String(cmd.String("template")),
			OauthConfig: o,
		}.Build())
		if err != nil {
			return err
		}

		fmt.Println("New link ID:", resp.GetLinkId())
		return nil
	},
}

var deleteLinkCommand = &cli.Command{
	Name:      "delete-link",
	Usage:     "Deletes a specific link's configuration",
	UsageText: "thrippy delete-link [global options] <link ID> [--allow-missing]",
	Category:  "link",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "allow-missing",
			Usage: "do not fail if the link does not exist",
		},
	},
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
		_, err = c.DeleteLink(ctx, thrippypb.DeleteLinkRequest_builder{
			LinkId:       proto.String(cmd.Args().First()),
			AllowMissing: proto.Bool(cmd.Bool("allow-missing")),
		}.Build())
		if err != nil {
			return err
		}

		fmt.Println("Link deleted successfully:", cmd.Args().First())
		return nil
	},
}

var getLinkCommand = &cli.Command{
	Name:      "get-link",
	Usage:     "Retrieves a specific link's configuration",
	UsageText: "thrippy get-link [global options] <link ID>",
	Category:  "link",
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
		resp, err := c.GetLink(ctx, thrippypb.GetLinkRequest_builder{
			LinkId: proto.String(cmd.Args().First()),
		}.Build())
		if err != nil {
			return err
		}

		fmt.Println("Template:  ", resp.GetTemplate())
		o := oauth.ToString(resp.GetOauthConfig())
		if o != "" {
			fmt.Println("")
			fmt.Println(o)
		}

		fmt.Println("\nExpected credential fields:")
		for _, cf := range resp.GetCredentialFields() {
			mod1 := "automatic"
			if cf.GetManual() {
				mod1 = "manual"
			}
			mod2 := "required"
			if cf.GetOptional() {
				mod2 = "optional"
			}
			fmt.Printf("- %-*s (%s, %s)\n", 25, cf.GetName(), mod1, mod2)
		}

		return nil
	},
}

func checkLinkIDArg(cmd *cli.Command) error {
	switch cmd.NArg() {
	case 0:
		return errors.New("missing link ID argument")
	case 1:
		// OK.
	default:
		return errors.New("too many arguments, expecting exactly one")
	}

	if _, err := shortuuid.DefaultEncoder.Decode(cmd.Args().First()); err != nil {
		return errors.New("invalid link ID argument")
	}
	return nil
}
