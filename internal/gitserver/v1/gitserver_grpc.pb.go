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
	GitserverService_BatchLog_FullMethodName                    = "/gitserver.v1.GitserverService/BatchLog"
	GitserverService_CreateCommitFromPatchBinary_FullMethodName = "/gitserver.v1.GitserverService/CreateCommitFromPatchBinary"
	GitserverService_DiskInfo_FullMethodName                    = "/gitserver.v1.GitserverService/DiskInfo"
	GitserverService_Exec_FullMethodName                        = "/gitserver.v1.GitserverService/Exec"
	GitserverService_GetObject_FullMethodName                   = "/gitserver.v1.GitserverService/GetObject"
	GitserverService_IsRepoCloneable_FullMethodName             = "/gitserver.v1.GitserverService/IsRepoCloneable"
	GitserverService_ListGitolite_FullMethodName                = "/gitserver.v1.GitserverService/ListGitolite"
	GitserverService_Search_FullMethodName                      = "/gitserver.v1.GitserverService/Search"
	GitserverService_Archive_FullMethodName                     = "/gitserver.v1.GitserverService/Archive"
	GitserverService_P4Exec_FullMethodName                      = "/gitserver.v1.GitserverService/P4Exec"
	GitserverService_RepoClone_FullMethodName                   = "/gitserver.v1.GitserverService/RepoClone"
	GitserverService_RepoCloneProgress_FullMethodName           = "/gitserver.v1.GitserverService/RepoCloneProgress"
	GitserverService_RepoDelete_FullMethodName                  = "/gitserver.v1.GitserverService/RepoDelete"
	GitserverService_RepoUpdate_FullMethodName                  = "/gitserver.v1.GitserverService/RepoUpdate"
	GitserverService_IsPerforcePathCloneable_FullMethodName     = "/gitserver.v1.GitserverService/IsPerforcePathCloneable"
)

// GitserverServiceClient is the client API for GitserverService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GitserverServiceClient interface {
	BatchLog(ctx context.Context, in *BatchLogRequest, opts ...grpc.CallOption) (*BatchLogResponse, error)
	CreateCommitFromPatchBinary(ctx context.Context, opts ...grpc.CallOption) (GitserverService_CreateCommitFromPatchBinaryClient, error)
	DiskInfo(ctx context.Context, in *DiskInfoRequest, opts ...grpc.CallOption) (*DiskInfoResponse, error)
	Exec(ctx context.Context, in *ExecRequest, opts ...grpc.CallOption) (GitserverService_ExecClient, error)
	GetObject(ctx context.Context, in *GetObjectRequest, opts ...grpc.CallOption) (*GetObjectResponse, error)
	IsRepoCloneable(ctx context.Context, in *IsRepoCloneableRequest, opts ...grpc.CallOption) (*IsRepoCloneableResponse, error)
	ListGitolite(ctx context.Context, in *ListGitoliteRequest, opts ...grpc.CallOption) (*ListGitoliteResponse, error)
	Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (GitserverService_SearchClient, error)
	Archive(ctx context.Context, in *ArchiveRequest, opts ...grpc.CallOption) (GitserverService_ArchiveClient, error)
	P4Exec(ctx context.Context, in *P4ExecRequest, opts ...grpc.CallOption) (GitserverService_P4ExecClient, error)
	RepoClone(ctx context.Context, in *RepoCloneRequest, opts ...grpc.CallOption) (*RepoCloneResponse, error)
	RepoCloneProgress(ctx context.Context, in *RepoCloneProgressRequest, opts ...grpc.CallOption) (*RepoCloneProgressResponse, error)
	RepoDelete(ctx context.Context, in *RepoDeleteRequest, opts ...grpc.CallOption) (*RepoDeleteResponse, error)
	RepoUpdate(ctx context.Context, in *RepoUpdateRequest, opts ...grpc.CallOption) (*RepoUpdateResponse, error)
	IsPerforcePathCloneable(ctx context.Context, in *IsPerforcePathCloneableRequest, opts ...grpc.CallOption) (*IsPerforcePathCloneableResponse, error)
}

type gitserverServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGitserverServiceClient(cc grpc.ClientConnInterface) GitserverServiceClient {
	return &gitserverServiceClient{cc}
}

func (c *gitserverServiceClient) BatchLog(ctx context.Context, in *BatchLogRequest, opts ...grpc.CallOption) (*BatchLogResponse, error) {
	out := new(BatchLogResponse)
	err := c.cc.Invoke(ctx, GitserverService_BatchLog_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) CreateCommitFromPatchBinary(ctx context.Context, opts ...grpc.CallOption) (GitserverService_CreateCommitFromPatchBinaryClient, error) {
	stream, err := c.cc.NewStream(ctx, &GitserverService_ServiceDesc.Streams[0], GitserverService_CreateCommitFromPatchBinary_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &gitserverServiceCreateCommitFromPatchBinaryClient{stream}
	return x, nil
}

type GitserverService_CreateCommitFromPatchBinaryClient interface {
	Send(*CreateCommitFromPatchBinaryRequest) error
	CloseAndRecv() (*CreateCommitFromPatchBinaryResponse, error)
	grpc.ClientStream
}

type gitserverServiceCreateCommitFromPatchBinaryClient struct {
	grpc.ClientStream
}

func (x *gitserverServiceCreateCommitFromPatchBinaryClient) Send(m *CreateCommitFromPatchBinaryRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *gitserverServiceCreateCommitFromPatchBinaryClient) CloseAndRecv() (*CreateCommitFromPatchBinaryResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(CreateCommitFromPatchBinaryResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *gitserverServiceClient) DiskInfo(ctx context.Context, in *DiskInfoRequest, opts ...grpc.CallOption) (*DiskInfoResponse, error) {
	out := new(DiskInfoResponse)
	err := c.cc.Invoke(ctx, GitserverService_DiskInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) Exec(ctx context.Context, in *ExecRequest, opts ...grpc.CallOption) (GitserverService_ExecClient, error) {
	stream, err := c.cc.NewStream(ctx, &GitserverService_ServiceDesc.Streams[1], GitserverService_Exec_FullMethodName, opts...)
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

func (c *gitserverServiceClient) GetObject(ctx context.Context, in *GetObjectRequest, opts ...grpc.CallOption) (*GetObjectResponse, error) {
	out := new(GetObjectResponse)
	err := c.cc.Invoke(ctx, GitserverService_GetObject_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) IsRepoCloneable(ctx context.Context, in *IsRepoCloneableRequest, opts ...grpc.CallOption) (*IsRepoCloneableResponse, error) {
	out := new(IsRepoCloneableResponse)
	err := c.cc.Invoke(ctx, GitserverService_IsRepoCloneable_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) ListGitolite(ctx context.Context, in *ListGitoliteRequest, opts ...grpc.CallOption) (*ListGitoliteResponse, error) {
	out := new(ListGitoliteResponse)
	err := c.cc.Invoke(ctx, GitserverService_ListGitolite_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (GitserverService_SearchClient, error) {
	stream, err := c.cc.NewStream(ctx, &GitserverService_ServiceDesc.Streams[2], GitserverService_Search_FullMethodName, opts...)
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
	stream, err := c.cc.NewStream(ctx, &GitserverService_ServiceDesc.Streams[3], GitserverService_Archive_FullMethodName, opts...)
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

func (c *gitserverServiceClient) P4Exec(ctx context.Context, in *P4ExecRequest, opts ...grpc.CallOption) (GitserverService_P4ExecClient, error) {
	stream, err := c.cc.NewStream(ctx, &GitserverService_ServiceDesc.Streams[4], GitserverService_P4Exec_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &gitserverServiceP4ExecClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type GitserverService_P4ExecClient interface {
	Recv() (*P4ExecResponse, error)
	grpc.ClientStream
}

type gitserverServiceP4ExecClient struct {
	grpc.ClientStream
}

func (x *gitserverServiceP4ExecClient) Recv() (*P4ExecResponse, error) {
	m := new(P4ExecResponse)
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

func (c *gitserverServiceClient) RepoCloneProgress(ctx context.Context, in *RepoCloneProgressRequest, opts ...grpc.CallOption) (*RepoCloneProgressResponse, error) {
	out := new(RepoCloneProgressResponse)
	err := c.cc.Invoke(ctx, GitserverService_RepoCloneProgress_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) RepoDelete(ctx context.Context, in *RepoDeleteRequest, opts ...grpc.CallOption) (*RepoDeleteResponse, error) {
	out := new(RepoDeleteResponse)
	err := c.cc.Invoke(ctx, GitserverService_RepoDelete_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) RepoUpdate(ctx context.Context, in *RepoUpdateRequest, opts ...grpc.CallOption) (*RepoUpdateResponse, error) {
	out := new(RepoUpdateResponse)
	err := c.cc.Invoke(ctx, GitserverService_RepoUpdate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) IsPerforcePathCloneable(ctx context.Context, in *IsPerforcePathCloneableRequest, opts ...grpc.CallOption) (*IsPerforcePathCloneableResponse, error) {
	out := new(IsPerforcePathCloneableResponse)
	err := c.cc.Invoke(ctx, GitserverService_IsPerforcePathCloneable_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GitserverServiceServer is the server API for GitserverService service.
// All implementations must embed UnimplementedGitserverServiceServer
// for forward compatibility
type GitserverServiceServer interface {
	BatchLog(context.Context, *BatchLogRequest) (*BatchLogResponse, error)
	CreateCommitFromPatchBinary(GitserverService_CreateCommitFromPatchBinaryServer) error
	DiskInfo(context.Context, *DiskInfoRequest) (*DiskInfoResponse, error)
	Exec(*ExecRequest, GitserverService_ExecServer) error
	GetObject(context.Context, *GetObjectRequest) (*GetObjectResponse, error)
	IsRepoCloneable(context.Context, *IsRepoCloneableRequest) (*IsRepoCloneableResponse, error)
	ListGitolite(context.Context, *ListGitoliteRequest) (*ListGitoliteResponse, error)
	Search(*SearchRequest, GitserverService_SearchServer) error
	Archive(*ArchiveRequest, GitserverService_ArchiveServer) error
	P4Exec(*P4ExecRequest, GitserverService_P4ExecServer) error
	RepoClone(context.Context, *RepoCloneRequest) (*RepoCloneResponse, error)
	RepoCloneProgress(context.Context, *RepoCloneProgressRequest) (*RepoCloneProgressResponse, error)
	RepoDelete(context.Context, *RepoDeleteRequest) (*RepoDeleteResponse, error)
	RepoUpdate(context.Context, *RepoUpdateRequest) (*RepoUpdateResponse, error)
	IsPerforcePathCloneable(context.Context, *IsPerforcePathCloneableRequest) (*IsPerforcePathCloneableResponse, error)
	mustEmbedUnimplementedGitserverServiceServer()
}

// UnimplementedGitserverServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGitserverServiceServer struct {
}

func (UnimplementedGitserverServiceServer) BatchLog(context.Context, *BatchLogRequest) (*BatchLogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchLog not implemented")
}
func (UnimplementedGitserverServiceServer) CreateCommitFromPatchBinary(GitserverService_CreateCommitFromPatchBinaryServer) error {
	return status.Errorf(codes.Unimplemented, "method CreateCommitFromPatchBinary not implemented")
}
func (UnimplementedGitserverServiceServer) DiskInfo(context.Context, *DiskInfoRequest) (*DiskInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DiskInfo not implemented")
}
func (UnimplementedGitserverServiceServer) Exec(*ExecRequest, GitserverService_ExecServer) error {
	return status.Errorf(codes.Unimplemented, "method Exec not implemented")
}
func (UnimplementedGitserverServiceServer) GetObject(context.Context, *GetObjectRequest) (*GetObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetObject not implemented")
}
func (UnimplementedGitserverServiceServer) IsRepoCloneable(context.Context, *IsRepoCloneableRequest) (*IsRepoCloneableResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsRepoCloneable not implemented")
}
func (UnimplementedGitserverServiceServer) ListGitolite(context.Context, *ListGitoliteRequest) (*ListGitoliteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListGitolite not implemented")
}
func (UnimplementedGitserverServiceServer) Search(*SearchRequest, GitserverService_SearchServer) error {
	return status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (UnimplementedGitserverServiceServer) Archive(*ArchiveRequest, GitserverService_ArchiveServer) error {
	return status.Errorf(codes.Unimplemented, "method Archive not implemented")
}
func (UnimplementedGitserverServiceServer) P4Exec(*P4ExecRequest, GitserverService_P4ExecServer) error {
	return status.Errorf(codes.Unimplemented, "method P4Exec not implemented")
}
func (UnimplementedGitserverServiceServer) RepoClone(context.Context, *RepoCloneRequest) (*RepoCloneResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RepoClone not implemented")
}
func (UnimplementedGitserverServiceServer) RepoCloneProgress(context.Context, *RepoCloneProgressRequest) (*RepoCloneProgressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RepoCloneProgress not implemented")
}
func (UnimplementedGitserverServiceServer) RepoDelete(context.Context, *RepoDeleteRequest) (*RepoDeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RepoDelete not implemented")
}
func (UnimplementedGitserverServiceServer) RepoUpdate(context.Context, *RepoUpdateRequest) (*RepoUpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RepoUpdate not implemented")
}
func (UnimplementedGitserverServiceServer) IsPerforcePathCloneable(context.Context, *IsPerforcePathCloneableRequest) (*IsPerforcePathCloneableResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsPerforcePathCloneable not implemented")
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

func _GitserverService_BatchLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).BatchLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_BatchLog_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitserverServiceServer).BatchLog(ctx, req.(*BatchLogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitserverService_CreateCommitFromPatchBinary_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GitserverServiceServer).CreateCommitFromPatchBinary(&gitserverServiceCreateCommitFromPatchBinaryServer{stream})
}

type GitserverService_CreateCommitFromPatchBinaryServer interface {
	SendAndClose(*CreateCommitFromPatchBinaryResponse) error
	Recv() (*CreateCommitFromPatchBinaryRequest, error)
	grpc.ServerStream
}

type gitserverServiceCreateCommitFromPatchBinaryServer struct {
	grpc.ServerStream
}

func (x *gitserverServiceCreateCommitFromPatchBinaryServer) SendAndClose(m *CreateCommitFromPatchBinaryResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *gitserverServiceCreateCommitFromPatchBinaryServer) Recv() (*CreateCommitFromPatchBinaryRequest, error) {
	m := new(CreateCommitFromPatchBinaryRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _GitserverService_DiskInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DiskInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).DiskInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_DiskInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitserverServiceServer).DiskInfo(ctx, req.(*DiskInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
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

func _GitserverService_GetObject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).GetObject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_GetObject_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitserverServiceServer).GetObject(ctx, req.(*GetObjectRequest))
	}
	return interceptor(ctx, in, info, handler)
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

func _GitserverService_ListGitolite_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListGitoliteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).ListGitolite(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_ListGitolite_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitserverServiceServer).ListGitolite(ctx, req.(*ListGitoliteRequest))
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

func _GitserverService_P4Exec_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(P4ExecRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GitserverServiceServer).P4Exec(m, &gitserverServiceP4ExecServer{stream})
}

type GitserverService_P4ExecServer interface {
	Send(*P4ExecResponse) error
	grpc.ServerStream
}

type gitserverServiceP4ExecServer struct {
	grpc.ServerStream
}

func (x *gitserverServiceP4ExecServer) Send(m *P4ExecResponse) error {
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

func _GitserverService_RepoCloneProgress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RepoCloneProgressRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).RepoCloneProgress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_RepoCloneProgress_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitserverServiceServer).RepoCloneProgress(ctx, req.(*RepoCloneProgressRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitserverService_RepoDelete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RepoDeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).RepoDelete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_RepoDelete_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitserverServiceServer).RepoDelete(ctx, req.(*RepoDeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitserverService_RepoUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RepoUpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).RepoUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_RepoUpdate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitserverServiceServer).RepoUpdate(ctx, req.(*RepoUpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GitserverService_IsPerforcePathCloneable_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsPerforcePathCloneableRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).IsPerforcePathCloneable(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_IsPerforcePathCloneable_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GitserverServiceServer).IsPerforcePathCloneable(ctx, req.(*IsPerforcePathCloneableRequest))
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
			MethodName: "BatchLog",
			Handler:    _GitserverService_BatchLog_Handler,
		},
		{
			MethodName: "DiskInfo",
			Handler:    _GitserverService_DiskInfo_Handler,
		},
		{
			MethodName: "GetObject",
			Handler:    _GitserverService_GetObject_Handler,
		},
		{
			MethodName: "IsRepoCloneable",
			Handler:    _GitserverService_IsRepoCloneable_Handler,
		},
		{
			MethodName: "ListGitolite",
			Handler:    _GitserverService_ListGitolite_Handler,
		},
		{
			MethodName: "RepoClone",
			Handler:    _GitserverService_RepoClone_Handler,
		},
		{
			MethodName: "RepoCloneProgress",
			Handler:    _GitserverService_RepoCloneProgress_Handler,
		},
		{
			MethodName: "RepoDelete",
			Handler:    _GitserverService_RepoDelete_Handler,
		},
		{
			MethodName: "RepoUpdate",
			Handler:    _GitserverService_RepoUpdate_Handler,
		},
		{
			MethodName: "IsPerforcePathCloneable",
			Handler:    _GitserverService_IsPerforcePathCloneable_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "CreateCommitFromPatchBinary",
			Handler:       _GitserverService_CreateCommitFromPatchBinary_Handler,
			ClientStreams: true,
		},
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
		{
			StreamName:    "P4Exec",
			Handler:       _GitserverService_P4Exec_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "gitserver.proto",
}
