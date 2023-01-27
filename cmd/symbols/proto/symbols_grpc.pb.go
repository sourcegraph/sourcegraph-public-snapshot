// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: symbols.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// SymbolsClient is the client API for Symbols service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SymbolsClient interface {
	Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SymbolsResponse, error)
	LocalCodeIntel(ctx context.Context, in *LocalCodeIntelRequest, opts ...grpc.CallOption) (*LocalCodeIntelResponse, error)
	ListLanguages(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ListLanguagesResponse, error)
	SymbolInfo(ctx context.Context, in *SymbolInfoRequest, opts ...grpc.CallOption) (*SymbolInfoResponse, error)
	Healthz(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type symbolsClient struct {
	cc grpc.ClientConnInterface
}

func NewSymbolsClient(cc grpc.ClientConnInterface) SymbolsClient {
	return &symbolsClient{cc}
}

func (c *symbolsClient) Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SymbolsResponse, error) {
	out := new(SymbolsResponse)
	err := c.cc.Invoke(ctx, "/symbols.v1.Symbols/Search", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *symbolsClient) LocalCodeIntel(ctx context.Context, in *LocalCodeIntelRequest, opts ...grpc.CallOption) (*LocalCodeIntelResponse, error) {
	out := new(LocalCodeIntelResponse)
	err := c.cc.Invoke(ctx, "/symbols.v1.Symbols/LocalCodeIntel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *symbolsClient) ListLanguages(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ListLanguagesResponse, error) {
	out := new(ListLanguagesResponse)
	err := c.cc.Invoke(ctx, "/symbols.v1.Symbols/ListLanguages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *symbolsClient) SymbolInfo(ctx context.Context, in *SymbolInfoRequest, opts ...grpc.CallOption) (*SymbolInfoResponse, error) {
	out := new(SymbolInfoResponse)
	err := c.cc.Invoke(ctx, "/symbols.v1.Symbols/SymbolInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *symbolsClient) Healthz(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/symbols.v1.Symbols/Healthz", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SymbolsServer is the server API for Symbols service.
// All implementations must embed UnimplementedSymbolsServer
// for forward compatibility
type SymbolsServer interface {
	Search(context.Context, *SearchRequest) (*SymbolsResponse, error)
	LocalCodeIntel(context.Context, *LocalCodeIntelRequest) (*LocalCodeIntelResponse, error)
	ListLanguages(context.Context, *emptypb.Empty) (*ListLanguagesResponse, error)
	SymbolInfo(context.Context, *SymbolInfoRequest) (*SymbolInfoResponse, error)
	Healthz(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedSymbolsServer()
}

// UnimplementedSymbolsServer must be embedded to have forward compatible implementations.
type UnimplementedSymbolsServer struct {
}

func (UnimplementedSymbolsServer) Search(context.Context, *SearchRequest) (*SymbolsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (UnimplementedSymbolsServer) LocalCodeIntel(context.Context, *LocalCodeIntelRequest) (*LocalCodeIntelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LocalCodeIntel not implemented")
}
func (UnimplementedSymbolsServer) ListLanguages(context.Context, *emptypb.Empty) (*ListLanguagesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListLanguages not implemented")
}
func (UnimplementedSymbolsServer) SymbolInfo(context.Context, *SymbolInfoRequest) (*SymbolInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SymbolInfo not implemented")
}
func (UnimplementedSymbolsServer) Healthz(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Healthz not implemented")
}
func (UnimplementedSymbolsServer) mustEmbedUnimplementedSymbolsServer() {}

// UnsafeSymbolsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SymbolsServer will
// result in compilation errors.
type UnsafeSymbolsServer interface {
	mustEmbedUnimplementedSymbolsServer()
}

func RegisterSymbolsServer(s grpc.ServiceRegistrar, srv SymbolsServer) {
	s.RegisterService(&Symbols_ServiceDesc, srv)
}

func _Symbols_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SymbolsServer).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/symbols.v1.Symbols/Search",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SymbolsServer).Search(ctx, req.(*SearchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Symbols_LocalCodeIntel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LocalCodeIntelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SymbolsServer).LocalCodeIntel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/symbols.v1.Symbols/LocalCodeIntel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SymbolsServer).LocalCodeIntel(ctx, req.(*LocalCodeIntelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Symbols_ListLanguages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SymbolsServer).ListLanguages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/symbols.v1.Symbols/ListLanguages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SymbolsServer).ListLanguages(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Symbols_SymbolInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SymbolInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SymbolsServer).SymbolInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/symbols.v1.Symbols/SymbolInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SymbolsServer).SymbolInfo(ctx, req.(*SymbolInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Symbols_Healthz_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SymbolsServer).Healthz(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/symbols.v1.Symbols/Healthz",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SymbolsServer).Healthz(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Symbols_ServiceDesc is the grpc.ServiceDesc for Symbols service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Symbols_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "symbols.v1.Symbols",
	HandlerType: (*SymbolsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Search",
			Handler:    _Symbols_Search_Handler,
		},
		{
			MethodName: "LocalCodeIntel",
			Handler:    _Symbols_LocalCodeIntel_Handler,
		},
		{
			MethodName: "ListLanguages",
			Handler:    _Symbols_ListLanguages_Handler,
		},
		{
			MethodName: "SymbolInfo",
			Handler:    _Symbols_SymbolInfo_Handler,
		},
		{
			MethodName: "Healthz",
			Handler:    _Symbols_Healthz_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "symbols.proto",
}
