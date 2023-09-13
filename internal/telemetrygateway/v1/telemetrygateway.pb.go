// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.29.1
// 	protoc        (unknown)
// source: telemetrygateway.proto

package v1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
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

type RecordEventsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Payload:
	//
	//	*RecordEventsRequest_Metadata_
	//	*RecordEventsRequest_Event
	Payload isRecordEventsRequest_Payload `protobuf_oneof:"payload"`
}

func (x *RecordEventsRequest) Reset() {
	*x = RecordEventsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_telemetrygateway_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecordEventsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecordEventsRequest) ProtoMessage() {}

func (x *RecordEventsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_telemetrygateway_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecordEventsRequest.ProtoReflect.Descriptor instead.
func (*RecordEventsRequest) Descriptor() ([]byte, []int) {
	return file_telemetrygateway_proto_rawDescGZIP(), []int{0}
}

func (m *RecordEventsRequest) GetPayload() isRecordEventsRequest_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *RecordEventsRequest) GetMetadata() *RecordEventsRequest_Metadata {
	if x, ok := x.GetPayload().(*RecordEventsRequest_Metadata_); ok {
		return x.Metadata
	}
	return nil
}

func (x *RecordEventsRequest) GetEvent() *Event {
	if x, ok := x.GetPayload().(*RecordEventsRequest_Event); ok {
		return x.Event
	}
	return nil
}

type isRecordEventsRequest_Payload interface {
	isRecordEventsRequest_Payload()
}

type RecordEventsRequest_Metadata_ struct {
	// Metadata about the events being recorded.
	Metadata *RecordEventsRequest_Metadata `protobuf:"bytes,1,opt,name=metadata,proto3,oneof"`
}

type RecordEventsRequest_Event struct {
	// Event to record.
	Event *Event `protobuf:"bytes,2,opt,name=event,proto3,oneof"`
}

func (*RecordEventsRequest_Metadata_) isRecordEventsRequest_Payload() {}

func (*RecordEventsRequest_Event) isRecordEventsRequest_Payload() {}

type RecordEventsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RecordEventsResponse) Reset() {
	*x = RecordEventsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_telemetrygateway_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecordEventsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecordEventsResponse) ProtoMessage() {}

func (x *RecordEventsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_telemetrygateway_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecordEventsResponse.ProtoReflect.Descriptor instead.
func (*RecordEventsResponse) Descriptor() ([]byte, []int) {
	return file_telemetrygateway_proto_rawDescGZIP(), []int{1}
}

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Timestamp of when the original event was recorded.
	Timestamp *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	// Feature associated with the event in camelCase, e.g. 'myFeature'.
	Feature string `protobuf:"bytes,2,opt,name=feature,proto3" json:"feature,omitempty"`
	// Action associated with the event in camelCase, e.g. 'pageView'.
	Action string `protobuf:"bytes,3,opt,name=action,proto3" json:"action,omitempty"`
	// Source of the event.
	Source *EventSource `protobuf:"bytes,4,opt,name=source,proto3" json:"source,omitempty"`
	// Parameters of the event.
	Parameters *EventParameters `protobuf:"bytes,5,opt,name=parameters,proto3" json:"parameters,omitempty"`
	// Optional user associated with the event.
	User *EventUser `protobuf:"bytes,6,opt,name=user,proto3,oneof" json:"user,omitempty"`
	// Optional marketing campaign tracking parameters.
	MarketingTracking *EventMarketingTracking `protobuf:"bytes,7,opt,name=marketing_tracking,json=marketingTracking,proto3,oneof" json:"marketing_tracking,omitempty"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_telemetrygateway_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_telemetrygateway_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_telemetrygateway_proto_rawDescGZIP(), []int{2}
}

func (x *Event) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *Event) GetFeature() string {
	if x != nil {
		return x.Feature
	}
	return ""
}

func (x *Event) GetAction() string {
	if x != nil {
		return x.Action
	}
	return ""
}

func (x *Event) GetSource() *EventSource {
	if x != nil {
		return x.Source
	}
	return nil
}

func (x *Event) GetParameters() *EventParameters {
	if x != nil {
		return x.Parameters
	}
	return nil
}

func (x *Event) GetUser() *EventUser {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *Event) GetMarketingTracking() *EventMarketingTracking {
	if x != nil {
		return x.MarketingTracking
	}
	return nil
}

type EventSource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Version of the Sourcegraph instance that received the event.
	ServerVersion string `protobuf:"bytes,1,opt,name=server_version,json=serverVersion,proto3" json:"server_version,omitempty"`
	// Source client of the event.
	Client *string `protobuf:"bytes,2,opt,name=client,proto3,oneof" json:"client,omitempty"`
	// Version of the source client of the event.
	ClientVersion *string `protobuf:"bytes,3,opt,name=client_version,json=clientVersion,proto3,oneof" json:"client_version,omitempty"`
}

func (x *EventSource) Reset() {
	*x = EventSource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_telemetrygateway_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventSource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventSource) ProtoMessage() {}

func (x *EventSource) ProtoReflect() protoreflect.Message {
	mi := &file_telemetrygateway_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventSource.ProtoReflect.Descriptor instead.
func (*EventSource) Descriptor() ([]byte, []int) {
	return file_telemetrygateway_proto_rawDescGZIP(), []int{3}
}

func (x *EventSource) GetServerVersion() string {
	if x != nil {
		return x.ServerVersion
	}
	return ""
}

func (x *EventSource) GetClient() string {
	if x != nil && x.Client != nil {
		return *x.Client
	}
	return ""
}

func (x *EventSource) GetClientVersion() string {
	if x != nil && x.ClientVersion != nil {
		return *x.ClientVersion
	}
	return ""
}

type EventParameters struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Version of the event parameters, used for indicating the "shape" of this
	// event's metadata, beginning at 0.
	Version int32 `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	// Strictly typed metadata, restricted to integer values.
	Metadata map[string]int64 `protobuf:"bytes,2,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	// Additional potentially sensitive metadata - i.e. not restricted to integer
	// values. By default, this metadata should not be assumed to be unsafe for
	// export from an instance, and should only be exported on an allowlist basis.
	PrivateMetadata *structpb.Struct `protobuf:"bytes,3,opt,name=private_metadata,json=privateMetadata,proto3,oneof" json:"private_metadata,omitempty"`
	// Billing-related metadata.
	BillingMetadata *EventBillingMetadata `protobuf:"bytes,4,opt,name=billing_metadata,json=billingMetadata,proto3,oneof" json:"billing_metadata,omitempty"`
}

func (x *EventParameters) Reset() {
	*x = EventParameters{}
	if protoimpl.UnsafeEnabled {
		mi := &file_telemetrygateway_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventParameters) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventParameters) ProtoMessage() {}

func (x *EventParameters) ProtoReflect() protoreflect.Message {
	mi := &file_telemetrygateway_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventParameters.ProtoReflect.Descriptor instead.
func (*EventParameters) Descriptor() ([]byte, []int) {
	return file_telemetrygateway_proto_rawDescGZIP(), []int{4}
}

func (x *EventParameters) GetVersion() int32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *EventParameters) GetMetadata() map[string]int64 {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *EventParameters) GetPrivateMetadata() *structpb.Struct {
	if x != nil {
		return x.PrivateMetadata
	}
	return nil
}

func (x *EventParameters) GetBillingMetadata() *EventBillingMetadata {
	if x != nil {
		return x.BillingMetadata
	}
	return nil
}

type EventBillingMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Billing product ID associated with the event.
	Product string `protobuf:"bytes,1,opt,name=product,proto3" json:"product,omitempty"`
	// Billing category ID the event falls into.
	Category string `protobuf:"bytes,2,opt,name=category,proto3" json:"category,omitempty"`
}

func (x *EventBillingMetadata) Reset() {
	*x = EventBillingMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_telemetrygateway_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventBillingMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventBillingMetadata) ProtoMessage() {}

func (x *EventBillingMetadata) ProtoReflect() protoreflect.Message {
	mi := &file_telemetrygateway_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventBillingMetadata.ProtoReflect.Descriptor instead.
func (*EventBillingMetadata) Descriptor() ([]byte, []int) {
	return file_telemetrygateway_proto_rawDescGZIP(), []int{5}
}

func (x *EventBillingMetadata) GetProduct() string {
	if x != nil {
		return x.Product
	}
	return ""
}

func (x *EventBillingMetadata) GetCategory() string {
	if x != nil {
		return x.Category
	}
	return ""
}

type EventUser struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Database user ID of signed in user.
	//
	// We use an int64 as an ID because in Sourcegraph, database user IDs are
	// always integers.
	UserId int64 `protobuf:"varint,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	// Randomized unique identifier for client (i.e. stored in localstorage in web
	// client).
	AnonymousUserId string `protobuf:"bytes,2,opt,name=anonymous_user_id,json=anonymousUserId,proto3" json:"anonymous_user_id,omitempty"`
}

func (x *EventUser) Reset() {
	*x = EventUser{}
	if protoimpl.UnsafeEnabled {
		mi := &file_telemetrygateway_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventUser) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventUser) ProtoMessage() {}

func (x *EventUser) ProtoReflect() protoreflect.Message {
	mi := &file_telemetrygateway_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventUser.ProtoReflect.Descriptor instead.
func (*EventUser) Descriptor() ([]byte, []int) {
	return file_telemetrygateway_proto_rawDescGZIP(), []int{6}
}

func (x *EventUser) GetUserId() int64 {
	if x != nil {
		return x.UserId
	}
	return 0
}

func (x *EventUser) GetAnonymousUserId() string {
	if x != nil {
		return x.AnonymousUserId
	}
	return ""
}

type EventMarketingTracking struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Url             string `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	FirstSourceUrl  string `protobuf:"bytes,2,opt,name=first_source_url,json=firstSourceUrl,proto3" json:"first_source_url,omitempty"`
	CohortId        string `protobuf:"bytes,3,opt,name=cohort_id,json=cohortId,proto3" json:"cohort_id,omitempty"`
	Referrer        string `protobuf:"bytes,4,opt,name=referrer,proto3" json:"referrer,omitempty"`
	LastSourceUrl   string `protobuf:"bytes,5,opt,name=last_source_url,json=lastSourceUrl,proto3" json:"last_source_url,omitempty"`
	DeviceSessionId string `protobuf:"bytes,6,opt,name=device_session_id,json=deviceSessionId,proto3" json:"device_session_id,omitempty"`
	SessionReferrer string `protobuf:"bytes,7,opt,name=session_referrer,json=sessionReferrer,proto3" json:"session_referrer,omitempty"`
	SessionFirstUrl string `protobuf:"bytes,8,opt,name=session_first_url,json=sessionFirstUrl,proto3" json:"session_first_url,omitempty"`
}

func (x *EventMarketingTracking) Reset() {
	*x = EventMarketingTracking{}
	if protoimpl.UnsafeEnabled {
		mi := &file_telemetrygateway_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EventMarketingTracking) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EventMarketingTracking) ProtoMessage() {}

func (x *EventMarketingTracking) ProtoReflect() protoreflect.Message {
	mi := &file_telemetrygateway_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EventMarketingTracking.ProtoReflect.Descriptor instead.
func (*EventMarketingTracking) Descriptor() ([]byte, []int) {
	return file_telemetrygateway_proto_rawDescGZIP(), []int{7}
}

func (x *EventMarketingTracking) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *EventMarketingTracking) GetFirstSourceUrl() string {
	if x != nil {
		return x.FirstSourceUrl
	}
	return ""
}

func (x *EventMarketingTracking) GetCohortId() string {
	if x != nil {
		return x.CohortId
	}
	return ""
}

func (x *EventMarketingTracking) GetReferrer() string {
	if x != nil {
		return x.Referrer
	}
	return ""
}

func (x *EventMarketingTracking) GetLastSourceUrl() string {
	if x != nil {
		return x.LastSourceUrl
	}
	return ""
}

func (x *EventMarketingTracking) GetDeviceSessionId() string {
	if x != nil {
		return x.DeviceSessionId
	}
	return ""
}

func (x *EventMarketingTracking) GetSessionReferrer() string {
	if x != nil {
		return x.SessionReferrer
	}
	return ""
}

func (x *EventMarketingTracking) GetSessionFirstUrl() string {
	if x != nil {
		return x.SessionFirstUrl
	}
	return ""
}

type RecordEventsRequest_Metadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Self-reported identifier for analytics purposes. Currently, this is an
	// analytics prefix and the instance license key hashed together.
	AnalyticsIdentifier string `protobuf:"bytes,1,opt,name=analytics_identifier,json=analyticsIdentifier,proto3" json:"analytics_identifier,omitempty"`
}

func (x *RecordEventsRequest_Metadata) Reset() {
	*x = RecordEventsRequest_Metadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_telemetrygateway_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecordEventsRequest_Metadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecordEventsRequest_Metadata) ProtoMessage() {}

func (x *RecordEventsRequest_Metadata) ProtoReflect() protoreflect.Message {
	mi := &file_telemetrygateway_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecordEventsRequest_Metadata.ProtoReflect.Descriptor instead.
func (*RecordEventsRequest_Metadata) Descriptor() ([]byte, []int) {
	return file_telemetrygateway_proto_rawDescGZIP(), []int{0, 0}
}

func (x *RecordEventsRequest_Metadata) GetAnalyticsIdentifier() string {
	if x != nil {
		return x.AnalyticsIdentifier
	}
	return ""
}

var File_telemetrygateway_proto protoreflect.FileDescriptor

var file_telemetrygateway_proto_rawDesc = []byte{
	0x0a, 0x16, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x67, 0x61, 0x74, 0x65, 0x77,
	0x61, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65,
	0x74, 0x72, 0x79, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe4, 0x01, 0x0a,
	0x13, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x4f, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x31, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74,
	0x72, 0x79, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x63,
	0x6f, 0x72, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x48, 0x00, 0x52, 0x08, 0x6d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x32, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79,
	0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74,
	0x48, 0x00, 0x52, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x1a, 0x3d, 0x0a, 0x08, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x31, 0x0a, 0x14, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x74, 0x69,
	0x63, 0x73, 0x5f, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x13, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x74, 0x69, 0x63, 0x73, 0x49, 0x64,
	0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x42, 0x09, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x22, 0x16, 0x0a, 0x14, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0xad, 0x03, 0x0a, 0x05,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x38, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12,
	0x18, 0x0a, 0x07, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x38, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x20, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x67, 0x61, 0x74,
	0x65, 0x77, 0x61, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x53, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x44, 0x0a, 0x0a, 0x70,
	0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x24, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x67, 0x61, 0x74, 0x65, 0x77,
	0x61, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x50, 0x61, 0x72, 0x61, 0x6d,
	0x65, 0x74, 0x65, 0x72, 0x73, 0x52, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72,
	0x73, 0x12, 0x37, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1e, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x67, 0x61, 0x74, 0x65, 0x77,
	0x61, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x55, 0x73, 0x65, 0x72, 0x48,
	0x00, 0x52, 0x04, 0x75, 0x73, 0x65, 0x72, 0x88, 0x01, 0x01, 0x12, 0x5f, 0x0a, 0x12, 0x6d, 0x61,
	0x72, 0x6b, 0x65, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x74, 0x72, 0x61, 0x63, 0x6b, 0x69, 0x6e, 0x67,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74,
	0x72, 0x79, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x4d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x69, 0x6e, 0x67, 0x54, 0x72, 0x61, 0x63, 0x6b,
	0x69, 0x6e, 0x67, 0x48, 0x01, 0x52, 0x11, 0x6d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x69, 0x6e, 0x67,
	0x54, 0x72, 0x61, 0x63, 0x6b, 0x69, 0x6e, 0x67, 0x88, 0x01, 0x01, 0x42, 0x07, 0x0a, 0x05, 0x5f,
	0x75, 0x73, 0x65, 0x72, 0x42, 0x15, 0x0a, 0x13, 0x5f, 0x6d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x69,
	0x6e, 0x67, 0x5f, 0x74, 0x72, 0x61, 0x63, 0x6b, 0x69, 0x6e, 0x67, 0x22, 0x9b, 0x01, 0x0a, 0x0b,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x25, 0x0a, 0x0e, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x1b, 0x0a, 0x06, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x00, 0x52, 0x06, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x88, 0x01, 0x01, 0x12,
	0x2a, 0x0a, 0x0e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52, 0x0d, 0x63, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x42, 0x09, 0x0a, 0x07, 0x5f,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x42, 0x11, 0x0a, 0x0f, 0x5f, 0x63, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x86, 0x03, 0x0a, 0x0f, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x12, 0x18, 0x0a,
	0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x4e, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x32, 0x2e, 0x74, 0x65, 0x6c, 0x65,
	0x6d, 0x65, 0x74, 0x72, 0x79, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x76, 0x31, 0x2e,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x2e,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x6d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x47, 0x0a, 0x10, 0x70, 0x72, 0x69, 0x76, 0x61,
	0x74, 0x65, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x48, 0x00, 0x52, 0x0f, 0x70, 0x72,
	0x69, 0x76, 0x61, 0x74, 0x65, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x88, 0x01, 0x01,
	0x12, 0x59, 0x0a, 0x10, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x5f, 0x6d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x74, 0x65, 0x6c,
	0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x76, 0x31,
	0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x42, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x48, 0x01, 0x52, 0x0f, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x88, 0x01, 0x01, 0x1a, 0x3b, 0x0a, 0x0d, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x13, 0x0a, 0x11, 0x5f, 0x70, 0x72, 0x69,
	0x76, 0x61, 0x74, 0x65, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x42, 0x13, 0x0a,
	0x11, 0x5f, 0x62, 0x69, 0x6c, 0x6c, 0x69, 0x6e, 0x67, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x22, 0x4c, 0x0a, 0x14, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x42, 0x69, 0x6c, 0x6c, 0x69,
	0x6e, 0x67, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x72,
	0x6f, 0x64, 0x75, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x72, 0x6f,
	0x64, 0x75, 0x63, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x61, 0x74, 0x65, 0x67, 0x6f, 0x72, 0x79,
	0x22, 0x50, 0x0a, 0x09, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x55, 0x73, 0x65, 0x72, 0x12, 0x17, 0x0a,
	0x07, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06,
	0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x2a, 0x0a, 0x11, 0x61, 0x6e, 0x6f, 0x6e, 0x79, 0x6d,
	0x6f, 0x75, 0x73, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0f, 0x61, 0x6e, 0x6f, 0x6e, 0x79, 0x6d, 0x6f, 0x75, 0x73, 0x55, 0x73, 0x65, 0x72,
	0x49, 0x64, 0x22, 0xb8, 0x02, 0x0a, 0x16, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x4d, 0x61, 0x72, 0x6b,
	0x65, 0x74, 0x69, 0x6e, 0x67, 0x54, 0x72, 0x61, 0x63, 0x6b, 0x69, 0x6e, 0x67, 0x12, 0x10, 0x0a,
	0x03, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12,
	0x28, 0x0a, 0x10, 0x66, 0x69, 0x72, 0x73, 0x74, 0x5f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f,
	0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x66, 0x69, 0x72, 0x73, 0x74,
	0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x55, 0x72, 0x6c, 0x12, 0x1b, 0x0a, 0x09, 0x63, 0x6f, 0x68,
	0x6f, 0x72, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x6f,
	0x68, 0x6f, 0x72, 0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72,
	0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72,
	0x65, 0x72, 0x12, 0x26, 0x0a, 0x0f, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x6c, 0x61, 0x73,
	0x74, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x55, 0x72, 0x6c, 0x12, 0x2a, 0x0a, 0x11, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x5f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x53, 0x65, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x29, 0x0a, 0x10, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x5f, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65,
	0x72, 0x12, 0x2a, 0x0a, 0x11, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x66, 0x69, 0x72,
	0x73, 0x74, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x73, 0x65,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x72, 0x73, 0x74, 0x55, 0x72, 0x6c, 0x32, 0x83, 0x01,
	0x0a, 0x18, 0x54, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x79, 0x47, 0x61, 0x74, 0x65,
	0x77, 0x61, 0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x67, 0x0a, 0x0c, 0x52, 0x65,
	0x63, 0x6f, 0x72, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x28, 0x2e, 0x74, 0x65, 0x6c,
	0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x76, 0x31,
	0x2e, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x29, 0x2e, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79,
	0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x72,
	0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x28, 0x01, 0x42, 0x41, 0x5a, 0x3f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x67, 0x72, 0x61, 0x70, 0x68, 0x2f, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x67, 0x72, 0x61, 0x70, 0x68, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e,
	0x61, 0x6c, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x67, 0x61, 0x74, 0x65,
	0x77, 0x61, 0x79, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_telemetrygateway_proto_rawDescOnce sync.Once
	file_telemetrygateway_proto_rawDescData = file_telemetrygateway_proto_rawDesc
)

func file_telemetrygateway_proto_rawDescGZIP() []byte {
	file_telemetrygateway_proto_rawDescOnce.Do(func() {
		file_telemetrygateway_proto_rawDescData = protoimpl.X.CompressGZIP(file_telemetrygateway_proto_rawDescData)
	})
	return file_telemetrygateway_proto_rawDescData
}

var file_telemetrygateway_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_telemetrygateway_proto_goTypes = []interface{}{
	(*RecordEventsRequest)(nil),          // 0: telemetrygateway.v1.RecordEventsRequest
	(*RecordEventsResponse)(nil),         // 1: telemetrygateway.v1.RecordEventsResponse
	(*Event)(nil),                        // 2: telemetrygateway.v1.Event
	(*EventSource)(nil),                  // 3: telemetrygateway.v1.EventSource
	(*EventParameters)(nil),              // 4: telemetrygateway.v1.EventParameters
	(*EventBillingMetadata)(nil),         // 5: telemetrygateway.v1.EventBillingMetadata
	(*EventUser)(nil),                    // 6: telemetrygateway.v1.EventUser
	(*EventMarketingTracking)(nil),       // 7: telemetrygateway.v1.EventMarketingTracking
	(*RecordEventsRequest_Metadata)(nil), // 8: telemetrygateway.v1.RecordEventsRequest.Metadata
	nil,                                  // 9: telemetrygateway.v1.EventParameters.MetadataEntry
	(*timestamppb.Timestamp)(nil),        // 10: google.protobuf.Timestamp
	(*structpb.Struct)(nil),              // 11: google.protobuf.Struct
}
var file_telemetrygateway_proto_depIdxs = []int32{
	8,  // 0: telemetrygateway.v1.RecordEventsRequest.metadata:type_name -> telemetrygateway.v1.RecordEventsRequest.Metadata
	2,  // 1: telemetrygateway.v1.RecordEventsRequest.event:type_name -> telemetrygateway.v1.Event
	10, // 2: telemetrygateway.v1.Event.timestamp:type_name -> google.protobuf.Timestamp
	3,  // 3: telemetrygateway.v1.Event.source:type_name -> telemetrygateway.v1.EventSource
	4,  // 4: telemetrygateway.v1.Event.parameters:type_name -> telemetrygateway.v1.EventParameters
	6,  // 5: telemetrygateway.v1.Event.user:type_name -> telemetrygateway.v1.EventUser
	7,  // 6: telemetrygateway.v1.Event.marketing_tracking:type_name -> telemetrygateway.v1.EventMarketingTracking
	9,  // 7: telemetrygateway.v1.EventParameters.metadata:type_name -> telemetrygateway.v1.EventParameters.MetadataEntry
	11, // 8: telemetrygateway.v1.EventParameters.private_metadata:type_name -> google.protobuf.Struct
	5,  // 9: telemetrygateway.v1.EventParameters.billing_metadata:type_name -> telemetrygateway.v1.EventBillingMetadata
	0,  // 10: telemetrygateway.v1.TelemeteryGatewayService.RecordEvents:input_type -> telemetrygateway.v1.RecordEventsRequest
	1,  // 11: telemetrygateway.v1.TelemeteryGatewayService.RecordEvents:output_type -> telemetrygateway.v1.RecordEventsResponse
	11, // [11:12] is the sub-list for method output_type
	10, // [10:11] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_telemetrygateway_proto_init() }
func file_telemetrygateway_proto_init() {
	if File_telemetrygateway_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_telemetrygateway_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecordEventsRequest); i {
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
		file_telemetrygateway_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecordEventsResponse); i {
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
		file_telemetrygateway_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
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
		file_telemetrygateway_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventSource); i {
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
		file_telemetrygateway_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventParameters); i {
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
		file_telemetrygateway_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventBillingMetadata); i {
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
		file_telemetrygateway_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventUser); i {
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
		file_telemetrygateway_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EventMarketingTracking); i {
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
		file_telemetrygateway_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecordEventsRequest_Metadata); i {
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
	file_telemetrygateway_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*RecordEventsRequest_Metadata_)(nil),
		(*RecordEventsRequest_Event)(nil),
	}
	file_telemetrygateway_proto_msgTypes[2].OneofWrappers = []interface{}{}
	file_telemetrygateway_proto_msgTypes[3].OneofWrappers = []interface{}{}
	file_telemetrygateway_proto_msgTypes[4].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_telemetrygateway_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_telemetrygateway_proto_goTypes,
		DependencyIndexes: file_telemetrygateway_proto_depIdxs,
		MessageInfos:      file_telemetrygateway_proto_msgTypes,
	}.Build()
	File_telemetrygateway_proto = out.File
	file_telemetrygateway_proto_rawDesc = nil
	file_telemetrygateway_proto_goTypes = nil
	file_telemetrygateway_proto_depIdxs = nil
}
