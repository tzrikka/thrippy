package main

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sort"

	"github.com/urfave/cli/v3"
	"google.golang.org/protobuf/proto"

	"github.com/tzrikka/thrippy/pkg/client"
	thrippypb "github.com/tzrikka/thrippy/proto/thrippy/v1"
)

var getMetaCommand = &cli.Command{
	Name:      "get-meta",
	Usage:     "Retrieves all saved metadata for a specific link",
	UsageText: "thrippy get-meta [global options] <link ID>",
	Category:  "link metadata",
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
		resp, err := c.GetMetadata(ctx, thrippypb.GetMetadataRequest_builder{
			LinkId: proto.String(cmd.Args().First()),
		}.Build())
		if err != nil {
			return err
		}

		// Maximum length of metadata keys (for pretty-printing).
		l := 0
		kv := resp.GetMetadata()
		for k := range kv {
			if len(k) > l {
				l = len(k)
			}
		}

		// Sort metadata keys before enumerating them.
		ks := slices.Collect(maps.Keys(kv))
		sort.Strings(ks)
		for _, k := range ks {
			fmt.Printf("- %-*s  %s\n", l, k, kv[k])
		}

		return nil
	},
}
