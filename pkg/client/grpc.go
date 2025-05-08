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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/tzrikka/trippy/pkg/oauth"
	trippypb "github.com/tzrikka/trippy/proto/trippy/v1"
)

const (
	timeout = time.Second * 3
)

// Connection creates a gRPC client connection to the specified address.
func Connection(addr string) (*grpc.ClientConn, error) {
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	return grpc.NewClient(addr, creds)
}

// LinkOAuthConfig returns the OAuth configuration for a given link ID.
// This function reports gRPC errors, and invalid OAuth configurations,
// but if the link or its OAuth configuration are not found it returns nil.
func LinkOAuthConfig(ctx context.Context, grpcAddr, linkID string) (*oauth.Config, error) {
	l := zerolog.Ctx(ctx)

	conn, err := Connection(grpcAddr)
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
