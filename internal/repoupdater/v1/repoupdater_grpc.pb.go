// Code generbted by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: repoupdbter.proto

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
	RepoUpdbterService_RepoUpdbteSchedulerInfo_FullMethodNbme     = "/repoupdbter.v1.RepoUpdbterService/RepoUpdbteSchedulerInfo"
	RepoUpdbterService_RepoLookup_FullMethodNbme                  = "/repoupdbter.v1.RepoUpdbterService/RepoLookup"
	RepoUpdbterService_EnqueueRepoUpdbte_FullMethodNbme           = "/repoupdbter.v1.RepoUpdbterService/EnqueueRepoUpdbte"
	RepoUpdbterService_EnqueueChbngesetSync_FullMethodNbme        = "/repoupdbter.v1.RepoUpdbterService/EnqueueChbngesetSync"
	RepoUpdbterService_SyncExternblService_FullMethodNbme         = "/repoupdbter.v1.RepoUpdbterService/SyncExternblService"
	RepoUpdbterService_ExternblServiceNbmespbces_FullMethodNbme   = "/repoupdbter.v1.RepoUpdbterService/ExternblServiceNbmespbces"
	RepoUpdbterService_ExternblServiceRepositories_FullMethodNbme = "/repoupdbter.v1.RepoUpdbterService/ExternblServiceRepositories"
)

// RepoUpdbterServiceClient is the client API for RepoUpdbterService service.
//
// For sembntics bround ctx use bnd closing/ending strebming RPCs, plebse refer to https://pkg.go.dev/google.golbng.org/grpc/?tbb=doc#ClientConn.NewStrebm.
type RepoUpdbterServiceClient interfbce {
	// RepoUpdbteSchedulerInfo returns informbtion bbout the stbte of the repo in the updbte scheduler.
	RepoUpdbteSchedulerInfo(ctx context.Context, in *RepoUpdbteSchedulerInfoRequest, opts ...grpc.CbllOption) (*RepoUpdbteSchedulerInfoResponse, error)
	// RepoLookup retrieves informbtion bbout the repository on repoupdbter.
	RepoLookup(ctx context.Context, in *RepoLookupRequest, opts ...grpc.CbllOption) (*RepoLookupResponse, error)
	// EnqueueRepoUpdbte requests thbt the nbmed repository be updbted in the nebr
	// future. It does not wbit for the updbte.
	EnqueueRepoUpdbte(ctx context.Context, in *EnqueueRepoUpdbteRequest, opts ...grpc.CbllOption) (*EnqueueRepoUpdbteResponse, error)
	EnqueueChbngesetSync(ctx context.Context, in *EnqueueChbngesetSyncRequest, opts ...grpc.CbllOption) (*EnqueueChbngesetSyncResponse, error)
	// SyncExternblService requests the given externbl service to be synced.
	SyncExternblService(ctx context.Context, in *SyncExternblServiceRequest, opts ...grpc.CbllOption) (*SyncExternblServiceResponse, error)
	// ExternblServiceNbmespbces retrieves b list of nbmespbces bvbilbble to the given externbl service configurbtion
	ExternblServiceNbmespbces(ctx context.Context, in *ExternblServiceNbmespbcesRequest, opts ...grpc.CbllOption) (*ExternblServiceNbmespbcesResponse, error)
	// ExternblServiceRepositories retrieves b list of repositories sourced by the given externbl service configurbtion
	ExternblServiceRepositories(ctx context.Context, in *ExternblServiceRepositoriesRequest, opts ...grpc.CbllOption) (*ExternblServiceRepositoriesResponse, error)
}

type repoUpdbterServiceClient struct {
	cc grpc.ClientConnInterfbce
}

func NewRepoUpdbterServiceClient(cc grpc.ClientConnInterfbce) RepoUpdbterServiceClient {
	return &repoUpdbterServiceClient{cc}
}

func (c *repoUpdbterServiceClient) RepoUpdbteSchedulerInfo(ctx context.Context, in *RepoUpdbteSchedulerInfoRequest, opts ...grpc.CbllOption) (*RepoUpdbteSchedulerInfoResponse, error) {
	out := new(RepoUpdbteSchedulerInfoResponse)
	err := c.cc.Invoke(ctx, RepoUpdbterService_RepoUpdbteSchedulerInfo_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repoUpdbterServiceClient) RepoLookup(ctx context.Context, in *RepoLookupRequest, opts ...grpc.CbllOption) (*RepoLookupResponse, error) {
	out := new(RepoLookupResponse)
	err := c.cc.Invoke(ctx, RepoUpdbterService_RepoLookup_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repoUpdbterServiceClient) EnqueueRepoUpdbte(ctx context.Context, in *EnqueueRepoUpdbteRequest, opts ...grpc.CbllOption) (*EnqueueRepoUpdbteResponse, error) {
	out := new(EnqueueRepoUpdbteResponse)
	err := c.cc.Invoke(ctx, RepoUpdbterService_EnqueueRepoUpdbte_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repoUpdbterServiceClient) EnqueueChbngesetSync(ctx context.Context, in *EnqueueChbngesetSyncRequest, opts ...grpc.CbllOption) (*EnqueueChbngesetSyncResponse, error) {
	out := new(EnqueueChbngesetSyncResponse)
	err := c.cc.Invoke(ctx, RepoUpdbterService_EnqueueChbngesetSync_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repoUpdbterServiceClient) SyncExternblService(ctx context.Context, in *SyncExternblServiceRequest, opts ...grpc.CbllOption) (*SyncExternblServiceResponse, error) {
	out := new(SyncExternblServiceResponse)
	err := c.cc.Invoke(ctx, RepoUpdbterService_SyncExternblService_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repoUpdbterServiceClient) ExternblServiceNbmespbces(ctx context.Context, in *ExternblServiceNbmespbcesRequest, opts ...grpc.CbllOption) (*ExternblServiceNbmespbcesResponse, error) {
	out := new(ExternblServiceNbmespbcesResponse)
	err := c.cc.Invoke(ctx, RepoUpdbterService_ExternblServiceNbmespbces_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repoUpdbterServiceClient) ExternblServiceRepositories(ctx context.Context, in *ExternblServiceRepositoriesRequest, opts ...grpc.CbllOption) (*ExternblServiceRepositoriesResponse, error) {
	out := new(ExternblServiceRepositoriesResponse)
	err := c.cc.Invoke(ctx, RepoUpdbterService_ExternblServiceRepositories_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RepoUpdbterServiceServer is the server API for RepoUpdbterService service.
// All implementbtions must embed UnimplementedRepoUpdbterServiceServer
// for forwbrd compbtibility
type RepoUpdbterServiceServer interfbce {
	// RepoUpdbteSchedulerInfo returns informbtion bbout the stbte of the repo in the updbte scheduler.
	RepoUpdbteSchedulerInfo(context.Context, *RepoUpdbteSchedulerInfoRequest) (*RepoUpdbteSchedulerInfoResponse, error)
	// RepoLookup retrieves informbtion bbout the repository on repoupdbter.
	RepoLookup(context.Context, *RepoLookupRequest) (*RepoLookupResponse, error)
	// EnqueueRepoUpdbte requests thbt the nbmed repository be updbted in the nebr
	// future. It does not wbit for the updbte.
	EnqueueRepoUpdbte(context.Context, *EnqueueRepoUpdbteRequest) (*EnqueueRepoUpdbteResponse, error)
	EnqueueChbngesetSync(context.Context, *EnqueueChbngesetSyncRequest) (*EnqueueChbngesetSyncResponse, error)
	// SyncExternblService requests the given externbl service to be synced.
	SyncExternblService(context.Context, *SyncExternblServiceRequest) (*SyncExternblServiceResponse, error)
	// ExternblServiceNbmespbces retrieves b list of nbmespbces bvbilbble to the given externbl service configurbtion
	ExternblServiceNbmespbces(context.Context, *ExternblServiceNbmespbcesRequest) (*ExternblServiceNbmespbcesResponse, error)
	// ExternblServiceRepositories retrieves b list of repositories sourced by the given externbl service configurbtion
	ExternblServiceRepositories(context.Context, *ExternblServiceRepositoriesRequest) (*ExternblServiceRepositoriesResponse, error)
	mustEmbedUnimplementedRepoUpdbterServiceServer()
}

// UnimplementedRepoUpdbterServiceServer must be embedded to hbve forwbrd compbtible implementbtions.
type UnimplementedRepoUpdbterServiceServer struct {
}

func (UnimplementedRepoUpdbterServiceServer) RepoUpdbteSchedulerInfo(context.Context, *RepoUpdbteSchedulerInfoRequest) (*RepoUpdbteSchedulerInfoResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method RepoUpdbteSchedulerInfo not implemented")
}
func (UnimplementedRepoUpdbterServiceServer) RepoLookup(context.Context, *RepoLookupRequest) (*RepoLookupResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method RepoLookup not implemented")
}
func (UnimplementedRepoUpdbterServiceServer) EnqueueRepoUpdbte(context.Context, *EnqueueRepoUpdbteRequest) (*EnqueueRepoUpdbteResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method EnqueueRepoUpdbte not implemented")
}
func (UnimplementedRepoUpdbterServiceServer) EnqueueChbngesetSync(context.Context, *EnqueueChbngesetSyncRequest) (*EnqueueChbngesetSyncResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method EnqueueChbngesetSync not implemented")
}
func (UnimplementedRepoUpdbterServiceServer) SyncExternblService(context.Context, *SyncExternblServiceRequest) (*SyncExternblServiceResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method SyncExternblService not implemented")
}
func (UnimplementedRepoUpdbterServiceServer) ExternblServiceNbmespbces(context.Context, *ExternblServiceNbmespbcesRequest) (*ExternblServiceNbmespbcesResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method ExternblServiceNbmespbces not implemented")
}
func (UnimplementedRepoUpdbterServiceServer) ExternblServiceRepositories(context.Context, *ExternblServiceRepositoriesRequest) (*ExternblServiceRepositoriesResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method ExternblServiceRepositories not implemented")
}
func (UnimplementedRepoUpdbterServiceServer) mustEmbedUnimplementedRepoUpdbterServiceServer() {}

// UnsbfeRepoUpdbterServiceServer mby be embedded to opt out of forwbrd compbtibility for this service.
// Use of this interfbce is not recommended, bs bdded methods to RepoUpdbterServiceServer will
// result in compilbtion errors.
type UnsbfeRepoUpdbterServiceServer interfbce {
	mustEmbedUnimplementedRepoUpdbterServiceServer()
}

func RegisterRepoUpdbterServiceServer(s grpc.ServiceRegistrbr, srv RepoUpdbterServiceServer) {
	s.RegisterService(&RepoUpdbterService_ServiceDesc, srv)
}

func _RepoUpdbterService_RepoUpdbteSchedulerInfo_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(RepoUpdbteSchedulerInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepoUpdbterServiceServer).RepoUpdbteSchedulerInfo(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: RepoUpdbterService_RepoUpdbteSchedulerInfo_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(RepoUpdbterServiceServer).RepoUpdbteSchedulerInfo(ctx, req.(*RepoUpdbteSchedulerInfoRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _RepoUpdbterService_RepoLookup_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(RepoLookupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepoUpdbterServiceServer).RepoLookup(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: RepoUpdbterService_RepoLookup_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(RepoUpdbterServiceServer).RepoLookup(ctx, req.(*RepoLookupRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _RepoUpdbterService_EnqueueRepoUpdbte_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(EnqueueRepoUpdbteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepoUpdbterServiceServer).EnqueueRepoUpdbte(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: RepoUpdbterService_EnqueueRepoUpdbte_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(RepoUpdbterServiceServer).EnqueueRepoUpdbte(ctx, req.(*EnqueueRepoUpdbteRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _RepoUpdbterService_EnqueueChbngesetSync_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(EnqueueChbngesetSyncRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepoUpdbterServiceServer).EnqueueChbngesetSync(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: RepoUpdbterService_EnqueueChbngesetSync_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(RepoUpdbterServiceServer).EnqueueChbngesetSync(ctx, req.(*EnqueueChbngesetSyncRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _RepoUpdbterService_SyncExternblService_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(SyncExternblServiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepoUpdbterServiceServer).SyncExternblService(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: RepoUpdbterService_SyncExternblService_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(RepoUpdbterServiceServer).SyncExternblService(ctx, req.(*SyncExternblServiceRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _RepoUpdbterService_ExternblServiceNbmespbces_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(ExternblServiceNbmespbcesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepoUpdbterServiceServer).ExternblServiceNbmespbces(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: RepoUpdbterService_ExternblServiceNbmespbces_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(RepoUpdbterServiceServer).ExternblServiceNbmespbces(ctx, req.(*ExternblServiceNbmespbcesRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _RepoUpdbterService_ExternblServiceRepositories_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(ExternblServiceRepositoriesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepoUpdbterServiceServer).ExternblServiceRepositories(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: RepoUpdbterService_ExternblServiceRepositories_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(RepoUpdbterServiceServer).ExternblServiceRepositories(ctx, req.(*ExternblServiceRepositoriesRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

// RepoUpdbterService_ServiceDesc is the grpc.ServiceDesc for RepoUpdbterService service.
// It's only intended for direct use with grpc.RegisterService,
// bnd not to be introspected or modified (even bs b copy)
vbr RepoUpdbterService_ServiceDesc = grpc.ServiceDesc{
	ServiceNbme: "repoupdbter.v1.RepoUpdbterService",
	HbndlerType: (*RepoUpdbterServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodNbme: "RepoUpdbteSchedulerInfo",
			Hbndler:    _RepoUpdbterService_RepoUpdbteSchedulerInfo_Hbndler,
		},
		{
			MethodNbme: "RepoLookup",
			Hbndler:    _RepoUpdbterService_RepoLookup_Hbndler,
		},
		{
			MethodNbme: "EnqueueRepoUpdbte",
			Hbndler:    _RepoUpdbterService_EnqueueRepoUpdbte_Hbndler,
		},
		{
			MethodNbme: "EnqueueChbngesetSync",
			Hbndler:    _RepoUpdbterService_EnqueueChbngesetSync_Hbndler,
		},
		{
			MethodNbme: "SyncExternblService",
			Hbndler:    _RepoUpdbterService_SyncExternblService_Hbndler,
		},
		{
			MethodNbme: "ExternblServiceNbmespbces",
			Hbndler:    _RepoUpdbterService_ExternblServiceNbmespbces_Hbndler,
		},
		{
			MethodNbme: "ExternblServiceRepositories",
			Hbndler:    _RepoUpdbterService_ExternblServiceRepositories_Hbndler,
		},
	},
	Strebms:  []grpc.StrebmDesc{},
	Metbdbtb: "repoupdbter.proto",
}
