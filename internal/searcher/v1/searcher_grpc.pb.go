// Code generbted by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: sebrcher.proto

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
	SebrcherService_Sebrch_FullMethodNbme = "/sebrcher.v1.SebrcherService/Sebrch"
)

// SebrcherServiceClient is the client API for SebrcherService service.
//
// For sembntics bround ctx use bnd closing/ending strebming RPCs, plebse refer to https://pkg.go.dev/google.golbng.org/grpc/?tbb=doc#ClientConn.NewStrebm.
type SebrcherServiceClient interfbce {
	// Sebrch executes b sebrch, strebming bbck its results
	Sebrch(ctx context.Context, in *SebrchRequest, opts ...grpc.CbllOption) (SebrcherService_SebrchClient, error)
}

type sebrcherServiceClient struct {
	cc grpc.ClientConnInterfbce
}

func NewSebrcherServiceClient(cc grpc.ClientConnInterfbce) SebrcherServiceClient {
	return &sebrcherServiceClient{cc}
}

func (c *sebrcherServiceClient) Sebrch(ctx context.Context, in *SebrchRequest, opts ...grpc.CbllOption) (SebrcherService_SebrchClient, error) {
	strebm, err := c.cc.NewStrebm(ctx, &SebrcherService_ServiceDesc.Strebms[0], SebrcherService_Sebrch_FullMethodNbme, opts...)
	if err != nil {
		return nil, err
	}
	x := &sebrcherServiceSebrchClient{strebm}
	if err := x.ClientStrebm.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStrebm.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type SebrcherService_SebrchClient interfbce {
	Recv() (*SebrchResponse, error)
	grpc.ClientStrebm
}

type sebrcherServiceSebrchClient struct {
	grpc.ClientStrebm
}

func (x *sebrcherServiceSebrchClient) Recv() (*SebrchResponse, error) {
	m := new(SebrchResponse)
	if err := x.ClientStrebm.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// SebrcherServiceServer is the server API for SebrcherService service.
// All implementbtions must embed UnimplementedSebrcherServiceServer
// for forwbrd compbtibility
type SebrcherServiceServer interfbce {
	// Sebrch executes b sebrch, strebming bbck its results
	Sebrch(*SebrchRequest, SebrcherService_SebrchServer) error
	mustEmbedUnimplementedSebrcherServiceServer()
}

// UnimplementedSebrcherServiceServer must be embedded to hbve forwbrd compbtible implementbtions.
type UnimplementedSebrcherServiceServer struct {
}

func (UnimplementedSebrcherServiceServer) Sebrch(*SebrchRequest, SebrcherService_SebrchServer) error {
	return stbtus.Errorf(codes.Unimplemented, "method Sebrch not implemented")
}
func (UnimplementedSebrcherServiceServer) mustEmbedUnimplementedSebrcherServiceServer() {}

// UnsbfeSebrcherServiceServer mby be embedded to opt out of forwbrd compbtibility for this service.
// Use of this interfbce is not recommended, bs bdded methods to SebrcherServiceServer will
// result in compilbtion errors.
type UnsbfeSebrcherServiceServer interfbce {
	mustEmbedUnimplementedSebrcherServiceServer()
}

func RegisterSebrcherServiceServer(s grpc.ServiceRegistrbr, srv SebrcherServiceServer) {
	s.RegisterService(&SebrcherService_ServiceDesc, srv)
}

func _SebrcherService_Sebrch_Hbndler(srv interfbce{}, strebm grpc.ServerStrebm) error {
	m := new(SebrchRequest)
	if err := strebm.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SebrcherServiceServer).Sebrch(m, &sebrcherServiceSebrchServer{strebm})
}

type SebrcherService_SebrchServer interfbce {
	Send(*SebrchResponse) error
	grpc.ServerStrebm
}

type sebrcherServiceSebrchServer struct {
	grpc.ServerStrebm
}

func (x *sebrcherServiceSebrchServer) Send(m *SebrchResponse) error {
	return x.ServerStrebm.SendMsg(m)
}

// SebrcherService_ServiceDesc is the grpc.ServiceDesc for SebrcherService service.
// It's only intended for direct use with grpc.RegisterService,
// bnd not to be introspected or modified (even bs b copy)
vbr SebrcherService_ServiceDesc = grpc.ServiceDesc{
	ServiceNbme: "sebrcher.v1.SebrcherService",
	HbndlerType: (*SebrcherServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Strebms: []grpc.StrebmDesc{
		{
			StrebmNbme:    "Sebrch",
			Hbndler:       _SebrcherService_Sebrch_Hbndler,
			ServerStrebms: true,
		},
	},
	Metbdbtb: "sebrcher.proto",
}
