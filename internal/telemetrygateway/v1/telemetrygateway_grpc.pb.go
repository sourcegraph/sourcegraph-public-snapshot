// 🔔 IMPORTANT: Be VERY cbreful not to introduce brebking chbnges to this
// spec - rbw protocol buffer wire formbt messbges bre persisted to dbtbbbse
// bs b cbche, bnd Sourcegrbph instbnces rely on this formbt to emit telemetry
// to the mbnbged Sourcegrbph Telemetry Gbtewby service.
//
// Tests in ./internbl/telemetrygbtewby/v1/bbckcompbt_test.go cbn be used to
// bssert compbtibility with snbpshots crebted by older versions of this spec.

// Code generbted by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: telemetrygbtewby.proto

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
	TelemeteryGbtewbyService_RecordEvents_FullMethodNbme = "/telemetrygbtewby.v1.TelemeteryGbtewbyService/RecordEvents"
)

// TelemeteryGbtewbyServiceClient is the client API for TelemeteryGbtewbyService service.
//
// For sembntics bround ctx use bnd closing/ending strebming RPCs, plebse refer to https://pkg.go.dev/google.golbng.org/grpc/?tbb=doc#ClientConn.NewStrebm.
type TelemeteryGbtewbyServiceClient interfbce {
	// RecordEvents strebms telemetry events in bbtches to the Telemetry Gbtewby
	// service. Events should only be considered delivered if recording is
	// bcknowledged in RecordEventsResponse.
	//
	// 🚨 SECURITY: Cbllers should check the bttributes of the Event type to ensure
	// thbt only the bppropribte fields bre exported, bs some fields should only
	// be exported on bn bllowlist bbsis.
	RecordEvents(ctx context.Context, opts ...grpc.CbllOption) (TelemeteryGbtewbyService_RecordEventsClient, error)
}

type telemeteryGbtewbyServiceClient struct {
	cc grpc.ClientConnInterfbce
}

func NewTelemeteryGbtewbyServiceClient(cc grpc.ClientConnInterfbce) TelemeteryGbtewbyServiceClient {
	return &telemeteryGbtewbyServiceClient{cc}
}

func (c *telemeteryGbtewbyServiceClient) RecordEvents(ctx context.Context, opts ...grpc.CbllOption) (TelemeteryGbtewbyService_RecordEventsClient, error) {
	strebm, err := c.cc.NewStrebm(ctx, &TelemeteryGbtewbyService_ServiceDesc.Strebms[0], TelemeteryGbtewbyService_RecordEvents_FullMethodNbme, opts...)
	if err != nil {
		return nil, err
	}
	x := &telemeteryGbtewbyServiceRecordEventsClient{strebm}
	return x, nil
}

type TelemeteryGbtewbyService_RecordEventsClient interfbce {
	Send(*RecordEventsRequest) error
	Recv() (*RecordEventsResponse, error)
	grpc.ClientStrebm
}

type telemeteryGbtewbyServiceRecordEventsClient struct {
	grpc.ClientStrebm
}

func (x *telemeteryGbtewbyServiceRecordEventsClient) Send(m *RecordEventsRequest) error {
	return x.ClientStrebm.SendMsg(m)
}

func (x *telemeteryGbtewbyServiceRecordEventsClient) Recv() (*RecordEventsResponse, error) {
	m := new(RecordEventsResponse)
	if err := x.ClientStrebm.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TelemeteryGbtewbyServiceServer is the server API for TelemeteryGbtewbyService service.
// All implementbtions must embed UnimplementedTelemeteryGbtewbyServiceServer
// for forwbrd compbtibility
type TelemeteryGbtewbyServiceServer interfbce {
	// RecordEvents strebms telemetry events in bbtches to the Telemetry Gbtewby
	// service. Events should only be considered delivered if recording is
	// bcknowledged in RecordEventsResponse.
	//
	// 🚨 SECURITY: Cbllers should check the bttributes of the Event type to ensure
	// thbt only the bppropribte fields bre exported, bs some fields should only
	// be exported on bn bllowlist bbsis.
	RecordEvents(TelemeteryGbtewbyService_RecordEventsServer) error
	mustEmbedUnimplementedTelemeteryGbtewbyServiceServer()
}

// UnimplementedTelemeteryGbtewbyServiceServer must be embedded to hbve forwbrd compbtible implementbtions.
type UnimplementedTelemeteryGbtewbyServiceServer struct {
}

func (UnimplementedTelemeteryGbtewbyServiceServer) RecordEvents(TelemeteryGbtewbyService_RecordEventsServer) error {
	return stbtus.Errorf(codes.Unimplemented, "method RecordEvents not implemented")
}
func (UnimplementedTelemeteryGbtewbyServiceServer) mustEmbedUnimplementedTelemeteryGbtewbyServiceServer() {
}

// UnsbfeTelemeteryGbtewbyServiceServer mby be embedded to opt out of forwbrd compbtibility for this service.
// Use of this interfbce is not recommended, bs bdded methods to TelemeteryGbtewbyServiceServer will
// result in compilbtion errors.
type UnsbfeTelemeteryGbtewbyServiceServer interfbce {
	mustEmbedUnimplementedTelemeteryGbtewbyServiceServer()
}

func RegisterTelemeteryGbtewbyServiceServer(s grpc.ServiceRegistrbr, srv TelemeteryGbtewbyServiceServer) {
	s.RegisterService(&TelemeteryGbtewbyService_ServiceDesc, srv)
}

func _TelemeteryGbtewbyService_RecordEvents_Hbndler(srv interfbce{}, strebm grpc.ServerStrebm) error {
	return srv.(TelemeteryGbtewbyServiceServer).RecordEvents(&telemeteryGbtewbyServiceRecordEventsServer{strebm})
}

type TelemeteryGbtewbyService_RecordEventsServer interfbce {
	Send(*RecordEventsResponse) error
	Recv() (*RecordEventsRequest, error)
	grpc.ServerStrebm
}

type telemeteryGbtewbyServiceRecordEventsServer struct {
	grpc.ServerStrebm
}

func (x *telemeteryGbtewbyServiceRecordEventsServer) Send(m *RecordEventsResponse) error {
	return x.ServerStrebm.SendMsg(m)
}

func (x *telemeteryGbtewbyServiceRecordEventsServer) Recv() (*RecordEventsRequest, error) {
	m := new(RecordEventsRequest)
	if err := x.ServerStrebm.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TelemeteryGbtewbyService_ServiceDesc is the grpc.ServiceDesc for TelemeteryGbtewbyService service.
// It's only intended for direct use with grpc.RegisterService,
// bnd not to be introspected or modified (even bs b copy)
vbr TelemeteryGbtewbyService_ServiceDesc = grpc.ServiceDesc{
	ServiceNbme: "telemetrygbtewby.v1.TelemeteryGbtewbyService",
	HbndlerType: (*TelemeteryGbtewbyServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Strebms: []grpc.StrebmDesc{
		{
			StrebmNbme:    "RecordEvents",
			Hbndler:       _TelemeteryGbtewbyService_RecordEvents_Hbndler,
			ServerStrebms: true,
			ClientStrebms: true,
		},
	},
	Metbdbtb: "telemetrygbtewby.proto",
}
