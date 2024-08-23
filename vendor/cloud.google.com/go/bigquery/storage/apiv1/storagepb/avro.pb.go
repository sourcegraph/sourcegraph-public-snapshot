// Copyright 2023 Google LLC
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
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.2
// source: google/cloud/bigquery/storage/v1/avro.proto

package storagepb

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Avro schema.
type AvroSchema struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Json serialized schema, as described at
	// https://avro.apache.org/docs/1.8.1/spec.html.
	Schema string `protobuf:"bytes,1,opt,name=schema,proto3" json:"schema,omitempty"`
}

func (x *AvroSchema) Reset() {
	*x = AvroSchema{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_bigquery_storage_v1_avro_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AvroSchema) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AvroSchema) ProtoMessage() {}

func (x *AvroSchema) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_bigquery_storage_v1_avro_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AvroSchema.ProtoReflect.Descriptor instead.
func (*AvroSchema) Descriptor() ([]byte, []int) {
	return file_google_cloud_bigquery_storage_v1_avro_proto_rawDescGZIP(), []int{0}
}

func (x *AvroSchema) GetSchema() string {
	if x != nil {
		return x.Schema
	}
	return ""
}

// Avro rows.
type AvroRows struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Binary serialized rows in a block.
	SerializedBinaryRows []byte `protobuf:"bytes,1,opt,name=serialized_binary_rows,json=serializedBinaryRows,proto3" json:"serialized_binary_rows,omitempty"`
	// [Deprecated] The count of rows in the returning block.
	// Please use the format-independent ReadRowsResponse.row_count instead.
	//
	// Deprecated: Marked as deprecated in google/cloud/bigquery/storage/v1/avro.proto.
	RowCount int64 `protobuf:"varint,2,opt,name=row_count,json=rowCount,proto3" json:"row_count,omitempty"`
}

func (x *AvroRows) Reset() {
	*x = AvroRows{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_bigquery_storage_v1_avro_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AvroRows) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AvroRows) ProtoMessage() {}

func (x *AvroRows) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_bigquery_storage_v1_avro_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AvroRows.ProtoReflect.Descriptor instead.
func (*AvroRows) Descriptor() ([]byte, []int) {
	return file_google_cloud_bigquery_storage_v1_avro_proto_rawDescGZIP(), []int{1}
}

func (x *AvroRows) GetSerializedBinaryRows() []byte {
	if x != nil {
		return x.SerializedBinaryRows
	}
	return nil
}

// Deprecated: Marked as deprecated in google/cloud/bigquery/storage/v1/avro.proto.
func (x *AvroRows) GetRowCount() int64 {
	if x != nil {
		return x.RowCount
	}
	return 0
}

// Contains options specific to Avro Serialization.
type AvroSerializationOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Enable displayName attribute in Avro schema.
	//
	// The Avro specification requires field names to be alphanumeric.  By
	// default, in cases when column names do not conform to these requirements
	// (e.g. non-ascii unicode codepoints) and Avro is requested as an output
	// format, the CreateReadSession call will fail.
	//
	// Setting this field to true, populates avro field names with a placeholder
	// value and populates a "displayName" attribute for every avro field with the
	// original column name.
	EnableDisplayNameAttribute bool `protobuf:"varint,1,opt,name=enable_display_name_attribute,json=enableDisplayNameAttribute,proto3" json:"enable_display_name_attribute,omitempty"`
}

func (x *AvroSerializationOptions) Reset() {
	*x = AvroSerializationOptions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_bigquery_storage_v1_avro_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AvroSerializationOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AvroSerializationOptions) ProtoMessage() {}

func (x *AvroSerializationOptions) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_bigquery_storage_v1_avro_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AvroSerializationOptions.ProtoReflect.Descriptor instead.
func (*AvroSerializationOptions) Descriptor() ([]byte, []int) {
	return file_google_cloud_bigquery_storage_v1_avro_proto_rawDescGZIP(), []int{2}
}

func (x *AvroSerializationOptions) GetEnableDisplayNameAttribute() bool {
	if x != nil {
		return x.EnableDisplayNameAttribute
	}
	return false
}

var File_google_cloud_bigquery_storage_v1_avro_proto protoreflect.FileDescriptor

var file_google_cloud_bigquery_storage_v1_avro_proto_rawDesc = []byte{
	0x0a, 0x2b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x62,
	0x69, 0x67, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2f,
	0x76, 0x31, 0x2f, 0x61, 0x76, 0x72, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x20, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x62, 0x69, 0x67, 0x71,
	0x75, 0x65, 0x72, 0x79, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x76, 0x31, 0x22,
	0x24, 0x0a, 0x0a, 0x41, 0x76, 0x72, 0x6f, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x12, 0x16, 0x0a,
	0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x22, 0x61, 0x0a, 0x08, 0x41, 0x76, 0x72, 0x6f, 0x52, 0x6f, 0x77,
	0x73, 0x12, 0x34, 0x0a, 0x16, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x5f,
	0x62, 0x69, 0x6e, 0x61, 0x72, 0x79, 0x5f, 0x72, 0x6f, 0x77, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x14, 0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x42, 0x69, 0x6e,
	0x61, 0x72, 0x79, 0x52, 0x6f, 0x77, 0x73, 0x12, 0x1f, 0x0a, 0x09, 0x72, 0x6f, 0x77, 0x5f, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x42, 0x02, 0x18, 0x01, 0x52, 0x08,
	0x72, 0x6f, 0x77, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x5d, 0x0a, 0x18, 0x41, 0x76, 0x72, 0x6f,
	0x53, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x12, 0x41, 0x0a, 0x1d, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x64,
	0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x5f, 0x61, 0x74, 0x74, 0x72,
	0x69, 0x62, 0x75, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x1a, 0x65, 0x6e, 0x61,
	0x62, 0x6c, 0x65, 0x44, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x41, 0x74,
	0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x65, 0x42, 0xb9, 0x01, 0x0a, 0x24, 0x63, 0x6f, 0x6d, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x62, 0x69, 0x67,
	0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x76, 0x31,
	0x42, 0x09, 0x41, 0x76, 0x72, 0x6f, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x67, 0x6f, 0x2f, 0x62, 0x69, 0x67, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2f, 0x73, 0x74, 0x6f, 0x72,
	0x61, 0x67, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67,
	0x65, 0x70, 0x62, 0x3b, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x70, 0x62, 0xaa, 0x02, 0x20,
	0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x42, 0x69, 0x67,
	0x51, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x56, 0x31,
	0xca, 0x02, 0x20, 0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x5c, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x5c,
	0x42, 0x69, 0x67, 0x51, 0x75, 0x65, 0x72, 0x79, 0x5c, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65,
	0x5c, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_cloud_bigquery_storage_v1_avro_proto_rawDescOnce sync.Once
	file_google_cloud_bigquery_storage_v1_avro_proto_rawDescData = file_google_cloud_bigquery_storage_v1_avro_proto_rawDesc
)

func file_google_cloud_bigquery_storage_v1_avro_proto_rawDescGZIP() []byte {
	file_google_cloud_bigquery_storage_v1_avro_proto_rawDescOnce.Do(func() {
		file_google_cloud_bigquery_storage_v1_avro_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_cloud_bigquery_storage_v1_avro_proto_rawDescData)
	})
	return file_google_cloud_bigquery_storage_v1_avro_proto_rawDescData
}

var file_google_cloud_bigquery_storage_v1_avro_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_google_cloud_bigquery_storage_v1_avro_proto_goTypes = []interface{}{
	(*AvroSchema)(nil),               // 0: google.cloud.bigquery.storage.v1.AvroSchema
	(*AvroRows)(nil),                 // 1: google.cloud.bigquery.storage.v1.AvroRows
	(*AvroSerializationOptions)(nil), // 2: google.cloud.bigquery.storage.v1.AvroSerializationOptions
}
var file_google_cloud_bigquery_storage_v1_avro_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_google_cloud_bigquery_storage_v1_avro_proto_init() }
func file_google_cloud_bigquery_storage_v1_avro_proto_init() {
	if File_google_cloud_bigquery_storage_v1_avro_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_cloud_bigquery_storage_v1_avro_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AvroSchema); i {
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
		file_google_cloud_bigquery_storage_v1_avro_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AvroRows); i {
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
		file_google_cloud_bigquery_storage_v1_avro_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AvroSerializationOptions); i {
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
			RawDescriptor: file_google_cloud_bigquery_storage_v1_avro_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_cloud_bigquery_storage_v1_avro_proto_goTypes,
		DependencyIndexes: file_google_cloud_bigquery_storage_v1_avro_proto_depIdxs,
		MessageInfos:      file_google_cloud_bigquery_storage_v1_avro_proto_msgTypes,
	}.Build()
	File_google_cloud_bigquery_storage_v1_avro_proto = out.File
	file_google_cloud_bigquery_storage_v1_avro_proto_rawDesc = nil
	file_google_cloud_bigquery_storage_v1_avro_proto_goTypes = nil
	file_google_cloud_bigquery_storage_v1_avro_proto_depIdxs = nil
}
