package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/lithammer/shortuuid/v4"
	"github.com/urfave/cli/v3"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
	intlinks "github.com/tzrikka/thrippy/internal/links"
	"github.com/tzrikka/thrippy/internal/logger"
	"github.com/tzrikka/thrippy/pkg/links"
	"github.com/tzrikka/thrippy/pkg/oauth"
	"github.com/tzrikka/thrippy/pkg/secrets"
)

type grpcServer struct {
	thrippypb.UnimplementedThrippyServiceServer

	sm secrets.Manager
}

// startGRPCServer starts a gRPC server for the [Thrippy service].
// This is non-blocking, in order to let Thrippy run an HTTP server as well.
//
// [Thrippy service]: https://github.com/tzrikka/thrippy-api/blob/main/proto/thrippy/v1/thrippy.proto
func startGRPCServer(ctx context.Context, cmd *cli.Command, sm secrets.Manager) (string, error) {
	lc := net.ListenConfig{}
	addr := cmd.String("grpc-addr")
	lis, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		slog.Error("failed to listen on gRPC address", slog.Any("error", err), slog.String("address", addr))
		return "", err
	}

	srv := grpc.NewServer(GRPCCreds(ctx, cmd)...)
	thrippypb.RegisterThrippyServiceServer(srv, &grpcServer{sm: sm})
	go func() {
		err = srv.Serve(lis)
		if err != nil {
			logger.FatalError(ctx, "gRPC serving error", err)
		}
	}()

	slog.Info("gRPC server listening on " + lis.Addr().String())
	return lis.Addr().String(), nil
}

func (s *grpcServer) CreateLink(ctx context.Context, in *thrippypb.CreateLinkRequest) (*thrippypb.CreateLinkResponse, error) {
	id := shortuuid.New()
	l := logger.FromContext(ctx).With(slog.String("grpc_handler", "CreateLink"), slog.String("link_id", id))
	l.Debug("received gRPC request")

	// Parse the input.
	t := in.GetTemplate()
	if _, ok := links.Templates[t]; !ok {
		l.Warn("invalid template", slog.String("template", t))
		return nil, status.Error(codes.InvalidArgument, "invalid template")
	}

	o := oauth.FromProto(in.GetOauthConfig())
	templ, ok := links.Templates[t]
	intlinks.ModifyOAuthByTemplate(o, templ, ok)
	if o != nil && o.Config.Endpoint.AuthURL != "" && o.Config.ClientID == "" {
		l.Warn("missing OAuth client ID")
		return nil, status.Error(codes.InvalidArgument, "missing OAuth client ID")
	}

	// Save the input template.
	if err := s.sm.Set(ctx, id+"/template", t); err != nil {
		l.Error("secrets manager write error", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "secrets manager write error")
	}

	// Save the parsed OAuth configuration, if there is one.
	if o.IsUsable() {
		j, err := o.ToJSON()
		if err != nil {
			l.Error("failed to convert OAuth proto into JSON", slog.Any("error", err))
			return nil, status.Error(codes.Internal, "secrets manager parse error")
		}

		if err := s.sm.Set(ctx, id+"/oauth", j); err != nil {
			l.Error("secrets manager write error", slog.Any("error", err))
			return nil, status.Error(codes.Internal, "secrets manager write error")
		}
	}

	return thrippypb.CreateLinkResponse_builder{
		LinkId:           proto.String(id),
		CredentialFields: links.Templates[t].CredFields(),
	}.Build(), nil
}

func (s *grpcServer) DeleteLink(ctx context.Context, in *thrippypb.DeleteLinkRequest) (*thrippypb.DeleteLinkResponse, error) {
	id := in.GetLinkId()
	l := logger.FromContext(ctx).With(slog.String("grpc_handler", "DeleteLink"), slog.String("link_id", id))
	l.Debug("received gRPC request")

	if id == "" {
		l.Warn("missing ID")
		return nil, status.Error(codes.InvalidArgument, "missing ID")
	}
	if _, err := shortuuid.DefaultEncoder.Decode(id); err != nil {
		l.Warn("ID is an invalid short UUID", slog.Any("error", err))
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}

	t, err := s.sm.Get(ctx, id+"/template")
	if err != nil {
		l.Error("secrets manager read error", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "secrets manager read error")
	}

	if t == "" {
		if in.GetAllowMissing() {
			return &thrippypb.DeleteLinkResponse{}, nil
		}
		l.Warn("link not found")
		return nil, status.Error(codes.NotFound, "link not found")
	}

	if err := s.sm.Delete(ctx, id+"/creds"); err != nil {
		l.Error("secrets manager delete error: creds", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "secrets manager delete error: creds")
	}
	if err := s.sm.Delete(ctx, id+"/meta"); err != nil {
		l.Error("secrets manager delete error: meta", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "secrets manager delete error: meta")
	}
	if err := s.sm.Delete(ctx, id+"/oauth"); err != nil {
		l.Error("secrets manager delete error: oauth", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "secrets manager delete error: oauth")
	}
	if err := s.sm.Delete(ctx, id+"/template"); err != nil {
		l.Error("secrets manager delete error: template", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "secrets manager delete error: template")
	}

	return &thrippypb.DeleteLinkResponse{}, nil
}

func (s *grpcServer) GetLink(ctx context.Context, in *thrippypb.GetLinkRequest) (*thrippypb.GetLinkResponse, error) {
	id := in.GetLinkId()
	l := logger.FromContext(ctx).With(slog.String("grpc_handler", "GetLink"), slog.String("link_id", id))
	l.Debug("received gRPC request")

	if id == "" {
		l.Warn("missing ID")
		return nil, status.Error(codes.InvalidArgument, "missing ID")
	}
	if _, err := shortuuid.DefaultEncoder.Decode(id); err != nil {
		l.Warn("ID is an invalid short UUID", slog.Any("error", err))
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}

	ctx = logger.InContext(ctx, l)
	t, o, err := s.templateAndOAuth(ctx, id)
	if err != nil {
		return nil, err
	}

	return thrippypb.GetLinkResponse_builder{
		Template:         proto.String(t),
		OauthConfig:      o,
		CredentialFields: links.Templates[t].CredFields(),
	}.Build(), nil
}

func (s *grpcServer) SetCredentials(ctx context.Context, in *thrippypb.SetCredentialsRequest) (*thrippypb.SetCredentialsResponse, error) {
	id := in.GetLinkId()
	l := logger.FromContext(ctx).With(slog.String("grpc_handler", "SetCredentials"), slog.String("link_id", id))
	l.Debug("received gRPC request")

	if id == "" {
		l.Warn("missing ID")
		return nil, status.Error(codes.InvalidArgument, "missing ID")
	}
	if _, err := shortuuid.DefaultEncoder.Decode(id); err != nil {
		l.Warn("ID is an invalid short UUID", slog.Any("error", err))
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}

	ctx = logger.InContext(ctx, l)
	template, oauthProto, err := s.templateAndOAuth(ctx, id)
	if err != nil {
		return nil, err
	}

	// OAuth-based links: change the nonce, now that the old one was used successfully.
	o := oauth.FromProto(oauthProto)
	if o.IsUsable() {
		j, err := o.ToJSON()
		if err != nil {
			l.Error("failed to convert OAuth proto into JSON", slog.Any("error", err))
			return nil, status.Error(codes.Internal, "secrets manager parse error")
		}

		if err := s.sm.Set(ctx, id+"/oauth", j); err != nil {
			l.Error("secrets manager write error", slog.Any("error", err))
			return nil, status.Error(codes.Internal, "secrets manager write error")
		}
	}

	// Credentials to store: either an OAuth token or a generic string map.
	// For OAuth tokens: persist extra secrets if already set.
	token := in.GetToken()
	m := in.GetGenericCreds()
	if strings.Contains(template, "oauth") {
		if token == nil {
			token = thrippypb.OAuthToken_builder{Raw: m}.Build()
		} else {
			token.SetRaw(s.getRaw(ctx, id))
		}
	}

	j, err := protojson.Marshal(token)
	if err != nil {
		l.Error("failed to convert proto into JSON", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "secrets manager parse error")
	}

	if len(j) <= 2 {
		j, err = json.Marshal(m)
		if err != nil {
			l.Error("failed to convert credentials into JSON", slog.Any("error", err))
			return nil, status.Error(codes.Internal, "secrets manager parse error")
		}
	}

	// Check the usability of the provided credentials, retrieve their metadata, and save both.
	metadata, err := links.Templates[template].Check(ctx, m, o, oauth.TokenFromProto(token))
	if err != nil {
		l.Error("failed to check credentials / extract metadata", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "credentials check error: "+err.Error())
	}

	if err := s.sm.Set(ctx, id+"/creds", string(j)); err != nil {
		l.Error("secrets manager write error", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "secrets manager write error")
	}

	if metadata != "" {
		if err := s.sm.Set(ctx, id+"/meta", metadata); err != nil {
			l.Error("secrets manager write error", slog.Any("error", err))
			return nil, status.Error(codes.Internal, "secrets manager write error")
		}
	}

	return &thrippypb.SetCredentialsResponse{}, nil
}

func (s *grpcServer) templateAndOAuth(ctx context.Context, id string) (string, *thrippypb.OAuthConfig, error) {
	l := logger.FromContext(ctx)

	t, err := s.sm.Get(ctx, id+"/template")
	if err != nil {
		l.Error("secrets manager read error", slog.Any("error", err))
		return "", nil, status.Error(codes.Internal, "secrets manager read error")
	}
	if t == "" {
		l.Warn("link not found")
		return "", nil, status.Error(codes.NotFound, "link not found")
	}

	o, err := s.sm.Get(ctx, id+"/oauth")
	if err != nil {
		l.Error("secrets manager read error", slog.Any("error", err))
		return "", nil, status.Error(codes.Internal, "secrets manager read error")
	}

	var m *thrippypb.OAuthConfig
	if o != "" {
		m = &thrippypb.OAuthConfig{}
		if err = protojson.Unmarshal([]byte(o), m); err != nil {
			l.Error("failed to convert JSON into proto", slog.Any("error", err))
			return "", nil, status.Error(codes.Internal, "secrets manager parse error")
		}
	}

	return t, m, nil
}

// getRaw retrieves the "raw" credentials map from an OAuth token stored in the
// secrets manager. This is used to preserve extra secrets alongside OAuth tokens.
// If there is any error, or if there are no extra secrets, this function returns nil.
func (s *grpcServer) getRaw(ctx context.Context, id string) map[string]string {
	j, err := s.sm.Get(ctx, id+"/creds")
	if err != nil {
		return nil
	}

	var token map[string]any
	if err := json.Unmarshal([]byte(j), &token); err != nil {
		return nil
	}

	raw, ok := token["raw"].(map[string]any)
	if !ok {
		return nil
	}

	m := make(map[string]string, len(raw))
	for k, v := range raw {
		m[k] = fmt.Sprintf("%v", v)
	}
	return m
}

func (s *grpcServer) GetCredentials(ctx context.Context, in *thrippypb.GetCredentialsRequest) (*thrippypb.GetCredentialsResponse, error) {
	id := in.GetLinkId()
	l := logger.FromContext(ctx).With(slog.String("grpc_handler", "GetCredentials"), slog.String("link_id", id))

	ctx = logger.InContext(ctx, l)
	ma, err := s.getSecrets(ctx, id, "/creds")
	if err != nil {
		return nil, err
	}

	// Refresh OAuth token, if needed.
	if t, ok := oauth.TokenFromMap(ma); ok && !t.Valid() {
		if updated, err := s.refreshOAuthToken(ctx, id, t); err == nil {
			ma = updated
		}
	}

	ms := make(map[string]string, len(ma))
	for k, v := range ma {
		if k != "raw" {
			ms[k] = fmt.Sprintf("%v", v)
			continue
		}

		// Flatten extra secrets from an OAuth token's "raw" map, but
		// in a limited way, to prevent the possibility of overwriting.
		raw, ok := v.(map[string]any)
		if !ok {
			continue
		}
		if ws, ok := raw["signing_secret"].(string); ok { // Slack.
			ms["signing_secret"] = ws
		}
		if ws, ok := raw["webhook_secret"].(string); ok { // Bitbucket, GitHub.
			ms["webhook_secret"] = ws
		}
	}

	return thrippypb.GetCredentialsResponse_builder{Credentials: ms}.Build(), nil
}

func (s *grpcServer) GetMetadata(ctx context.Context, in *thrippypb.GetMetadataRequest) (*thrippypb.GetMetadataResponse, error) {
	id := in.GetLinkId()
	l := logger.FromContext(ctx).With(slog.String("grpc_handler", "GetMetadata"), slog.String("link_id", id))

	ctx = logger.InContext(ctx, l)
	ma, err := s.getSecrets(ctx, id, "/meta")
	if err != nil {
		return nil, err
	}

	ms := make(map[string]string, len(ma))
	for k, v := range ma {
		ms[k] = fmt.Sprintf("%v", v)
	}

	return thrippypb.GetMetadataResponse_builder{Metadata: ms}.Build(), nil
}

func (s *grpcServer) getSecrets(ctx context.Context, linkID, keySuffix string) (map[string]any, error) {
	l := logger.FromContext(ctx)
	l.Debug("received gRPC request")

	if linkID == "" {
		l.Warn("missing ID")
		return nil, status.Error(codes.InvalidArgument, "missing ID")
	}
	if _, err := shortuuid.DefaultEncoder.Decode(linkID); err != nil {
		l.Warn("ID is an invalid short UUID", slog.Any("error", err))
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}

	j, err := s.sm.Get(ctx, linkID+keySuffix)
	if err != nil {
		l.Error("secrets manager read error", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "secrets manager read error")
	}

	var m map[string]any
	if j != "" {
		if err := json.Unmarshal([]byte(j), &m); err != nil {
			l.Error("failed to convert JSON into map", slog.Any("error", err))
			return nil, status.Error(codes.Internal, "secrets manager parse error")
		}
	}

	return m, nil
}

func (s *grpcServer) refreshOAuthToken(ctx context.Context, id string, t *oauth2.Token) (map[string]any, error) {
	l := logger.FromContext(ctx)

	jsonConfig, err := s.sm.Get(ctx, id+"/oauth")
	if err != nil {
		l.Error("secrets manager read error", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "secrets manager read error")
	}

	o := &thrippypb.OAuthConfig{}
	if err := protojson.Unmarshal([]byte(jsonConfig), o); err != nil {
		l.Error("failed to convert JSON into proto", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "secrets manager parse error")
	}

	m, err := oauth.FromProto(o).RefreshToken(ctx, t, false)
	if err != nil {
		l.Error("failed to refresh OAuth token", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "OAuth token refresh error")
	}

	if raw := s.getRaw(ctx, id); raw != nil {
		rawAny := make(map[string]any, len(raw))
		for k, v := range raw {
			rawAny[k] = v
		}
		m["raw"] = rawAny
	}

	jsonToken, err := json.Marshal(m)
	if err != nil {
		l.Error("failed to convert map into JSON", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "secrets manager parse error")
	}

	if err := s.sm.Set(ctx, id+"/creds", string(jsonToken)); err != nil {
		l.Error("secrets manager write error", slog.Any("error", err))
		return nil, status.Error(codes.Internal, "secrets manager write error")
	}

	return m, nil
}
