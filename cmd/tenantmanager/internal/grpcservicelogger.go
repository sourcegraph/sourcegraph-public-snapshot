package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	proto "github.com/sourcegraph/sourcegraph/cmd/tenantmanager/shared/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/grpcutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type loggingTenantManagerServiceServer struct {
	base   proto.TenantManagerServiceServer
	logger log.Logger

	proto.UnsafeTenantManagerServiceServer
}

func doLog(logger log.Logger, fullMethod string, statusCode codes.Code, traceID string, duration time.Duration, requestFields ...log.Field) {
	server, method := grpcutil.SplitMethodName(fullMethod)

	fields := []log.Field{
		log.String("server", server),
		log.String("method", method),
		log.String("status", statusCode.String()),
		log.String("traceID", traceID),
		log.Duration("duration", duration),
	}

	if len(requestFields) > 0 {
		fields = append(fields, log.Object("request", requestFields...))
	} else {
		fields = append(fields, log.String("request", "<empty>"))
	}

	logger.Debug(fmt.Sprintf("Handled %s RPC", method), fields...)
}

func (l *loggingTenantManagerServiceServer) DeleteTenant(ctx context.Context, request *proto.DeleteTenantRequest) (resp *proto.DeleteTenantResponse, err error) {
	start := time.Now()

	defer func() {
		elapsed := time.Since(start)

		doLog(
			l.logger,

			proto.TenantManagerService_DeleteTenant_FullMethodName,
			status.Code(err),
			trace.Context(ctx).TraceID,
			elapsed,

			deleteTenantRequestToLogFields(request)...,
		)

	}()
	return l.base.DeleteTenant(ctx, request)
}

func deleteTenantRequestToLogFields(req *proto.DeleteTenantRequest) []log.Field {
	return []log.Field{
		log.Int("id", int(req.GetId())),
		log.String("name", req.GetName()),
	}
}

func (l *loggingTenantManagerServiceServer) CreateTenant(ctx context.Context, request *proto.CreateTenantRequest) (resp *proto.CreateTenantResponse, err error) {
	start := time.Now()

	defer func() {
		elapsed := time.Since(start)

		doLog(
			l.logger,

			proto.TenantManagerService_CreateTenant_FullMethodName,
			status.Code(err),
			trace.Context(ctx).TraceID,
			elapsed,

			createTenantRequestToLogFields(request)...,
		)
	}()

	return l.base.CreateTenant(ctx, request)
}

func createTenantRequestToLogFields(req *proto.CreateTenantRequest) []log.Field {
	return []log.Field{
		log.String("name", req.GetName()),
	}
}

func (l *loggingTenantManagerServiceServer) ListTenants(ctx context.Context, request *proto.ListTenantsRequest) (resp *proto.ListTenantsResponse, err error) {
	start := time.Now()

	defer func() {
		elapsed := time.Since(start)

		doLog(
			l.logger,

			proto.TenantManagerService_ListTenants_FullMethodName,
			status.Code(err),
			trace.Context(ctx).TraceID,
			elapsed,

			listTenantsRequestToLogFields(request)...,
		)
	}()

	return l.base.ListTenants(ctx, request)
}

func listTenantsRequestToLogFields(req *proto.ListTenantsRequest) []log.Field {
	return []log.Field{
		log.Int("page_size", int(req.GetPageSize())),
		log.String("page_token", req.GetPageToken()),
	}
}

var (
	_ proto.TenantManagerServiceServer = &loggingTenantManagerServiceServer{}
)
