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
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"

	"github.com/tzrikka/thrippy/pkg/client"
	"github.com/tzrikka/thrippy/pkg/links"
	"github.com/tzrikka/thrippy/pkg/oauth"
	thrippypb "github.com/tzrikka/thrippy/proto/thrippy/v1"
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
	UsageText: "thrippy create-link [global options] --template <...> [--oauth <...>]",
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
			Name:    "oauth",
			Aliases: []string{"o"},
			Usage:   `"thrippy.v1.OAuthConfig" proto message`,
			Validator: func(v string) error {
				return prototext.Unmarshal([]byte(v), &thrippypb.OAuthConfig{})
			},
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		conn, err := client.Connection(cmd.String("grpc-addr"), client.Creds(cmd))
		if err != nil {
			return err
		}
		defer conn.Close()

		// Syntax already checked by the flag's validator
		// (semantics will be checked by the server).
		o := &thrippypb.OAuthConfig{}
		_ = prototext.Unmarshal([]byte(cmd.String("oauth")), o)

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

var getLinkCommand = &cli.Command{
	Name:      "get-link",
	Usage:     "Retrieves a specific link's configuration",
	UsageText: "thrippy get-link [global options] <link ID>",
	Category:  "link",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if err := checkLinkIDArg(cmd); err != nil {
			return err
		}

		conn, err := client.Connection(cmd.String("grpc-addr"), client.Creds(cmd))
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

		fmt.Println("Template: ", resp.GetTemplate())
		o := oauth.ToString(resp.GetOauthConfig())
		if o != "" {
			fmt.Println("")
			fmt.Println(o)
		}
		fmt.Println("\nExpected credential fields:", resp.GetCredentialFields())

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
