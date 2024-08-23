// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.21.12
// source: internal/proto/v2/s2a_context/s2a_context.proto

package s2a_context_go_proto

import (
	common_go_proto "github.com/google/s2a-go/internal/proto/common_go_proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type S2AContext struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The SPIFFE ID from the peer leaf certificate, if present.
	//
	// This field is only populated if the leaf certificate is a valid SPIFFE
	// SVID; in particular, there is a unique URI SAN and this URI SAN is a valid
	// SPIFFE ID.
	LeafCertSpiffeId string `protobuf:"bytes,1,opt,name=leaf_cert_spiffe_id,json=leafCertSpiffeId,proto3" json:"leaf_cert_spiffe_id,omitempty"`
	// The URIs that are present in the SubjectAltName extension of the peer leaf
	// certificate.
	//
	// Note that the extracted URIs are not validated and may not be properly
	// formatted.
	LeafCertUris []string `protobuf:"bytes,2,rep,name=leaf_cert_uris,json=leafCertUris,proto3" json:"leaf_cert_uris,omitempty"`
	// The DNSNames that are present in the SubjectAltName extension of the peer
	// leaf certificate.
	LeafCertDnsnames []string `protobuf:"bytes,3,rep,name=leaf_cert_dnsnames,json=leafCertDnsnames,proto3" json:"leaf_cert_dnsnames,omitempty"`
	// The (ordered) list of fingerprints in the certificate chain used to verify
	// the given leaf certificate. The order MUST be from leaf certificate
	// fingerprint to root certificate fingerprint.
	//
	// A fingerprint is the base-64 encoding of the SHA256 hash of the
	// DER-encoding of a certificate. The list MAY be populated even if the peer
	// certificate chain was NOT validated successfully.
	PeerCertificateChainFingerprints []string `protobuf:"bytes,4,rep,name=peer_certificate_chain_fingerprints,json=peerCertificateChainFingerprints,proto3" json:"peer_certificate_chain_fingerprints,omitempty"`
	// The local identity used during session setup.
	LocalIdentity *common_go_proto.Identity `protobuf:"bytes,5,opt,name=local_identity,json=localIdentity,proto3" json:"local_identity,omitempty"`
	// The SHA256 hash of the DER-encoding of the local leaf certificate used in
	// the handshake.
	LocalLeafCertFingerprint []byte `protobuf:"bytes,6,opt,name=local_leaf_cert_fingerprint,json=localLeafCertFingerprint,proto3" json:"local_leaf_cert_fingerprint,omitempty"`
}

func (x *S2AContext) Reset() {
	*x = S2AContext{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_v2_s2a_context_s2a_context_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *S2AContext) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*S2AContext) ProtoMessage() {}

func (x *S2AContext) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_v2_s2a_context_s2a_context_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use S2AContext.ProtoReflect.Descriptor instead.
func (*S2AContext) Descriptor() ([]byte, []int) {
	return file_internal_proto_v2_s2a_context_s2a_context_proto_rawDescGZIP(), []int{0}
}

func (x *S2AContext) GetLeafCertSpiffeId() string {
	if x != nil {
		return x.LeafCertSpiffeId
	}
	return ""
}

func (x *S2AContext) GetLeafCertUris() []string {
	if x != nil {
		return x.LeafCertUris
	}
	return nil
}

func (x *S2AContext) GetLeafCertDnsnames() []string {
	if x != nil {
		return x.LeafCertDnsnames
	}
	return nil
}

func (x *S2AContext) GetPeerCertificateChainFingerprints() []string {
	if x != nil {
		return x.PeerCertificateChainFingerprints
	}
	return nil
}

func (x *S2AContext) GetLocalIdentity() *common_go_proto.Identity {
	if x != nil {
		return x.LocalIdentity
	}
	return nil
}

func (x *S2AContext) GetLocalLeafCertFingerprint() []byte {
	if x != nil {
		return x.LocalLeafCertFingerprint
	}
	return nil
}

var File_internal_proto_v2_s2a_context_s2a_context_proto protoreflect.FileDescriptor

var file_internal_proto_v2_s2a_context_s2a_context_proto_rawDesc = []byte{
	0x0a, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x76, 0x32, 0x2f, 0x73, 0x32, 0x61, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x2f,
	0x73, 0x32, 0x61, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x0c, 0x73, 0x32, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x76, 0x32, 0x1a,
	0x22, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0xd9, 0x02, 0x0a, 0x0a, 0x53, 0x32, 0x41, 0x43, 0x6f, 0x6e, 0x74, 0x65,
	0x78, 0x74, 0x12, 0x2d, 0x0a, 0x13, 0x6c, 0x65, 0x61, 0x66, 0x5f, 0x63, 0x65, 0x72, 0x74, 0x5f,
	0x73, 0x70, 0x69, 0x66, 0x66, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x10, 0x6c, 0x65, 0x61, 0x66, 0x43, 0x65, 0x72, 0x74, 0x53, 0x70, 0x69, 0x66, 0x66, 0x65, 0x49,
	0x64, 0x12, 0x24, 0x0a, 0x0e, 0x6c, 0x65, 0x61, 0x66, 0x5f, 0x63, 0x65, 0x72, 0x74, 0x5f, 0x75,
	0x72, 0x69, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0c, 0x6c, 0x65, 0x61, 0x66, 0x43,
	0x65, 0x72, 0x74, 0x55, 0x72, 0x69, 0x73, 0x12, 0x2c, 0x0a, 0x12, 0x6c, 0x65, 0x61, 0x66, 0x5f,
	0x63, 0x65, 0x72, 0x74, 0x5f, 0x64, 0x6e, 0x73, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x03, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x10, 0x6c, 0x65, 0x61, 0x66, 0x43, 0x65, 0x72, 0x74, 0x44, 0x6e, 0x73,
	0x6e, 0x61, 0x6d, 0x65, 0x73, 0x12, 0x4d, 0x0a, 0x23, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x63, 0x65,
	0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f,
	0x66, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x70, 0x72, 0x69, 0x6e, 0x74, 0x73, 0x18, 0x04, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x20, 0x70, 0x65, 0x65, 0x72, 0x43, 0x65, 0x72, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x65, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x46, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x70, 0x72,
	0x69, 0x6e, 0x74, 0x73, 0x12, 0x3a, 0x0a, 0x0e, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x5f, 0x69, 0x64,
	0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x73,
	0x32, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74,
	0x79, 0x52, 0x0d, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x12, 0x3d, 0x0a, 0x1b, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x5f, 0x6c, 0x65, 0x61, 0x66, 0x5f, 0x63,
	0x65, 0x72, 0x74, 0x5f, 0x66, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x70, 0x72, 0x69, 0x6e, 0x74, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x18, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x4c, 0x65, 0x61, 0x66,
	0x43, 0x65, 0x72, 0x74, 0x46, 0x69, 0x6e, 0x67, 0x65, 0x72, 0x70, 0x72, 0x69, 0x6e, 0x74, 0x42,
	0x3e, 0x5a, 0x3c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x73, 0x32, 0x61, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x76, 0x32, 0x2f, 0x73, 0x32, 0x61, 0x5f, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x67, 0x6f, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_proto_v2_s2a_context_s2a_context_proto_rawDescOnce sync.Once
	file_internal_proto_v2_s2a_context_s2a_context_proto_rawDescData = file_internal_proto_v2_s2a_context_s2a_context_proto_rawDesc
)

func file_internal_proto_v2_s2a_context_s2a_context_proto_rawDescGZIP() []byte {
	file_internal_proto_v2_s2a_context_s2a_context_proto_rawDescOnce.Do(func() {
		file_internal_proto_v2_s2a_context_s2a_context_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_proto_v2_s2a_context_s2a_context_proto_rawDescData)
	})
	return file_internal_proto_v2_s2a_context_s2a_context_proto_rawDescData
}

var file_internal_proto_v2_s2a_context_s2a_context_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_internal_proto_v2_s2a_context_s2a_context_proto_goTypes = []interface{}{
	(*S2AContext)(nil),               // 0: s2a.proto.v2.S2AContext
	(*common_go_proto.Identity)(nil), // 1: s2a.proto.Identity
}
var file_internal_proto_v2_s2a_context_s2a_context_proto_depIdxs = []int32{
	1, // 0: s2a.proto.v2.S2AContext.local_identity:type_name -> s2a.proto.Identity
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_internal_proto_v2_s2a_context_s2a_context_proto_init() }
func file_internal_proto_v2_s2a_context_s2a_context_proto_init() {
	if File_internal_proto_v2_s2a_context_s2a_context_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_proto_v2_s2a_context_s2a_context_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*S2AContext); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_proto_v2_s2a_context_s2a_context_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_proto_v2_s2a_context_s2a_context_proto_goTypes,
		DependencyIndexes: file_internal_proto_v2_s2a_context_s2a_context_proto_depIdxs,
		MessageInfos:      file_internal_proto_v2_s2a_context_s2a_context_proto_msgTypes,
	}.Build()
	File_internal_proto_v2_s2a_context_s2a_context_proto = out.File
	file_internal_proto_v2_s2a_context_s2a_context_proto_rawDesc = nil
	file_internal_proto_v2_s2a_context_s2a_context_proto_goTypes = nil
	file_internal_proto_v2_s2a_context_s2a_context_proto_depIdxs = nil
}
