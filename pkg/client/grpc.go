// Package client provides minimal and lightweight wrappers for some
// gRPC client functionalities. It is meant to facilitate code reuse,
// not to provide a complete native layer on top of the Trippy gRPC
// service (proto/.../trippy.proto).
package client

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/tzrikka/trippy/pkg/oauth"
	trippypb "github.com/tzrikka/trippy/proto/trippy/v1"
)

const (
	timeout = time.Second * 3
)

// Creds initializes gRPC client credentials, based on CLI flags.
func Creds(cmd *cli.Command) credentials.TransportCredentials {
	if cmd.Bool("dev") {
		return InsecureCreds()
	}

	// With flags defined in main.go:
	// https://grpc.io/docs/guides/auth/
	// https://grpc.io/docs/languages/go/alts/
	// https://github.com/grpc/grpc-go/tree/master/examples/features/authentication
	// https://github.com/grpc/grpc-go/tree/master/examples/features/encryption
	panic("not implemented yet")
}

func InsecureCreds() credentials.TransportCredentials {
	return insecure.NewCredentials()
}

// Connection creates a gRPC client connection to the given address.
// It supports both secure and insecure connections, based on the given credentials.
func Connection(addr string, creds credentials.TransportCredentials) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
}

// LinkOAuthConfig returns the OAuth configuration for a given link ID.
// This function reports gRPC errors, and invalid OAuth configurations,
// but if the link or its OAuth configuration are not found it returns nil.
func LinkOAuthConfig(ctx context.Context, grpcAddr string, creds credentials.TransportCredentials, linkID string) (*oauth.Config, error) {
	l := zerolog.Ctx(ctx)

	conn, err := Connection(grpcAddr, creds)
	if err != nil {
		l.Error().Stack().Err(err).Send()
		return nil, err
	}
	defer conn.Close()

	c := trippypb.NewTrippyServiceClient(conn)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := c.GetLink(ctx, trippypb.GetLinkRequest_builder{
		LinkId: proto.String(linkID),
	}.Build())
	if err != nil {
		if status.Code(err) != codes.NotFound {
			l.Error().Stack().Err(err).Send()
			return nil, err
		}
		return nil, nil
	}

	o := oauth.FromProto(resp.GetOauthConfig())
	if o != nil && (o.Config.ClientID == "" || o.Config.ClientSecret == "") {
		l.Error().Stack().Msg("empty OAuth client ID and/or secret")
		return nil, errors.New("empty OAuth client ID and/or secret")
	}

	return o, nil
}
