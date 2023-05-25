// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: gitserver.proto

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
	GitserverService_Exec_FullMethodName            = "/gitserver.v1.GitserverService/Exec"
	GitserverService_IsRepoCloneable_FullMethodName = "/gitserver.v1.GitserverService/IsRepoCloneable"
	GitserverService_Search_FullMethodName          = "/gitserver.v1.GitserverService/Search"
	GitserverService_Archive_FullMethodName         = "/gitserver.v1.GitserverService/Archive"
	GitserverService_RepoClone_FullMethodName       = "/gitserver.v1.GitserverService/RepoClone"
	GitserverService_ReposStats_FullMethodName      = "/gitserver.v1.GitserverService/ReposStats"
)

// GitserverServiceClient is the client API for GitserverService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GitserverServiceClient interface {
	Exec(ctx context.Context, in *ExecRequest, opts ...grpc.CallOption) (GitserverService_ExecClient, error)
	IsRepoCloneable(ctx context.Context, in *IsRepoCloneableRequest, opts ...grpc.CallOption) (*IsRepoCloneableResponse, error)
	Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (GitserverService_SearchClient, error)
	Archive(ctx context.Context, in *ArchiveRequest, opts ...grpc.CallOption) (GitserverService_ArchiveClient, error)
	RepoClone(ctx context.Context, in *RepoCloneRequest, opts ...grpc.CallOption) (*RepoCloneResponse, error)
	ReposStats(ctx context.Context, in *ReposStatsRequest, opts ...grpc.CallOption) (*ReposStatsResponse, error)
}

type gitserverServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGitserverServiceClient(cc grpc.ClientConnInterface) GitserverServiceClient {
	return &gitserverServiceClient{cc}
}

func (c *gitserverServiceClient) Exec(ctx context.Context, in *ExecRequest, opts ...grpc.CallOption) (GitserverService_ExecClient, error) {
	stream, err := c.cc.NewStream(ctx, &GitserverService_ServiceDesc.Streams[0], GitserverService_Exec_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &gitserverServiceExecClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type GitserverService_ExecClient interface {
	Recv() (*ExecResponse, error)
	grpc.ClientStream
}

type gitserverServiceExecClient struct {
	grpc.ClientStream
}

func (x *gitserverServiceExecClient) Recv() (*ExecResponse, error) {
	m := new(ExecResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *gitserverServiceClient) IsRepoCloneable(ctx context.Context, in *IsRepoCloneableRequest, opts ...grpc.CallOption) (*IsRepoCloneableResponse, error) {
	out := new(IsRepoCloneableResponse)
	err := c.cc.Invoke(ctx, GitserverService_IsRepoCloneable_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (GitserverService_SearchClient, error) {
	stream, err := c.cc.NewStream(ctx, &GitserverService_ServiceDesc.Streams[1], GitserverService_Search_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &gitserverServiceSearchClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type GitserverService_SearchClient interface {
	Recv() (*SearchResponse, error)
	grpc.ClientStream
}

type gitserverServiceSearchClient struct {
	grpc.ClientStream
}

func (x *gitserverServiceSearchClient) Recv() (*SearchResponse, error) {
	m := new(SearchResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *gitserverServiceClient) Archive(ctx context.Context, in *ArchiveRequest, opts ...grpc.CallOption) (GitserverService_ArchiveClient, error) {
	stream, err := c.cc.NewStream(ctx, &GitserverService_ServiceDesc.Streams[2], GitserverService_Archive_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &gitserverServiceArchiveClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type GitserverService_ArchiveClient interface {
	Recv() (*ArchiveResponse, error)
	grpc.ClientStream
}

type gitserverServiceArchiveClient struct {
	grpc.ClientStream
}

func (x *gitserverServiceArchiveClient) Recv() (*ArchiveResponse, error) {
	m := new(ArchiveResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *gitserverServiceClient) RepoClone(ctx context.Context, in *RepoCloneRequest, opts ...grpc.CallOption) (*RepoCloneResponse, error) {
	out := new(RepoCloneResponse)
	err := c.cc.Invoke(ctx, GitserverService_RepoClone_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) ReposStats(ctx context.Context, in *ReposStatsRequest, opts ...grpc.CallOption) (*ReposStatsResponse, error) {
	out := new(ReposStatsResponse)
	err := c.cc.Invoke(ctx, GitserverService_ReposStats_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GitserverServiceServer is the server API for GitserverService service.
// All implementations must embed UnimplementedGitserverServiceServer
// for forward compatibility
type GitserverServiceServer interface {
	Exec(*ExecRequest, GitserverService_ExecServer) error
	IsRepoCloneable(context.Context, *IsRepoCloneableRequest) (*IsRepoCloneableResponse, error)
	Search(*SearchRequest, GitserverService_SearchServer) error
	Archive(*ArchiveRequest, GitserverService_ArchiveServer) error
	RepoClone(context.Context, *RepoCloneRequest) (*RepoCloneResponse, error)
	ReposStats(context.Context, *ReposStatsRequest) (*ReposStatsResponse, error)
	mustEmbedUnimplementedGitserverServiceServer()
}

// UnimplementedGitserverServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGitserverServiceServer struct {
}

func (UnimplementedGitserverServiceServer) Exec(*ExecRequest, GitserverService_ExecServer) error {
	return status.Errorf(codes.Unimplemented, "method Exec not implemented")
}
func (UnimplementedGitserverServiceServer) IsRepoCloneable(context.Context, *IsRepoCloneableRequest) (*IsRepoCloneableResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsRepoCloneable not implemented")
}
func (UnimplementedGitserverServiceServer) Search(*SearchRequest, GitserverService_SearchServer) error {
	return status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (UnimplementedGitserverServiceServer) Archive(*ArchiveRequest, GitserverService_ArchiveServer) error {
	return status.Errorf(codes.Unimplemented, "method Archive not implemented")
}
func (UnimplementedGitserverServiceServer) RepoClone(context.Context, *RepoCloneRequest) (*RepoCloneResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RepoClone not implemented")
}
func (UnimplementedGitserverServiceServer) ReposStats(context.Context, *ReposStatsRequest) (*ReposStatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReposStats not implemented")
}
func (UnimplementedGitserverServiceServer) mustEmbedUnimplementedGitserverServiceServer() {}

// UnsafeGitserverServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GitserverServiceServer will
// result in compilation errors.
type UnsafeGitserverServiceServer interface {
	mustEmbedUnimplementedGitserverServiceServer()
}

func RegisterGitserverServiceServer(s grpc.ServiceRegistrar, srv GitserverServiceServer) {
	s.RegisterService(&GitserverService_ServiceDesc, srv)
}

func _GitserverService_Exec_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ExecRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GitserverServiceServer).Exec(m, &gitserverServiceExecServer{stream})
}

type GitserverService_ExecServer interface {
	Send(*ExecResponse) error
	grpc.ServerStream
}

type gitserverServiceExecServer struct {
	grpc.ServerStream
}

func (x *gitserverServiceExecServer) Send(m *ExecResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _GitserverService_IsRepoCloneable_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsRepoCloneableRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).IsRepoCloneable(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_IsRepoCloneable_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitserverServiceServer).IsRepoCloneable(ctx, req.(*IsRepoCloneableRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitserverService_Search_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SearchRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GitserverServiceServer).Search(m, &gitserverServiceSearchServer{stream})
}

type GitserverService_SearchServer interface {
	Send(*SearchResponse) error
	grpc.ServerStream
}

type gitserverServiceSearchServer struct {
	grpc.ServerStream
}

func (x *gitserverServiceSearchServer) Send(m *SearchResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _GitserverService_Archive_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ArchiveRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GitserverServiceServer).Archive(m, &gitserverServiceArchiveServer{stream})
}

type GitserverService_ArchiveServer interface {
	Send(*ArchiveResponse) error
	grpc.ServerStream
}

type gitserverServiceArchiveServer struct {
	grpc.ServerStream
}

func (x *gitserverServiceArchiveServer) Send(m *ArchiveResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _GitserverService_RepoClone_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RepoCloneRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).RepoClone(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_RepoClone_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitserverServiceServer).RepoClone(ctx, req.(*RepoCloneRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitserverService_ReposStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReposStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).ReposStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_ReposStats_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitserverServiceServer).ReposStats(ctx, req.(*ReposStatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GitserverService_ServiceDesc is the grpc.ServiceDesc for GitserverService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GitserverService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gitserver.v1.GitserverService",
	HandlerType: (*GitserverServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "IsRepoCloneable",
			Handler:    _GitserverService_IsRepoCloneable_Handler,
		},
		{
			MethodName: "RepoClone",
			Handler:    _GitserverService_RepoClone_Handler,
		},
		{
			MethodName: "ReposStats",
			Handler:    _GitserverService_ReposStats_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Exec",
			Handler:       _GitserverService_Exec_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Search",
			Handler:       _GitserverService_Search_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Archive",
			Handler:       _GitserverService_Archive_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "gitserver.proto",
}
