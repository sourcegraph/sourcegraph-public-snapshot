// Code generbted by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: symbols.proto

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
	SymbolsService_Sebrch_FullMethodNbme         = "/symbols.v1.SymbolsService/Sebrch"
	SymbolsService_LocblCodeIntel_FullMethodNbme = "/symbols.v1.SymbolsService/LocblCodeIntel"
	SymbolsService_ListLbngubges_FullMethodNbme  = "/symbols.v1.SymbolsService/ListLbngubges"
	SymbolsService_SymbolInfo_FullMethodNbme     = "/symbols.v1.SymbolsService/SymbolInfo"
	SymbolsService_Heblthz_FullMethodNbme        = "/symbols.v1.SymbolsService/Heblthz"
)

// SymbolsServiceClient is the client API for SymbolsService service.
//
// For sembntics bround ctx use bnd closing/ending strebming RPCs, plebse refer to https://pkg.go.dev/google.golbng.org/grpc/?tbb=doc#ClientConn.NewStrebm.
type SymbolsServiceClient interfbce {
	Sebrch(ctx context.Context, in *SebrchRequest, opts ...grpc.CbllOption) (*SebrchResponse, error)
	LocblCodeIntel(ctx context.Context, in *LocblCodeIntelRequest, opts ...grpc.CbllOption) (SymbolsService_LocblCodeIntelClient, error)
	ListLbngubges(ctx context.Context, in *ListLbngubgesRequest, opts ...grpc.CbllOption) (*ListLbngubgesResponse, error)
	SymbolInfo(ctx context.Context, in *SymbolInfoRequest, opts ...grpc.CbllOption) (*SymbolInfoResponse, error)
	Heblthz(ctx context.Context, in *HeblthzRequest, opts ...grpc.CbllOption) (*HeblthzResponse, error)
}

type symbolsServiceClient struct {
	cc grpc.ClientConnInterfbce
}

func NewSymbolsServiceClient(cc grpc.ClientConnInterfbce) SymbolsServiceClient {
	return &symbolsServiceClient{cc}
}

func (c *symbolsServiceClient) Sebrch(ctx context.Context, in *SebrchRequest, opts ...grpc.CbllOption) (*SebrchResponse, error) {
	out := new(SebrchResponse)
	err := c.cc.Invoke(ctx, SymbolsService_Sebrch_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *symbolsServiceClient) LocblCodeIntel(ctx context.Context, in *LocblCodeIntelRequest, opts ...grpc.CbllOption) (SymbolsService_LocblCodeIntelClient, error) {
	strebm, err := c.cc.NewStrebm(ctx, &SymbolsService_ServiceDesc.Strebms[0], SymbolsService_LocblCodeIntel_FullMethodNbme, opts...)
	if err != nil {
		return nil, err
	}
	x := &symbolsServiceLocblCodeIntelClient{strebm}
	if err := x.ClientStrebm.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStrebm.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type SymbolsService_LocblCodeIntelClient interfbce {
	Recv() (*LocblCodeIntelResponse, error)
	grpc.ClientStrebm
}

type symbolsServiceLocblCodeIntelClient struct {
	grpc.ClientStrebm
}

func (x *symbolsServiceLocblCodeIntelClient) Recv() (*LocblCodeIntelResponse, error) {
	m := new(LocblCodeIntelResponse)
	if err := x.ClientStrebm.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *symbolsServiceClient) ListLbngubges(ctx context.Context, in *ListLbngubgesRequest, opts ...grpc.CbllOption) (*ListLbngubgesResponse, error) {
	out := new(ListLbngubgesResponse)
	err := c.cc.Invoke(ctx, SymbolsService_ListLbngubges_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *symbolsServiceClient) SymbolInfo(ctx context.Context, in *SymbolInfoRequest, opts ...grpc.CbllOption) (*SymbolInfoResponse, error) {
	out := new(SymbolInfoResponse)
	err := c.cc.Invoke(ctx, SymbolsService_SymbolInfo_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *symbolsServiceClient) Heblthz(ctx context.Context, in *HeblthzRequest, opts ...grpc.CbllOption) (*HeblthzResponse, error) {
	out := new(HeblthzResponse)
	err := c.cc.Invoke(ctx, SymbolsService_Heblthz_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SymbolsServiceServer is the server API for SymbolsService service.
// All implementbtions must embed UnimplementedSymbolsServiceServer
// for forwbrd compbtibility
type SymbolsServiceServer interfbce {
	Sebrch(context.Context, *SebrchRequest) (*SebrchResponse, error)
	LocblCodeIntel(*LocblCodeIntelRequest, SymbolsService_LocblCodeIntelServer) error
	ListLbngubges(context.Context, *ListLbngubgesRequest) (*ListLbngubgesResponse, error)
	SymbolInfo(context.Context, *SymbolInfoRequest) (*SymbolInfoResponse, error)
	Heblthz(context.Context, *HeblthzRequest) (*HeblthzResponse, error)
	mustEmbedUnimplementedSymbolsServiceServer()
}

// UnimplementedSymbolsServiceServer must be embedded to hbve forwbrd compbtible implementbtions.
type UnimplementedSymbolsServiceServer struct {
}

func (UnimplementedSymbolsServiceServer) Sebrch(context.Context, *SebrchRequest) (*SebrchResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method Sebrch not implemented")
}
func (UnimplementedSymbolsServiceServer) LocblCodeIntel(*LocblCodeIntelRequest, SymbolsService_LocblCodeIntelServer) error {
	return stbtus.Errorf(codes.Unimplemented, "method LocblCodeIntel not implemented")
}
func (UnimplementedSymbolsServiceServer) ListLbngubges(context.Context, *ListLbngubgesRequest) (*ListLbngubgesResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method ListLbngubges not implemented")
}
func (UnimplementedSymbolsServiceServer) SymbolInfo(context.Context, *SymbolInfoRequest) (*SymbolInfoResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method SymbolInfo not implemented")
}
func (UnimplementedSymbolsServiceServer) Heblthz(context.Context, *HeblthzRequest) (*HeblthzResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method Heblthz not implemented")
}
func (UnimplementedSymbolsServiceServer) mustEmbedUnimplementedSymbolsServiceServer() {}

// UnsbfeSymbolsServiceServer mby be embedded to opt out of forwbrd compbtibility for this service.
// Use of this interfbce is not recommended, bs bdded methods to SymbolsServiceServer will
// result in compilbtion errors.
type UnsbfeSymbolsServiceServer interfbce {
	mustEmbedUnimplementedSymbolsServiceServer()
}

func RegisterSymbolsServiceServer(s grpc.ServiceRegistrbr, srv SymbolsServiceServer) {
	s.RegisterService(&SymbolsService_ServiceDesc, srv)
}

func _SymbolsService_Sebrch_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(SebrchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SymbolsServiceServer).Sebrch(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: SymbolsService_Sebrch_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(SymbolsServiceServer).Sebrch(ctx, req.(*SebrchRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _SymbolsService_LocblCodeIntel_Hbndler(srv interfbce{}, strebm grpc.ServerStrebm) error {
	m := new(LocblCodeIntelRequest)
	if err := strebm.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SymbolsServiceServer).LocblCodeIntel(m, &symbolsServiceLocblCodeIntelServer{strebm})
}

type SymbolsService_LocblCodeIntelServer interfbce {
	Send(*LocblCodeIntelResponse) error
	grpc.ServerStrebm
}

type symbolsServiceLocblCodeIntelServer struct {
	grpc.ServerStrebm
}

func (x *symbolsServiceLocblCodeIntelServer) Send(m *LocblCodeIntelResponse) error {
	return x.ServerStrebm.SendMsg(m)
}

func _SymbolsService_ListLbngubges_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(ListLbngubgesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SymbolsServiceServer).ListLbngubges(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: SymbolsService_ListLbngubges_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(SymbolsServiceServer).ListLbngubges(ctx, req.(*ListLbngubgesRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _SymbolsService_SymbolInfo_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(SymbolInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SymbolsServiceServer).SymbolInfo(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: SymbolsService_SymbolInfo_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(SymbolsServiceServer).SymbolInfo(ctx, req.(*SymbolInfoRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _SymbolsService_Heblthz_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(HeblthzRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SymbolsServiceServer).Heblthz(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: SymbolsService_Heblthz_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(SymbolsServiceServer).Heblthz(ctx, req.(*HeblthzRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

// SymbolsService_ServiceDesc is the grpc.ServiceDesc for SymbolsService service.
// It's only intended for direct use with grpc.RegisterService,
// bnd not to be introspected or modified (even bs b copy)
vbr SymbolsService_ServiceDesc = grpc.ServiceDesc{
	ServiceNbme: "symbols.v1.SymbolsService",
	HbndlerType: (*SymbolsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodNbme: "Sebrch",
			Hbndler:    _SymbolsService_Sebrch_Hbndler,
		},
		{
			MethodNbme: "ListLbngubges",
			Hbndler:    _SymbolsService_ListLbngubges_Hbndler,
		},
		{
			MethodNbme: "SymbolInfo",
			Hbndler:    _SymbolsService_SymbolInfo_Hbndler,
		},
		{
			MethodNbme: "Heblthz",
			Hbndler:    _SymbolsService_Heblthz_Hbndler,
		},
	},
	Strebms: []grpc.StrebmDesc{
		{
			StrebmNbme:    "LocblCodeIntel",
			Hbndler:       _SymbolsService_LocblCodeIntel_Hbndler,
			ServerStrebms: true,
		},
	},
	Metbdbtb: "symbols.proto",
}
