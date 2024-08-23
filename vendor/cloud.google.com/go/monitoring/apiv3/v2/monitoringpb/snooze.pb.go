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
// source: google/monitoring/v3/snooze.proto

package monitoringpb

import (
	reflect "reflect"
	sync "sync"

	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// A `Snooze` will prevent any alerts from being opened, and close any that
// are already open. The `Snooze` will work on alerts that match the
// criteria defined in the `Snooze`. The `Snooze` will be active from
// `interval.start_time` through `interval.end_time`.
type Snooze struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Required. The name of the `Snooze`. The format is:
	//
	//	projects/[PROJECT_ID_OR_NUMBER]/snoozes/[SNOOZE_ID]
	//
	// The ID of the `Snooze` will be generated by the system.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Required. This defines the criteria for applying the `Snooze`. See
	// `Criteria` for more information.
	Criteria *Snooze_Criteria `protobuf:"bytes,3,opt,name=criteria,proto3" json:"criteria,omitempty"`
	// Required. The `Snooze` will be active from `interval.start_time` through
	// `interval.end_time`.
	// `interval.start_time` cannot be in the past. There is a 15 second clock
	// skew to account for the time it takes for a request to reach the API from
	// the UI.
	Interval *TimeInterval `protobuf:"bytes,4,opt,name=interval,proto3" json:"interval,omitempty"`
	// Required. A display name for the `Snooze`. This can be, at most, 512
	// unicode characters.
	DisplayName string `protobuf:"bytes,5,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
}

func (x *Snooze) Reset() {
	*x = Snooze{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_monitoring_v3_snooze_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Snooze) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Snooze) ProtoMessage() {}

func (x *Snooze) ProtoReflect() protoreflect.Message {
	mi := &file_google_monitoring_v3_snooze_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Snooze.ProtoReflect.Descriptor instead.
func (*Snooze) Descriptor() ([]byte, []int) {
	return file_google_monitoring_v3_snooze_proto_rawDescGZIP(), []int{0}
}

func (x *Snooze) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Snooze) GetCriteria() *Snooze_Criteria {
	if x != nil {
		return x.Criteria
	}
	return nil
}

func (x *Snooze) GetInterval() *TimeInterval {
	if x != nil {
		return x.Interval
	}
	return nil
}

func (x *Snooze) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

// Criteria specific to the `AlertPolicy`s that this `Snooze` applies to. The
// `Snooze` will suppress alerts that come from one of the `AlertPolicy`s
// whose names are supplied.
type Snooze_Criteria struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The specific `AlertPolicy` names for the alert that should be snoozed.
	// The format is:
	//
	//	projects/[PROJECT_ID_OR_NUMBER]/alertPolicies/[POLICY_ID]
	//
	// There is a limit of 16 policies per snooze. This limit is checked during
	// snooze creation.
	Policies []string `protobuf:"bytes,1,rep,name=policies,proto3" json:"policies,omitempty"`
}

func (x *Snooze_Criteria) Reset() {
	*x = Snooze_Criteria{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_monitoring_v3_snooze_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Snooze_Criteria) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Snooze_Criteria) ProtoMessage() {}

func (x *Snooze_Criteria) ProtoReflect() protoreflect.Message {
	mi := &file_google_monitoring_v3_snooze_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Snooze_Criteria.ProtoReflect.Descriptor instead.
func (*Snooze_Criteria) Descriptor() ([]byte, []int) {
	return file_google_monitoring_v3_snooze_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Snooze_Criteria) GetPolicies() []string {
	if x != nil {
		return x.Policies
	}
	return nil
}

var File_google_monitoring_v3_snooze_proto protoreflect.FileDescriptor

var file_google_monitoring_v3_snooze_proto_rawDesc = []byte{
	0x0a, 0x21, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72,
	0x69, 0x6e, 0x67, 0x2f, 0x76, 0x33, 0x2f, 0x73, 0x6e, 0x6f, 0x6f, 0x7a, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x14, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x6d, 0x6f, 0x6e, 0x69,
	0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x33, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61,
	0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x21, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x6d, 0x6f,
	0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x33, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf6, 0x02, 0x0a, 0x06, 0x53, 0x6e, 0x6f,
	0x6f, 0x7a, 0x65, 0x12, 0x17, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x46, 0x0a, 0x08,
	0x63, 0x72, 0x69, 0x74, 0x65, 0x72, 0x69, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69,
	0x6e, 0x67, 0x2e, 0x76, 0x33, 0x2e, 0x53, 0x6e, 0x6f, 0x6f, 0x7a, 0x65, 0x2e, 0x43, 0x72, 0x69,
	0x74, 0x65, 0x72, 0x69, 0x61, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x08, 0x63, 0x72, 0x69, 0x74,
	0x65, 0x72, 0x69, 0x61, 0x12, 0x43, 0x0a, 0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x33, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52,
	0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x12, 0x26, 0x0a, 0x0c, 0x64, 0x69, 0x73,
	0x70, 0x6c, 0x61, 0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x03, 0xe0, 0x41, 0x02, 0x52, 0x0b, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x4e, 0x61, 0x6d,
	0x65, 0x1a, 0x52, 0x0a, 0x08, 0x43, 0x72, 0x69, 0x74, 0x65, 0x72, 0x69, 0x61, 0x12, 0x46, 0x0a,
	0x08, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x69, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x42,
	0x2a, 0xfa, 0x41, 0x27, 0x0a, 0x25, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x41, 0x6c, 0x65, 0x72, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x08, 0x70, 0x6f, 0x6c,
	0x69, 0x63, 0x69, 0x65, 0x73, 0x3a, 0x4a, 0xea, 0x41, 0x47, 0x0a, 0x20, 0x6d, 0x6f, 0x6e, 0x69,
	0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61, 0x70, 0x69,
	0x73, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x53, 0x6e, 0x6f, 0x6f, 0x7a, 0x65, 0x12, 0x23, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x2f, 0x7b, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x7d,
	0x2f, 0x73, 0x6e, 0x6f, 0x6f, 0x7a, 0x65, 0x73, 0x2f, 0x7b, 0x73, 0x6e, 0x6f, 0x6f, 0x7a, 0x65,
	0x7d, 0x42, 0xc6, 0x01, 0x0a, 0x18, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x33, 0x42, 0x0b,
	0x53, 0x6e, 0x6f, 0x6f, 0x7a, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x41, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x67, 0x6f, 0x2f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x61, 0x70,
	0x69, 0x76, 0x33, 0x2f, 0x76, 0x32, 0x2f, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e,
	0x67, 0x70, 0x62, 0x3b, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x70, 0x62,
	0xaa, 0x02, 0x1a, 0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x2e, 0x56, 0x33, 0xca, 0x02, 0x1a,
	0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x5c, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x5c, 0x4d, 0x6f, 0x6e,
	0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x5c, 0x56, 0x33, 0xea, 0x02, 0x1d, 0x47, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x3a, 0x3a, 0x43, 0x6c, 0x6f, 0x75, 0x64, 0x3a, 0x3a, 0x4d, 0x6f, 0x6e, 0x69,
	0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x3a, 0x3a, 0x56, 0x33, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_google_monitoring_v3_snooze_proto_rawDescOnce sync.Once
	file_google_monitoring_v3_snooze_proto_rawDescData = file_google_monitoring_v3_snooze_proto_rawDesc
)

func file_google_monitoring_v3_snooze_proto_rawDescGZIP() []byte {
	file_google_monitoring_v3_snooze_proto_rawDescOnce.Do(func() {
		file_google_monitoring_v3_snooze_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_monitoring_v3_snooze_proto_rawDescData)
	})
	return file_google_monitoring_v3_snooze_proto_rawDescData
}

var file_google_monitoring_v3_snooze_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_google_monitoring_v3_snooze_proto_goTypes = []interface{}{
	(*Snooze)(nil),          // 0: google.monitoring.v3.Snooze
	(*Snooze_Criteria)(nil), // 1: google.monitoring.v3.Snooze.Criteria
	(*TimeInterval)(nil),    // 2: google.monitoring.v3.TimeInterval
}
var file_google_monitoring_v3_snooze_proto_depIdxs = []int32{
	1, // 0: google.monitoring.v3.Snooze.criteria:type_name -> google.monitoring.v3.Snooze.Criteria
	2, // 1: google.monitoring.v3.Snooze.interval:type_name -> google.monitoring.v3.TimeInterval
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_google_monitoring_v3_snooze_proto_init() }
func file_google_monitoring_v3_snooze_proto_init() {
	if File_google_monitoring_v3_snooze_proto != nil {
		return
	}
	file_google_monitoring_v3_common_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_google_monitoring_v3_snooze_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Snooze); i {
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
		file_google_monitoring_v3_snooze_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Snooze_Criteria); i {
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
			RawDescriptor: file_google_monitoring_v3_snooze_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_monitoring_v3_snooze_proto_goTypes,
		DependencyIndexes: file_google_monitoring_v3_snooze_proto_depIdxs,
		MessageInfos:      file_google_monitoring_v3_snooze_proto_msgTypes,
	}.Build()
	File_google_monitoring_v3_snooze_proto = out.File
	file_google_monitoring_v3_snooze_proto_rawDesc = nil
	file_google_monitoring_v3_snooze_proto_goTypes = nil
	file_google_monitoring_v3_snooze_proto_depIdxs = nil
}
