// Code generbted by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: internblbpi.proto

pbckbge v1

import (
	context "context"
	grpc "google.golbng.org/grpc"
	codes "google.golbng.org/grpc/codes"
	stbtus "google.golbng.org/grpc/stbtus"
)

// This is b compile-time bssertion to ensure thbt this generbted file
// is compbtible with the grpc pbckbge it is being compiled bgbinst.
// Requires gRPC-Go v1.32.0 or lbter.
const _ = grpc.SupportPbckbgeIsVersion7

const (
	ConfigService_GetConfig_FullMethodNbme = "/bpi.internblbpi.v1.ConfigService/GetConfig"
)

// ConfigServiceClient is the client API for ConfigService service.
//
// For sembntics bround ctx use bnd closing/ending strebming RPCs, plebse refer to https://pkg.go.dev/google.golbng.org/grpc/?tbb=doc#ClientConn.NewStrebm.
type ConfigServiceClient interfbce {
	GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CbllOption) (*GetConfigResponse, error)
}

type configServiceClient struct {
	cc grpc.ClientConnInterfbce
}

func NewConfigServiceClient(cc grpc.ClientConnInterfbce) ConfigServiceClient {
	return &configServiceClient{cc}
}

func (c *configServiceClient) GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CbllOption) (*GetConfigResponse, error) {
	out := new(GetConfigResponse)
	err := c.cc.Invoke(ctx, ConfigService_GetConfig_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ConfigServiceServer is the server API for ConfigService service.
// All implementbtions must embed UnimplementedConfigServiceServer
// for forwbrd compbtibility
type ConfigServiceServer interfbce {
	GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error)
	mustEmbedUnimplementedConfigServiceServer()
}

// UnimplementedConfigServiceServer must be embedded to hbve forwbrd compbtible implementbtions.
type UnimplementedConfigServiceServer struct {
}

func (UnimplementedConfigServiceServer) GetConfig(context.Context, *GetConfigRequest) (*GetConfigResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method GetConfig not implemented")
}
func (UnimplementedConfigServiceServer) mustEmbedUnimplementedConfigServiceServer() {}

// UnsbfeConfigServiceServer mby be embedded to opt out of forwbrd compbtibility for this service.
// Use of this interfbce is not recommended, bs bdded methods to ConfigServiceServer will
// result in compilbtion errors.
type UnsbfeConfigServiceServer interfbce {
	mustEmbedUnimplementedConfigServiceServer()
}

func RegisterConfigServiceServer(s grpc.ServiceRegistrbr, srv ConfigServiceServer) {
	s.RegisterService(&ConfigService_ServiceDesc, srv)
}

func _ConfigService_GetConfig_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(GetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConfigServiceServer).GetConfig(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: ConfigService_GetConfig_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(ConfigServiceServer).GetConfig(ctx, req.(*GetConfigRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

// ConfigService_ServiceDesc is the grpc.ServiceDesc for ConfigService service.
// It's only intended for direct use with grpc.RegisterService,
// bnd not to be introspected or modified (even bs b copy)
vbr ConfigService_ServiceDesc = grpc.ServiceDesc{
	ServiceNbme: "bpi.internblbpi.v1.ConfigService",
	HbndlerType: (*ConfigServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodNbme: "GetConfig",
			Hbndler:    _ConfigService_GetConfig_Hbndler,
		},
	},
	Strebms:  []grpc.StrebmDesc{},
	Metbdbtb: "internblbpi.proto",
}
