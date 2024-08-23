package builder

import (
	"fmt"

	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/jhump/protoreflect/desc"
)

// FieldType represents the type of a field or extension. It can represent a
// message or enum type or any of the scalar types supported by protobufs.
//
// Message and enum types can reference a message or enum builder. A type that
// refers to a built message or enum descriptor is called an "imported" type.
//
// There are numerous factory methods for creating FieldType instances.
type FieldType struct {
	fieldType       descriptorpb.FieldDescriptorProto_Type
	foreignMsgType  *desc.MessageDescriptor
	localMsgType    *MessageBuilder
	foreignEnumType *desc.EnumDescriptor
	localEnumType   *EnumBuilder
}

// GetType returns the enum value indicating the type of the field. If the type
// is a message (or group) or enum type, GetTypeName provides the name of the
// referenced type.
func (ft *FieldType) GetType() descriptorpb.FieldDescriptorProto_Type {
	return ft.fieldType
}

// GetTypeName returns the fully-qualified name of the referenced message or
// enum type. It returns an empty string if this type does not represent a
// message or enum type.
func (ft *FieldType) GetTypeName() string {
	if ft.foreignMsgType != nil {
		return ft.foreignMsgType.GetFullyQualifiedName()
	} else if ft.foreignEnumType != nil {
		return ft.foreignEnumType.GetFullyQualifiedName()
	} else if ft.localMsgType != nil {
		return GetFullyQualifiedName(ft.localMsgType)
	} else if ft.localEnumType != nil {
		return GetFullyQualifiedName(ft.localEnumType)
	} else {
		return ""
	}
}

var scalarTypes = map[descriptorpb.FieldDescriptorProto_Type]*FieldType{
	descriptorpb.FieldDescriptorProto_TYPE_BOOL:     {fieldType: descriptorpb.FieldDescriptorProto_TYPE_BOOL},
	descriptorpb.FieldDescriptorProto_TYPE_INT32:    {fieldType: descriptorpb.FieldDescriptorProto_TYPE_INT32},
	descriptorpb.FieldDescriptorProto_TYPE_INT64:    {fieldType: descriptorpb.FieldDescriptorProto_TYPE_INT64},
	descriptorpb.FieldDescriptorProto_TYPE_SINT32:   {fieldType: descriptorpb.FieldDescriptorProto_TYPE_SINT32},
	descriptorpb.FieldDescriptorProto_TYPE_SINT64:   {fieldType: descriptorpb.FieldDescriptorProto_TYPE_SINT64},
	descriptorpb.FieldDescriptorProto_TYPE_UINT32:   {fieldType: descriptorpb.FieldDescriptorProto_TYPE_UINT32},
	descriptorpb.FieldDescriptorProto_TYPE_UINT64:   {fieldType: descriptorpb.FieldDescriptorProto_TYPE_UINT64},
	descriptorpb.FieldDescriptorProto_TYPE_FIXED32:  {fieldType: descriptorpb.FieldDescriptorProto_TYPE_FIXED32},
	descriptorpb.FieldDescriptorProto_TYPE_FIXED64:  {fieldType: descriptorpb.FieldDescriptorProto_TYPE_FIXED64},
	descriptorpb.FieldDescriptorProto_TYPE_SFIXED32: {fieldType: descriptorpb.FieldDescriptorProto_TYPE_SFIXED32},
	descriptorpb.FieldDescriptorProto_TYPE_SFIXED64: {fieldType: descriptorpb.FieldDescriptorProto_TYPE_SFIXED64},
	descriptorpb.FieldDescriptorProto_TYPE_FLOAT:    {fieldType: descriptorpb.FieldDescriptorProto_TYPE_FLOAT},
	descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:   {fieldType: descriptorpb.FieldDescriptorProto_TYPE_DOUBLE},
	descriptorpb.FieldDescriptorProto_TYPE_STRING:   {fieldType: descriptorpb.FieldDescriptorProto_TYPE_STRING},
	descriptorpb.FieldDescriptorProto_TYPE_BYTES:    {fieldType: descriptorpb.FieldDescriptorProto_TYPE_BYTES},
}

// FieldTypeScalar returns a FieldType for the given scalar type. If the given
// type is not scalar (e.g. it is a message, group, or enum) than this function
// will panic.
func FieldTypeScalar(t descriptorpb.FieldDescriptorProto_Type) *FieldType {
	if ft, ok := scalarTypes[t]; ok {
		return ft
	}
	panic(fmt.Sprintf("field %v is not scalar", t))
}

// FieldTypeInt32 returns a FieldType for the int32 scalar type.
func FieldTypeInt32() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_INT32)
}

// FieldTypeUInt32 returns a FieldType for the uint32 scalar type.
func FieldTypeUInt32() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_UINT32)
}

// FieldTypeSInt32 returns a FieldType for the sint32 scalar type.
func FieldTypeSInt32() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_SINT32)
}

// FieldTypeFixed32 returns a FieldType for the fixed32 scalar type.
func FieldTypeFixed32() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_FIXED32)
}

// FieldTypeSFixed32 returns a FieldType for the sfixed32 scalar type.
func FieldTypeSFixed32() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_SFIXED32)
}

// FieldTypeInt64 returns a FieldType for the int64 scalar type.
func FieldTypeInt64() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_INT64)
}

// FieldTypeUInt64 returns a FieldType for the uint64 scalar type.
func FieldTypeUInt64() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_UINT64)
}

// FieldTypeSInt64 returns a FieldType for the sint64 scalar type.
func FieldTypeSInt64() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_SINT64)
}

// FieldTypeFixed64 returns a FieldType for the fixed64 scalar type.
func FieldTypeFixed64() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_FIXED64)
}

// FieldTypeSFixed64 returns a FieldType for the sfixed64 scalar type.
func FieldTypeSFixed64() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_SFIXED64)
}

// FieldTypeFloat returns a FieldType for the float scalar type.
func FieldTypeFloat() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_FLOAT)
}

// FieldTypeDouble returns a FieldType for the double scalar type.
func FieldTypeDouble() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_DOUBLE)
}

// FieldTypeBool returns a FieldType for the bool scalar type.
func FieldTypeBool() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_BOOL)
}

// FieldTypeString returns a FieldType for the string scalar type.
func FieldTypeString() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_STRING)
}

// FieldTypeBytes returns a FieldType for the bytes scalar type.
func FieldTypeBytes() *FieldType {
	return FieldTypeScalar(descriptorpb.FieldDescriptorProto_TYPE_BYTES)
}

// FieldTypeMessage returns a FieldType for the given message type.
func FieldTypeMessage(mb *MessageBuilder) *FieldType {
	return &FieldType{
		fieldType:    descriptorpb.FieldDescriptorProto_TYPE_MESSAGE,
		localMsgType: mb,
	}
}

// FieldTypeImportedMessage returns a FieldType that references the given
// message descriptor.
func FieldTypeImportedMessage(md *desc.MessageDescriptor) *FieldType {
	return &FieldType{
		fieldType:      descriptorpb.FieldDescriptorProto_TYPE_MESSAGE,
		foreignMsgType: md,
	}
}

// FieldTypeEnum returns a FieldType for the given enum type.
func FieldTypeEnum(eb *EnumBuilder) *FieldType {
	return &FieldType{
		fieldType:     descriptorpb.FieldDescriptorProto_TYPE_ENUM,
		localEnumType: eb,
	}
}

// FieldTypeImportedEnum returns a FieldType that references the given enum
// descriptor.
func FieldTypeImportedEnum(ed *desc.EnumDescriptor) *FieldType {
	return &FieldType{
		fieldType:       descriptorpb.FieldDescriptorProto_TYPE_ENUM,
		foreignEnumType: ed,
	}
}

func fieldTypeFromDescriptor(fld *desc.FieldDescriptor) *FieldType {
	switch fld.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_GROUP:
		return &FieldType{fieldType: descriptorpb.FieldDescriptorProto_TYPE_GROUP, foreignMsgType: fld.GetMessageType()}
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		return FieldTypeImportedMessage(fld.GetMessageType())
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		return FieldTypeImportedEnum(fld.GetEnumType())
	default:
		return FieldTypeScalar(fld.GetType())
	}
}

// RpcType represents the type of an RPC request or response. The only allowed
// types are messages, but can be streams or unary messages.
//
// Message types can reference a message builder. A type that refers to a built
// message descriptor is called an "imported" type.
//
// To create an RpcType, see RpcTypeMessage and RpcTypeImportedMessage.
type RpcType struct {
	IsStream bool

	foreignType *desc.MessageDescriptor
	localType   *MessageBuilder
}

// RpcTypeMessage creates an RpcType that refers to the given message builder.
func RpcTypeMessage(mb *MessageBuilder, stream bool) *RpcType {
	return &RpcType{
		IsStream:  stream,
		localType: mb,
	}
}

// RpcTypeImportedMessage creates an RpcType that refers to the given message
// descriptor.
func RpcTypeImportedMessage(md *desc.MessageDescriptor, stream bool) *RpcType {
	return &RpcType{
		IsStream:    stream,
		foreignType: md,
	}
}

// GetTypeName returns the fully qualified name of the message type to which
// this RpcType refers.
func (rt *RpcType) GetTypeName() string {
	if rt.foreignType != nil {
		return rt.foreignType.GetFullyQualifiedName()
	} else {
		return GetFullyQualifiedName(rt.localType)
	}
}
