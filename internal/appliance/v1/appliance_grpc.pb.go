// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: appliance.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	ApplianceService_GetApplianceVersion_FullMethodName = "/appliance.v1.ApplianceService/GetApplianceVersion"
	ApplianceService_GetApplianceStage_FullMethodName   = "/appliance.v1.ApplianceService/GetApplianceStage"
)

// ApplianceServiceClient is the client API for ApplianceService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ApplianceServiceClient interface {
	GetApplianceVersion(ctx context.Context, in *GetApplianceVersionRequest, opts ...grpc.CallOption) (*GetApplianceVersionResponse, error)
	GetApplianceStage(ctx context.Context, in *GetApplianceStageRequest, opts ...grpc.CallOption) (*GetApplianceStageResponse, error)
}

type applianceServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewApplianceServiceClient(cc grpc.ClientConnInterface) ApplianceServiceClient {
	return &applianceServiceClient{cc}
}

func (c *applianceServiceClient) GetApplianceVersion(ctx context.Context, in *GetApplianceVersionRequest, opts ...grpc.CallOption) (*GetApplianceVersionResponse, error) {
	out := new(GetApplianceVersionResponse)
	err := c.cc.Invoke(ctx, ApplianceService_GetApplianceVersion_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applianceServiceClient) GetApplianceStage(ctx context.Context, in *GetApplianceStageRequest, opts ...grpc.CallOption) (*GetApplianceStageResponse, error) {
	out := new(GetApplianceStageResponse)
	err := c.cc.Invoke(ctx, ApplianceService_GetApplianceStage_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ApplianceServiceServer is the server API for ApplianceService service.
// All implementations must embed UnimplementedApplianceServiceServer
// for forward compatibility
type ApplianceServiceServer interface {
	GetApplianceVersion(context.Context, *GetApplianceVersionRequest) (*GetApplianceVersionResponse, error)
	GetApplianceStage(context.Context, *GetApplianceStageRequest) (*GetApplianceStageResponse, error)
	mustEmbedUnimplementedApplianceServiceServer()
}

// UnimplementedApplianceServiceServer must be embedded to have forward compatible implementations.
type UnimplementedApplianceServiceServer struct {
}

func (UnimplementedApplianceServiceServer) GetApplianceVersion(context.Context, *GetApplianceVersionRequest) (*GetApplianceVersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetApplianceVersion not implemented")
}
func (UnimplementedApplianceServiceServer) GetApplianceStage(context.Context, *GetApplianceStageRequest) (*GetApplianceStageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetApplianceStage not implemented")
}
func (UnimplementedApplianceServiceServer) mustEmbedUnimplementedApplianceServiceServer() {}

// UnsafeApplianceServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ApplianceServiceServer will
// result in compilation errors.
type UnsafeApplianceServiceServer interface {
	mustEmbedUnimplementedApplianceServiceServer()
}

func RegisterApplianceServiceServer(s grpc.ServiceRegistrar, srv ApplianceServiceServer) {
	s.RegisterService(&ApplianceService_ServiceDesc, srv)
}

func _ApplianceService_GetApplianceVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetApplianceVersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplianceServiceServer).GetApplianceVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ApplianceService_GetApplianceVersion_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplianceServiceServer).GetApplianceVersion(ctx, req.(*GetApplianceVersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ApplianceService_GetApplianceStage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetApplianceStageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplianceServiceServer).GetApplianceStage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ApplianceService_GetApplianceStage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplianceServiceServer).GetApplianceStage(ctx, req.(*GetApplianceStageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ApplianceService_ServiceDesc is the grpc.ServiceDesc for ApplianceService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ApplianceService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "appliance.v1.ApplianceService",
	HandlerType: (*ApplianceServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetApplianceVersion",
			Handler:    _ApplianceService_GetApplianceVersion_Handler,
		},
		{
			MethodName: "GetApplianceStage",
			Handler:    _ApplianceService_GetApplianceStage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "appliance.proto",
}
