// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v4.24.4
// source: google/api/httpbody.proto

package httpbody

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Message that represents an arbitrary HTTP body. It should only be used for
// payload formats that can't be represented as JSON, such as raw binary or
// an HTML page.
//
// This message can be used both in streaming and non-streaming API methods in
// the request as well as the response.
//
// It can be used as a top-level request field, which is convenient if one
// wants to extract parameters from either the URL or HTTP template into the
// request fields and also want access to the raw HTTP body.
//
// Example:
//
//	message GetResourceRequest {
//	  // A unique request id.
//	  string request_id = 1;
//
//	  // The raw HTTP body is bound to this field.
//	  google.api.HttpBody http_body = 2;
//
//	}
//
//	service ResourceService {
//	  rpc GetResource(GetResourceRequest)
//	    returns (google.api.HttpBody);
//	  rpc UpdateResource(google.api.HttpBody)
//	    returns (google.protobuf.Empty);
//
//	}
//
// Example with streaming methods:
//
//	service CaldavService {
//	  rpc GetCalendar(stream google.api.HttpBody)
//	    returns (stream google.api.HttpBody);
//	  rpc UpdateCalendar(stream google.api.HttpBody)
//	    returns (stream google.api.HttpBody);
//
//	}
//
// Use of this type only changes how the request and response bodies are
// handled, all other features will continue to work unchanged.
type HttpBody struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The HTTP Content-Type header value specifying the content type of the body.
	ContentType string `protobuf:"bytes,1,opt,name=content_type,json=contentType,proto3" json:"content_type,omitempty"`
	// The HTTP request/response body as raw binary.
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	// Application specific response metadata. Must be set in the first response
	// for streaming APIs.
	Extensions []*anypb.Any `protobuf:"bytes,3,rep,name=extensions,proto3" json:"extensions,omitempty"`
}

func (x *HttpBody) Reset() {
	*x = HttpBody{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_api_httpbody_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HttpBody) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HttpBody) ProtoMessage() {}

func (x *HttpBody) ProtoReflect() protoreflect.Message {
	mi := &file_google_api_httpbody_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HttpBody.ProtoReflect.Descriptor instead.
func (*HttpBody) Descriptor() ([]byte, []int) {
	return file_google_api_httpbody_proto_rawDescGZIP(), []int{0}
}

func (x *HttpBody) GetContentType() string {
	if x != nil {
		return x.ContentType
	}
	return ""
}

func (x *HttpBody) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *HttpBody) GetExtensions() []*anypb.Any {
	if x != nil {
		return x.Extensions
	}
	return nil
}

var File_google_api_httpbody_proto protoreflect.FileDescriptor

var file_google_api_httpbody_proto_rawDesc = []byte{
	0x0a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x68, 0x74, 0x74,
	0x70, 0x62, 0x6f, 0x64, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x77, 0x0a, 0x08, 0x48, 0x74, 0x74, 0x70, 0x42, 0x6f, 0x64, 0x79, 0x12, 0x21,
	0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x34, 0x0a, 0x0a, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69,
	0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52,
	0x0a, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x42, 0x68, 0x0a, 0x0e, 0x63,
	0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x42, 0x0d, 0x48,
	0x74, 0x74, 0x70, 0x42, 0x6f, 0x64, 0x79, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3b,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x2e, 0x6f, 0x72,
	0x67, 0x2f, 0x67, 0x65, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x68, 0x74, 0x74, 0x70, 0x62, 0x6f,
	0x64, 0x79, 0x3b, 0x68, 0x74, 0x74, 0x70, 0x62, 0x6f, 0x64, 0x79, 0xf8, 0x01, 0x01, 0xa2, 0x02,
	0x04, 0x47, 0x41, 0x50, 0x49, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_api_httpbody_proto_rawDescOnce sync.Once
	file_google_api_httpbody_proto_rawDescData = file_google_api_httpbody_proto_rawDesc
)

func file_google_api_httpbody_proto_rawDescGZIP() []byte {
	file_google_api_httpbody_proto_rawDescOnce.Do(func() {
		file_google_api_httpbody_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_api_httpbody_proto_rawDescData)
	})
	return file_google_api_httpbody_proto_rawDescData
}

var file_google_api_httpbody_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_google_api_httpbody_proto_goTypes = []interface{}{
	(*HttpBody)(nil),  // 0: google.api.HttpBody
	(*anypb.Any)(nil), // 1: google.protobuf.Any
}
var file_google_api_httpbody_proto_depIdxs = []int32{
	1, // 0: google.api.HttpBody.extensions:type_name -> google.protobuf.Any
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_google_api_httpbody_proto_init() }
func file_google_api_httpbody_proto_init() {
	if File_google_api_httpbody_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_api_httpbody_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HttpBody); i {
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
			RawDescriptor: file_google_api_httpbody_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_api_httpbody_proto_goTypes,
		DependencyIndexes: file_google_api_httpbody_proto_depIdxs,
		MessageInfos:      file_google_api_httpbody_proto_msgTypes,
	}.Build()
	File_google_api_httpbody_proto = out.File
	file_google_api_httpbody_proto_rawDesc = nil
	file_google_api_httpbody_proto_goTypes = nil
	file_google_api_httpbody_proto_depIdxs = nil
}
