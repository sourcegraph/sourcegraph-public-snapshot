// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: subscriptions.proto

package v1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// SubscriptionsServiceName is the fully-qualified name of the SubscriptionsService service.
	SubscriptionsServiceName = "enterpriseportal.subscriptions.v1.SubscriptionsService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// SubscriptionsServiceGetEnterpriseSubscriptionProcedure is the fully-qualified name of the
	// SubscriptionsService's GetEnterpriseSubscription RPC.
	SubscriptionsServiceGetEnterpriseSubscriptionProcedure = "/enterpriseportal.subscriptions.v1.SubscriptionsService/GetEnterpriseSubscription"
	// SubscriptionsServiceListEnterpriseSubscriptionsProcedure is the fully-qualified name of the
	// SubscriptionsService's ListEnterpriseSubscriptions RPC.
	SubscriptionsServiceListEnterpriseSubscriptionsProcedure = "/enterpriseportal.subscriptions.v1.SubscriptionsService/ListEnterpriseSubscriptions"
	// SubscriptionsServiceListEnterpriseSubscriptionLicensesProcedure is the fully-qualified name of
	// the SubscriptionsService's ListEnterpriseSubscriptionLicenses RPC.
	SubscriptionsServiceListEnterpriseSubscriptionLicensesProcedure = "/enterpriseportal.subscriptions.v1.SubscriptionsService/ListEnterpriseSubscriptionLicenses"
	// SubscriptionsServiceCreateEnterpriseSubscriptionLicenseProcedure is the fully-qualified name of
	// the SubscriptionsService's CreateEnterpriseSubscriptionLicense RPC.
	SubscriptionsServiceCreateEnterpriseSubscriptionLicenseProcedure = "/enterpriseportal.subscriptions.v1.SubscriptionsService/CreateEnterpriseSubscriptionLicense"
	// SubscriptionsServiceRevokeEnterpriseSubscriptionLicenseProcedure is the fully-qualified name of
	// the SubscriptionsService's RevokeEnterpriseSubscriptionLicense RPC.
	SubscriptionsServiceRevokeEnterpriseSubscriptionLicenseProcedure = "/enterpriseportal.subscriptions.v1.SubscriptionsService/RevokeEnterpriseSubscriptionLicense"
	// SubscriptionsServiceUpdateEnterpriseSubscriptionProcedure is the fully-qualified name of the
	// SubscriptionsService's UpdateEnterpriseSubscription RPC.
	SubscriptionsServiceUpdateEnterpriseSubscriptionProcedure = "/enterpriseportal.subscriptions.v1.SubscriptionsService/UpdateEnterpriseSubscription"
	// SubscriptionsServiceArchiveEnterpriseSubscriptionProcedure is the fully-qualified name of the
	// SubscriptionsService's ArchiveEnterpriseSubscription RPC.
	SubscriptionsServiceArchiveEnterpriseSubscriptionProcedure = "/enterpriseportal.subscriptions.v1.SubscriptionsService/ArchiveEnterpriseSubscription"
	// SubscriptionsServiceCreateEnterpriseSubscriptionProcedure is the fully-qualified name of the
	// SubscriptionsService's CreateEnterpriseSubscription RPC.
	SubscriptionsServiceCreateEnterpriseSubscriptionProcedure = "/enterpriseportal.subscriptions.v1.SubscriptionsService/CreateEnterpriseSubscription"
	// SubscriptionsServiceUpdateSubscriptionMembershipProcedure is the fully-qualified name of the
	// SubscriptionsService's UpdateSubscriptionMembership RPC.
	SubscriptionsServiceUpdateSubscriptionMembershipProcedure = "/enterpriseportal.subscriptions.v1.SubscriptionsService/UpdateSubscriptionMembership"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	subscriptionsServiceServiceDescriptor                                   = v1.File_subscriptions_proto.Services().ByName("SubscriptionsService")
	subscriptionsServiceGetEnterpriseSubscriptionMethodDescriptor           = subscriptionsServiceServiceDescriptor.Methods().ByName("GetEnterpriseSubscription")
	subscriptionsServiceListEnterpriseSubscriptionsMethodDescriptor         = subscriptionsServiceServiceDescriptor.Methods().ByName("ListEnterpriseSubscriptions")
	subscriptionsServiceListEnterpriseSubscriptionLicensesMethodDescriptor  = subscriptionsServiceServiceDescriptor.Methods().ByName("ListEnterpriseSubscriptionLicenses")
	subscriptionsServiceCreateEnterpriseSubscriptionLicenseMethodDescriptor = subscriptionsServiceServiceDescriptor.Methods().ByName("CreateEnterpriseSubscriptionLicense")
	subscriptionsServiceRevokeEnterpriseSubscriptionLicenseMethodDescriptor = subscriptionsServiceServiceDescriptor.Methods().ByName("RevokeEnterpriseSubscriptionLicense")
	subscriptionsServiceUpdateEnterpriseSubscriptionMethodDescriptor        = subscriptionsServiceServiceDescriptor.Methods().ByName("UpdateEnterpriseSubscription")
	subscriptionsServiceArchiveEnterpriseSubscriptionMethodDescriptor       = subscriptionsServiceServiceDescriptor.Methods().ByName("ArchiveEnterpriseSubscription")
	subscriptionsServiceCreateEnterpriseSubscriptionMethodDescriptor        = subscriptionsServiceServiceDescriptor.Methods().ByName("CreateEnterpriseSubscription")
	subscriptionsServiceUpdateSubscriptionMembershipMethodDescriptor        = subscriptionsServiceServiceDescriptor.Methods().ByName("UpdateSubscriptionMembership")
)

// SubscriptionsServiceClient is a client for the
// enterpriseportal.subscriptions.v1.SubscriptionsService service.
type SubscriptionsServiceClient interface {
	// GetEnterpriseSubscription retrieves an exact match on an Enterprise subscription.
	GetEnterpriseSubscription(context.Context, *connect.Request[v1.GetEnterpriseSubscriptionRequest]) (*connect.Response[v1.GetEnterpriseSubscriptionResponse], error)
	// ListEnterpriseSubscriptions queries for Enterprise subscriptions.
	ListEnterpriseSubscriptions(context.Context, *connect.Request[v1.ListEnterpriseSubscriptionsRequest]) (*connect.Response[v1.ListEnterpriseSubscriptionsResponse], error)
	// ListEnterpriseSubscriptionLicenses queries for licenses associated with
	// Enterprise subscription licenses, with the ability to list licenses across
	// all subscriptions, or just a specific subscription.
	//
	// Each subscription owns a collection of licenses, typically a series of
	// licenses with the most recent one being a subscription's active license.
	ListEnterpriseSubscriptionLicenses(context.Context, *connect.Request[v1.ListEnterpriseSubscriptionLicensesRequest]) (*connect.Response[v1.ListEnterpriseSubscriptionLicensesResponse], error)
	// CreateEnterpriseSubscription creates license for an Enterprise subscription.
	CreateEnterpriseSubscriptionLicense(context.Context, *connect.Request[v1.CreateEnterpriseSubscriptionLicenseRequest]) (*connect.Response[v1.CreateEnterpriseSubscriptionLicenseResponse], error)
	// RevokeEnterpriseSubscriptionLicense revokes an existing license for an
	// Enterprise subscription, permanently disabling its use for features
	// managed by Sourcegraph. Revocation cannot be undone.
	RevokeEnterpriseSubscriptionLicense(context.Context, *connect.Request[v1.RevokeEnterpriseSubscriptionLicenseRequest]) (*connect.Response[v1.RevokeEnterpriseSubscriptionLicenseResponse], error)
	// UpdateEnterpriseSubscription updates an existing Enterprise subscription.
	// Only properties specified by the update_mask are applied.
	UpdateEnterpriseSubscription(context.Context, *connect.Request[v1.UpdateEnterpriseSubscriptionRequest]) (*connect.Response[v1.UpdateEnterpriseSubscriptionResponse], error)
	// ArchiveEnterpriseSubscriptionRequest archives an existing Enterprise
	// subscription. This is a permanent operation, and cannot be undone.
	//
	// Archiving a subscription also immediately and permanently revokes all
	// associated licenses.
	ArchiveEnterpriseSubscription(context.Context, *connect.Request[v1.ArchiveEnterpriseSubscriptionRequest]) (*connect.Response[v1.ArchiveEnterpriseSubscriptionResponse], error)
	// CreateEnterpriseSubscription creates an Enterprise subscription.
	CreateEnterpriseSubscription(context.Context, *connect.Request[v1.CreateEnterpriseSubscriptionRequest]) (*connect.Response[v1.CreateEnterpriseSubscriptionResponse], error)
	// UpdateSubscriptionMembership updates a subscription membership. It creates
	// a new one if it does not exist and allow_missing is set to true.
	UpdateSubscriptionMembership(context.Context, *connect.Request[v1.UpdateSubscriptionMembershipRequest]) (*connect.Response[v1.UpdateSubscriptionMembershipResponse], error)
}

// NewSubscriptionsServiceClient constructs a client for the
// enterpriseportal.subscriptions.v1.SubscriptionsService service. By default, it uses the Connect
// protocol with the binary Protobuf Codec, asks for gzipped responses, and sends uncompressed
// requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewSubscriptionsServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) SubscriptionsServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &subscriptionsServiceClient{
		getEnterpriseSubscription: connect.NewClient[v1.GetEnterpriseSubscriptionRequest, v1.GetEnterpriseSubscriptionResponse](
			httpClient,
			baseURL+SubscriptionsServiceGetEnterpriseSubscriptionProcedure,
			connect.WithSchema(subscriptionsServiceGetEnterpriseSubscriptionMethodDescriptor),
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		listEnterpriseSubscriptions: connect.NewClient[v1.ListEnterpriseSubscriptionsRequest, v1.ListEnterpriseSubscriptionsResponse](
			httpClient,
			baseURL+SubscriptionsServiceListEnterpriseSubscriptionsProcedure,
			connect.WithSchema(subscriptionsServiceListEnterpriseSubscriptionsMethodDescriptor),
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		listEnterpriseSubscriptionLicenses: connect.NewClient[v1.ListEnterpriseSubscriptionLicensesRequest, v1.ListEnterpriseSubscriptionLicensesResponse](
			httpClient,
			baseURL+SubscriptionsServiceListEnterpriseSubscriptionLicensesProcedure,
			connect.WithSchema(subscriptionsServiceListEnterpriseSubscriptionLicensesMethodDescriptor),
			connect.WithIdempotency(connect.IdempotencyNoSideEffects),
			connect.WithClientOptions(opts...),
		),
		createEnterpriseSubscriptionLicense: connect.NewClient[v1.CreateEnterpriseSubscriptionLicenseRequest, v1.CreateEnterpriseSubscriptionLicenseResponse](
			httpClient,
			baseURL+SubscriptionsServiceCreateEnterpriseSubscriptionLicenseProcedure,
			connect.WithSchema(subscriptionsServiceCreateEnterpriseSubscriptionLicenseMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		revokeEnterpriseSubscriptionLicense: connect.NewClient[v1.RevokeEnterpriseSubscriptionLicenseRequest, v1.RevokeEnterpriseSubscriptionLicenseResponse](
			httpClient,
			baseURL+SubscriptionsServiceRevokeEnterpriseSubscriptionLicenseProcedure,
			connect.WithSchema(subscriptionsServiceRevokeEnterpriseSubscriptionLicenseMethodDescriptor),
			connect.WithIdempotency(connect.IdempotencyIdempotent),
			connect.WithClientOptions(opts...),
		),
		updateEnterpriseSubscription: connect.NewClient[v1.UpdateEnterpriseSubscriptionRequest, v1.UpdateEnterpriseSubscriptionResponse](
			httpClient,
			baseURL+SubscriptionsServiceUpdateEnterpriseSubscriptionProcedure,
			connect.WithSchema(subscriptionsServiceUpdateEnterpriseSubscriptionMethodDescriptor),
			connect.WithIdempotency(connect.IdempotencyIdempotent),
			connect.WithClientOptions(opts...),
		),
		archiveEnterpriseSubscription: connect.NewClient[v1.ArchiveEnterpriseSubscriptionRequest, v1.ArchiveEnterpriseSubscriptionResponse](
			httpClient,
			baseURL+SubscriptionsServiceArchiveEnterpriseSubscriptionProcedure,
			connect.WithSchema(subscriptionsServiceArchiveEnterpriseSubscriptionMethodDescriptor),
			connect.WithIdempotency(connect.IdempotencyIdempotent),
			connect.WithClientOptions(opts...),
		),
		createEnterpriseSubscription: connect.NewClient[v1.CreateEnterpriseSubscriptionRequest, v1.CreateEnterpriseSubscriptionResponse](
			httpClient,
			baseURL+SubscriptionsServiceCreateEnterpriseSubscriptionProcedure,
			connect.WithSchema(subscriptionsServiceCreateEnterpriseSubscriptionMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		updateSubscriptionMembership: connect.NewClient[v1.UpdateSubscriptionMembershipRequest, v1.UpdateSubscriptionMembershipResponse](
			httpClient,
			baseURL+SubscriptionsServiceUpdateSubscriptionMembershipProcedure,
			connect.WithSchema(subscriptionsServiceUpdateSubscriptionMembershipMethodDescriptor),
			connect.WithIdempotency(connect.IdempotencyIdempotent),
			connect.WithClientOptions(opts...),
		),
	}
}

// subscriptionsServiceClient implements SubscriptionsServiceClient.
type subscriptionsServiceClient struct {
	getEnterpriseSubscription           *connect.Client[v1.GetEnterpriseSubscriptionRequest, v1.GetEnterpriseSubscriptionResponse]
	listEnterpriseSubscriptions         *connect.Client[v1.ListEnterpriseSubscriptionsRequest, v1.ListEnterpriseSubscriptionsResponse]
	listEnterpriseSubscriptionLicenses  *connect.Client[v1.ListEnterpriseSubscriptionLicensesRequest, v1.ListEnterpriseSubscriptionLicensesResponse]
	createEnterpriseSubscriptionLicense *connect.Client[v1.CreateEnterpriseSubscriptionLicenseRequest, v1.CreateEnterpriseSubscriptionLicenseResponse]
	revokeEnterpriseSubscriptionLicense *connect.Client[v1.RevokeEnterpriseSubscriptionLicenseRequest, v1.RevokeEnterpriseSubscriptionLicenseResponse]
	updateEnterpriseSubscription        *connect.Client[v1.UpdateEnterpriseSubscriptionRequest, v1.UpdateEnterpriseSubscriptionResponse]
	archiveEnterpriseSubscription       *connect.Client[v1.ArchiveEnterpriseSubscriptionRequest, v1.ArchiveEnterpriseSubscriptionResponse]
	createEnterpriseSubscription        *connect.Client[v1.CreateEnterpriseSubscriptionRequest, v1.CreateEnterpriseSubscriptionResponse]
	updateSubscriptionMembership        *connect.Client[v1.UpdateSubscriptionMembershipRequest, v1.UpdateSubscriptionMembershipResponse]
}

// GetEnterpriseSubscription calls
// enterpriseportal.subscriptions.v1.SubscriptionsService.GetEnterpriseSubscription.
func (c *subscriptionsServiceClient) GetEnterpriseSubscription(ctx context.Context, req *connect.Request[v1.GetEnterpriseSubscriptionRequest]) (*connect.Response[v1.GetEnterpriseSubscriptionResponse], error) {
	return c.getEnterpriseSubscription.CallUnary(ctx, req)
}

// ListEnterpriseSubscriptions calls
// enterpriseportal.subscriptions.v1.SubscriptionsService.ListEnterpriseSubscriptions.
func (c *subscriptionsServiceClient) ListEnterpriseSubscriptions(ctx context.Context, req *connect.Request[v1.ListEnterpriseSubscriptionsRequest]) (*connect.Response[v1.ListEnterpriseSubscriptionsResponse], error) {
	return c.listEnterpriseSubscriptions.CallUnary(ctx, req)
}

// ListEnterpriseSubscriptionLicenses calls
// enterpriseportal.subscriptions.v1.SubscriptionsService.ListEnterpriseSubscriptionLicenses.
func (c *subscriptionsServiceClient) ListEnterpriseSubscriptionLicenses(ctx context.Context, req *connect.Request[v1.ListEnterpriseSubscriptionLicensesRequest]) (*connect.Response[v1.ListEnterpriseSubscriptionLicensesResponse], error) {
	return c.listEnterpriseSubscriptionLicenses.CallUnary(ctx, req)
}

// CreateEnterpriseSubscriptionLicense calls
// enterpriseportal.subscriptions.v1.SubscriptionsService.CreateEnterpriseSubscriptionLicense.
func (c *subscriptionsServiceClient) CreateEnterpriseSubscriptionLicense(ctx context.Context, req *connect.Request[v1.CreateEnterpriseSubscriptionLicenseRequest]) (*connect.Response[v1.CreateEnterpriseSubscriptionLicenseResponse], error) {
	return c.createEnterpriseSubscriptionLicense.CallUnary(ctx, req)
}

// RevokeEnterpriseSubscriptionLicense calls
// enterpriseportal.subscriptions.v1.SubscriptionsService.RevokeEnterpriseSubscriptionLicense.
func (c *subscriptionsServiceClient) RevokeEnterpriseSubscriptionLicense(ctx context.Context, req *connect.Request[v1.RevokeEnterpriseSubscriptionLicenseRequest]) (*connect.Response[v1.RevokeEnterpriseSubscriptionLicenseResponse], error) {
	return c.revokeEnterpriseSubscriptionLicense.CallUnary(ctx, req)
}

// UpdateEnterpriseSubscription calls
// enterpriseportal.subscriptions.v1.SubscriptionsService.UpdateEnterpriseSubscription.
func (c *subscriptionsServiceClient) UpdateEnterpriseSubscription(ctx context.Context, req *connect.Request[v1.UpdateEnterpriseSubscriptionRequest]) (*connect.Response[v1.UpdateEnterpriseSubscriptionResponse], error) {
	return c.updateEnterpriseSubscription.CallUnary(ctx, req)
}

// ArchiveEnterpriseSubscription calls
// enterpriseportal.subscriptions.v1.SubscriptionsService.ArchiveEnterpriseSubscription.
func (c *subscriptionsServiceClient) ArchiveEnterpriseSubscription(ctx context.Context, req *connect.Request[v1.ArchiveEnterpriseSubscriptionRequest]) (*connect.Response[v1.ArchiveEnterpriseSubscriptionResponse], error) {
	return c.archiveEnterpriseSubscription.CallUnary(ctx, req)
}

// CreateEnterpriseSubscription calls
// enterpriseportal.subscriptions.v1.SubscriptionsService.CreateEnterpriseSubscription.
func (c *subscriptionsServiceClient) CreateEnterpriseSubscription(ctx context.Context, req *connect.Request[v1.CreateEnterpriseSubscriptionRequest]) (*connect.Response[v1.CreateEnterpriseSubscriptionResponse], error) {
	return c.createEnterpriseSubscription.CallUnary(ctx, req)
}

// UpdateSubscriptionMembership calls
// enterpriseportal.subscriptions.v1.SubscriptionsService.UpdateSubscriptionMembership.
func (c *subscriptionsServiceClient) UpdateSubscriptionMembership(ctx context.Context, req *connect.Request[v1.UpdateSubscriptionMembershipRequest]) (*connect.Response[v1.UpdateSubscriptionMembershipResponse], error) {
	return c.updateSubscriptionMembership.CallUnary(ctx, req)
}

// SubscriptionsServiceHandler is an implementation of the
// enterpriseportal.subscriptions.v1.SubscriptionsService service.
type SubscriptionsServiceHandler interface {
	// GetEnterpriseSubscription retrieves an exact match on an Enterprise subscription.
	GetEnterpriseSubscription(context.Context, *connect.Request[v1.GetEnterpriseSubscriptionRequest]) (*connect.Response[v1.GetEnterpriseSubscriptionResponse], error)
	// ListEnterpriseSubscriptions queries for Enterprise subscriptions.
	ListEnterpriseSubscriptions(context.Context, *connect.Request[v1.ListEnterpriseSubscriptionsRequest]) (*connect.Response[v1.ListEnterpriseSubscriptionsResponse], error)
	// ListEnterpriseSubscriptionLicenses queries for licenses associated with
	// Enterprise subscription licenses, with the ability to list licenses across
	// all subscriptions, or just a specific subscription.
	//
	// Each subscription owns a collection of licenses, typically a series of
	// licenses with the most recent one being a subscription's active license.
	ListEnterpriseSubscriptionLicenses(context.Context, *connect.Request[v1.ListEnterpriseSubscriptionLicensesRequest]) (*connect.Response[v1.ListEnterpriseSubscriptionLicensesResponse], error)
	// CreateEnterpriseSubscription creates license for an Enterprise subscription.
	CreateEnterpriseSubscriptionLicense(context.Context, *connect.Request[v1.CreateEnterpriseSubscriptionLicenseRequest]) (*connect.Response[v1.CreateEnterpriseSubscriptionLicenseResponse], error)
	// RevokeEnterpriseSubscriptionLicense revokes an existing license for an
	// Enterprise subscription, permanently disabling its use for features
	// managed by Sourcegraph. Revocation cannot be undone.
	RevokeEnterpriseSubscriptionLicense(context.Context, *connect.Request[v1.RevokeEnterpriseSubscriptionLicenseRequest]) (*connect.Response[v1.RevokeEnterpriseSubscriptionLicenseResponse], error)
	// UpdateEnterpriseSubscription updates an existing Enterprise subscription.
	// Only properties specified by the update_mask are applied.
	UpdateEnterpriseSubscription(context.Context, *connect.Request[v1.UpdateEnterpriseSubscriptionRequest]) (*connect.Response[v1.UpdateEnterpriseSubscriptionResponse], error)
	// ArchiveEnterpriseSubscriptionRequest archives an existing Enterprise
	// subscription. This is a permanent operation, and cannot be undone.
	//
	// Archiving a subscription also immediately and permanently revokes all
	// associated licenses.
	ArchiveEnterpriseSubscription(context.Context, *connect.Request[v1.ArchiveEnterpriseSubscriptionRequest]) (*connect.Response[v1.ArchiveEnterpriseSubscriptionResponse], error)
	// CreateEnterpriseSubscription creates an Enterprise subscription.
	CreateEnterpriseSubscription(context.Context, *connect.Request[v1.CreateEnterpriseSubscriptionRequest]) (*connect.Response[v1.CreateEnterpriseSubscriptionResponse], error)
	// UpdateSubscriptionMembership updates a subscription membership. It creates
	// a new one if it does not exist and allow_missing is set to true.
	UpdateSubscriptionMembership(context.Context, *connect.Request[v1.UpdateSubscriptionMembershipRequest]) (*connect.Response[v1.UpdateSubscriptionMembershipResponse], error)
}

// NewSubscriptionsServiceHandler builds an HTTP handler from the service implementation. It returns
// the path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewSubscriptionsServiceHandler(svc SubscriptionsServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	subscriptionsServiceGetEnterpriseSubscriptionHandler := connect.NewUnaryHandler(
		SubscriptionsServiceGetEnterpriseSubscriptionProcedure,
		svc.GetEnterpriseSubscription,
		connect.WithSchema(subscriptionsServiceGetEnterpriseSubscriptionMethodDescriptor),
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	subscriptionsServiceListEnterpriseSubscriptionsHandler := connect.NewUnaryHandler(
		SubscriptionsServiceListEnterpriseSubscriptionsProcedure,
		svc.ListEnterpriseSubscriptions,
		connect.WithSchema(subscriptionsServiceListEnterpriseSubscriptionsMethodDescriptor),
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	subscriptionsServiceListEnterpriseSubscriptionLicensesHandler := connect.NewUnaryHandler(
		SubscriptionsServiceListEnterpriseSubscriptionLicensesProcedure,
		svc.ListEnterpriseSubscriptionLicenses,
		connect.WithSchema(subscriptionsServiceListEnterpriseSubscriptionLicensesMethodDescriptor),
		connect.WithIdempotency(connect.IdempotencyNoSideEffects),
		connect.WithHandlerOptions(opts...),
	)
	subscriptionsServiceCreateEnterpriseSubscriptionLicenseHandler := connect.NewUnaryHandler(
		SubscriptionsServiceCreateEnterpriseSubscriptionLicenseProcedure,
		svc.CreateEnterpriseSubscriptionLicense,
		connect.WithSchema(subscriptionsServiceCreateEnterpriseSubscriptionLicenseMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	subscriptionsServiceRevokeEnterpriseSubscriptionLicenseHandler := connect.NewUnaryHandler(
		SubscriptionsServiceRevokeEnterpriseSubscriptionLicenseProcedure,
		svc.RevokeEnterpriseSubscriptionLicense,
		connect.WithSchema(subscriptionsServiceRevokeEnterpriseSubscriptionLicenseMethodDescriptor),
		connect.WithIdempotency(connect.IdempotencyIdempotent),
		connect.WithHandlerOptions(opts...),
	)
	subscriptionsServiceUpdateEnterpriseSubscriptionHandler := connect.NewUnaryHandler(
		SubscriptionsServiceUpdateEnterpriseSubscriptionProcedure,
		svc.UpdateEnterpriseSubscription,
		connect.WithSchema(subscriptionsServiceUpdateEnterpriseSubscriptionMethodDescriptor),
		connect.WithIdempotency(connect.IdempotencyIdempotent),
		connect.WithHandlerOptions(opts...),
	)
	subscriptionsServiceArchiveEnterpriseSubscriptionHandler := connect.NewUnaryHandler(
		SubscriptionsServiceArchiveEnterpriseSubscriptionProcedure,
		svc.ArchiveEnterpriseSubscription,
		connect.WithSchema(subscriptionsServiceArchiveEnterpriseSubscriptionMethodDescriptor),
		connect.WithIdempotency(connect.IdempotencyIdempotent),
		connect.WithHandlerOptions(opts...),
	)
	subscriptionsServiceCreateEnterpriseSubscriptionHandler := connect.NewUnaryHandler(
		SubscriptionsServiceCreateEnterpriseSubscriptionProcedure,
		svc.CreateEnterpriseSubscription,
		connect.WithSchema(subscriptionsServiceCreateEnterpriseSubscriptionMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	subscriptionsServiceUpdateSubscriptionMembershipHandler := connect.NewUnaryHandler(
		SubscriptionsServiceUpdateSubscriptionMembershipProcedure,
		svc.UpdateSubscriptionMembership,
		connect.WithSchema(subscriptionsServiceUpdateSubscriptionMembershipMethodDescriptor),
		connect.WithIdempotency(connect.IdempotencyIdempotent),
		connect.WithHandlerOptions(opts...),
	)
	return "/enterpriseportal.subscriptions.v1.SubscriptionsService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case SubscriptionsServiceGetEnterpriseSubscriptionProcedure:
			subscriptionsServiceGetEnterpriseSubscriptionHandler.ServeHTTP(w, r)
		case SubscriptionsServiceListEnterpriseSubscriptionsProcedure:
			subscriptionsServiceListEnterpriseSubscriptionsHandler.ServeHTTP(w, r)
		case SubscriptionsServiceListEnterpriseSubscriptionLicensesProcedure:
			subscriptionsServiceListEnterpriseSubscriptionLicensesHandler.ServeHTTP(w, r)
		case SubscriptionsServiceCreateEnterpriseSubscriptionLicenseProcedure:
			subscriptionsServiceCreateEnterpriseSubscriptionLicenseHandler.ServeHTTP(w, r)
		case SubscriptionsServiceRevokeEnterpriseSubscriptionLicenseProcedure:
			subscriptionsServiceRevokeEnterpriseSubscriptionLicenseHandler.ServeHTTP(w, r)
		case SubscriptionsServiceUpdateEnterpriseSubscriptionProcedure:
			subscriptionsServiceUpdateEnterpriseSubscriptionHandler.ServeHTTP(w, r)
		case SubscriptionsServiceArchiveEnterpriseSubscriptionProcedure:
			subscriptionsServiceArchiveEnterpriseSubscriptionHandler.ServeHTTP(w, r)
		case SubscriptionsServiceCreateEnterpriseSubscriptionProcedure:
			subscriptionsServiceCreateEnterpriseSubscriptionHandler.ServeHTTP(w, r)
		case SubscriptionsServiceUpdateSubscriptionMembershipProcedure:
			subscriptionsServiceUpdateSubscriptionMembershipHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedSubscriptionsServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedSubscriptionsServiceHandler struct{}

func (UnimplementedSubscriptionsServiceHandler) GetEnterpriseSubscription(context.Context, *connect.Request[v1.GetEnterpriseSubscriptionRequest]) (*connect.Response[v1.GetEnterpriseSubscriptionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("enterpriseportal.subscriptions.v1.SubscriptionsService.GetEnterpriseSubscription is not implemented"))
}

func (UnimplementedSubscriptionsServiceHandler) ListEnterpriseSubscriptions(context.Context, *connect.Request[v1.ListEnterpriseSubscriptionsRequest]) (*connect.Response[v1.ListEnterpriseSubscriptionsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("enterpriseportal.subscriptions.v1.SubscriptionsService.ListEnterpriseSubscriptions is not implemented"))
}

func (UnimplementedSubscriptionsServiceHandler) ListEnterpriseSubscriptionLicenses(context.Context, *connect.Request[v1.ListEnterpriseSubscriptionLicensesRequest]) (*connect.Response[v1.ListEnterpriseSubscriptionLicensesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("enterpriseportal.subscriptions.v1.SubscriptionsService.ListEnterpriseSubscriptionLicenses is not implemented"))
}

func (UnimplementedSubscriptionsServiceHandler) CreateEnterpriseSubscriptionLicense(context.Context, *connect.Request[v1.CreateEnterpriseSubscriptionLicenseRequest]) (*connect.Response[v1.CreateEnterpriseSubscriptionLicenseResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("enterpriseportal.subscriptions.v1.SubscriptionsService.CreateEnterpriseSubscriptionLicense is not implemented"))
}

func (UnimplementedSubscriptionsServiceHandler) RevokeEnterpriseSubscriptionLicense(context.Context, *connect.Request[v1.RevokeEnterpriseSubscriptionLicenseRequest]) (*connect.Response[v1.RevokeEnterpriseSubscriptionLicenseResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("enterpriseportal.subscriptions.v1.SubscriptionsService.RevokeEnterpriseSubscriptionLicense is not implemented"))
}

func (UnimplementedSubscriptionsServiceHandler) UpdateEnterpriseSubscription(context.Context, *connect.Request[v1.UpdateEnterpriseSubscriptionRequest]) (*connect.Response[v1.UpdateEnterpriseSubscriptionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("enterpriseportal.subscriptions.v1.SubscriptionsService.UpdateEnterpriseSubscription is not implemented"))
}

func (UnimplementedSubscriptionsServiceHandler) ArchiveEnterpriseSubscription(context.Context, *connect.Request[v1.ArchiveEnterpriseSubscriptionRequest]) (*connect.Response[v1.ArchiveEnterpriseSubscriptionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("enterpriseportal.subscriptions.v1.SubscriptionsService.ArchiveEnterpriseSubscription is not implemented"))
}

func (UnimplementedSubscriptionsServiceHandler) CreateEnterpriseSubscription(context.Context, *connect.Request[v1.CreateEnterpriseSubscriptionRequest]) (*connect.Response[v1.CreateEnterpriseSubscriptionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("enterpriseportal.subscriptions.v1.SubscriptionsService.CreateEnterpriseSubscription is not implemented"))
}

func (UnimplementedSubscriptionsServiceHandler) UpdateSubscriptionMembership(context.Context, *connect.Request[v1.UpdateSubscriptionMembershipRequest]) (*connect.Response[v1.UpdateSubscriptionMembershipResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("enterpriseportal.subscriptions.v1.SubscriptionsService.UpdateSubscriptionMembership is not implemented"))
}
