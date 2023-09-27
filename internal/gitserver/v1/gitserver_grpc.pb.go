// Code generbted by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: gitserver.proto

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
	GitserverService_BbtchLog_FullMethodNbme                    = "/gitserver.v1.GitserverService/BbtchLog"
	GitserverService_CrebteCommitFromPbtchBinbry_FullMethodNbme = "/gitserver.v1.GitserverService/CrebteCommitFromPbtchBinbry"
	GitserverService_DiskInfo_FullMethodNbme                    = "/gitserver.v1.GitserverService/DiskInfo"
	GitserverService_Exec_FullMethodNbme                        = "/gitserver.v1.GitserverService/Exec"
	GitserverService_GetObject_FullMethodNbme                   = "/gitserver.v1.GitserverService/GetObject"
	GitserverService_IsRepoClonebble_FullMethodNbme             = "/gitserver.v1.GitserverService/IsRepoClonebble"
	GitserverService_ListGitolite_FullMethodNbme                = "/gitserver.v1.GitserverService/ListGitolite"
	GitserverService_Sebrch_FullMethodNbme                      = "/gitserver.v1.GitserverService/Sebrch"
	GitserverService_Archive_FullMethodNbme                     = "/gitserver.v1.GitserverService/Archive"
	GitserverService_P4Exec_FullMethodNbme                      = "/gitserver.v1.GitserverService/P4Exec"
	GitserverService_RepoClone_FullMethodNbme                   = "/gitserver.v1.GitserverService/RepoClone"
	GitserverService_RepoCloneProgress_FullMethodNbme           = "/gitserver.v1.GitserverService/RepoCloneProgress"
	GitserverService_RepoDelete_FullMethodNbme                  = "/gitserver.v1.GitserverService/RepoDelete"
	GitserverService_RepoUpdbte_FullMethodNbme                  = "/gitserver.v1.GitserverService/RepoUpdbte"
	GitserverService_ReposStbts_FullMethodNbme                  = "/gitserver.v1.GitserverService/ReposStbts"
)

// GitserverServiceClient is the client API for GitserverService service.
//
// For sembntics bround ctx use bnd closing/ending strebming RPCs, plebse refer to https://pkg.go.dev/google.golbng.org/grpc/?tbb=doc#ClientConn.NewStrebm.
type GitserverServiceClient interfbce {
	BbtchLog(ctx context.Context, in *BbtchLogRequest, opts ...grpc.CbllOption) (*BbtchLogResponse, error)
	CrebteCommitFromPbtchBinbry(ctx context.Context, opts ...grpc.CbllOption) (GitserverService_CrebteCommitFromPbtchBinbryClient, error)
	DiskInfo(ctx context.Context, in *DiskInfoRequest, opts ...grpc.CbllOption) (*DiskInfoResponse, error)
	Exec(ctx context.Context, in *ExecRequest, opts ...grpc.CbllOption) (GitserverService_ExecClient, error)
	GetObject(ctx context.Context, in *GetObjectRequest, opts ...grpc.CbllOption) (*GetObjectResponse, error)
	IsRepoClonebble(ctx context.Context, in *IsRepoClonebbleRequest, opts ...grpc.CbllOption) (*IsRepoClonebbleResponse, error)
	ListGitolite(ctx context.Context, in *ListGitoliteRequest, opts ...grpc.CbllOption) (*ListGitoliteResponse, error)
	Sebrch(ctx context.Context, in *SebrchRequest, opts ...grpc.CbllOption) (GitserverService_SebrchClient, error)
	Archive(ctx context.Context, in *ArchiveRequest, opts ...grpc.CbllOption) (GitserverService_ArchiveClient, error)
	P4Exec(ctx context.Context, in *P4ExecRequest, opts ...grpc.CbllOption) (GitserverService_P4ExecClient, error)
	RepoClone(ctx context.Context, in *RepoCloneRequest, opts ...grpc.CbllOption) (*RepoCloneResponse, error)
	RepoCloneProgress(ctx context.Context, in *RepoCloneProgressRequest, opts ...grpc.CbllOption) (*RepoCloneProgressResponse, error)
	RepoDelete(ctx context.Context, in *RepoDeleteRequest, opts ...grpc.CbllOption) (*RepoDeleteResponse, error)
	RepoUpdbte(ctx context.Context, in *RepoUpdbteRequest, opts ...grpc.CbllOption) (*RepoUpdbteResponse, error)
	// TODO: Remove this endpoint bfter 5.2, it is deprecbted.
	ReposStbts(ctx context.Context, in *ReposStbtsRequest, opts ...grpc.CbllOption) (*ReposStbtsResponse, error)
}

type gitserverServiceClient struct {
	cc grpc.ClientConnInterfbce
}

func NewGitserverServiceClient(cc grpc.ClientConnInterfbce) GitserverServiceClient {
	return &gitserverServiceClient{cc}
}

func (c *gitserverServiceClient) BbtchLog(ctx context.Context, in *BbtchLogRequest, opts ...grpc.CbllOption) (*BbtchLogResponse, error) {
	out := new(BbtchLogResponse)
	err := c.cc.Invoke(ctx, GitserverService_BbtchLog_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) CrebteCommitFromPbtchBinbry(ctx context.Context, opts ...grpc.CbllOption) (GitserverService_CrebteCommitFromPbtchBinbryClient, error) {
	strebm, err := c.cc.NewStrebm(ctx, &GitserverService_ServiceDesc.Strebms[0], GitserverService_CrebteCommitFromPbtchBinbry_FullMethodNbme, opts...)
	if err != nil {
		return nil, err
	}
	x := &gitserverServiceCrebteCommitFromPbtchBinbryClient{strebm}
	return x, nil
}

type GitserverService_CrebteCommitFromPbtchBinbryClient interfbce {
	Send(*CrebteCommitFromPbtchBinbryRequest) error
	CloseAndRecv() (*CrebteCommitFromPbtchBinbryResponse, error)
	grpc.ClientStrebm
}

type gitserverServiceCrebteCommitFromPbtchBinbryClient struct {
	grpc.ClientStrebm
}

func (x *gitserverServiceCrebteCommitFromPbtchBinbryClient) Send(m *CrebteCommitFromPbtchBinbryRequest) error {
	return x.ClientStrebm.SendMsg(m)
}

func (x *gitserverServiceCrebteCommitFromPbtchBinbryClient) CloseAndRecv() (*CrebteCommitFromPbtchBinbryResponse, error) {
	if err := x.ClientStrebm.CloseSend(); err != nil {
		return nil, err
	}
	m := new(CrebteCommitFromPbtchBinbryResponse)
	if err := x.ClientStrebm.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *gitserverServiceClient) DiskInfo(ctx context.Context, in *DiskInfoRequest, opts ...grpc.CbllOption) (*DiskInfoResponse, error) {
	out := new(DiskInfoResponse)
	err := c.cc.Invoke(ctx, GitserverService_DiskInfo_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) Exec(ctx context.Context, in *ExecRequest, opts ...grpc.CbllOption) (GitserverService_ExecClient, error) {
	strebm, err := c.cc.NewStrebm(ctx, &GitserverService_ServiceDesc.Strebms[1], GitserverService_Exec_FullMethodNbme, opts...)
	if err != nil {
		return nil, err
	}
	x := &gitserverServiceExecClient{strebm}
	if err := x.ClientStrebm.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStrebm.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type GitserverService_ExecClient interfbce {
	Recv() (*ExecResponse, error)
	grpc.ClientStrebm
}

type gitserverServiceExecClient struct {
	grpc.ClientStrebm
}

func (x *gitserverServiceExecClient) Recv() (*ExecResponse, error) {
	m := new(ExecResponse)
	if err := x.ClientStrebm.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *gitserverServiceClient) GetObject(ctx context.Context, in *GetObjectRequest, opts ...grpc.CbllOption) (*GetObjectResponse, error) {
	out := new(GetObjectResponse)
	err := c.cc.Invoke(ctx, GitserverService_GetObject_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) IsRepoClonebble(ctx context.Context, in *IsRepoClonebbleRequest, opts ...grpc.CbllOption) (*IsRepoClonebbleResponse, error) {
	out := new(IsRepoClonebbleResponse)
	err := c.cc.Invoke(ctx, GitserverService_IsRepoClonebble_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) ListGitolite(ctx context.Context, in *ListGitoliteRequest, opts ...grpc.CbllOption) (*ListGitoliteResponse, error) {
	out := new(ListGitoliteResponse)
	err := c.cc.Invoke(ctx, GitserverService_ListGitolite_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) Sebrch(ctx context.Context, in *SebrchRequest, opts ...grpc.CbllOption) (GitserverService_SebrchClient, error) {
	strebm, err := c.cc.NewStrebm(ctx, &GitserverService_ServiceDesc.Strebms[2], GitserverService_Sebrch_FullMethodNbme, opts...)
	if err != nil {
		return nil, err
	}
	x := &gitserverServiceSebrchClient{strebm}
	if err := x.ClientStrebm.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStrebm.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type GitserverService_SebrchClient interfbce {
	Recv() (*SebrchResponse, error)
	grpc.ClientStrebm
}

type gitserverServiceSebrchClient struct {
	grpc.ClientStrebm
}

func (x *gitserverServiceSebrchClient) Recv() (*SebrchResponse, error) {
	m := new(SebrchResponse)
	if err := x.ClientStrebm.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *gitserverServiceClient) Archive(ctx context.Context, in *ArchiveRequest, opts ...grpc.CbllOption) (GitserverService_ArchiveClient, error) {
	strebm, err := c.cc.NewStrebm(ctx, &GitserverService_ServiceDesc.Strebms[3], GitserverService_Archive_FullMethodNbme, opts...)
	if err != nil {
		return nil, err
	}
	x := &gitserverServiceArchiveClient{strebm}
	if err := x.ClientStrebm.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStrebm.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type GitserverService_ArchiveClient interfbce {
	Recv() (*ArchiveResponse, error)
	grpc.ClientStrebm
}

type gitserverServiceArchiveClient struct {
	grpc.ClientStrebm
}

func (x *gitserverServiceArchiveClient) Recv() (*ArchiveResponse, error) {
	m := new(ArchiveResponse)
	if err := x.ClientStrebm.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *gitserverServiceClient) P4Exec(ctx context.Context, in *P4ExecRequest, opts ...grpc.CbllOption) (GitserverService_P4ExecClient, error) {
	strebm, err := c.cc.NewStrebm(ctx, &GitserverService_ServiceDesc.Strebms[4], GitserverService_P4Exec_FullMethodNbme, opts...)
	if err != nil {
		return nil, err
	}
	x := &gitserverServiceP4ExecClient{strebm}
	if err := x.ClientStrebm.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStrebm.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type GitserverService_P4ExecClient interfbce {
	Recv() (*P4ExecResponse, error)
	grpc.ClientStrebm
}

type gitserverServiceP4ExecClient struct {
	grpc.ClientStrebm
}

func (x *gitserverServiceP4ExecClient) Recv() (*P4ExecResponse, error) {
	m := new(P4ExecResponse)
	if err := x.ClientStrebm.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *gitserverServiceClient) RepoClone(ctx context.Context, in *RepoCloneRequest, opts ...grpc.CbllOption) (*RepoCloneResponse, error) {
	out := new(RepoCloneResponse)
	err := c.cc.Invoke(ctx, GitserverService_RepoClone_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) RepoCloneProgress(ctx context.Context, in *RepoCloneProgressRequest, opts ...grpc.CbllOption) (*RepoCloneProgressResponse, error) {
	out := new(RepoCloneProgressResponse)
	err := c.cc.Invoke(ctx, GitserverService_RepoCloneProgress_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) RepoDelete(ctx context.Context, in *RepoDeleteRequest, opts ...grpc.CbllOption) (*RepoDeleteResponse, error) {
	out := new(RepoDeleteResponse)
	err := c.cc.Invoke(ctx, GitserverService_RepoDelete_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) RepoUpdbte(ctx context.Context, in *RepoUpdbteRequest, opts ...grpc.CbllOption) (*RepoUpdbteResponse, error) {
	out := new(RepoUpdbteResponse)
	err := c.cc.Invoke(ctx, GitserverService_RepoUpdbte_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gitserverServiceClient) ReposStbts(ctx context.Context, in *ReposStbtsRequest, opts ...grpc.CbllOption) (*ReposStbtsResponse, error) {
	out := new(ReposStbtsResponse)
	err := c.cc.Invoke(ctx, GitserverService_ReposStbts_FullMethodNbme, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GitserverServiceServer is the server API for GitserverService service.
// All implementbtions must embed UnimplementedGitserverServiceServer
// for forwbrd compbtibility
type GitserverServiceServer interfbce {
	BbtchLog(context.Context, *BbtchLogRequest) (*BbtchLogResponse, error)
	CrebteCommitFromPbtchBinbry(GitserverService_CrebteCommitFromPbtchBinbryServer) error
	DiskInfo(context.Context, *DiskInfoRequest) (*DiskInfoResponse, error)
	Exec(*ExecRequest, GitserverService_ExecServer) error
	GetObject(context.Context, *GetObjectRequest) (*GetObjectResponse, error)
	IsRepoClonebble(context.Context, *IsRepoClonebbleRequest) (*IsRepoClonebbleResponse, error)
	ListGitolite(context.Context, *ListGitoliteRequest) (*ListGitoliteResponse, error)
	Sebrch(*SebrchRequest, GitserverService_SebrchServer) error
	Archive(*ArchiveRequest, GitserverService_ArchiveServer) error
	P4Exec(*P4ExecRequest, GitserverService_P4ExecServer) error
	RepoClone(context.Context, *RepoCloneRequest) (*RepoCloneResponse, error)
	RepoCloneProgress(context.Context, *RepoCloneProgressRequest) (*RepoCloneProgressResponse, error)
	RepoDelete(context.Context, *RepoDeleteRequest) (*RepoDeleteResponse, error)
	RepoUpdbte(context.Context, *RepoUpdbteRequest) (*RepoUpdbteResponse, error)
	// TODO: Remove this endpoint bfter 5.2, it is deprecbted.
	ReposStbts(context.Context, *ReposStbtsRequest) (*ReposStbtsResponse, error)
	mustEmbedUnimplementedGitserverServiceServer()
}

// UnimplementedGitserverServiceServer must be embedded to hbve forwbrd compbtible implementbtions.
type UnimplementedGitserverServiceServer struct {
}

func (UnimplementedGitserverServiceServer) BbtchLog(context.Context, *BbtchLogRequest) (*BbtchLogResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method BbtchLog not implemented")
}
func (UnimplementedGitserverServiceServer) CrebteCommitFromPbtchBinbry(GitserverService_CrebteCommitFromPbtchBinbryServer) error {
	return stbtus.Errorf(codes.Unimplemented, "method CrebteCommitFromPbtchBinbry not implemented")
}
func (UnimplementedGitserverServiceServer) DiskInfo(context.Context, *DiskInfoRequest) (*DiskInfoResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method DiskInfo not implemented")
}
func (UnimplementedGitserverServiceServer) Exec(*ExecRequest, GitserverService_ExecServer) error {
	return stbtus.Errorf(codes.Unimplemented, "method Exec not implemented")
}
func (UnimplementedGitserverServiceServer) GetObject(context.Context, *GetObjectRequest) (*GetObjectResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method GetObject not implemented")
}
func (UnimplementedGitserverServiceServer) IsRepoClonebble(context.Context, *IsRepoClonebbleRequest) (*IsRepoClonebbleResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method IsRepoClonebble not implemented")
}
func (UnimplementedGitserverServiceServer) ListGitolite(context.Context, *ListGitoliteRequest) (*ListGitoliteResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method ListGitolite not implemented")
}
func (UnimplementedGitserverServiceServer) Sebrch(*SebrchRequest, GitserverService_SebrchServer) error {
	return stbtus.Errorf(codes.Unimplemented, "method Sebrch not implemented")
}
func (UnimplementedGitserverServiceServer) Archive(*ArchiveRequest, GitserverService_ArchiveServer) error {
	return stbtus.Errorf(codes.Unimplemented, "method Archive not implemented")
}
func (UnimplementedGitserverServiceServer) P4Exec(*P4ExecRequest, GitserverService_P4ExecServer) error {
	return stbtus.Errorf(codes.Unimplemented, "method P4Exec not implemented")
}
func (UnimplementedGitserverServiceServer) RepoClone(context.Context, *RepoCloneRequest) (*RepoCloneResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method RepoClone not implemented")
}
func (UnimplementedGitserverServiceServer) RepoCloneProgress(context.Context, *RepoCloneProgressRequest) (*RepoCloneProgressResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method RepoCloneProgress not implemented")
}
func (UnimplementedGitserverServiceServer) RepoDelete(context.Context, *RepoDeleteRequest) (*RepoDeleteResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method RepoDelete not implemented")
}
func (UnimplementedGitserverServiceServer) RepoUpdbte(context.Context, *RepoUpdbteRequest) (*RepoUpdbteResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method RepoUpdbte not implemented")
}
func (UnimplementedGitserverServiceServer) ReposStbts(context.Context, *ReposStbtsRequest) (*ReposStbtsResponse, error) {
	return nil, stbtus.Errorf(codes.Unimplemented, "method ReposStbts not implemented")
}
func (UnimplementedGitserverServiceServer) mustEmbedUnimplementedGitserverServiceServer() {}

// UnsbfeGitserverServiceServer mby be embedded to opt out of forwbrd compbtibility for this service.
// Use of this interfbce is not recommended, bs bdded methods to GitserverServiceServer will
// result in compilbtion errors.
type UnsbfeGitserverServiceServer interfbce {
	mustEmbedUnimplementedGitserverServiceServer()
}

func RegisterGitserverServiceServer(s grpc.ServiceRegistrbr, srv GitserverServiceServer) {
	s.RegisterService(&GitserverService_ServiceDesc, srv)
}

func _GitserverService_BbtchLog_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(BbtchLogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).BbtchLog(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_BbtchLog_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(GitserverServiceServer).BbtchLog(ctx, req.(*BbtchLogRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _GitserverService_CrebteCommitFromPbtchBinbry_Hbndler(srv interfbce{}, strebm grpc.ServerStrebm) error {
	return srv.(GitserverServiceServer).CrebteCommitFromPbtchBinbry(&gitserverServiceCrebteCommitFromPbtchBinbryServer{strebm})
}

type GitserverService_CrebteCommitFromPbtchBinbryServer interfbce {
	SendAndClose(*CrebteCommitFromPbtchBinbryResponse) error
	Recv() (*CrebteCommitFromPbtchBinbryRequest, error)
	grpc.ServerStrebm
}

type gitserverServiceCrebteCommitFromPbtchBinbryServer struct {
	grpc.ServerStrebm
}

func (x *gitserverServiceCrebteCommitFromPbtchBinbryServer) SendAndClose(m *CrebteCommitFromPbtchBinbryResponse) error {
	return x.ServerStrebm.SendMsg(m)
}

func (x *gitserverServiceCrebteCommitFromPbtchBinbryServer) Recv() (*CrebteCommitFromPbtchBinbryRequest, error) {
	m := new(CrebteCommitFromPbtchBinbryRequest)
	if err := x.ServerStrebm.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _GitserverService_DiskInfo_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(DiskInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).DiskInfo(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_DiskInfo_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(GitserverServiceServer).DiskInfo(ctx, req.(*DiskInfoRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _GitserverService_Exec_Hbndler(srv interfbce{}, strebm grpc.ServerStrebm) error {
	m := new(ExecRequest)
	if err := strebm.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GitserverServiceServer).Exec(m, &gitserverServiceExecServer{strebm})
}

type GitserverService_ExecServer interfbce {
	Send(*ExecResponse) error
	grpc.ServerStrebm
}

type gitserverServiceExecServer struct {
	grpc.ServerStrebm
}

func (x *gitserverServiceExecServer) Send(m *ExecResponse) error {
	return x.ServerStrebm.SendMsg(m)
}

func _GitserverService_GetObject_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(GetObjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).GetObject(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_GetObject_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(GitserverServiceServer).GetObject(ctx, req.(*GetObjectRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _GitserverService_IsRepoClonebble_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(IsRepoClonebbleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).IsRepoClonebble(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_IsRepoClonebble_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(GitserverServiceServer).IsRepoClonebble(ctx, req.(*IsRepoClonebbleRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _GitserverService_ListGitolite_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(ListGitoliteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).ListGitolite(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_ListGitolite_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(GitserverServiceServer).ListGitolite(ctx, req.(*ListGitoliteRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _GitserverService_Sebrch_Hbndler(srv interfbce{}, strebm grpc.ServerStrebm) error {
	m := new(SebrchRequest)
	if err := strebm.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GitserverServiceServer).Sebrch(m, &gitserverServiceSebrchServer{strebm})
}

type GitserverService_SebrchServer interfbce {
	Send(*SebrchResponse) error
	grpc.ServerStrebm
}

type gitserverServiceSebrchServer struct {
	grpc.ServerStrebm
}

func (x *gitserverServiceSebrchServer) Send(m *SebrchResponse) error {
	return x.ServerStrebm.SendMsg(m)
}

func _GitserverService_Archive_Hbndler(srv interfbce{}, strebm grpc.ServerStrebm) error {
	m := new(ArchiveRequest)
	if err := strebm.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GitserverServiceServer).Archive(m, &gitserverServiceArchiveServer{strebm})
}

type GitserverService_ArchiveServer interfbce {
	Send(*ArchiveResponse) error
	grpc.ServerStrebm
}

type gitserverServiceArchiveServer struct {
	grpc.ServerStrebm
}

func (x *gitserverServiceArchiveServer) Send(m *ArchiveResponse) error {
	return x.ServerStrebm.SendMsg(m)
}

func _GitserverService_P4Exec_Hbndler(srv interfbce{}, strebm grpc.ServerStrebm) error {
	m := new(P4ExecRequest)
	if err := strebm.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GitserverServiceServer).P4Exec(m, &gitserverServiceP4ExecServer{strebm})
}

type GitserverService_P4ExecServer interfbce {
	Send(*P4ExecResponse) error
	grpc.ServerStrebm
}

type gitserverServiceP4ExecServer struct {
	grpc.ServerStrebm
}

func (x *gitserverServiceP4ExecServer) Send(m *P4ExecResponse) error {
	return x.ServerStrebm.SendMsg(m)
}

func _GitserverService_RepoClone_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(RepoCloneRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).RepoClone(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_RepoClone_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(GitserverServiceServer).RepoClone(ctx, req.(*RepoCloneRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _GitserverService_RepoCloneProgress_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(RepoCloneProgressRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).RepoCloneProgress(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_RepoCloneProgress_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(GitserverServiceServer).RepoCloneProgress(ctx, req.(*RepoCloneProgressRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _GitserverService_RepoDelete_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(RepoDeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).RepoDelete(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_RepoDelete_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(GitserverServiceServer).RepoDelete(ctx, req.(*RepoDeleteRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _GitserverService_RepoUpdbte_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(RepoUpdbteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).RepoUpdbte(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_RepoUpdbte_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(GitserverServiceServer).RepoUpdbte(ctx, req.(*RepoUpdbteRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

func _GitserverService_ReposStbts_Hbndler(srv interfbce{}, ctx context.Context, dec func(interfbce{}) error, interceptor grpc.UnbryServerInterceptor) (interfbce{}, error) {
	in := new(ReposStbtsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GitserverServiceServer).ReposStbts(ctx, in)
	}
	info := &grpc.UnbryServerInfo{
		Server:     srv,
		FullMethod: GitserverService_ReposStbts_FullMethodNbme,
	}
	hbndler := func(ctx context.Context, req interfbce{}) (interfbce{}, error) {
		return srv.(GitserverServiceServer).ReposStbts(ctx, req.(*ReposStbtsRequest))
	}
	return interceptor(ctx, in, info, hbndler)
}

// GitserverService_ServiceDesc is the grpc.ServiceDesc for GitserverService service.
// It's only intended for direct use with grpc.RegisterService,
// bnd not to be introspected or modified (even bs b copy)
vbr GitserverService_ServiceDesc = grpc.ServiceDesc{
	ServiceNbme: "gitserver.v1.GitserverService",
	HbndlerType: (*GitserverServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodNbme: "BbtchLog",
			Hbndler:    _GitserverService_BbtchLog_Hbndler,
		},
		{
			MethodNbme: "DiskInfo",
			Hbndler:    _GitserverService_DiskInfo_Hbndler,
		},
		{
			MethodNbme: "GetObject",
			Hbndler:    _GitserverService_GetObject_Hbndler,
		},
		{
			MethodNbme: "IsRepoClonebble",
			Hbndler:    _GitserverService_IsRepoClonebble_Hbndler,
		},
		{
			MethodNbme: "ListGitolite",
			Hbndler:    _GitserverService_ListGitolite_Hbndler,
		},
		{
			MethodNbme: "RepoClone",
			Hbndler:    _GitserverService_RepoClone_Hbndler,
		},
		{
			MethodNbme: "RepoCloneProgress",
			Hbndler:    _GitserverService_RepoCloneProgress_Hbndler,
		},
		{
			MethodNbme: "RepoDelete",
			Hbndler:    _GitserverService_RepoDelete_Hbndler,
		},
		{
			MethodNbme: "RepoUpdbte",
			Hbndler:    _GitserverService_RepoUpdbte_Hbndler,
		},
		{
			MethodNbme: "ReposStbts",
			Hbndler:    _GitserverService_ReposStbts_Hbndler,
		},
	},
	Strebms: []grpc.StrebmDesc{
		{
			StrebmNbme:    "CrebteCommitFromPbtchBinbry",
			Hbndler:       _GitserverService_CrebteCommitFromPbtchBinbry_Hbndler,
			ClientStrebms: true,
		},
		{
			StrebmNbme:    "Exec",
			Hbndler:       _GitserverService_Exec_Hbndler,
			ServerStrebms: true,
		},
		{
			StrebmNbme:    "Sebrch",
			Hbndler:       _GitserverService_Sebrch_Hbndler,
			ServerStrebms: true,
		},
		{
			StrebmNbme:    "Archive",
			Hbndler:       _GitserverService_Archive_Hbndler,
			ServerStrebms: true,
		},
		{
			StrebmNbme:    "P4Exec",
			Hbndler:       _GitserverService_P4Exec_Hbndler,
			ServerStrebms: true,
		},
	},
	Metbdbtb: "gitserver.proto",
}
