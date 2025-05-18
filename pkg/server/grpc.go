package server

import (
	"context"
	"encoding/json"
	"net"

	"github.com/lithammer/shortuuid/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/tzrikka/thrippy/pkg/links"
	"github.com/tzrikka/thrippy/pkg/oauth"
	"github.com/tzrikka/thrippy/pkg/secrets"
	thrippypb "github.com/tzrikka/thrippy/proto/thrippy/v1"
)

type grpcServer struct {
	thrippypb.UnimplementedThrippyServiceServer
	sm secrets.Manager
}

// startGRPCServer starts a gRPC server for the [Thrippy service]. This
// is non-blocking, in order to let Thrippy run an HTTP server as well.
//
// [Thrippy service]: https://github.com/tzrikka/thrippy/blob/main/proto/thrippy/v1/thrippy.proto
func startGRPCServer(sm secrets.Manager, addr string) (string, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Err(err).Send()
		return "", err
	}

	srv := grpc.NewServer()
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

	t := in.GetTemplate()
	if _, ok := links.Templates[t]; !ok {
		l.Warn().Str("template", t).Msg("invalid template")
		return nil, status.Error(codes.InvalidArgument, "invalid template")
	}

	o := oauth.FromProto(in.GetOauthConfig())
	links.ModifyOAuthByTemplate(o, t)
	if o != nil && o.Config.Endpoint.AuthURL != "" && o.Config.ClientID == "" {
		l.Warn().Msg("missing OAuth client ID")
		return nil, status.Error(codes.InvalidArgument, "missing OAuth client ID")
	}

	if err := s.sm.Set(ctx, id+"/template", t); err != nil {
		l.Err(err).Msg("secrets manager write error")
		return nil, status.Error(codes.Internal, "secrets manager write error")
	}

	j, err := o.ToJSON()
	if err != nil {
		l.Err(err).Msg("failed to convert proto into JSON")
		return nil, status.Error(codes.Internal, "secrets manager parse error")
	}

	if len(j) > 2 { // Save only non-empty OAuth configs.
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

	// Credentials to store: either an OAuth token or a
	// generic string map - whichever of them isn't empty.
	token := in.GetToken()
	m := in.GetGenericCreds()
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

	metadata, err := links.Templates[template].Check(ctx, m, oauth.FromProto(c), oauth.TokenFromProto(token))
	if err != nil {
		l.Err(err).Msg("failed to check credentials / extract metadata")
		return nil, status.Error(codes.Internal, "credentials check error: "+err.Error())
	}

	if err := s.sm.Set(ctx, id+"/creds", string(j)); err != nil {
		l.Err(err).Msg("secrets manager write error")
		return nil, status.Error(codes.Internal, "secrets manager write error")
	}

	if err := s.sm.Set(ctx, id+"/meta", metadata); err != nil {
		l.Err(err).Msg("secrets manager write error")
		return nil, status.Error(codes.Internal, "secrets manager write error")
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
		err = protojson.Unmarshal([]byte(o), m)
		if err != nil {
			l.Err(err).Msg("failed to convert JSON into proto")
			return "", nil, status.Error(codes.Internal, "secrets manager parse error")
		}
	}

	return t, m, nil
}

func (s *grpcServer) GetCredentials(ctx context.Context, in *thrippypb.GetCredentialsRequest) (*thrippypb.GetCredentialsResponse, error) {
	id := in.GetLinkId()
	l := log.With().Str("grpc_method", "GetCredentials").Str("id", id).Logger()

	m, err := s.getSecrets(l.WithContext(ctx), id, "/creds")
	if err != nil {
		return nil, err
	}

	return thrippypb.GetCredentialsResponse_builder{Credentials: m}.Build(), nil
}

func (s *grpcServer) GetMetadata(ctx context.Context, in *thrippypb.GetMetadataRequest) (*thrippypb.GetMetadataResponse, error) {
	id := in.GetLinkId()
	l := log.With().Str("grpc_method", "GetMetadata").Str("id", id).Logger()

	m, err := s.getSecrets(l.WithContext(ctx), id, "/meta")
	if err != nil {
		return nil, err
	}

	return thrippypb.GetMetadataResponse_builder{Metadata: m}.Build(), nil
}

func (s *grpcServer) getSecrets(ctx context.Context, linkID, keySuffix string) (map[string]string, error) {
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

	var m map[string]string
	if j != "" {
		if err := json.Unmarshal([]byte(j), &m); err != nil {
			l.Err(err).Msg("failed to convert JSON into map")
			return nil, status.Error(codes.Internal, "secrets manager parse error")
		}
	}

	return m, nil
}
