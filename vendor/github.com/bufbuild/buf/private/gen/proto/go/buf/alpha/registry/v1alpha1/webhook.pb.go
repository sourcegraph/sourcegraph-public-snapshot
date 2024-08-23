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
// source: buf/alpha/registry/v1alpha1/webhook.proto

package registryv1alpha1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// WebhookEvent contains the currently supported webhook event types.
type WebhookEvent int32

const (
	// WEBHOOK_EVENT_UNSPECIFIED is a safe noop default for webhook events
	// subscription. It will trigger an error if trying to register a webhook with
	// this event.
	WebhookEvent_WEBHOOK_EVENT_UNSPECIFIED WebhookEvent = 0
	// WEBHOOK_EVENT_REPOSITORY_PUSH is emitted whenever a successful buf push is
	// completed for a specific repository.
	WebhookEvent_WEBHOOK_EVENT_REPOSITORY_PUSH WebhookEvent = 1
)

// Enum value maps for WebhookEvent.
var (
	WebhookEvent_name = map[int32]string{
		0: "WEBHOOK_EVENT_UNSPECIFIED",
		1: "WEBHOOK_EVENT_REPOSITORY_PUSH",
	}
	WebhookEvent_value = map[string]int32{
		"WEBHOOK_EVENT_UNSPECIFIED":     0,
		"WEBHOOK_EVENT_REPOSITORY_PUSH": 1,
	}
)

func (x WebhookEvent) Enum() *WebhookEvent {
	p := new(WebhookEvent)
	*p = x
	return p
}

func (x WebhookEvent) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (WebhookEvent) Descriptor() protoreflect.EnumDescriptor {
	return file_buf_alpha_registry_v1alpha1_webhook_proto_enumTypes[0].Descriptor()
}

func (WebhookEvent) Type() protoreflect.EnumType {
	return &file_buf_alpha_registry_v1alpha1_webhook_proto_enumTypes[0]
}

func (x WebhookEvent) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use WebhookEvent.Descriptor instead.
func (WebhookEvent) EnumDescriptor() ([]byte, []int) {
	return file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescGZIP(), []int{0}
}

// CreateWebhookRequest is the proto request representation of a
// webhook request body.
type CreateWebhookRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The event to subscribe to for the given repository.
	WebhookEvent WebhookEvent `protobuf:"varint,1,opt,name=webhook_event,json=webhookEvent,proto3,enum=buf.alpha.registry.v1alpha1.WebhookEvent" json:"webhook_event,omitempty"`
	// The owner name of the repository in the corresponding subscription request.
	OwnerName string `protobuf:"bytes,2,opt,name=owner_name,json=ownerName,proto3" json:"owner_name,omitempty"`
	// The repository name that the subscriber wishes create a subscription for.
	RepositoryName string `protobuf:"bytes,3,opt,name=repository_name,json=repositoryName,proto3" json:"repository_name,omitempty"`
	// The subscriber's callback URL where notifications should be delivered.
	CallbackUrl string `protobuf:"bytes,4,opt,name=callback_url,json=callbackUrl,proto3" json:"callback_url,omitempty"`
}

func (x *CreateWebhookRequest) Reset() {
	*x = CreateWebhookRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateWebhookRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateWebhookRequest) ProtoMessage() {}

func (x *CreateWebhookRequest) ProtoReflect() protoreflect.Message {
	mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateWebhookRequest.ProtoReflect.Descriptor instead.
func (*CreateWebhookRequest) Descriptor() ([]byte, []int) {
	return file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescGZIP(), []int{0}
}

func (x *CreateWebhookRequest) GetWebhookEvent() WebhookEvent {
	if x != nil {
		return x.WebhookEvent
	}
	return WebhookEvent_WEBHOOK_EVENT_UNSPECIFIED
}

func (x *CreateWebhookRequest) GetOwnerName() string {
	if x != nil {
		return x.OwnerName
	}
	return ""
}

func (x *CreateWebhookRequest) GetRepositoryName() string {
	if x != nil {
		return x.RepositoryName
	}
	return ""
}

func (x *CreateWebhookRequest) GetCallbackUrl() string {
	if x != nil {
		return x.CallbackUrl
	}
	return ""
}

// CreateWebhookResponse is the proto response representation
// of a webhook request.
type CreateWebhookResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Created webhook subscription.
	Webhook *Webhook `protobuf:"bytes,1,opt,name=webhook,proto3" json:"webhook,omitempty"`
}

func (x *CreateWebhookResponse) Reset() {
	*x = CreateWebhookResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateWebhookResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateWebhookResponse) ProtoMessage() {}

func (x *CreateWebhookResponse) ProtoReflect() protoreflect.Message {
	mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateWebhookResponse.ProtoReflect.Descriptor instead.
func (*CreateWebhookResponse) Descriptor() ([]byte, []int) {
	return file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescGZIP(), []int{1}
}

func (x *CreateWebhookResponse) GetWebhook() *Webhook {
	if x != nil {
		return x.Webhook
	}
	return nil
}

// DeleteWebhookRequest is the request for unsubscribing to a webhook.
type DeleteWebhookRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The id of the webhook subscription to delete.
	WebhookId string `protobuf:"bytes,1,opt,name=webhook_id,json=webhookId,proto3" json:"webhook_id,omitempty"`
}

func (x *DeleteWebhookRequest) Reset() {
	*x = DeleteWebhookRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteWebhookRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteWebhookRequest) ProtoMessage() {}

func (x *DeleteWebhookRequest) ProtoReflect() protoreflect.Message {
	mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteWebhookRequest.ProtoReflect.Descriptor instead.
func (*DeleteWebhookRequest) Descriptor() ([]byte, []int) {
	return file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescGZIP(), []int{2}
}

func (x *DeleteWebhookRequest) GetWebhookId() string {
	if x != nil {
		return x.WebhookId
	}
	return ""
}

// DeleteWebhookResponse is the response for unsubscribing
// from a webhook.
type DeleteWebhookResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DeleteWebhookResponse) Reset() {
	*x = DeleteWebhookResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteWebhookResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteWebhookResponse) ProtoMessage() {}

func (x *DeleteWebhookResponse) ProtoReflect() protoreflect.Message {
	mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteWebhookResponse.ProtoReflect.Descriptor instead.
func (*DeleteWebhookResponse) Descriptor() ([]byte, []int) {
	return file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescGZIP(), []int{3}
}

// ListWebhooksRequest is the request to get the
// list of subscribed webhooks for a given repository.
type ListWebhooksRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The repository name given in the corresponding subscription request.
	RepositoryName string `protobuf:"bytes,1,opt,name=repository_name,json=repositoryName,proto3" json:"repository_name,omitempty"`
	// The owner associated with the repository.
	OwnerName string `protobuf:"bytes,2,opt,name=owner_name,json=ownerName,proto3" json:"owner_name,omitempty"`
	// The page token for paginating.
	PageToken string `protobuf:"bytes,3,opt,name=page_token,json=pageToken,proto3" json:"page_token,omitempty"`
}

func (x *ListWebhooksRequest) Reset() {
	*x = ListWebhooksRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListWebhooksRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListWebhooksRequest) ProtoMessage() {}

func (x *ListWebhooksRequest) ProtoReflect() protoreflect.Message {
	mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListWebhooksRequest.ProtoReflect.Descriptor instead.
func (*ListWebhooksRequest) Descriptor() ([]byte, []int) {
	return file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescGZIP(), []int{4}
}

func (x *ListWebhooksRequest) GetRepositoryName() string {
	if x != nil {
		return x.RepositoryName
	}
	return ""
}

func (x *ListWebhooksRequest) GetOwnerName() string {
	if x != nil {
		return x.OwnerName
	}
	return ""
}

func (x *ListWebhooksRequest) GetPageToken() string {
	if x != nil {
		return x.PageToken
	}
	return ""
}

// ListWebhooksResponse is the response for the list of
// subscribed webhooks for a given repository.
type ListWebhooksResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The list of subscribed webhooks for a given repository.
	Webhooks []*Webhook `protobuf:"bytes,1,rep,name=webhooks,proto3" json:"webhooks,omitempty"`
	// The next page token for paginating.
	NextPageToken string `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
}

func (x *ListWebhooksResponse) Reset() {
	*x = ListWebhooksResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListWebhooksResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListWebhooksResponse) ProtoMessage() {}

func (x *ListWebhooksResponse) ProtoReflect() protoreflect.Message {
	mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListWebhooksResponse.ProtoReflect.Descriptor instead.
func (*ListWebhooksResponse) Descriptor() ([]byte, []int) {
	return file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescGZIP(), []int{5}
}

func (x *ListWebhooksResponse) GetWebhooks() []*Webhook {
	if x != nil {
		return x.Webhooks
	}
	return nil
}

func (x *ListWebhooksResponse) GetNextPageToken() string {
	if x != nil {
		return x.NextPageToken
	}
	return ""
}

// Webhook is the representation of a webhook repository event subscription.
type Webhook struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The event associated with the subscription id.
	Event WebhookEvent `protobuf:"varint,1,opt,name=event,proto3,enum=buf.alpha.registry.v1alpha1.WebhookEvent" json:"event,omitempty"`
	// The id of the associated subscription.
	WebhookId string `protobuf:"bytes,2,opt,name=webhook_id,json=webhookId,proto3" json:"webhook_id,omitempty"`
	// The webhook creation timestamp.
	CreateTime *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	// The webhook last updated timestamp.
	UpdateTime *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty"`
	// The webhook repository name.
	RepositoryName string `protobuf:"bytes,5,opt,name=repository_name,json=repositoryName,proto3" json:"repository_name,omitempty"`
	// The webhook repository owner name.
	OwnerName string `protobuf:"bytes,6,opt,name=owner_name,json=ownerName,proto3" json:"owner_name,omitempty"`
	// The subscriber's callback URL where notifications are delivered. Currently
	// we only support Connect-powered backends with application/proto as the
	// content type. Make sure that your URL ends with
	// "/buf.alpha.webhook.v1alpha1.EventService/Event". For more information
	// about Connect, see https://connect.build.
	CallbackUrl string `protobuf:"bytes,7,opt,name=callback_url,json=callbackUrl,proto3" json:"callback_url,omitempty"`
}

func (x *Webhook) Reset() {
	*x = Webhook{}
	if protoimpl.UnsafeEnabled {
		mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Webhook) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Webhook) ProtoMessage() {}

func (x *Webhook) ProtoReflect() protoreflect.Message {
	mi := &file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Webhook.ProtoReflect.Descriptor instead.
func (*Webhook) Descriptor() ([]byte, []int) {
	return file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescGZIP(), []int{6}
}

func (x *Webhook) GetEvent() WebhookEvent {
	if x != nil {
		return x.Event
	}
	return WebhookEvent_WEBHOOK_EVENT_UNSPECIFIED
}

func (x *Webhook) GetWebhookId() string {
	if x != nil {
		return x.WebhookId
	}
	return ""
}

func (x *Webhook) GetCreateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CreateTime
	}
	return nil
}

func (x *Webhook) GetUpdateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.UpdateTime
	}
	return nil
}

func (x *Webhook) GetRepositoryName() string {
	if x != nil {
		return x.RepositoryName
	}
	return ""
}

func (x *Webhook) GetOwnerName() string {
	if x != nil {
		return x.OwnerName
	}
	return ""
}

func (x *Webhook) GetCallbackUrl() string {
	if x != nil {
		return x.CallbackUrl
	}
	return ""
}

var File_buf_alpha_registry_v1alpha1_webhook_proto protoreflect.FileDescriptor

var file_buf_alpha_registry_v1alpha1_webhook_proto_rawDesc = []byte{
	0x0a, 0x29, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2f, 0x72, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x77, 0x65,
	0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1b, 0x62, 0x75, 0x66,
	0x2e, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e,
	0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd1, 0x01, 0x0a, 0x14, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x4e, 0x0a, 0x0d, 0x77, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x5f, 0x65, 0x76,
	0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x29, 0x2e, 0x62, 0x75, 0x66, 0x2e,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x52, 0x0c, 0x77, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x4e, 0x61, 0x6d,
	0x65, 0x12, 0x27, 0x0a, 0x0f, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x5f,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x72, 0x65, 0x70, 0x6f,
	0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x61,
	0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0b, 0x63, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x55, 0x72, 0x6c, 0x22, 0x57, 0x0a,
	0x15, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3e, 0x0a, 0x07, 0x77, 0x65, 0x62, 0x68, 0x6f, 0x6f,
	0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x62, 0x75, 0x66, 0x2e, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x07, 0x77,
	0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x22, 0x35, 0x0a, 0x14, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d,
	0x0a, 0x0a, 0x77, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x77, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x49, 0x64, 0x22, 0x17, 0x0a,
	0x15, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x7c, 0x0a, 0x13, 0x4c, 0x69, 0x73, 0x74, 0x57, 0x65,
	0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a,
	0x0f, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f,
	0x72, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x5f,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6f, 0x77, 0x6e, 0x65,
	0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f,
	0x6b, 0x65, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x61, 0x67, 0x65, 0x54,
	0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x80, 0x01, 0x0a, 0x14, 0x4c, 0x69, 0x73, 0x74, 0x57, 0x65, 0x62,
	0x68, 0x6f, 0x6f, 0x6b, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x40, 0x0a,
	0x08, 0x77, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x24, 0x2e, 0x62, 0x75, 0x66, 0x2e, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x72, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x57, 0x65,
	0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x08, 0x77, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x73, 0x12,
	0x26, 0x0a, 0x0f, 0x6e, 0x65, 0x78, 0x74, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b,
	0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x6e, 0x65, 0x78, 0x74, 0x50, 0x61,
	0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0xce, 0x02, 0x0a, 0x07, 0x57, 0x65, 0x62, 0x68,
	0x6f, 0x6f, 0x6b, 0x12, 0x3f, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x29, 0x2e, 0x62, 0x75, 0x66, 0x2e, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x72,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x2e, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x05, 0x65,
	0x76, 0x65, 0x6e, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x77, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x5f,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x77, 0x65, 0x62, 0x68, 0x6f, 0x6f,
	0x6b, 0x49, 0x64, 0x12, 0x3b, 0x0a, 0x0b, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65,
	0x12, 0x3b, 0x0a, 0x0b, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x52, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x27, 0x0a,
	0x0f, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f,
	0x72, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x5f,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6f, 0x77, 0x6e, 0x65,
	0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63,
	0x6b, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x61, 0x6c,
	0x6c, 0x62, 0x61, 0x63, 0x6b, 0x55, 0x72, 0x6c, 0x2a, 0x50, 0x0a, 0x0c, 0x57, 0x65, 0x62, 0x68,
	0x6f, 0x6f, 0x6b, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x1d, 0x0a, 0x19, 0x57, 0x45, 0x42, 0x48,
	0x4f, 0x4f, 0x4b, 0x5f, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43,
	0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x21, 0x0a, 0x1d, 0x57, 0x45, 0x42, 0x48, 0x4f,
	0x4f, 0x4b, 0x5f, 0x45, 0x56, 0x45, 0x4e, 0x54, 0x5f, 0x52, 0x45, 0x50, 0x4f, 0x53, 0x49, 0x54,
	0x4f, 0x52, 0x59, 0x5f, 0x50, 0x55, 0x53, 0x48, 0x10, 0x01, 0x32, 0x84, 0x03, 0x0a, 0x0e, 0x57,
	0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x7b, 0x0a,
	0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x12, 0x31,
	0x2e, 0x62, 0x75, 0x66, 0x2e, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x32, 0x2e, 0x62, 0x75, 0x66, 0x2e, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x72, 0x65,
	0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x03, 0x90, 0x02, 0x02, 0x12, 0x7b, 0x0a, 0x0d, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x12, 0x31, 0x2e, 0x62, 0x75,
	0x66, 0x2e, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79,
	0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x32,
	0x2e, 0x62, 0x75, 0x66, 0x2e, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x44, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x03, 0x90, 0x02, 0x02, 0x12, 0x78, 0x0a, 0x0c, 0x4c, 0x69, 0x73, 0x74, 0x57,
	0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x73, 0x12, 0x30, 0x2e, 0x62, 0x75, 0x66, 0x2e, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f,
	0x6b, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x31, 0x2e, 0x62, 0x75, 0x66, 0x2e,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x57, 0x65, 0x62, 0x68,
	0x6f, 0x6f, 0x6b, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x03, 0x90, 0x02,
	0x01, 0x42, 0x99, 0x02, 0x0a, 0x1f, 0x63, 0x6f, 0x6d, 0x2e, 0x62, 0x75, 0x66, 0x2e, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x2e, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x42, 0x0c, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x59, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x62, 0x75, 0x66, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2f, 0x62, 0x75, 0x66, 0x2f, 0x70,
	0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x67, 0x6f, 0x2f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2f, 0x72, 0x65,
	0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x3b,
	0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0xa2, 0x02, 0x03, 0x42, 0x41, 0x52, 0xaa, 0x02, 0x1b, 0x42, 0x75, 0x66, 0x2e, 0x41, 0x6c, 0x70,
	0x68, 0x61, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x2e, 0x56, 0x31, 0x61, 0x6c,
	0x70, 0x68, 0x61, 0x31, 0xca, 0x02, 0x1b, 0x42, 0x75, 0x66, 0x5c, 0x41, 0x6c, 0x70, 0x68, 0x61,
	0x5c, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x5c, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0xe2, 0x02, 0x27, 0x42, 0x75, 0x66, 0x5c, 0x41, 0x6c, 0x70, 0x68, 0x61, 0x5c, 0x52,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x79, 0x5c, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31,
	0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x1e, 0x42,
	0x75, 0x66, 0x3a, 0x3a, 0x41, 0x6c, 0x70, 0x68, 0x61, 0x3a, 0x3a, 0x52, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x72, 0x79, 0x3a, 0x3a, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescOnce sync.Once
	file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescData = file_buf_alpha_registry_v1alpha1_webhook_proto_rawDesc
)

func file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescGZIP() []byte {
	file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescOnce.Do(func() {
		file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescData = protoimpl.X.CompressGZIP(file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescData)
	})
	return file_buf_alpha_registry_v1alpha1_webhook_proto_rawDescData
}

var file_buf_alpha_registry_v1alpha1_webhook_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_buf_alpha_registry_v1alpha1_webhook_proto_goTypes = []interface{}{
	(WebhookEvent)(0),             // 0: buf.alpha.registry.v1alpha1.WebhookEvent
	(*CreateWebhookRequest)(nil),  // 1: buf.alpha.registry.v1alpha1.CreateWebhookRequest
	(*CreateWebhookResponse)(nil), // 2: buf.alpha.registry.v1alpha1.CreateWebhookResponse
	(*DeleteWebhookRequest)(nil),  // 3: buf.alpha.registry.v1alpha1.DeleteWebhookRequest
	(*DeleteWebhookResponse)(nil), // 4: buf.alpha.registry.v1alpha1.DeleteWebhookResponse
	(*ListWebhooksRequest)(nil),   // 5: buf.alpha.registry.v1alpha1.ListWebhooksRequest
	(*ListWebhooksResponse)(nil),  // 6: buf.alpha.registry.v1alpha1.ListWebhooksResponse
	(*Webhook)(nil),               // 7: buf.alpha.registry.v1alpha1.Webhook
	(*timestamppb.Timestamp)(nil), // 8: google.protobuf.Timestamp
}
var file_buf_alpha_registry_v1alpha1_webhook_proto_depIdxs = []int32{
	0, // 0: buf.alpha.registry.v1alpha1.CreateWebhookRequest.webhook_event:type_name -> buf.alpha.registry.v1alpha1.WebhookEvent
	7, // 1: buf.alpha.registry.v1alpha1.CreateWebhookResponse.webhook:type_name -> buf.alpha.registry.v1alpha1.Webhook
	7, // 2: buf.alpha.registry.v1alpha1.ListWebhooksResponse.webhooks:type_name -> buf.alpha.registry.v1alpha1.Webhook
	0, // 3: buf.alpha.registry.v1alpha1.Webhook.event:type_name -> buf.alpha.registry.v1alpha1.WebhookEvent
	8, // 4: buf.alpha.registry.v1alpha1.Webhook.create_time:type_name -> google.protobuf.Timestamp
	8, // 5: buf.alpha.registry.v1alpha1.Webhook.update_time:type_name -> google.protobuf.Timestamp
	1, // 6: buf.alpha.registry.v1alpha1.WebhookService.CreateWebhook:input_type -> buf.alpha.registry.v1alpha1.CreateWebhookRequest
	3, // 7: buf.alpha.registry.v1alpha1.WebhookService.DeleteWebhook:input_type -> buf.alpha.registry.v1alpha1.DeleteWebhookRequest
	5, // 8: buf.alpha.registry.v1alpha1.WebhookService.ListWebhooks:input_type -> buf.alpha.registry.v1alpha1.ListWebhooksRequest
	2, // 9: buf.alpha.registry.v1alpha1.WebhookService.CreateWebhook:output_type -> buf.alpha.registry.v1alpha1.CreateWebhookResponse
	4, // 10: buf.alpha.registry.v1alpha1.WebhookService.DeleteWebhook:output_type -> buf.alpha.registry.v1alpha1.DeleteWebhookResponse
	6, // 11: buf.alpha.registry.v1alpha1.WebhookService.ListWebhooks:output_type -> buf.alpha.registry.v1alpha1.ListWebhooksResponse
	9, // [9:12] is the sub-list for method output_type
	6, // [6:9] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_buf_alpha_registry_v1alpha1_webhook_proto_init() }
func file_buf_alpha_registry_v1alpha1_webhook_proto_init() {
	if File_buf_alpha_registry_v1alpha1_webhook_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateWebhookRequest); i {
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
		file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateWebhookResponse); i {
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
		file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteWebhookRequest); i {
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
		file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteWebhookResponse); i {
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
		file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListWebhooksRequest); i {
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
		file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListWebhooksResponse); i {
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
		file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Webhook); i {
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
			RawDescriptor: file_buf_alpha_registry_v1alpha1_webhook_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_buf_alpha_registry_v1alpha1_webhook_proto_goTypes,
		DependencyIndexes: file_buf_alpha_registry_v1alpha1_webhook_proto_depIdxs,
		EnumInfos:         file_buf_alpha_registry_v1alpha1_webhook_proto_enumTypes,
		MessageInfos:      file_buf_alpha_registry_v1alpha1_webhook_proto_msgTypes,
	}.Build()
	File_buf_alpha_registry_v1alpha1_webhook_proto = out.File
	file_buf_alpha_registry_v1alpha1_webhook_proto_rawDesc = nil
	file_buf_alpha_registry_v1alpha1_webhook_proto_goTypes = nil
	file_buf_alpha_registry_v1alpha1_webhook_proto_depIdxs = nil
}
