// Copyright 2020-2023 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: buf/alpha/registry/v1alpha1/image.proto

package registryv1alpha1

import (
	v1 "github.com/bufbuild/buf/private/gen/proto/go/buf/alpha/image/v1"
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

// ImageMask is used in GetImageRequest to specify which parts of an image
// should be masked in responses.
type ImageMask int32

const (
	ImageMask_IMAGE_MASK_UNSPECIFIED ImageMask = 0
	// IMAGE_MASK_MESSAGES refers to ImageFile's `google.protobuf.DescriptorProto
	// message_type` field.
	ImageMask_IMAGE_MASK_MESSAGES ImageMask = 1
	// IMAGE_MASK_ENUMS refers to ImageFile's `google.protobuf.EnumDescriptorProto
	// enum_type` field.
	ImageMask_IMAGE_MASK_ENUMS ImageMask = 2
	// IMAGE_MASK_SERVICES refers to ImageFile's
	// `google.protobuf.ServiceDescriptorProto service` field.
	ImageMask_IMAGE_MASK_SERVICES ImageMask = 3
)

// Enum value maps for ImageMask.
var (
	ImageMask_name = map[int32]string{
		0: "IMAGE_MASK_UNSPECIFIED",
		1: "IMAGE_MASK_MESSAGES",
		2: "IMAGE_MASK_ENUMS",
		3: "IMAGE_MASK_SERVICES",
	}
	ImageMask_value = map[string]int32{
		"IMAGE_MASK_UNSPECIFIED": 0,
		"IMAGE_MASK_MESSAGES":    1,
		"IMAGE_MASK_ENUMS":       2,
		"IMAGE_MASK_SERVICES":    3,
	}
)

func (x ImageMask) Enum() *ImageMask {
	p := new(ImageMask)
	*p = x
	return p
}

func (x ImageMask) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ImageMask) Descriptor() protoreflect.EnumDescriptor {
	return file_buf_alpha_registry_v1alpha1_image_proto_enumTypes[0].Descriptor()
}

func (ImageMask) Type() protoreflect.EnumType {
	return &file_buf_alpha_registry_v1alpha1_image_proto_enumTypes[0]
}

func (x ImageMask) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ImageMask.Descriptor instead.
func (ImageMask) EnumDescriptor() ([]byte, []int) {
	return file_buf_alpha_registry_v1alpha1_image_proto_rawDescGZIP(), []int{0}
}

type GetImageRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Owner      string `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Repository string `protobuf:"bytes,2,opt,name=repository,proto3" json:"repository,omitempty"`
	Reference  string `protobuf:"bytes,3,opt,name=reference,proto3" json:"reference,omitempty"`
	// Exclude files from imported buf modules in this image.
	ExcludeImports bool `protobuf:"varint,4,opt,name=exclude_imports,json=excludeImports,proto3" json:"exclude_imports,omitempty"`
	// Exclude source_code_info fields from each ImageFile.
	ExcludeSourceInfo bool `protobuf:"varint,5,opt,name=exclude_source_info,json=excludeSourceInfo,proto3" json:"exclude_source_info,omitempty"`
	// When specified the returned image will only contain the necessary files and
	// descriptors in those files to describe these types. Accepts messages, enums
	// and services. All types must be defined in the buf module, types in
	// dependencies are not accepted.
	//
	// At this time specifying `types` requires `exclude_source_info` to be set to
	// true.
	Types []string `protobuf:"bytes,6,rep,name=types,proto3" json:"types,omitempty"`
	// When not empty, the returned image's files will only include
	// *DescriptorProto fields for the elements specified here. The masks are
	// applied without regard for dependencies between types. For example, if
	// `IMAGE_MASK_MESSAGES` is specified without `IMAGE_MASK_ENUMS` the resulting
	// image will NOT contain enum definitions even if they are referenced from
	// message fields.
	IncludeMask []ImageMask `protobuf:"varint,7,rep,packed,name=include_mask,json=includeMask,proto3,enum=buf.alpha.registry.v1alpha1.ImageMask" json:"include_mask,omitempty"`
}

func (x *GetImageRequest) Reset() {
	*x = GetImageRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_buf_alpha_registry_v1alpha1_image_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetImageRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetImageRequest) ProtoMessage() {}

func (x *GetImageRequest) ProtoReflect() protoreflect.Message {
	mi := &file_buf_alpha_registry_v1alpha1_image_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetImageRequest.ProtoReflect.Descriptor instead.
func (*GetImageRequest) Descriptor() ([]byte, []int) {
	return file_buf_alpha_registry_v1alpha1_image_proto_rawDescGZIP(), []int{0}
}

func (x *GetImageRequest) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *GetImageRequest) GetRepository() string {
	if x != nil {
		return x.Repository
	}
	return ""
}

func (x *GetImageRequest) GetReference() string {
	if x != nil {
		return x.Reference
	}
	return ""
}

func (x *GetImageRequest) GetExcludeImports() bool {
	if x != nil {
		return x.ExcludeImports
	}
	return false
}

func (x *GetImageRequest) GetExcludeSourceInfo() bool {
	if x != nil {
		return x.ExcludeSourceInfo
	}
	return false
}

func (x *GetImageRequest) GetTypes() []string {
	if x != nil {
		return x.Types
	}
	return nil
}

func (x *GetImageRequest) GetIncludeMask() []ImageMask {
	if x != nil {
		return x.IncludeMask
	}
	return nil
}

type GetImageResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Image *v1.Image `protobuf:"bytes,1,opt,name=image,proto3" json:"image,omitempty"`
}

func (x *GetImageResponse) Reset() {
	*x = GetImageResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_buf_alpha_registry_v1alpha1_image_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetImageResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetImageResponse) ProtoMessage() {}

func (x *GetImageResponse) ProtoReflect() protoreflect.Message {
	mi := &file_buf_alpha_registry_v1alpha1_image_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetImageResponse.ProtoReflect.Descriptor instead.
func (*GetImageResponse) Descriptor() ([]byte, []int) {
	return file_buf_alpha_registry_v1alpha1_image_proto_rawDescGZIP(), []int{1}
}

func (x *GetImageResponse) GetImage() *v1.Image {
	if x != nil {
		return x.Image
	}
	return nil
}

var File_buf_alpha_registry_v1alpha1_image_proto protoreflect.FileDescriptor

var file_buf_alpha_registry_v1alpha1_image_proto_rawDesc = []byte{
	0x0a, 0x27, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2f, 0x72, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x69, 0x6d,
	0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1b, 0x62, 0x75, 0x66, 0x2e, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1e, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x2f, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x69, 0x6d, 0x61, 0x67, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9f, 0x02, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x49, 0x6d,
	0x61, 0x67, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x6f, 0x77,
	0x6e, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72,
	0x12, 0x1e, 0x0a, 0x0a, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79,
	0x12, 0x1c, 0x0a, 0x09, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x27,
	0x0a, 0x0f, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x5f, 0x69, 0x6d, 0x70, 0x6f, 0x72, 0x74,
	0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0e, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65,
	0x49, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x12, 0x2e, 0x0a, 0x13, 0x65, 0x78, 0x63, 0x6c, 0x75,
	0x64, 0x65, 0x5f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x11, 0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x53, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x79, 0x70, 0x65, 0x73,
	0x18, 0x06, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x74, 0x79, 0x70, 0x65, 0x73, 0x12, 0x49, 0x0a,
	0x0c, 0x69, 0x6e, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x5f, 0x6d, 0x61, 0x73, 0x6b, 0x18, 0x07, 0x20,
	0x03, 0x28, 0x0e, 0x32, 0x26, 0x2e, 0x62, 0x75, 0x66, 0x2e, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e,
	0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x2e, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x4d, 0x61, 0x73, 0x6b, 0x52, 0x0b, 0x69, 0x6e, 0x63,
	0x6c, 0x75, 0x64, 0x65, 0x4d, 0x61, 0x73, 0x6b, 0x22, 0x43, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x49,
	0x6d, 0x61, 0x67, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2f, 0x0a, 0x05,
	0x69, 0x6d, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x62, 0x75,
	0x66, 0x2e, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x05, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x2a, 0x6f, 0x0a,
	0x09, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x4d, 0x61, 0x73, 0x6b, 0x12, 0x1a, 0x0a, 0x16, 0x49, 0x4d,
	0x41, 0x47, 0x45, 0x5f, 0x4d, 0x41, 0x53, 0x4b, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49,
	0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x17, 0x0a, 0x13, 0x49, 0x4d, 0x41, 0x47, 0x45, 0x5f,
	0x4d, 0x41, 0x53, 0x4b, 0x5f, 0x4d, 0x45, 0x53, 0x53, 0x41, 0x47, 0x45, 0x53, 0x10, 0x01, 0x12,
	0x14, 0x0a, 0x10, 0x49, 0x4d, 0x41, 0x47, 0x45, 0x5f, 0x4d, 0x41, 0x53, 0x4b, 0x5f, 0x45, 0x4e,
	0x55, 0x4d, 0x53, 0x10, 0x02, 0x12, 0x17, 0x0a, 0x13, 0x49, 0x4d, 0x41, 0x47, 0x45, 0x5f, 0x4d,
	0x41, 0x53, 0x4b, 0x5f, 0x53, 0x45, 0x52, 0x56, 0x49, 0x43, 0x45, 0x53, 0x10, 0x03, 0x32, 0x7c,
	0x0a, 0x0c, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x6c,
	0x0a, 0x08, 0x47, 0x65, 0x74, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x12, 0x2c, 0x2e, 0x62, 0x75, 0x66,
	0x2e, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6d, 0x61, 0x67,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e, 0x62, 0x75, 0x66, 0x2e, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x03, 0x90, 0x02, 0x01, 0x42, 0x97, 0x02, 0x0a,
	0x1f, 0x63, 0x6f, 0x6d, 0x2e, 0x62, 0x75, 0x66, 0x2e, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x72,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x42, 0x0a, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x59,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x75, 0x66, 0x62, 0x75,
	0x69, 0x6c, 0x64, 0x2f, 0x62, 0x75, 0x66, 0x2f, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x2f,
	0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x62, 0x75, 0x66,
	0x2f, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x3b, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x79, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xa2, 0x02, 0x03, 0x42, 0x41, 0x52, 0xaa,
	0x02, 0x1b, 0x42, 0x75, 0x66, 0x2e, 0x41, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x52, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x72, 0x79, 0x2e, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xca, 0x02, 0x1b,
	0x42, 0x75, 0x66, 0x5c, 0x41, 0x6c, 0x70, 0x68, 0x61, 0x5c, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x72, 0x79, 0x5c, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xe2, 0x02, 0x27, 0x42, 0x75,
	0x66, 0x5c, 0x41, 0x6c, 0x70, 0x68, 0x61, 0x5c, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x5c, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x1e, 0x42, 0x75, 0x66, 0x3a, 0x3a, 0x41, 0x6c, 0x70,
	0x68, 0x61, 0x3a, 0x3a, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x3a, 0x3a, 0x56, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_buf_alpha_registry_v1alpha1_image_proto_rawDescOnce sync.Once
	file_buf_alpha_registry_v1alpha1_image_proto_rawDescData = file_buf_alpha_registry_v1alpha1_image_proto_rawDesc
)

func file_buf_alpha_registry_v1alpha1_image_proto_rawDescGZIP() []byte {
	file_buf_alpha_registry_v1alpha1_image_proto_rawDescOnce.Do(func() {
		file_buf_alpha_registry_v1alpha1_image_proto_rawDescData = protoimpl.X.CompressGZIP(file_buf_alpha_registry_v1alpha1_image_proto_rawDescData)
	})
	return file_buf_alpha_registry_v1alpha1_image_proto_rawDescData
}

var file_buf_alpha_registry_v1alpha1_image_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_buf_alpha_registry_v1alpha1_image_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_buf_alpha_registry_v1alpha1_image_proto_goTypes = []interface{}{
	(ImageMask)(0),           // 0: buf.alpha.registry.v1alpha1.ImageMask
	(*GetImageRequest)(nil),  // 1: buf.alpha.registry.v1alpha1.GetImageRequest
	(*GetImageResponse)(nil), // 2: buf.alpha.registry.v1alpha1.GetImageResponse
	(*v1.Image)(nil),         // 3: buf.alpha.image.v1.Image
}
var file_buf_alpha_registry_v1alpha1_image_proto_depIdxs = []int32{
	0, // 0: buf.alpha.registry.v1alpha1.GetImageRequest.include_mask:type_name -> buf.alpha.registry.v1alpha1.ImageMask
	3, // 1: buf.alpha.registry.v1alpha1.GetImageResponse.image:type_name -> buf.alpha.image.v1.Image
	1, // 2: buf.alpha.registry.v1alpha1.ImageService.GetImage:input_type -> buf.alpha.registry.v1alpha1.GetImageRequest
	2, // 3: buf.alpha.registry.v1alpha1.ImageService.GetImage:output_type -> buf.alpha.registry.v1alpha1.GetImageResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_buf_alpha_registry_v1alpha1_image_proto_init() }
func file_buf_alpha_registry_v1alpha1_image_proto_init() {
	if File_buf_alpha_registry_v1alpha1_image_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_buf_alpha_registry_v1alpha1_image_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetImageRequest); i {
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
		file_buf_alpha_registry_v1alpha1_image_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetImageResponse); i {
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
			RawDescriptor: file_buf_alpha_registry_v1alpha1_image_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_buf_alpha_registry_v1alpha1_image_proto_goTypes,
		DependencyIndexes: file_buf_alpha_registry_v1alpha1_image_proto_depIdxs,
		EnumInfos:         file_buf_alpha_registry_v1alpha1_image_proto_enumTypes,
		MessageInfos:      file_buf_alpha_registry_v1alpha1_image_proto_msgTypes,
	}.Build()
	File_buf_alpha_registry_v1alpha1_image_proto = out.File
	file_buf_alpha_registry_v1alpha1_image_proto_rawDesc = nil
	file_buf_alpha_registry_v1alpha1_image_proto_goTypes = nil
	file_buf_alpha_registry_v1alpha1_image_proto_depIdxs = nil
}
