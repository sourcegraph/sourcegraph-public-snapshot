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
// source: google/api/expr/v1alpha1/explain.proto

package expr

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

// Values of intermediate expressions produced when evaluating expression.
// Deprecated, use `EvalState` instead.
//
// Deprecated: Do not use.
type Explain struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// All of the observed values.
	//
	// The field value_index is an index in the values list.
	// Separating values from steps is needed to remove redundant values.
	Values []*Value `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
	// List of steps.
	//
	// Repeated evaluations of the same expression generate new ExprStep
	// instances. The order of such ExprStep instances matches the order of
	// elements returned by Comprehension.iter_range.
	ExprSteps []*Explain_ExprStep `protobuf:"bytes,2,rep,name=expr_steps,json=exprSteps,proto3" json:"expr_steps,omitempty"`
}

func (x *Explain) Reset() {
	*x = Explain{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_api_expr_v1alpha1_explain_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Explain) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Explain) ProtoMessage() {}

func (x *Explain) ProtoReflect() protoreflect.Message {
	mi := &file_google_api_expr_v1alpha1_explain_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Explain.ProtoReflect.Descriptor instead.
func (*Explain) Descriptor() ([]byte, []int) {
	return file_google_api_expr_v1alpha1_explain_proto_rawDescGZIP(), []int{0}
}

func (x *Explain) GetValues() []*Value {
	if x != nil {
		return x.Values
	}
	return nil
}

func (x *Explain) GetExprSteps() []*Explain_ExprStep {
	if x != nil {
		return x.ExprSteps
	}
	return nil
}

// ID and value index of one step.
type Explain_ExprStep struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ID of corresponding Expr node.
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	// Index of the value in the values list.
	ValueIndex int32 `protobuf:"varint,2,opt,name=value_index,json=valueIndex,proto3" json:"value_index,omitempty"`
}

func (x *Explain_ExprStep) Reset() {
	*x = Explain_ExprStep{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_api_expr_v1alpha1_explain_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Explain_ExprStep) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Explain_ExprStep) ProtoMessage() {}

func (x *Explain_ExprStep) ProtoReflect() protoreflect.Message {
	mi := &file_google_api_expr_v1alpha1_explain_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Explain_ExprStep.ProtoReflect.Descriptor instead.
func (*Explain_ExprStep) Descriptor() ([]byte, []int) {
	return file_google_api_expr_v1alpha1_explain_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Explain_ExprStep) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Explain_ExprStep) GetValueIndex() int32 {
	if x != nil {
		return x.ValueIndex
	}
	return 0
}

var File_google_api_expr_v1alpha1_explain_proto protoreflect.FileDescriptor

var file_google_api_expr_v1alpha1_explain_proto_rawDesc = []byte{
	0x0a, 0x26, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x65, 0x78, 0x70,
	0x72, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x65, 0x78, 0x70, 0x6c, 0x61,
	0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x65, 0x78, 0x70, 0x72, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x1a, 0x24, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x65,
	0x78, 0x70, 0x72, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xce, 0x01, 0x0a, 0x07, 0x45, 0x78, 0x70,
	0x6c, 0x61, 0x69, 0x6e, 0x12, 0x37, 0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x65, 0x78, 0x70, 0x72, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x12, 0x49, 0x0a,
	0x0a, 0x65, 0x78, 0x70, 0x72, 0x5f, 0x73, 0x74, 0x65, 0x70, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x2a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x65,
	0x78, 0x70, 0x72, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x45, 0x78, 0x70,
	0x6c, 0x61, 0x69, 0x6e, 0x2e, 0x45, 0x78, 0x70, 0x72, 0x53, 0x74, 0x65, 0x70, 0x52, 0x09, 0x65,
	0x78, 0x70, 0x72, 0x53, 0x74, 0x65, 0x70, 0x73, 0x1a, 0x3b, 0x0a, 0x08, 0x45, 0x78, 0x70, 0x72,
	0x53, 0x74, 0x65, 0x70, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x69, 0x6e,
	0x64, 0x65, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x49, 0x6e, 0x64, 0x65, 0x78, 0x3a, 0x02, 0x18, 0x01, 0x42, 0x6f, 0x0a, 0x1c, 0x63, 0x6f, 0x6d,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x65, 0x78, 0x70, 0x72,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x42, 0x0c, 0x45, 0x78, 0x70, 0x6c, 0x61,
	0x69, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3c, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x67, 0x65, 0x6e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61, 0x70, 0x69, 0x73,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x65, 0x78, 0x70, 0x72, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x3b, 0x65, 0x78, 0x70, 0x72, 0xf8, 0x01, 0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_google_api_expr_v1alpha1_explain_proto_rawDescOnce sync.Once
	file_google_api_expr_v1alpha1_explain_proto_rawDescData = file_google_api_expr_v1alpha1_explain_proto_rawDesc
)

func file_google_api_expr_v1alpha1_explain_proto_rawDescGZIP() []byte {
	file_google_api_expr_v1alpha1_explain_proto_rawDescOnce.Do(func() {
		file_google_api_expr_v1alpha1_explain_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_api_expr_v1alpha1_explain_proto_rawDescData)
	})
	return file_google_api_expr_v1alpha1_explain_proto_rawDescData
}

var file_google_api_expr_v1alpha1_explain_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_google_api_expr_v1alpha1_explain_proto_goTypes = []interface{}{
	(*Explain)(nil),          // 0: google.api.expr.v1alpha1.Explain
	(*Explain_ExprStep)(nil), // 1: google.api.expr.v1alpha1.Explain.ExprStep
	(*Value)(nil),            // 2: google.api.expr.v1alpha1.Value
}
var file_google_api_expr_v1alpha1_explain_proto_depIdxs = []int32{
	2, // 0: google.api.expr.v1alpha1.Explain.values:type_name -> google.api.expr.v1alpha1.Value
	1, // 1: google.api.expr.v1alpha1.Explain.expr_steps:type_name -> google.api.expr.v1alpha1.Explain.ExprStep
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_google_api_expr_v1alpha1_explain_proto_init() }
func file_google_api_expr_v1alpha1_explain_proto_init() {
	if File_google_api_expr_v1alpha1_explain_proto != nil {
		return
	}
	file_google_api_expr_v1alpha1_value_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_google_api_expr_v1alpha1_explain_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Explain); i {
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
		file_google_api_expr_v1alpha1_explain_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Explain_ExprStep); i {
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
			RawDescriptor: file_google_api_expr_v1alpha1_explain_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_api_expr_v1alpha1_explain_proto_goTypes,
		DependencyIndexes: file_google_api_expr_v1alpha1_explain_proto_depIdxs,
		MessageInfos:      file_google_api_expr_v1alpha1_explain_proto_msgTypes,
	}.Build()
	File_google_api_expr_v1alpha1_explain_proto = out.File
	file_google_api_expr_v1alpha1_explain_proto_rawDesc = nil
	file_google_api_expr_v1alpha1_explain_proto_goTypes = nil
	file_google_api_expr_v1alpha1_explain_proto_depIdxs = nil
}
