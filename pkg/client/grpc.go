package client

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
	"github.com/tzrikka/thrippy/internal/logger"
	"github.com/tzrikka/thrippy/pkg/oauth"
)

const (
	timeout = 3 * time.Second
)

// Connection creates a gRPC client connection to the given address.
// It supports both secure and insecure connections, based on the given credentials.
func Connection(addr string, creds credentials.TransportCredentials) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
}

// LinkOAuthConfig returns the OAuth configuration for a given link ID.
// This function reports gRPC errors, and invalid OAuth configurations,
// but if the link or its OAuth configuration are not found it returns nil.
func LinkOAuthConfig(ctx context.Context, grpcAddr string, creds credentials.TransportCredentials, linkID string) (*oauth.Config, error) {
	l := logger.FromContext(ctx)

	conn, err := Connection(grpcAddr, creds)
	if err != nil {
		l.Error("gRPC connection error", slog.Any("error", err))
		return nil, err
	}
	defer conn.Close()

	c := thrippypb.NewThrippyServiceClient(conn)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := c.GetLink(ctx, thrippypb.GetLinkRequest_builder{LinkId: new(linkID)}.Build())
	if err != nil {
		if status.Code(err) != codes.NotFound {
			l.Error("bad response from gRPC service", slog.Any("error", err), slog.String("client_method", "GetLink"))
			return nil, err
		}
		return nil, nil
	}

	o := oauth.FromProto(resp.GetOauthConfig())
	if o != nil && o.Config.ClientID == "" {
		l.Error("empty OAuth client ID")
		return nil, errors.New("empty OAuth client ID")
	}

	return o, nil
}

// AddGitHubCreds adds the given GitHub base URL and app installation ID to the given
// link's existing credentials. This also includes settings new metadata for the link.
func AddGitHubCreds(ctx context.Context, grpcAddr string, creds credentials.TransportCredentials, linkID, installID, url string) error {
	l := logger.FromContext(ctx)

	conn, err := Connection(grpcAddr, creds)
	if err != nil {
		l.Error("gRPC connection error", slog.Any("error", err))
		return err
	}
	defer conn.Close()

	c := thrippypb.NewThrippyServiceClient(conn)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := c.GetCredentials(ctx, thrippypb.GetCredentialsRequest_builder{LinkId: new(linkID)}.Build())
	if err != nil {
		l.Error("bad response from gRPC service", slog.Any("error", err), slog.String("client_method", "GetCredentials"))
		return err
	}

	m := resp.GetCredentials()
	m["install_id"] = installID
	m["api_base_url"] = url

	req := thrippypb.SetCredentialsRequest_builder{LinkId: new(linkID), GenericCreds: m}.Build()
	if _, err = c.SetCredentials(ctx, req); err != nil {
		l.Error("bad response from gRPC service", slog.Any("error", err), slog.String("client_method", "SetCredentials"))
		return err
	}

	return nil
}

// SetOAuthCreds checks and saves the given OAuth token as the credentials
// of the given link. This also includes settings new metadata for the link.
func SetOAuthCreds(ctx context.Context, grpcAddr string, creds credentials.TransportCredentials, linkID string, t *oauth2.Token) error {
	l := logger.FromContext(ctx)

	conn, err := Connection(grpcAddr, creds)
	if err != nil {
		l.Error("gRPC connection error", slog.Any("error", err))
		return err
	}
	defer conn.Close()

	c := thrippypb.NewThrippyServiceClient(conn)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req := thrippypb.SetCredentialsRequest_builder{LinkId: new(linkID), Token: oauth.TokenToProto(t)}.Build()
	if _, err = c.SetCredentials(ctx, req); err != nil {
		l.Error("bad response from gRPC service", slog.Any("error", err), slog.String("client_method", "SetCredentials"))
		return err
	}

	return nil
}
