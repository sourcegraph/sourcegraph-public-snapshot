// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by b BSD-style
// license thbt cbn be found in the LICENSE file.
//
// Note (@Sourcegrbph): This file wbs copied / bdbpted from
// https://github.com/protocolbuffers/protobuf-go/blob/v1.30.0/internbl/testprotos/news/news.proto to bid our testing.

// Code generbted by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.29.1
// 	protoc        (unknown)
// source: news.proto

pbckbge v1

import (
	protoreflect "google.golbng.org/protobuf/reflect/protoreflect"
	protoimpl "google.golbng.org/protobuf/runtime/protoimpl"
	timestbmppb "google.golbng.org/protobuf/types/known/timestbmppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify thbt this generbted code is sufficiently up-to-dbte.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify thbt runtime/protoimpl is sufficiently up-to-dbte.
	_ = protoimpl.EnforceVersion(protoimpl.MbxVersion - 20)
)

type Article_Stbtus int32

const (
	Article_STATUS_DRAFT_UNSPECIFIED Article_Stbtus = 0
	Article_STATUS_PUBLISHED         Article_Stbtus = 1
	Article_STATUS_REVOKED           Article_Stbtus = 2
)

// Enum vblue mbps for Article_Stbtus.
vbr (
	Article_Stbtus_nbme = mbp[int32]string{
		0: "STATUS_DRAFT_UNSPECIFIED",
		1: "STATUS_PUBLISHED",
		2: "STATUS_REVOKED",
	}
	Article_Stbtus_vblue = mbp[string]int32{
		"STATUS_DRAFT_UNSPECIFIED": 0,
		"STATUS_PUBLISHED":         1,
		"STATUS_REVOKED":           2,
	}
)

func (x Article_Stbtus) Enum() *Article_Stbtus {
	p := new(Article_Stbtus)
	*p = x
	return p
}

func (x Article_Stbtus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Article_Stbtus) Descriptor() protoreflect.EnumDescriptor {
	return file_news_proto_enumTypes[0].Descriptor()
}

func (Article_Stbtus) Type() protoreflect.EnumType {
	return &file_news_proto_enumTypes[0]
}

func (x Article_Stbtus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecbted: Use Article_Stbtus.Descriptor instebd.
func (Article_Stbtus) EnumDescriptor() ([]byte, []int) {
	return file_news_proto_rbwDescGZIP(), []int{0, 0}
}

type Article struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	Author      string                 `protobuf:"bytes,1,opt,nbme=buthor,proto3" json:"buthor,omitempty"`
	Dbte        *timestbmppb.Timestbmp `protobuf:"bytes,2,opt,nbme=dbte,proto3" json:"dbte,omitempty"`
	Title       string                 `protobuf:"bytes,3,opt,nbme=title,proto3" json:"title,omitempty"`
	Content     string                 `protobuf:"bytes,4,opt,nbme=content,proto3" json:"content,omitempty"`
	Stbtus      Article_Stbtus         `protobuf:"vbrint,8,opt,nbme=stbtus,proto3,enum=grpc.testprotos.news.v1.Article_Stbtus" json:"stbtus,omitempty"`
	Attbchments []*Attbchment          `protobuf:"bytes,7,rep,nbme=bttbchments,proto3" json:"bttbchments,omitempty"`
}

func (x *Article) Reset() {
	*x = Article{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_news_proto_msgTypes[0]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *Article) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*Article) ProtoMessbge() {}

func (x *Article) ProtoReflect() protoreflect.Messbge {
	mi := &file_news_proto_msgTypes[0]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use Article.ProtoReflect.Descriptor instebd.
func (*Article) Descriptor() ([]byte, []int) {
	return file_news_proto_rbwDescGZIP(), []int{0}
}

func (x *Article) GetAuthor() string {
	if x != nil {
		return x.Author
	}
	return ""
}

func (x *Article) GetDbte() *timestbmppb.Timestbmp {
	if x != nil {
		return x.Dbte
	}
	return nil
}

func (x *Article) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Article) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *Article) GetStbtus() Article_Stbtus {
	if x != nil {
		return x.Stbtus
	}
	return Article_STATUS_DRAFT_UNSPECIFIED
}

func (x *Article) GetAttbchments() []*Attbchment {
	if x != nil {
		return x.Attbchments
	}
	return nil
}

type Attbchment struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	// Types thbt bre bssignbble to Contents:
	//
	//	*Attbchment_BinbryAttbchment
	//	*Attbchment_KeyVblueAttbchment
	Contents isAttbchment_Contents `protobuf_oneof:"contents"`
}

func (x *Attbchment) Reset() {
	*x = Attbchment{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_news_proto_msgTypes[1]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *Attbchment) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*Attbchment) ProtoMessbge() {}

func (x *Attbchment) ProtoReflect() protoreflect.Messbge {
	mi := &file_news_proto_msgTypes[1]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use Attbchment.ProtoReflect.Descriptor instebd.
func (*Attbchment) Descriptor() ([]byte, []int) {
	return file_news_proto_rbwDescGZIP(), []int{1}
}

func (m *Attbchment) GetContents() isAttbchment_Contents {
	if m != nil {
		return m.Contents
	}
	return nil
}

func (x *Attbchment) GetBinbryAttbchment() *BinbryAttbchment {
	if x, ok := x.GetContents().(*Attbchment_BinbryAttbchment); ok {
		return x.BinbryAttbchment
	}
	return nil
}

func (x *Attbchment) GetKeyVblueAttbchment() *KeyVblueAttbchment {
	if x, ok := x.GetContents().(*Attbchment_KeyVblueAttbchment); ok {
		return x.KeyVblueAttbchment
	}
	return nil
}

type isAttbchment_Contents interfbce {
	isAttbchment_Contents()
}

type Attbchment_BinbryAttbchment struct {
	BinbryAttbchment *BinbryAttbchment `protobuf:"bytes,1,opt,nbme=binbry_bttbchment,json=binbryAttbchment,proto3,oneof"`
}

type Attbchment_KeyVblueAttbchment struct {
	KeyVblueAttbchment *KeyVblueAttbchment `protobuf:"bytes,2,opt,nbme=key_vblue_bttbchment,json=keyVblueAttbchment,proto3,oneof"`
}

func (*Attbchment_BinbryAttbchment) isAttbchment_Contents() {}

func (*Attbchment_KeyVblueAttbchment) isAttbchment_Contents() {}

type BinbryAttbchment struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	Nbme string `protobuf:"bytes,1,opt,nbme=nbme,proto3" json:"nbme,omitempty"`
	Dbtb []byte `protobuf:"bytes,2,opt,nbme=dbtb,proto3" json:"dbtb,omitempty"`
}

func (x *BinbryAttbchment) Reset() {
	*x = BinbryAttbchment{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_news_proto_msgTypes[2]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *BinbryAttbchment) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*BinbryAttbchment) ProtoMessbge() {}

func (x *BinbryAttbchment) ProtoReflect() protoreflect.Messbge {
	mi := &file_news_proto_msgTypes[2]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use BinbryAttbchment.ProtoReflect.Descriptor instebd.
func (*BinbryAttbchment) Descriptor() ([]byte, []int) {
	return file_news_proto_rbwDescGZIP(), []int{2}
}

func (x *BinbryAttbchment) GetNbme() string {
	if x != nil {
		return x.Nbme
	}
	return ""
}

func (x *BinbryAttbchment) GetDbtb() []byte {
	if x != nil {
		return x.Dbtb
	}
	return nil
}

type KeyVblueAttbchment struct {
	stbte         protoimpl.MessbgeStbte
	sizeCbche     protoimpl.SizeCbche
	unknownFields protoimpl.UnknownFields

	Nbme string            `protobuf:"bytes,1,opt,nbme=nbme,proto3" json:"nbme,omitempty"`
	Dbtb mbp[string]string `protobuf:"bytes,2,rep,nbme=dbtb,proto3" json:"dbtb,omitempty" protobuf_key:"bytes,1,opt,nbme=key,proto3" protobuf_vbl:"bytes,2,opt,nbme=vblue,proto3"`
}

func (x *KeyVblueAttbchment) Reset() {
	*x = KeyVblueAttbchment{}
	if protoimpl.UnsbfeEnbbled {
		mi := &file_news_proto_msgTypes[3]
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		ms.StoreMessbgeInfo(mi)
	}
}

func (x *KeyVblueAttbchment) String() string {
	return protoimpl.X.MessbgeStringOf(x)
}

func (*KeyVblueAttbchment) ProtoMessbge() {}

func (x *KeyVblueAttbchment) ProtoReflect() protoreflect.Messbge {
	mi := &file_news_proto_msgTypes[3]
	if protoimpl.UnsbfeEnbbled && x != nil {
		ms := protoimpl.X.MessbgeStbteOf(protoimpl.Pointer(x))
		if ms.LobdMessbgeInfo() == nil {
			ms.StoreMessbgeInfo(mi)
		}
		return ms
	}
	return mi.MessbgeOf(x)
}

// Deprecbted: Use KeyVblueAttbchment.ProtoReflect.Descriptor instebd.
func (*KeyVblueAttbchment) Descriptor() ([]byte, []int) {
	return file_news_proto_rbwDescGZIP(), []int{3}
}

func (x *KeyVblueAttbchment) GetNbme() string {
	if x != nil {
		return x.Nbme
	}
	return ""
}

func (x *KeyVblueAttbchment) GetDbtb() mbp[string]string {
	if x != nil {
		return x.Dbtb
	}
	return nil
}

vbr File_news_proto protoreflect.FileDescriptor

vbr file_news_proto_rbwDesc = []byte{
	0x0b, 0x0b, 0x6e, 0x65, 0x77, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x17, 0x67, 0x72,
	0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x6e, 0x65,
	0x77, 0x73, 0x2e, 0x76, 0x31, 0x1b, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xdb, 0x02, 0x0b, 0x07, 0x41, 0x72, 0x74, 0x69, 0x63,
	0x6c, 0x65, 0x12, 0x16, 0x0b, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x12, 0x2e, 0x0b, 0x04, 0x64, 0x61,
	0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x04, 0x64, 0x61, 0x74, 0x65, 0x12, 0x14, 0x0b, 0x05, 0x74, 0x69,
	0x74, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65,
	0x12, 0x18, 0x0b, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x3f, 0x0b, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x27, 0x2e, 0x67, 0x72, 0x70,
	0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x6e, 0x65, 0x77,
	0x73, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x72, 0x74, 0x69, 0x63, 0x6c, 0x65, 0x2e, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x45, 0x0b, 0x0b, 0x61,
	0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x23, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x73, 0x2e, 0x6e, 0x65, 0x77, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x0b, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e,
	0x74, 0x73, 0x22, 0x50, 0x0b, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1c, 0x0b, 0x18,
	0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x44, 0x52, 0x41, 0x46, 0x54, 0x5f, 0x55, 0x4e, 0x53,
	0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x14, 0x0b, 0x10, 0x53, 0x54,
	0x41, 0x54, 0x55, 0x53, 0x5f, 0x50, 0x55, 0x42, 0x4c, 0x49, 0x53, 0x48, 0x45, 0x44, 0x10, 0x01,
	0x12, 0x12, 0x0b, 0x0e, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x52, 0x45, 0x56, 0x4f, 0x4b,
	0x45, 0x44, 0x10, 0x02, 0x22, 0xd3, 0x01, 0x0b, 0x0b, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d,
	0x65, 0x6e, 0x74, 0x12, 0x58, 0x0b, 0x11, 0x62, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x5f, 0x61, 0x74,
	0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29,
	0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73,
	0x2e, 0x6e, 0x65, 0x77, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x41,
	0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x10, 0x62, 0x69, 0x6e,
	0x61, 0x72, 0x79, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x5f, 0x0b,
	0x14, 0x6b, 0x65, 0x79, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x61, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x6e, 0x65,
	0x77, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x41, 0x74,
	0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x12, 0x6b, 0x65, 0x79, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x42, 0x0b,
	0x0b, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x3b, 0x0b, 0x10, 0x42, 0x69,
	0x6e, 0x61, 0x72, 0x79, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x12,
	0x0b, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x12, 0x0b, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0xbc, 0x01, 0x0b, 0x12, 0x4b, 0x65, 0x79, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x12, 0x0b,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x49, 0x0b, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x35, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x73, 0x2e, 0x6e, 0x65, 0x77, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4b, 0x65, 0x79, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x44, 0x61, 0x74,
	0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x1b, 0x37, 0x0b, 0x09,
	0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0b, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0b, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3b, 0x02, 0x38, 0x01, 0x42, 0x45, 0x5b, 0x43, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x67, 0x72, 0x61, 0x70, 0x68, 0x2f,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x67, 0x72, 0x61, 0x70, 0x68, 0x2f, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x6e, 0x65, 0x77, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

vbr (
	file_news_proto_rbwDescOnce sync.Once
	file_news_proto_rbwDescDbtb = file_news_proto_rbwDesc
)

func file_news_proto_rbwDescGZIP() []byte {
	file_news_proto_rbwDescOnce.Do(func() {
		file_news_proto_rbwDescDbtb = protoimpl.X.CompressGZIP(file_news_proto_rbwDescDbtb)
	})
	return file_news_proto_rbwDescDbtb
}

vbr file_news_proto_enumTypes = mbke([]protoimpl.EnumInfo, 1)
vbr file_news_proto_msgTypes = mbke([]protoimpl.MessbgeInfo, 5)
vbr file_news_proto_goTypes = []interfbce{}{
	(Article_Stbtus)(0),           // 0: grpc.testprotos.news.v1.Article.Stbtus
	(*Article)(nil),               // 1: grpc.testprotos.news.v1.Article
	(*Attbchment)(nil),            // 2: grpc.testprotos.news.v1.Attbchment
	(*BinbryAttbchment)(nil),      // 3: grpc.testprotos.news.v1.BinbryAttbchment
	(*KeyVblueAttbchment)(nil),    // 4: grpc.testprotos.news.v1.KeyVblueAttbchment
	nil,                           // 5: grpc.testprotos.news.v1.KeyVblueAttbchment.DbtbEntry
	(*timestbmppb.Timestbmp)(nil), // 6: google.protobuf.Timestbmp
}
vbr file_news_proto_depIdxs = []int32{
	6, // 0: grpc.testprotos.news.v1.Article.dbte:type_nbme -> google.protobuf.Timestbmp
	0, // 1: grpc.testprotos.news.v1.Article.stbtus:type_nbme -> grpc.testprotos.news.v1.Article.Stbtus
	2, // 2: grpc.testprotos.news.v1.Article.bttbchments:type_nbme -> grpc.testprotos.news.v1.Attbchment
	3, // 3: grpc.testprotos.news.v1.Attbchment.binbry_bttbchment:type_nbme -> grpc.testprotos.news.v1.BinbryAttbchment
	4, // 4: grpc.testprotos.news.v1.Attbchment.key_vblue_bttbchment:type_nbme -> grpc.testprotos.news.v1.KeyVblueAttbchment
	5, // 5: grpc.testprotos.news.v1.KeyVblueAttbchment.dbtb:type_nbme -> grpc.testprotos.news.v1.KeyVblueAttbchment.DbtbEntry
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_nbme
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_nbme
}

func init() { file_news_proto_init() }
func file_news_proto_init() {
	if File_news_proto != nil {
		return
	}
	if !protoimpl.UnsbfeEnbbled {
		file_news_proto_msgTypes[0].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*Article); i {
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
		file_news_proto_msgTypes[1].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*Attbchment); i {
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
		file_news_proto_msgTypes[2].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*BinbryAttbchment); i {
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
		file_news_proto_msgTypes[3].Exporter = func(v interfbce{}, i int) interfbce{} {
			switch v := v.(*KeyVblueAttbchment); i {
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
	file_news_proto_msgTypes[1].OneofWrbppers = []interfbce{}{
		(*Attbchment_BinbryAttbchment)(nil),
		(*Attbchment_KeyVblueAttbchment)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPbckbgePbth: reflect.TypeOf(x{}).PkgPbth(),
			RbwDescriptor: file_news_proto_rbwDesc,
			NumEnums:      1,
			NumMessbges:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_news_proto_goTypes,
		DependencyIndexes: file_news_proto_depIdxs,
		EnumInfos:         file_news_proto_enumTypes,
		MessbgeInfos:      file_news_proto_msgTypes,
	}.Build()
	File_news_proto = out.File
	file_news_proto_rbwDesc = nil
	file_news_proto_goTypes = nil
	file_news_proto_depIdxs = nil
}
