package server

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/openfga/openfga/internal/utils"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/openfga/openfga/internal/condition"
	"github.com/openfga/openfga/internal/graph"
	"github.com/openfga/openfga/pkg/middleware/validator"
	"github.com/openfga/openfga/pkg/telemetry"
	"github.com/openfga/openfga/pkg/tuple"

	"github.com/openfga/openfga/pkg/server/commands/listusers"
	serverErrors "github.com/openfga/openfga/pkg/server/errors"
	"github.com/openfga/openfga/pkg/typesystem"
)

// ListUsers returns all users (e.g. subjects) matching a specific user filter criteria
// that have a specific relation with some object.
func (s *Server) ListUsers(
	ctx context.Context,
	req *openfgav1.ListUsersRequest,
) (*openfgav1.ListUsersResponse, error) {
	if !s.IsExperimentallyEnabled(ExperimentalEnableListUsers) {
		return nil, status.Error(codes.Unimplemented, "ListUsers is not enabled. It can be enabled for experimental use by passing the `--experimentals enable-list-users` configuration option when running OpenFGA server")
	}

	start := time.Now()
	ctx, span := tracer.Start(ctx, "ListUsers", trace.WithAttributes(
		attribute.String("store_id", req.GetStoreId()),
		attribute.String("object", tuple.BuildObject(req.GetObject().GetType(), req.GetObject().GetId())),
		attribute.String("relation", req.GetRelation()),
		attribute.String("user_filters", userFiltersToString(req.GetUserFilters())),
	))
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	const methodName = "listusers"

	typesys, err := s.resolveTypesystem(ctx, req.GetStoreId(), req.GetAuthorizationModelId())
	if err != nil {
		return nil, err
	}

	err = listusers.ValidateListUsersRequest(ctx, req, typesys)
	if err != nil {
		return nil, err
	}

	ctx = typesystem.ContextWithTypesystem(ctx, typesys)

	listUsersQuery := listusers.NewListUsersQuery(s.datastore,
		listusers.WithResolveNodeLimit(s.resolveNodeLimit),
		listusers.WithResolveNodeBreadthLimit(s.resolveNodeBreadthLimit),
		listusers.WithListUsersQueryLogger(s.logger),
		listusers.WithListUsersMaxResults(s.listUsersMaxResults),
		listusers.WithListUsersDeadline(s.listUsersDeadline),
		listusers.WithListUsersMaxConcurrentReads(s.maxConcurrentReadsForListUsers),
	)

	resp, err := listUsersQuery.ListUsers(ctx, req)
	if err != nil {
		telemetry.TraceError(span, err)

		switch {
		case errors.Is(err, graph.ErrResolutionDepthExceeded):
			return nil, serverErrors.AuthorizationModelResolutionTooComplex
		case errors.Is(err, condition.ErrEvaluationFailed):
			return nil, serverErrors.ValidationError(err)
		default:
			return nil, serverErrors.HandleError("", err)
		}
	}

	datastoreQueryCount := float64(resp.Metadata.DatastoreQueryCount)

	grpc_ctxtags.Extract(ctx).Set(datastoreQueryCountHistogramName, datastoreQueryCount)
	span.SetAttributes(attribute.Float64(datastoreQueryCountHistogramName, datastoreQueryCount))
	datastoreQueryCountHistogram.WithLabelValues(
		s.serviceName,
		methodName,
	).Observe(datastoreQueryCount)

	dispatchCount := float64(resp.Metadata.DispatchCounter.Load())
	grpc_ctxtags.Extract(ctx).Set(dispatchCountHistogramName, dispatchCount)
	span.SetAttributes(attribute.Float64(dispatchCountHistogramName, dispatchCount))
	dispatchCountHistogram.WithLabelValues(
		s.serviceName,
		methodName,
	).Observe(dispatchCount)

	requestDurationHistogram.WithLabelValues(
		s.serviceName,
		methodName,
		utils.Bucketize(uint(datastoreQueryCount), s.requestDurationByQueryHistogramBuckets),
		utils.Bucketize(uint(dispatchCount), s.requestDurationByDispatchCountHistogramBuckets),
	).Observe(float64(time.Since(start).Milliseconds()))

	return &openfgav1.ListUsersResponse{
		Users:         resp.GetUsers(),
		ExcludedUsers: resp.GetExcludedUsers(),
	}, nil
}

func userFiltersToString(filter []*openfgav1.UserTypeFilter) string {
	var s strings.Builder
	for _, f := range filter {
		s.WriteString(f.GetType())
		if f.GetRelation() != "" {
			s.WriteString("#" + f.GetRelation())
		}
	}
	return s.String()
}
