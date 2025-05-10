package server

import (
	"context"
	"encoding/json"
	"net"

	"github.com/lithammer/shortuuid/v4"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/tzrikka/trippy/pkg/links"
	"github.com/tzrikka/trippy/pkg/oauth"
	"github.com/tzrikka/trippy/pkg/secrets"
	trippypb "github.com/tzrikka/trippy/proto/trippy/v1"
)

type grpcServer struct {
	trippypb.UnimplementedTrippyServiceServer
	sm secrets.Manager
}

// startGRPCServer starts a gRPC server for the Trippy service (proto/.../trippy.proto).
// This function is non-blocking, in order to let Trippy run an HTTP server as well.
func startGRPCServer(sm secrets.Manager, addr string) (string, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Err(err).Send()
		return "", err
	}

	srv := grpc.NewServer()
	trippypb.RegisterTrippyServiceServer(srv, &grpcServer{sm: sm})
	go func() {
		err = srv.Serve(lis)
		if err != nil {
			log.Fatal().Err(err).Msg("gRPC serving error")
		}
	}()

	log.Info().Msgf("gRPC server listening on %s", lis.Addr().String())
	return lis.Addr().String(), nil
}

func (s *grpcServer) CreateLink(ctx context.Context, in *trippypb.CreateLinkRequest) (*trippypb.CreateLinkResponse, error) {
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
	if o != nil && o.Config.Endpoint.AuthURL != "" && (o.Config.ClientID == "" || o.Config.ClientSecret == "") {
		l.Warn().Msg("missing OAuth client ID and/or secret")
		return nil, status.Error(codes.InvalidArgument, "missing OAuth client ID and/or secret")
	}

	if err := s.sm.Set(ctx, id+"/template", t); err != nil {
		l.Err(err).Msg("secrets manager write error")
		return nil, status.Error(codes.Internal, "secrets manager write error")
	}

	j, err := o.ToJSON()
	if err != nil {
		l.Err(err).Msg("failed to transform proto into JSON")
		return nil, status.Error(codes.Internal, "secrets manager parse error")
	}

	if len(j) > 2 { // Save only non-empty OAuth configs.
		if err := s.sm.Set(ctx, id+"/oauth", j); err != nil {
			l.Err(err).Msg("secrets manager write error")
			return nil, status.Error(codes.Internal, "secrets manager write error")
		}
	}

	l.Trace().Msg("secrets manager write success")
	return trippypb.CreateLinkResponse_builder{
		LinkId:           proto.String(id),
		CredentialFields: links.Templates[t].CredsFields,
	}.Build(), nil
}

func (s *grpcServer) GetLink(ctx context.Context, in *trippypb.GetLinkRequest) (*trippypb.GetLinkResponse, error) {
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

	t, err := s.sm.Get(ctx, id+"/template")
	if err != nil {
		l.Err(err).Msg("secrets manager read error")
		return nil, status.Error(codes.Internal, "secrets manager read error")
	}
	if t == "" {
		l.Warn().Msg("link not found")
		return nil, status.Error(codes.NotFound, "link not found")
	}

	o, err := s.sm.Get(ctx, id+"/oauth")
	if err != nil {
		l.Err(err).Msg("secrets manager read error")
		return nil, status.Error(codes.Internal, "secrets manager read error")
	}

	var m *trippypb.OAuthConfig
	if o != "" {
		m = &trippypb.OAuthConfig{}
		err = protojson.Unmarshal([]byte(o), m)
		if err != nil {
			l.Err(err).Msg("failed to transform JSON into proto")
			return nil, status.Error(codes.Internal, "secrets manager parse error")
		}
	}

	return trippypb.GetLinkResponse_builder{
		Template:         proto.String(t),
		OauthConfig:      m,
		CredentialFields: links.Templates[t].CredsFields,
	}.Build(), nil
}

func (s *grpcServer) SetCredentials(ctx context.Context, in *trippypb.SetCredentialsRequest) (*trippypb.SetCredentialsResponse, error) {
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

	j, err := json.Marshal(in.GetCredentials())
	if err != nil {
		l.Err(err).Msg("failed to transform map into JSON")
		return nil, status.Error(codes.Internal, "secrets manager parse error")
	}

	if len(j) > 2 { // Save only non-empty maps.
		if err := s.sm.Set(ctx, id+"/creds", string(j)); err != nil {
			l.Err(err).Msg("secrets manager write error")
			return nil, status.Error(codes.Internal, "secrets manager write error")
		}
	}

	l.Trace().Msg("secrets manager write success")
	return &trippypb.SetCredentialsResponse{}, nil
}

func (s *grpcServer) GetCredentials(ctx context.Context, in *trippypb.GetCredentialsRequest) (*trippypb.GetCredentialsResponse, error) {
	id := in.GetLinkId()
	l := log.With().Str("grpc_method", "GetCredentials").Str("id", id).Logger()
	l.Debug().Msg("received gRPC request")

	if id == "" {
		l.Warn().Msg("missing ID")
		return nil, status.Error(codes.InvalidArgument, "missing ID")
	}
	if _, err := shortuuid.DefaultEncoder.Decode(id); err != nil {
		l.Warn().Err(err).Msg("ID is an invalid short UUID")
		return nil, status.Error(codes.InvalidArgument, "invalid ID")
	}

	j, err := s.sm.Get(ctx, id+"/creds")
	if err != nil {
		l.Err(err).Msg("secrets manager read error")
		return nil, status.Error(codes.Internal, "secrets manager read error")
	}

	var m map[string]string
	if j != "" {
		if err := json.Unmarshal([]byte(j), &m); err != nil {
			l.Err(err).Msg("failed to transform JSON into map")
			return nil, status.Error(codes.Internal, "secrets manager parse error")
		}
	}

	return trippypb.GetCredentialsResponse_builder{Credentials: m}.Build(), nil
}
