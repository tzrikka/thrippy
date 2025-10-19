package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/lithammer/shortuuid/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	thrippypb "github.com/tzrikka/thrippy-api/thrippy/v1"
	intlinks "github.com/tzrikka/thrippy/internal/links"
	"github.com/tzrikka/thrippy/pkg/links"
	"github.com/tzrikka/thrippy/pkg/oauth"
	"github.com/tzrikka/thrippy/pkg/secrets"
)

type grpcServer struct {
	thrippypb.UnimplementedThrippyServiceServer

	sm secrets.Manager
}

// startGRPCServer starts a gRPC server for the [Thrippy service]. This
// is non-blocking, in order to let Thrippy run an HTTP server as well.
//
// [Thrippy service]: https://github.com/tzrikka/thrippy-api/blob/main/proto/thrippy/v1/thrippy.proto
func startGRPCServer(ctx context.Context, cmd *cli.Command, sm secrets.Manager) (string, error) {
	lc := net.ListenConfig{}
	lis, err := lc.Listen(ctx, "tcp", cmd.String("grpc-addr"))
	if err != nil {
		log.Err(err).Send()
		return "", err
	}

	srv := grpc.NewServer(GRPCCreds(cmd)...)
	thrippypb.RegisterThrippyServiceServer(srv, &grpcServer{sm: sm})
	go func() {
		err = srv.Serve(lis)
		if err != nil {
			log.Fatal().Err(err).Msg("gRPC serving error")
		}
	}()

	log.Info().Msgf("gRPC server listening on %s", lis.Addr().String())
	return lis.Addr().String(), nil
}

func (s *grpcServer) CreateLink(ctx context.Context, in *thrippypb.CreateLinkRequest) (*thrippypb.CreateLinkResponse, error) {
	id := shortuuid.New()
	l := log.With().Str("grpc_method", "CreateLink").Str("id", id).Logger()
	l.Debug().Msg("received gRPC request")

	// Parse the input.
	t := in.GetTemplate()
	if _, ok := links.Templates[t]; !ok {
		l.Warn().Str("template", t).Msg("invalid template")
		return nil, status.Error(codes.InvalidArgument, "invalid template")
	}

	o := oauth.FromProto(in.GetOauthConfig())
	templ, ok := links.Templates[t]
	intlinks.ModifyOAuthByTemplate(o, templ, ok)
	if o != nil && o.Config.Endpoint.AuthURL != "" && o.Config.ClientID == "" {
		l.Warn().Msg("missing OAuth client ID")
		return nil, status.Error(codes.InvalidArgument, "missing OAuth client ID")
	}

	// Save the input template.
	if err := s.sm.Set(ctx, id+"/template", t); err != nil {
		l.Err(err).Msg("secrets manager write error")
		return nil, status.Error(codes.Internal, "secrets manager write error")
	}

	// Save the parsed OAuth configuration, if there is one.
	if o.IsUsable() {
		j, err := o.ToJSON()
		if err != nil {
			l.Err(err).Msg("failed to convert proto into JSON")
			return nil, status.Error(codes.Internal, "secrets manager parse error")
		}

		if err := s.sm.Set(ctx, id+"/oauth", j); err != nil {
			l.Err(err).Msg("secrets manager write error")
			return nil, status.Error(codes.Internal, "secrets manager write error")
		}
	}

	l.Trace().Msg("secrets manager write success")
	return thrippypb.CreateLinkResponse_builder{
		LinkId:           proto.String(id),
		CredentialFields: links.Templates[t].CredFields(),
	}.Build(), nil
}

func (s *grpcServer) DeleteLink(ctx context.Context, in *thrippypb.DeleteLinkRequest) (*thrippypb.DeleteLinkResponse, error) {
	id := in.GetLinkId()
	l := log.With().Str("grpc_method", "DeleteLink").Str("id", id).Logger()
	l.Debug().Msg("received gRPC request")

	if id == "" {
		l.Warn().Msg("missing ID")
		return nil, status.Error(codes.InvalidArgument, "missing ID")
	}
	if _, err := shortuuid.DefaultEncoder.Decode(id); err != nil {
		l.Warn().Err(err).Msg("ID is an invalid short UUID")
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}

	t, err := s.sm.Get(ctx, id+"/template")
	if err != nil {
		l.Error().Stack().Err(err).Msg("secrets manager read error")
		return nil, status.Error(codes.Internal, "secrets manager read error")
	}

	if t == "" {
		if in.GetAllowMissing() {
			return &thrippypb.DeleteLinkResponse{}, nil
		} else {
			l.Warn().Stack().Msg("link not found")
			return nil, status.Error(codes.NotFound, "link not found")
		}
	}

	if err := s.sm.Delete(ctx, id+"/creds"); err != nil {
		l.Err(err).Msg("secrets manager delete error: creds")
		return nil, status.Error(codes.Internal, "secrets manager delete error: creds")
	}
	if err := s.sm.Delete(ctx, id+"/meta"); err != nil {
		l.Err(err).Msg("secrets manager delete error: meta")
		return nil, status.Error(codes.Internal, "secrets manager delete error: meta")
	}
	if err := s.sm.Delete(ctx, id+"/oauth"); err != nil {
		l.Err(err).Msg("secrets manager delete error: oauth")
		return nil, status.Error(codes.Internal, "secrets manager delete error: oauth")
	}
	if err := s.sm.Delete(ctx, id+"/template"); err != nil {
		l.Err(err).Msg("secrets manager delete error: template")
		return nil, status.Error(codes.Internal, "secrets manager delete error: template")
	}

	l.Trace().Msg("secrets manager delete success")
	return &thrippypb.DeleteLinkResponse{}, nil
}

func (s *grpcServer) GetLink(ctx context.Context, in *thrippypb.GetLinkRequest) (*thrippypb.GetLinkResponse, error) {
	id := in.GetLinkId()
	l := log.With().Str("grpc_method", "GetLink").Str("id", id).Logger()
	l.Debug().Msg("received gRPC request")

	if id == "" {
		l.Warn().Msg("missing ID")
		return nil, status.Error(codes.InvalidArgument, "missing ID")
	}
	if _, err := shortuuid.DefaultEncoder.Decode(id); err != nil {
		l.Warn().Err(err).Msg("ID is an invalid short UUID")
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}

	t, o, err := s.templateAndOAuth(l.WithContext(ctx), id)
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
	l := log.With().Str("grpc_method", "SetCredentials").Str("id", id).Logger()
	l.Debug().Msg("received gRPC request")

	if id == "" {
		l.Warn().Msg("missing ID")
		return nil, status.Error(codes.InvalidArgument, "missing ID")
	}
	if _, err := shortuuid.DefaultEncoder.Decode(id); err != nil {
		l.Warn().Err(err).Msg("ID is an invalid short UUID")
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}

	template, c, err := s.templateAndOAuth(l.WithContext(ctx), id)
	if err != nil {
		return nil, err
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
		l.Err(err).Msg("failed to convert proto into JSON")
		return nil, status.Error(codes.Internal, "secrets manager parse error")
	}

	if len(j) <= 2 {
		j, err = json.Marshal(m)
		if err != nil {
			l.Err(err).Msg("failed to convert credentials into JSON")
			return nil, status.Error(codes.Internal, "secrets manager parse error")
		}
	}

	// Check the usability of the provided credentials, retrieve their metadata, and save both.
	metadata, err := links.Templates[template].Check(ctx, m, oauth.FromProto(c), oauth.TokenFromProto(token))
	if err != nil {
		l.Err(err).Msg("failed to check credentials / extract metadata")
		return nil, status.Error(codes.Internal, "credentials check error: "+err.Error())
	}

	if err := s.sm.Set(ctx, id+"/creds", string(j)); err != nil {
		l.Err(err).Msg("secrets manager write error")
		return nil, status.Error(codes.Internal, "secrets manager write error")
	}

	if metadata != "" {
		if err := s.sm.Set(ctx, id+"/meta", metadata); err != nil {
			l.Err(err).Msg("secrets manager write error")
			return nil, status.Error(codes.Internal, "secrets manager write error")
		}
	}

	l.Trace().Msg("secrets manager write success")
	return &thrippypb.SetCredentialsResponse{}, nil
}

func (s *grpcServer) templateAndOAuth(ctx context.Context, id string) (string, *thrippypb.OAuthConfig, error) {
	l := zerolog.Ctx(ctx)

	t, err := s.sm.Get(ctx, id+"/template")
	if err != nil {
		l.Error().Stack().Err(err).Msg("secrets manager read error")
		return "", nil, status.Error(codes.Internal, "secrets manager read error")
	}
	if t == "" {
		l.Warn().Stack().Msg("link not found")
		return "", nil, status.Error(codes.NotFound, "link not found")
	}

	o, err := s.sm.Get(ctx, id+"/oauth")
	if err != nil {
		l.Error().Stack().Err(err).Msg("secrets manager read error")
		return "", nil, status.Error(codes.Internal, "secrets manager read error")
	}

	var m *thrippypb.OAuthConfig
	if o != "" {
		m = &thrippypb.OAuthConfig{}
		if err = protojson.Unmarshal([]byte(o), m); err != nil {
			l.Err(err).Msg("failed to convert JSON into proto")
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
	l := log.With().Str("grpc_method", "GetCredentials").Str("id", id).Logger()

	ctx = l.WithContext(ctx)
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

		// Flatten extra secrets from an OAuth token's "raw" map,
		// but in a limited way, to prevent overwriting.
		raw := v.(map[string]any)
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
	l := log.With().Str("grpc_method", "GetMetadata").Str("id", id).Logger()

	ma, err := s.getSecrets(l.WithContext(ctx), id, "/meta")
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
	l := zerolog.Ctx(ctx)
	l.Debug().Msg("received gRPC request")

	if linkID == "" {
		l.Warn().Msg("missing ID")
		return nil, status.Error(codes.InvalidArgument, "missing ID")
	}
	if _, err := shortuuid.DefaultEncoder.Decode(linkID); err != nil {
		l.Warn().Err(err).Msg("ID is an invalid short UUID")
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}

	j, err := s.sm.Get(ctx, linkID+keySuffix)
	if err != nil {
		l.Err(err).Msg("secrets manager read error")
		return nil, status.Error(codes.Internal, "secrets manager read error")
	}

	var m map[string]any
	if j != "" {
		if err := json.Unmarshal([]byte(j), &m); err != nil {
			l.Err(err).Msg("failed to convert JSON into map")
			return nil, status.Error(codes.Internal, "secrets manager parse error")
		}
	}

	return m, nil
}

func (s *grpcServer) refreshOAuthToken(ctx context.Context, id string, t *oauth2.Token) (map[string]any, error) {
	l := zerolog.Ctx(ctx)

	jsonConfig, err := s.sm.Get(ctx, id+"/oauth")
	if err != nil {
		l.Err(err).Msg("secrets manager read error")
		return nil, status.Error(codes.Internal, "secrets manager read error")
	}

	o := &thrippypb.OAuthConfig{}
	if err := protojson.Unmarshal([]byte(jsonConfig), o); err != nil {
		l.Err(err).Msg("failed to convert JSON into proto")
		return nil, status.Error(codes.Internal, "secrets manager parse error")
	}

	m, err := oauth.FromProto(o).RefreshToken(ctx, t, false)
	if err != nil {
		l.Err(err).Msg("failed to refresh OAuth token")
		return nil, status.Error(codes.Internal, "OAuth token refresh error")
	}

	if raw := s.getRaw(ctx, id); raw != nil {
		m["raw"] = raw
	}

	jsonToken, err := json.Marshal(m)
	if err != nil {
		l.Err(err).Msg("failed to convert map into JSON")
		return nil, status.Error(codes.Internal, "secrets manager parse error")
	}

	if err := s.sm.Set(ctx, id+"/creds", string(jsonToken)); err != nil {
		l.Err(err).Msg("secrets manager write error")
		return nil, status.Error(codes.Internal, "secrets manager write error")
	}

	return m, nil
}
