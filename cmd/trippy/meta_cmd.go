package main

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sort"

	"github.com/urfave/cli/v3"
	"google.golang.org/protobuf/proto"

	"github.com/tzrikka/trippy/pkg/client"
	trippypb "github.com/tzrikka/trippy/proto/trippy/v1"
)

var getMetaCommand = &cli.Command{
	Name:      "get-meta",
	Usage:     "Retrieves all saved metadata for a specific link",
	UsageText: "trippy get-meta [global options] <link ID>",
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

		c := trippypb.NewTrippyServiceClient(conn)
		resp, err := c.GetMetadata(ctx, trippypb.GetMetadataRequest_builder{
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
