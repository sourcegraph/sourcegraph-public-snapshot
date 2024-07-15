package internal

import (
	"context"
	"encoding/base64"

	"github.com/sourcegraph/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/cmd/tenantmanager/internal/store"
	proto "github.com/sourcegraph/sourcegraph/cmd/tenantmanager/shared/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GRPCTenantManagerServiceConfig struct {
	ExhaustiveRequestLoggingEnabled bool
}

func NewTenantManagerServiceServer(logger log.Logger, store store.Store, config *GRPCTenantManagerServiceConfig) proto.TenantManagerServiceServer {
	var srv proto.TenantManagerServiceServer = &tenantManagerServiceServer{
		logger: logger,
		store:  store,
	}

	if config.ExhaustiveRequestLoggingEnabled {
		logger := logger.Scoped("gRPCRequestLogger")

		srv = &loggingTenantManagerServiceServer{
			base:   srv,
			logger: logger,
		}
	}

	return srv
}

type tenantManagerServiceServer struct {
	logger log.Logger
	store  store.Store

	proto.UnimplementedTenantManagerServiceServer
}

var _ proto.TenantManagerServiceServer = &tenantManagerServiceServer{}

func (s *tenantManagerServiceServer) CreateTenant(ctx context.Context, req *proto.CreateTenantRequest) (*proto.CreateTenantResponse, error) {
	name := req.GetName()
	if name == "" {
		return nil, status.New(codes.InvalidArgument, "name must be specified").Err()
	}

	tnt, err := s.store.CreateTenant(ctx, name)
	if err != nil {
		return nil, status.New(codes.Internal, errors.Wrap(err, "failed to create tenant in database").Error()).Err()
	}

	return &proto.CreateTenantResponse{
		Tenant: &proto.Tenant{
			Id:   tnt.ID,
			Name: tnt.Name,
			// DisplayName: tnt.DisplayName,
			CreatedAt: timestamppb.New(tnt.CreatedAt),
			UpdatedAt: timestamppb.New(tnt.UpdatedAt),
		},
	}, nil
}

func (s *tenantManagerServiceServer) DeleteTenant(ctx context.Context, req *proto.DeleteTenantRequest) (*proto.DeleteTenantResponse, error) {
	if req.GetName() == "" && req.GetId() == 0 {
		return nil, status.New(codes.InvalidArgument, "name or id must be specified").Err()
	}

	id := req.GetId()

	if id == 0 {
		tnt, err := s.store.GetByName(ctx, req.GetName())
		if err != nil {
			// TODO: Handle notfound.
			return nil, status.New(codes.Internal, "failed to get tenant from database").Err()
		}
		id = tnt.ID
	}

	err := s.store.DeleteTenant(ctx, id)
	if err != nil {
		// TODO: handle notfound.
		err = errors.Wrap(err, "removing tenant")
		s.logger.Error("failed to delete tenant", log.Int("tenant", int(id)), log.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to delete tenant %d: %s", id, err)
	}

	s.logger.Info("tenant deleted", log.Int("tenant", int(id)))

	return &proto.DeleteTenantResponse{}, nil
}

func (s *tenantManagerServiceServer) ListTenants(ctx context.Context, req *proto.ListTenantsRequest) (*proto.ListTenantsResponse, error) {
	if req.GetPageSize() == 0 {
		return nil, status.New(codes.InvalidArgument, "page_size must be > 0").Err()
	}

	token, err := base64.StdEncoding.DecodeString(req.GetPageToken())
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid page_token").Err()
	}

	tenants, nextCursor, err := s.store.ListTenants(ctx, store.ListTenantsOptions{
		Cursor: string(token),
		Limit:  int(req.GetPageSize()),
	})
	if err != nil {
		return nil, status.New(codes.Internal, errors.Wrap(err, "failed to list tenants").Error()).Err()
	}

	pts := make([]*proto.Tenant, 0, len(tenants))
	for _, tnt := range tenants {
		pts = append(pts, &proto.Tenant{
			Id:   tnt.ID,
			Name: tnt.Name,
			// DisplayName: tnt.DisplayName,
			CreatedAt: timestamppb.New(tnt.CreatedAt),
			UpdatedAt: timestamppb.New(tnt.UpdatedAt),
		})
	}

	return &proto.ListTenantsResponse{
		Tenants:       pts,
		NextPageToken: base64.StdEncoding.EncodeToString([]byte(nextCursor)),
	}, nil
}
