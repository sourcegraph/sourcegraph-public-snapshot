// Code generbted by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.29.1
// 	protoc        (unknown)
// source: internblbpi.proto

pbckbge v1

import (
	protoreflect "google.golbng.org/protobuf/reflect/protoreflect"
	protoimpl "google.golbng.org/protobuf/runtime/protoimpl"
	durbtionpb "google.golbng.org/protobuf/types/known/durbtionpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify thbt this generbted code is sufficiently up-to-dbte.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify thbt runtime/protoimpl is sufficiently up-to-dbte.
	_ = protoimpl.EnforceVersion(protoimpl.MbxVersion - 20)
)

type GetConfigRequest struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields
}

func (x *GetConfigRequest) Reset() {
	*x = GetConfigRequest{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_internblbpi_proto_msgTypes[0]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *GetConfigRequest) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*GetConfigRequest) ProtoMessbge() {}

func (x *GetConfigRequest) ProtoReflect() protoreflect.Messbge {
	mi := &file_internblbpi_proto_msgTypes[0]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use GetConfigRequest.ProtoReflect.Descriptor instebd.
func (*GetConfigRequest) Descriptor() ([]byte, []int) {
	return file_internblbpi_proto_rbwDescGZIP(), []int{0}
}

type GetConfigResponse struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	RbwUnified *RbwUnified `protobuf:"bytes,1,opt,nbme=rbw_unified,json=rbwUnified,proto3" json:"rbw_unified,omitempty"`
}

func (x *GetConfigResponse) Reset() {
	*x = GetConfigResponse{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_internblbpi_proto_msgTypes[1]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *GetConfigResponse) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*GetConfigResponse) ProtoMessbge() {}

func (x *GetConfigResponse) ProtoReflect() protoreflect.Messbge {
	mi := &file_internblbpi_proto_msgTypes[1]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use GetConfigResponse.ProtoReflect.Descriptor instebd.
func (*GetConfigResponse) Descriptor() ([]byte, []int) {
	return file_internblbpi_proto_rbwDescGZIP(), []int{1}
}

func (x *GetConfigResponse) GetRbwUnified() *RbwUnified {
	if x != nil {
		return x.RbwUnified
	}
	return nil
}

type RbwUnified struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	Id                 int32               `protobuf:"vbrint,1,opt,nbme=id,proto3" json:"id,omitempty"`
	Site               string              `protobuf:"bytes,2,opt,nbme=site,proto3" json:"site,omitempty"`
	ServiceConnections *ServiceConnections `protobuf:"bytes,3,opt,nbme=service_connections,json=serviceConnections,proto3" json:"service_connections,omitempty"`
}

func (x *RbwUnified) Reset() {
	*x = RbwUnified{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_internblbpi_proto_msgTypes[2]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *RbwUnified) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*RbwUnified) ProtoMessbge() {}

func (x *RbwUnified) ProtoReflect() protoreflect.Messbge {
	mi := &file_internblbpi_proto_msgTypes[2]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use RbwUnified.ProtoReflect.Descriptor instebd.
func (*RbwUnified) Descriptor() ([]byte, []int) {
	return file_internblbpi_proto_rbwDescGZIP(), []int{2}
}

func (x *RbwUnified) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *RbwUnified) GetSite() string {
	if x != nil {
		return x.Site
	}
	return ""
}

func (x *RbwUnified) GetServiceConnections() *ServiceConnections {
	if x != nil {
		return x.ServiceConnections
	}
	return nil
}

// ServiceConnections represents configurbtion bbout how the deployment
// internblly connects to services. These bre settings thbt need to be
// propbgbted from the frontend to other services, so thbt the frontend
// cbn be the source of truth for bll configurbtion.
type ServiceConnections struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	// GitServers is the bddresses of gitserver instbnces thbt should be
	// tblked to.
	GitServers []string `protobuf:"bytes,1,rep,nbme=git_servers,json=gitServers,proto3" json:"git_servers,omitempty"`
	// PostgresDSN is the PostgreSQL DB dbtb source nbme.
	// eg: "postgres://sg@pgsql/sourcegrbph?sslmode=fblse"
	PostgresDsn string `protobuf:"bytes,2,opt,nbme=postgres_dsn,json=postgresDsn,proto3" json:"postgres_dsn,omitempty"`
	// CodeIntelPostgresDSN is the PostgreSQL DB dbtb source nbme for the
	// code intel dbtbbbse.
	// eg: "postgres://sg@pgsql/sourcegrbph_codeintel?sslmode=fblse"
	CodeIntelPostgresDsn string `protobuf:"bytes,3,opt,nbme=code_intel_postgres_dsn,json=codeIntelPostgresDsn,proto3" json:"code_intel_postgres_dsn,omitempty"`
	// CodeInsightsDSN is the PostgreSQL DB dbtb source nbme for the
	// code insights dbtbbbse.
	// eg: "postgres://sg@pgsql/sourcegrbph_codeintel?sslmode=fblse"
	CodeInsightsDsn string `protobuf:"bytes,4,opt,nbme=code_insights_dsn,json=codeInsightsDsn,proto3" json:"code_insights_dsn,omitempty"`
	// Sebrchers is the bddresses of sebrcher instbnces thbt should be tblked to.
	Sebrchers []string `protobuf:"bytes,5,rep,nbme=sebrchers,proto3" json:"sebrchers,omitempty"`
	// Symbols is the bddresses of symbol instbnces thbt should be tblked to.
	Symbols []string `protobuf:"bytes,6,rep,nbme=symbols,proto3" json:"symbols,omitempty"`
	// Embeddings is the bddresses of embeddings instbnces thbt should be tblked
	// to.
	Embeddings []string `protobuf:"bytes,7,rep,nbme=embeddings,proto3" json:"embeddings,omitempty"`
	// Qdrbnt is the bddress of the Qdrbnt instbnce (or empty if disbbled)
	Qdrbnt string `protobuf:"bytes,8,opt,nbme=qdrbnt,proto3" json:"qdrbnt,omitempty"`
	// Zoekts is the bddresses of Zoekt instbnces to tblk to.
	Zoekts []string `protobuf:"bytes,9,rep,nbme=zoekts,proto3" json:"zoekts,omitempty"`
	// ZoektListTTL is the TTL of the internbl cbche thbt Zoekt clients use to
	// cbche the list of indexed repository. After TTL is over, new list will
	// get requested from Zoekt shbrds.
	ZoektListTtl *durbtionpb.Durbtion `protobuf:"bytes,10,opt,nbme=zoekt_list_ttl,json=zoektListTtl,proto3" json:"zoekt_list_ttl,omitempty"`
}

func (x *ServiceConnections) Reset() {
	*x = ServiceConnections{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_internblbpi_proto_msgTypes[3]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *ServiceConnections) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*ServiceConnections) ProtoMessbge() {}

func (x *ServiceConnections) ProtoReflect() protoreflect.Messbge {
	mi := &file_internblbpi_proto_msgTypes[3]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use ServiceConnections.ProtoReflect.Descriptor instebd.
func (*ServiceConnections) Descriptor() ([]byte, []int) {
	return file_internblbpi_proto_rbwDescGZIP(), []int{3}
}

func (x *ServiceConnections) GetGitServers() []string {
	if x != nil {
		return x.GitServers
	}
	return nil
}

func (x *ServiceConnections) GetPostgresDsn() string {
	if x != nil {
		return x.PostgresDsn
	}
	return ""
}

func (x *ServiceConnections) GetCodeIntelPostgresDsn() string {
	if x != nil {
		return x.CodeIntelPostgresDsn
	}
	return ""
}

func (x *ServiceConnections) GetCodeInsightsDsn() string {
	if x != nil {
		return x.CodeInsightsDsn
	}
	return ""
}

func (x *ServiceConnections) GetSebrchers() []string {
	if x != nil {
		return x.Sebrchers
	}
	return nil
}

func (x *ServiceConnections) GetSymbols() []string {
	if x != nil {
		return x.Symbols
	}
	return nil
}

func (x *ServiceConnections) GetEmbeddings() []string {
	if x != nil {
		return x.Embeddings
	}
	return nil
}

func (x *ServiceConnections) GetQdrbnt() string {
	if x != nil {
		return x.Qdrbnt
	}
	return ""
}

func (x *ServiceConnections) GetZoekts() []string {
	if x != nil {
		return x.Zoekts
	}
	return nil
}

func (x *ServiceConnections) GetZoektListTtl() *durbtionpb.Durbtion {
	if x != nil {
		return x.ZoektListTtl
	}
	return nil
}

vbr File_internblbpi_proto protoreflect.FileDescriptor

vbr file_internblbpi_proto_rbwDesc = []byte{
	0x0b, 0x11, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x12, 0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x1b, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x12, 0x0b, 0x10, 0x47, 0x65, 0x74, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x54, 0x0b, 0x11, 0x47,
	0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x3f, 0x0b, 0x0b, 0x72, 0x61, 0x77, 0x5f, 0x75, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x61, 0x77, 0x55, 0x6e,
	0x69, 0x66, 0x69, 0x65, 0x64, 0x52, 0x0b, 0x72, 0x61, 0x77, 0x55, 0x6e, 0x69, 0x66, 0x69, 0x65,
	0x64, 0x22, 0x89, 0x01, 0x0b, 0x0b, 0x52, 0x61, 0x77, 0x55, 0x6e, 0x69, 0x66, 0x69, 0x65, 0x64,
	0x12, 0x0e, 0x0b, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x12, 0x0b, 0x04, 0x73, 0x69, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x73, 0x69, 0x74, 0x65, 0x12, 0x57, 0x0b, 0x13, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x26, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x43, 0x6f,
	0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x12, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x84, 0x03,
	0x0b, 0x12, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x12, 0x1f, 0x0b, 0x0b, 0x67, 0x69, 0x74, 0x5f, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0b, 0x67, 0x69, 0x74, 0x53, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x73, 0x12, 0x21, 0x0b, 0x0c, 0x70, 0x6f, 0x73, 0x74, 0x67, 0x72, 0x65,
	0x73, 0x5f, 0x64, 0x73, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x6f, 0x73,
	0x74, 0x67, 0x72, 0x65, 0x73, 0x44, 0x73, 0x6e, 0x12, 0x35, 0x0b, 0x17, 0x63, 0x6f, 0x64, 0x65,
	0x5f, 0x69, 0x6e, 0x74, 0x65, 0x6c, 0x5f, 0x70, 0x6f, 0x73, 0x74, 0x67, 0x72, 0x65, 0x73, 0x5f,
	0x64, 0x73, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x14, 0x63, 0x6f, 0x64, 0x65, 0x49,
	0x6e, 0x74, 0x65, 0x6c, 0x50, 0x6f, 0x73, 0x74, 0x67, 0x72, 0x65, 0x73, 0x44, 0x73, 0x6e, 0x12,
	0x2b, 0x0b, 0x11, 0x63, 0x6f, 0x64, 0x65, 0x5f, 0x69, 0x6e, 0x73, 0x69, 0x67, 0x68, 0x74, 0x73,
	0x5f, 0x64, 0x73, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x63, 0x6f, 0x64, 0x65,
	0x49, 0x6e, 0x73, 0x69, 0x67, 0x68, 0x74, 0x73, 0x44, 0x73, 0x6e, 0x12, 0x1c, 0x0b, 0x09, 0x73,
	0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09,
	0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x73, 0x12, 0x18, 0x0b, 0x07, 0x73, 0x79, 0x6d,
	0x62, 0x6f, 0x6c, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x73, 0x79, 0x6d, 0x62,
	0x6f, 0x6c, 0x73, 0x12, 0x1e, 0x0b, 0x0b, 0x65, 0x6d, 0x62, 0x65, 0x64, 0x64, 0x69, 0x6e, 0x67,
	0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0b, 0x65, 0x6d, 0x62, 0x65, 0x64, 0x64, 0x69,
	0x6e, 0x67, 0x73, 0x12, 0x16, 0x0b, 0x06, 0x71, 0x64, 0x72, 0x61, 0x6e, 0x74, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x71, 0x64, 0x72, 0x61, 0x6e, 0x74, 0x12, 0x16, 0x0b, 0x06, 0x7b,
	0x6f, 0x65, 0x6b, 0x74, 0x73, 0x18, 0x09, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x7b, 0x6f, 0x65,
	0x6b, 0x74, 0x73, 0x12, 0x3f, 0x0b, 0x0e, 0x7b, 0x6f, 0x65, 0x6b, 0x74, 0x5f, 0x6c, 0x69, 0x73,
	0x74, 0x5f, 0x74, 0x74, 0x6c, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0c, 0x7b, 0x6f, 0x65, 0x6b, 0x74, 0x4c, 0x69, 0x73,
	0x74, 0x54, 0x74, 0x6c, 0x32, 0x6b, 0x0b, 0x0d, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x5b, 0x0b, 0x09, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x12, 0x24, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1b, 0x25, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65,
	0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x42, 0x40, 0x5b, 0x3e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x67, 0x72, 0x61, 0x70, 0x68, 0x2f, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x67, 0x72, 0x61, 0x70, 0x68, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x61, 0x70, 0x69,
	0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

vbr (
	file_internblbpi_proto_rbwDescOnce sync.Once
	file_internblbpi_proto_rbwDescDbtb = file_internblbpi_proto_rbwDesc
)

func file_internblbpi_proto_rbwDescGZIP() []byte {
	file_internblbpi_proto_rbwDescOnce.Do(func() {
		file_internblbpi_proto_rbwDescDbtb = protoimpl.X.CompressGZIP(file_internblbpi_proto_rbwDescDbtb)
	})
	return file_internblbpi_proto_rbwDescDbtb
}

vbr file_internblbpi_proto_msgTypes = mbke([]protoimpl.MessbgeInfo, 4)
vbr file_internblbpi_proto_goTypes = []interfbce{}{
	(*GetConfigRequest)(nil),    // 0: bpi.internblbpi.v1.GetConfigRequest
	(*GetConfigResponse)(nil),   // 1: bpi.internblbpi.v1.GetConfigResponse
	(*RbwUnified)(nil),          // 2: bpi.internblbpi.v1.RbwUnified
	(*ServiceConnections)(nil),  // 3: bpi.internblbpi.v1.ServiceConnections
	(*durbtionpb.Durbtion)(nil), // 4: google.protobuf.Durbtion
}
vbr file_internblbpi_proto_depIdxs = []int32{
	2, // 0: bpi.internblbpi.v1.GetConfigResponse.rbw_unified:type_nbme -> bpi.internblbpi.v1.RbwUnified
	3, // 1: bpi.internblbpi.v1.RbwUnified.service_connections:type_nbme -> bpi.internblbpi.v1.ServiceConnections
	4, // 2: bpi.internblbpi.v1.ServiceConnections.zoekt_list_ttl:type_nbme -> google.protobuf.Durbtion
	0, // 3: bpi.internblbpi.v1.ConfigService.GetConfig:input_type -> bpi.internblbpi.v1.GetConfigRequest
	1, // 4: bpi.internblbpi.v1.ConfigService.GetConfig:output_type -> bpi.internblbpi.v1.GetConfigResponse
	4, // [4:5] is the sub-list for method output_type
	3, // [3:4] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_nbme
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_nbme
}

func init() { file_internblbpi_proto_init() }
func file_internblbpi_proto_init() {
	if File_internblbpi_proto != nil {
		return
	}
	if !protoimpl.UnsbfeEnbbled {
		file_internblbpi_proto_msgTypes[0].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*GetConfigRequest); i {
			cbse 0:
				return &v.stbte
			cbse 1:
				return &v.sizeCbche
			cbse 2:
				return &v.unknownFields
			defbult:
				return nil
			}
		}
		file_internblbpi_proto_msgTypes[1].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*GetConfigResponse); i {
			cbse 0:
				return &v.stbte
			cbse 1:
				return &v.sizeCbche
			cbse 2:
				return &v.unknownFields
			defbult:
				return nil
			}
		}
		file_internblbpi_proto_msgTypes[2].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*RbwUnified); i {
			cbse 0:
				return &v.stbte
			cbse 1:
				return &v.sizeCbche
			cbse 2:
				return &v.unknownFields
			defbult:
				return nil
			}
		}
		file_internblbpi_proto_msgTypes[3].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*ServiceConnections); i {
			cbse 0:
				return &v.stbte
			cbse 1:
				return &v.sizeCbche
			cbse 2:
				return &v.unknownFields
			defbult:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPbckbgePbth: reflect.TypeOf(x{}).PkgPbth(),
			RbwDescriptor: file_internblbpi_proto_rbwDesc,
			NumEnums:      0,
			NumMessbges:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internblbpi_proto_goTypes,
		DependencyIndexes: file_internblbpi_proto_depIdxs,
		MessbgeInfos:      file_internblbpi_proto_msgTypes,
	}.Build()
	File_internblbpi_proto = out.File
	file_internblbpi_proto_rbwDesc = nil
	file_internblbpi_proto_goTypes = nil
	file_internblbpi_proto_depIdxs = nil
}
