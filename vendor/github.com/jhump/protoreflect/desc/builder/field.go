package builder

import (
	"fmt"
	"strings"
	"unicode"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/internal"
)

// FieldBuilder is a builder used to construct a desc.FieldDescriptor. A field
// builder is used to create fields and extensions as well as map entry
// messages. It is also used to link groups (defined via a message builder) into
// an enclosing message, associating it with a group field.  A non-extension
// field builder *must* be added to a message before calling its Build() method.
//
// To create a new FieldBuilder, use NewField, NewMapField, NewGroupField,
// NewExtension, or NewExtensionImported (depending on the type of field being
// built).
type FieldBuilder struct {
	baseBuilder
	number int32

	// msgType is populated for fields that have a "private" message type that
	// isn't expected to be referenced elsewhere. This happens for map fields,
	// where the private message type represents the map entry, and for group
	// fields.
	msgType   *MessageBuilder
	fieldType *FieldType

	Options        *descriptorpb.FieldOptions
	Label          descriptorpb.FieldDescriptorProto_Label
	Proto3Optional bool
	Default        string
	JsonName       string

	foreignExtendee *desc.MessageDescriptor
	localExtendee   *MessageBuilder
}

// NewField creates a new FieldBuilder for a non-extension field with the given
// name and type. To create a map or group field, see NewMapField or
// NewGroupField respectively.
//
// The new field will be optional. See SetLabel, SetRepeated, and SetRequired
// for changing this aspect of the field. The new field's tag will be zero,
// which means it will be auto-assigned when the descriptor is built. Use
// SetNumber or TrySetNumber to assign an explicit tag number.
func NewField(name string, typ *FieldType) *FieldBuilder {
	flb := &FieldBuilder{
		baseBuilder: baseBuilderWithName(name),
		fieldType:   typ,
	}
	return flb
}

// NewMapField creates a new FieldBuilder for a non-extension field with the
// given name and whose type is a map of the given key and value types. Map keys
// can be any of the scalar integer types, booleans, or strings. If any other
// type is specified, this function will panic. Map values cannot be groups: if
// a group type is specified, this function will panic.
//
// When this field is added to a message, the associated map entry message type
// will also be added.
//
// The new field's tag will be zero, which means it will be auto-assigned when
// the descriptor is built. Use SetNumber or TrySetNumber to assign an explicit
// tag number.
func NewMapField(name string, keyTyp, valTyp *FieldType) *FieldBuilder {
	switch keyTyp.fieldType {
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL,
		descriptorpb.FieldDescriptorProto_TYPE_STRING,
		descriptorpb.FieldDescriptorProto_TYPE_INT32, descriptorpb.FieldDescriptorProto_TYPE_INT64,
		descriptorpb.FieldDescriptorProto_TYPE_SINT32, descriptorpb.FieldDescriptorProto_TYPE_SINT64,
		descriptorpb.FieldDescriptorProto_TYPE_UINT32, descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_FIXED32, descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32, descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		// allowed
	default:
		panic(fmt.Sprintf("Map types cannot have keys of type %v", keyTyp.fieldType))
	}
	if valTyp.fieldType == descriptorpb.FieldDescriptorProto_TYPE_GROUP {
		panic(fmt.Sprintf("Map types cannot have values of type %v", valTyp.fieldType))
	}
	entryMsg := NewMessage(entryTypeName(name))
	keyFlb := NewField("key", keyTyp)
	keyFlb.number = 1
	valFlb := NewField("value", valTyp)
	valFlb.number = 2
	entryMsg.AddField(keyFlb)
	entryMsg.AddField(valFlb)
	entryMsg.Options = &descriptorpb.MessageOptions{MapEntry: proto.Bool(true)}

	flb := NewField(name, FieldTypeMessage(entryMsg)).
		SetLabel(descriptorpb.FieldDescriptorProto_LABEL_REPEATED)
	flb.msgType = entryMsg
	entryMsg.setParent(flb)
	return flb
}

// NewGroupField creates a new FieldBuilder for a non-extension field whose type
// is a group with the given definition. The given message's name must start
// with a capital letter, and the resulting field will have the same name but
// converted to all lower-case. If a message is given with a name that starts
// with a lower-case letter, this function will panic.
//
// When this field is added to a message, the associated group message type will
// also be added.
//
// The new field will be optional. See SetLabel, SetRepeated, and SetRequired
// for changing this aspect of the field. The new field's tag will be zero,
// which means it will be auto-assigned when the descriptor is built. Use
// SetNumber or TrySetNumber to assign an explicit tag number.
func NewGroupField(mb *MessageBuilder) *FieldBuilder {
	if !unicode.IsUpper(rune(mb.name[0])) {
		panic(fmt.Sprintf("group name %s must start with a capital letter", mb.name))
	}
	Unlink(mb)

	ft := &FieldType{
		fieldType:    descriptorpb.FieldDescriptorProto_TYPE_GROUP,
		localMsgType: mb,
	}
	fieldName := strings.ToLower(mb.GetName())
	flb := NewField(fieldName, ft)
	flb.msgType = mb
	mb.setParent(flb)
	return flb
}

// NewExtension creates a new FieldBuilder for an extension field with the given
// name, tag, type, and extendee. The extendee given is a message builder.
//
// The new field will be optional. See SetLabel and SetRepeated for changing
// this aspect of the field.
func NewExtension(name string, tag int32, typ *FieldType, extendee *MessageBuilder) *FieldBuilder {
	if extendee == nil {
		panic("extendee cannot be nil")
	}
	flb := NewField(name, typ).SetNumber(tag)
	flb.localExtendee = extendee
	return flb
}

// NewExtensionImported creates a new FieldBuilder for an extension field with
// the given name, tag, type, and extendee. The extendee given is a message
// descriptor.
//
// The new field will be optional. See SetLabel and SetRepeated for changing
// this aspect of the field.
func NewExtensionImported(name string, tag int32, typ *FieldType, extendee *desc.MessageDescriptor) *FieldBuilder {
	if extendee == nil {
		panic("extendee cannot be nil")
	}
	flb := NewField(name, typ).SetNumber(tag)
	flb.foreignExtendee = extendee
	return flb
}

// FromField returns a FieldBuilder that is effectively a copy of the given
// descriptor.
//
// Note that it is not just the given field that is copied but its entire file.
// So the caller can get the parent element of the returned builder and the
// result would be a builder that is effectively a copy of the field
// descriptor's parent.
//
// This means that field builders created from descriptors do not need to be
// explicitly assigned to a file in order to preserve the original field's
// package name.
func FromField(fld *desc.FieldDescriptor) (*FieldBuilder, error) {
	if fb, err := FromFile(fld.GetFile()); err != nil {
		return nil, err
	} else if flb, ok := fb.findFullyQualifiedElement(fld.GetFullyQualifiedName()).(*FieldBuilder); ok {
		return flb, nil
	} else {
		return nil, fmt.Errorf("could not find field %s after converting file %q to builder", fld.GetFullyQualifiedName(), fld.GetFile().GetName())
	}
}

func fromField(fld *desc.FieldDescriptor) (*FieldBuilder, error) {
	ft := fieldTypeFromDescriptor(fld)
	flb := NewField(fld.GetName(), ft)
	flb.Options = fld.GetFieldOptions()
	flb.Label = fld.GetLabel()
	flb.Proto3Optional = fld.IsProto3Optional()
	flb.Default = fld.AsFieldDescriptorProto().GetDefaultValue()
	flb.JsonName = fld.GetJSONName()
	setComments(&flb.comments, fld.GetSourceInfo())

	if fld.IsExtension() {
		flb.foreignExtendee = fld.GetOwner()
	}
	if err := flb.TrySetNumber(fld.GetNumber()); err != nil {
		return nil, err
	}
	return flb, nil
}

// SetName changes this field's name, returning the field builder for method
// chaining. If the given new name is not valid (e.g. TrySetName would have
// returned an error) then this method will panic.
func (flb *FieldBuilder) SetName(newName string) *FieldBuilder {
	if err := flb.TrySetName(newName); err != nil {
		panic(err)
	}
	return flb
}

// TrySetName changes this field's name. It will return an error if the given
// new name is not a valid protobuf identifier or if the parent builder already
// has an element with the given name.
//
// If the field is a non-extension whose parent is a one-of, the one-of's
// enclosing message is checked for elements with a conflicting name. Despite
// the fact that one-of choices are modeled as children of the one-of builder,
// in the protobuf IDL they are actually all defined in the message's namespace.
func (flb *FieldBuilder) TrySetName(newName string) error {
	var oldMsgName string
	if flb.msgType != nil {
		if flb.fieldType.fieldType == descriptorpb.FieldDescriptorProto_TYPE_GROUP {
			return fmt.Errorf("cannot change name of group field %s; change name of group instead", GetFullyQualifiedName(flb))
		} else {
			oldMsgName = flb.msgType.name
			msgName := entryTypeName(newName)
			if err := flb.msgType.trySetNameInternal(msgName); err != nil {
				return err
			}
		}
	}
	if err := flb.baseBuilder.setName(flb, newName); err != nil {
		// undo change to map entry name
		if flb.msgType != nil && flb.fieldType.fieldType != descriptorpb.FieldDescriptorProto_TYPE_GROUP {
			flb.msgType.setNameInternal(oldMsgName)
		}
		return err
	}
	return nil
}

func (flb *FieldBuilder) trySetNameInternal(newName string) error {
	return flb.baseBuilder.setName(flb, newName)
}

func (flb *FieldBuilder) setNameInternal(newName string) {
	if err := flb.trySetNameInternal(newName); err != nil {
		panic(err)
	}
}

// SetComments sets the comments associated with the field. This method returns
// the field builder, for method chaining.
func (flb *FieldBuilder) SetComments(c Comments) *FieldBuilder {
	flb.comments = c
	return flb
}

func (flb *FieldBuilder) setParent(newParent Builder) {
	flb.baseBuilder.setParent(newParent)
}

// GetChildren returns any builders assigned to this field builder. The only
// kind of children a field can have are message types, that correspond to the
// field's map entry type or group type (for map and group fields respectively).
func (flb *FieldBuilder) GetChildren() []Builder {
	if flb.msgType != nil {
		return []Builder{flb.msgType}
	}
	return nil
}

func (flb *FieldBuilder) findChild(name string) Builder {
	if flb.msgType != nil && flb.msgType.name == name {
		return flb.msgType
	}
	return nil
}

func (flb *FieldBuilder) removeChild(b Builder) {
	if mb, ok := b.(*MessageBuilder); ok && mb == flb.msgType {
		flb.msgType = nil
		if p, ok := flb.parent.(*MessageBuilder); ok {
			delete(p.symbols, mb.GetName())
		}
	}
}

func (flb *FieldBuilder) renamedChild(b Builder, oldName string) error {
	if flb.msgType != nil {
		var oldFieldName string
		if flb.fieldType.fieldType == descriptorpb.FieldDescriptorProto_TYPE_GROUP {
			if !unicode.IsUpper(rune(b.GetName()[0])) {
				return fmt.Errorf("group name %s must start with capital letter", b.GetName())
			}
			// change field name to be lower-case form of group name
			oldFieldName = flb.name
			fieldName := strings.ToLower(b.GetName())
			if err := flb.trySetNameInternal(fieldName); err != nil {
				return err
			}
		}
		if p, ok := flb.parent.(*MessageBuilder); ok {
			if err := p.addSymbol(b); err != nil {
				if flb.fieldType.fieldType == descriptorpb.FieldDescriptorProto_TYPE_GROUP {
					// revert the field rename
					flb.setNameInternal(oldFieldName)
				}
				return err
			}
		}
	}
	return nil
}

// GetNumber returns this field's tag number, or zero if the tag number will be
// auto-assigned when the field descriptor is built.
func (flb *FieldBuilder) GetNumber() int32 {
	return flb.number
}

// SetNumber changes the numeric tag for this field and then returns the field,
// for method chaining. If the given new tag is not valid (e.g. TrySetNumber
// would have returned an error) then this method will panic.
func (flb *FieldBuilder) SetNumber(tag int32) *FieldBuilder {
	if err := flb.TrySetNumber(tag); err != nil {
		panic(err)
	}
	return flb
}

// TrySetNumber changes this field's tag number. It will return an error if the
// given new tag is out of valid range or (for non-extension fields) if the
// enclosing message already includes a field with the given tag.
//
// Non-extension fields can be set to zero, which means a proper tag number will
// be auto-assigned when the descriptor is built. Extension field tags, however,
// must be set to a valid non-zero value.
func (flb *FieldBuilder) TrySetNumber(tag int32) error {
	if tag == flb.number {
		return nil // no change
	}
	if tag < 0 {
		return fmt.Errorf("cannot set tag number for field %s to negative value %d", GetFullyQualifiedName(flb), tag)
	}
	if tag == 0 && flb.IsExtension() {
		return fmt.Errorf("cannot set tag number for extension %s; only regular fields can be auto-assigned", GetFullyQualifiedName(flb))
	}
	if tag >= internal.SpecialReservedStart && tag <= internal.SpecialReservedEnd {
		return fmt.Errorf("tag for field %s cannot be in special reserved range %d-%d", GetFullyQualifiedName(flb), internal.SpecialReservedStart, internal.SpecialReservedEnd)
	}
	if tag > internal.MaxTag {
		return fmt.Errorf("tag for field %s cannot be above max %d", GetFullyQualifiedName(flb), internal.MaxTag)
	}
	oldTag := flb.number
	flb.number = tag
	if flb.IsExtension() {
		// extension tags are not tracked by builders, so no more to do
		return nil
	}
	switch p := flb.parent.(type) {
	case *OneOfBuilder:
		m := p.parent()
		if m != nil {
			if err := m.addTag(flb); err != nil {
				flb.number = oldTag
				return err
			}
			delete(m.fieldTags, oldTag)
		}
	case *MessageBuilder:
		if err := p.addTag(flb); err != nil {
			flb.number = oldTag
			return err
		}
		delete(p.fieldTags, oldTag)
	}
	return nil
}

// SetOptions sets the field options for this field and returns the field, for
// method chaining.
func (flb *FieldBuilder) SetOptions(options *descriptorpb.FieldOptions) *FieldBuilder {
	flb.Options = options
	return flb
}

// SetLabel sets the label for this field, which can be optional, repeated, or
// required. It returns the field builder, for method chaining.
func (flb *FieldBuilder) SetLabel(lbl descriptorpb.FieldDescriptorProto_Label) *FieldBuilder {
	flb.Label = lbl
	return flb
}

// SetProto3Optional sets whether this is a proto3 optional field. It returns
// the field builder, for method chaining.
func (flb *FieldBuilder) SetProto3Optional(p3o bool) *FieldBuilder {
	flb.Proto3Optional = p3o
	return flb
}

// SetRepeated sets the label for this field to repeated. It returns the field
// builder, for method chaining.
func (flb *FieldBuilder) SetRepeated() *FieldBuilder {
	return flb.SetLabel(descriptorpb.FieldDescriptorProto_LABEL_REPEATED)
}

// SetRequired sets the label for this field to required. It returns the field
// builder, for method chaining.
func (flb *FieldBuilder) SetRequired() *FieldBuilder {
	return flb.SetLabel(descriptorpb.FieldDescriptorProto_LABEL_REQUIRED)
}

// SetOptional sets the label for this field to optional. It returns the field
// builder, for method chaining.
func (flb *FieldBuilder) SetOptional() *FieldBuilder {
	return flb.SetLabel(descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL)
}

// IsRepeated returns true if this field's label is repeated. Fields created via
// NewMapField will be repeated (since map's are represented "under the hood" as
// a repeated field of map entry messages).
func (flb *FieldBuilder) IsRepeated() bool {
	return flb.Label == descriptorpb.FieldDescriptorProto_LABEL_REPEATED
}

// IsRequired returns true if this field's label is required.
func (flb *FieldBuilder) IsRequired() bool {
	return flb.Label == descriptorpb.FieldDescriptorProto_LABEL_REQUIRED
}

// IsOptional returns true if this field's label is optional.
func (flb *FieldBuilder) IsOptional() bool {
	return flb.Label == descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
}

// IsMap returns true if this field is a map field.
func (flb *FieldBuilder) IsMap() bool {
	return flb.IsRepeated() &&
		flb.msgType != nil &&
		flb.fieldType.fieldType != descriptorpb.FieldDescriptorProto_TYPE_GROUP &&
		flb.msgType.Options != nil &&
		flb.msgType.Options.GetMapEntry()
}

// GetType returns the field's type.
func (flb *FieldBuilder) GetType() *FieldType {
	return flb.fieldType
}

// SetType changes the field's type and returns the field builder, for method
// chaining.
func (flb *FieldBuilder) SetType(ft *FieldType) *FieldBuilder {
	flb.fieldType = ft
	if flb.msgType != nil && flb.msgType != ft.localMsgType {
		Unlink(flb.msgType)
	}
	return flb
}

// SetDefaultValue changes the field's type and returns the field builder, for
// method chaining.
func (flb *FieldBuilder) SetDefaultValue(defValue string) *FieldBuilder {
	flb.Default = defValue
	return flb
}

// SetJsonName sets the name used in the field's JSON representation and then
// returns the field builder, for method chaining.
func (flb *FieldBuilder) SetJsonName(jsonName string) *FieldBuilder {
	flb.JsonName = jsonName
	return flb
}

// IsExtension returns true if this is an extension field.
func (flb *FieldBuilder) IsExtension() bool {
	return flb.localExtendee != nil || flb.foreignExtendee != nil
}

// GetExtendeeTypeName returns the fully qualified name of the extended message
// or it returns an empty string if this is not an extension field.
func (flb *FieldBuilder) GetExtendeeTypeName() string {
	if flb.foreignExtendee != nil {
		return flb.foreignExtendee.GetFullyQualifiedName()
	} else if flb.localExtendee != nil {
		return GetFullyQualifiedName(flb.localExtendee)
	} else {
		return ""
	}
}

func (flb *FieldBuilder) buildProto(path []int32, sourceInfo *descriptorpb.SourceCodeInfo, isMessageSet bool) (*descriptorpb.FieldDescriptorProto, error) {
	addCommentsTo(sourceInfo, path, &flb.comments)

	isProto3 := flb.GetFile().IsProto3
	if flb.Proto3Optional {
		if !isProto3 {
			return nil, fmt.Errorf("field %s is not in a proto3 syntax file but is marked as a proto3 optional field", GetFullyQualifiedName(flb))
		}
		if flb.IsExtension() {
			return nil, fmt.Errorf("field %s: extensions cannot be proto3 optional fields", GetFullyQualifiedName(flb))
		}
		if _, ok := flb.GetParent().(*OneOfBuilder); ok {
			return nil, fmt.Errorf("field %s: proto3 optional fields cannot belong to a oneof", GetFullyQualifiedName(flb))
		}
	}

	var lbl *descriptorpb.FieldDescriptorProto_Label
	if int32(flb.Label) != 0 {
		if isProto3 && flb.Label == descriptorpb.FieldDescriptorProto_LABEL_REQUIRED {
			return nil, fmt.Errorf("field %s: proto3 does not allow required fields", GetFullyQualifiedName(flb))
		}
		lbl = flb.Label.Enum()
	}
	var typeName *string
	tn := flb.fieldType.GetTypeName()
	if tn != "" {
		typeName = proto.String("." + tn)
	}
	var extendee *string
	if flb.IsExtension() {
		extendee = proto.String("." + flb.GetExtendeeTypeName())
	}
	jsName := flb.JsonName
	if jsName == "" {
		jsName = internal.JsonName(flb.name)
	}
	var def *string
	if flb.Default != "" {
		def = proto.String(flb.Default)
	}
	var proto3Optional *bool
	if flb.Proto3Optional {
		proto3Optional = proto.Bool(true)
	}

	maxTag := internal.GetMaxTag(isMessageSet)
	if flb.number > maxTag {
		return nil, fmt.Errorf("tag for field %s cannot be above max %d", GetFullyQualifiedName(flb), maxTag)
	}

	fd := &descriptorpb.FieldDescriptorProto{
		Name:           proto.String(flb.name),
		Number:         proto.Int32(flb.number),
		Options:        flb.Options,
		Label:          lbl,
		Type:           flb.fieldType.fieldType.Enum(),
		TypeName:       typeName,
		JsonName:       proto.String(jsName),
		DefaultValue:   def,
		Extendee:       extendee,
		Proto3Optional: proto3Optional,
	}
	return fd, nil
}

// Build constructs a field descriptor based on the contents of this field
// builder. If there are any problems constructing the descriptor, including
// resolving symbols referenced by the builder or failing to meet certain
// validation rules, an error is returned.
func (flb *FieldBuilder) Build() (*desc.FieldDescriptor, error) {
	fld, err := flb.BuildDescriptor()
	if err != nil {
		return nil, err
	}
	return fld.(*desc.FieldDescriptor), nil
}

// BuildDescriptor constructs a field descriptor based on the contents of this
// field builder. Most usages will prefer Build() instead, whose return type is
// a concrete descriptor type. This method is present to satisfy the Builder
// interface.
func (flb *FieldBuilder) BuildDescriptor() (desc.Descriptor, error) {
	return doBuild(flb, BuilderOptions{})
}

// OneOfBuilder is a builder used to construct a desc.OneOfDescriptor. A one-of
// builder *must* be added to a message before calling its Build() method.
//
// To create a new OneOfBuilder, use NewOneOf.
type OneOfBuilder struct {
	baseBuilder

	Options *descriptorpb.OneofOptions

	choices []*FieldBuilder
	symbols map[string]*FieldBuilder
}

// NewOneOf creates a new OneOfBuilder for a one-of with the given name.
func NewOneOf(name string) *OneOfBuilder {
	return &OneOfBuilder{
		baseBuilder: baseBuilderWithName(name),
		symbols:     map[string]*FieldBuilder{},
	}
}

// FromOneOf returns a OneOfBuilder that is effectively a copy of the given
// descriptor.
//
// Note that it is not just the given one-of that is copied but its entire file.
// So the caller can get the parent element of the returned builder and the
// result would be a builder that is effectively a copy of the one-of
// descriptor's parent message.
//
// This means that one-of builders created from descriptors do not need to be
// explicitly assigned to a file in order to preserve the original one-of's
// package name.
//
// This function returns an error if the given descriptor is synthetic.
func FromOneOf(ood *desc.OneOfDescriptor) (*OneOfBuilder, error) {
	if ood.IsSynthetic() {
		return nil, fmt.Errorf("one-of %s is synthetic", ood.GetFullyQualifiedName())
	}
	if fb, err := FromFile(ood.GetFile()); err != nil {
		return nil, err
	} else if oob, ok := fb.findFullyQualifiedElement(ood.GetFullyQualifiedName()).(*OneOfBuilder); ok {
		return oob, nil
	} else {
		return nil, fmt.Errorf("could not find one-of %s after converting file %q to builder", ood.GetFullyQualifiedName(), ood.GetFile().GetName())
	}
}

func fromOneOf(ood *desc.OneOfDescriptor) (*OneOfBuilder, error) {
	oob := NewOneOf(ood.GetName())
	oob.Options = ood.GetOneOfOptions()
	setComments(&oob.comments, ood.GetSourceInfo())

	for _, fld := range ood.GetChoices() {
		if flb, err := fromField(fld); err != nil {
			return nil, err
		} else if err := oob.TryAddChoice(flb); err != nil {
			return nil, err
		}
	}

	return oob, nil
}

// SetName changes this one-of's name, returning the one-of builder for method
// chaining. If the given new name is not valid (e.g. TrySetName would have
// returned an error) then this method will panic.
func (oob *OneOfBuilder) SetName(newName string) *OneOfBuilder {
	if err := oob.TrySetName(newName); err != nil {
		panic(err)
	}
	return oob
}

// TrySetName changes this one-of's name. It will return an error if the given
// new name is not a valid protobuf identifier or if the parent message builder
// already has an element with the given name.
func (oob *OneOfBuilder) TrySetName(newName string) error {
	return oob.baseBuilder.setName(oob, newName)
}

// SetComments sets the comments associated with the one-of. This method
// returns the one-of builder, for method chaining.
func (oob *OneOfBuilder) SetComments(c Comments) *OneOfBuilder {
	oob.comments = c
	return oob
}

// GetChildren returns any builders assigned to this one-of builder. These will
// be choices for the one-of, each of which will be a field builder.
func (oob *OneOfBuilder) GetChildren() []Builder {
	var ch []Builder
	for _, evb := range oob.choices {
		ch = append(ch, evb)
	}
	return ch
}

func (oob *OneOfBuilder) parent() *MessageBuilder {
	if oob.baseBuilder.parent == nil {
		return nil
	}
	return oob.baseBuilder.parent.(*MessageBuilder)
}

func (oob *OneOfBuilder) findChild(name string) Builder {
	// in terms of finding a child by qualified name, fields in the
	// one-of are considered children of the message, not the one-of
	return nil
}

func (oob *OneOfBuilder) removeChild(b Builder) {
	if p, ok := b.GetParent().(*OneOfBuilder); !ok || p != oob {
		return
	}

	if oob.parent() != nil {
		// remove from message's name and tag maps
		flb := b.(*FieldBuilder)
		delete(oob.parent().fieldTags, flb.GetNumber())
		delete(oob.parent().symbols, flb.GetName())
	}

	oob.choices = deleteBuilder(b.GetName(), oob.choices).([]*FieldBuilder)
	delete(oob.symbols, b.GetName())
	b.setParent(nil)
}

func (oob *OneOfBuilder) renamedChild(b Builder, oldName string) error {
	if p, ok := b.GetParent().(*OneOfBuilder); !ok || p != oob {
		return nil
	}

	if err := oob.addSymbol(b.(*FieldBuilder)); err != nil {
		return err
	}

	// update message's name map (to make sure new field name doesn't
	// collide with other kinds of elements in the message)
	if oob.parent() != nil {
		if err := oob.parent().addSymbol(b); err != nil {
			delete(oob.symbols, b.GetName())
			return err
		}
		delete(oob.parent().symbols, oldName)
	}

	delete(oob.symbols, oldName)
	return nil
}

func (oob *OneOfBuilder) addSymbol(b *FieldBuilder) error {
	if _, ok := oob.symbols[b.GetName()]; ok {
		return fmt.Errorf("one-of %s already contains field named %q", GetFullyQualifiedName(oob), b.GetName())
	}
	oob.symbols[b.GetName()] = b
	return nil
}

// GetChoice returns the field with the given name. If no such field exists in
// the one-of, nil is returned.
func (oob *OneOfBuilder) GetChoice(name string) *FieldBuilder {
	return oob.symbols[name]
}

// RemoveChoice removes the field with the given name. If no such field exists
// in the one-of, this is a no-op. This returns the one-of builder, for method
// chaining.
func (oob *OneOfBuilder) RemoveChoice(name string) *OneOfBuilder {
	oob.TryRemoveChoice(name)
	return oob
}

// TryRemoveChoice removes the field with the given name and returns false if
// the one-of has no such field.
func (oob *OneOfBuilder) TryRemoveChoice(name string) bool {
	if flb, ok := oob.symbols[name]; ok {
		oob.removeChild(flb)
		return true
	}
	return false
}

// AddChoice adds the given field to this one-of. If an error prevents the field
// from being added, this method panics. If the given field is an extension,
// this method panics. If the given field is a group or map field or if it is
// not optional (e.g. it is required or repeated), this method panics. This
// returns the one-of builder, for method chaining.
func (oob *OneOfBuilder) AddChoice(flb *FieldBuilder) *OneOfBuilder {
	if err := oob.TryAddChoice(flb); err != nil {
		panic(err)
	}
	return oob
}

// TryAddChoice adds the given field to this one-of, returning any error that
// prevents the field from being added (such as a name collision with another
// element already added to the enclosing message). An error is returned if the
// given field is an extension field, a map or group field, or repeated or
// required.
func (oob *OneOfBuilder) TryAddChoice(flb *FieldBuilder) error {
	if flb.IsExtension() {
		return fmt.Errorf("field %s is an extension, not a regular field", flb.GetName())
	}
	if flb.msgType != nil && flb.fieldType.fieldType != descriptorpb.FieldDescriptorProto_TYPE_GROUP {
		return fmt.Errorf("cannot add a map field %q to one-of %s", flb.name, GetFullyQualifiedName(oob))
	}
	if flb.IsRepeated() || flb.IsRequired() {
		return fmt.Errorf("fields in a one-of must be optional, %s is %v", flb.name, flb.Label)
	}
	if err := oob.addSymbol(flb); err != nil {
		return err
	}
	mb := oob.parent()
	if mb != nil {
		// If we are moving field from a message to a one-of that belongs to the
		// same message, we have to use different order of operations to prevent
		// failure (otherwise, it looks like it's being added twice).
		// (We do similar if moving the other direction, from the one-of into
		// the message to which one-of belongs.)
		needToUnlinkFirst := mb.isPresentButNotChild(flb)
		if needToUnlinkFirst {
			Unlink(flb)
			mb.registerField(flb)
		} else {
			if err := mb.registerField(flb); err != nil {
				delete(oob.symbols, flb.GetName())
				return err
			}
			Unlink(flb)
		}
	}
	flb.setParent(oob)
	oob.choices = append(oob.choices, flb)
	return nil
}

// SetOptions sets the one-of options for this one-of and returns the one-of,
// for method chaining.
func (oob *OneOfBuilder) SetOptions(options *descriptorpb.OneofOptions) *OneOfBuilder {
	oob.Options = options
	return oob
}

func (oob *OneOfBuilder) buildProto(path []int32, sourceInfo *descriptorpb.SourceCodeInfo) (*descriptorpb.OneofDescriptorProto, error) {
	addCommentsTo(sourceInfo, path, &oob.comments)

	for _, flb := range oob.choices {
		if flb.IsRepeated() || flb.IsRequired() {
			return nil, fmt.Errorf("fields in a one-of must be optional, %s is %v", GetFullyQualifiedName(flb), flb.Label)
		}
	}

	return &descriptorpb.OneofDescriptorProto{
		Name:    proto.String(oob.name),
		Options: oob.Options,
	}, nil
}

// Build constructs a one-of descriptor based on the contents of this one-of
// builder. If there are any problems constructing the descriptor, including
// resolving symbols referenced by the builder or failing to meet certain
// validation rules, an error is returned.
func (oob *OneOfBuilder) Build() (*desc.OneOfDescriptor, error) {
	ood, err := oob.BuildDescriptor()
	if err != nil {
		return nil, err
	}
	return ood.(*desc.OneOfDescriptor), nil
}

// BuildDescriptor constructs a one-of descriptor based on the contents of this
// one-of builder. Most usages will prefer Build() instead, whose return type is
// a concrete descriptor type. This method is present to satisfy the Builder
// interface.
func (oob *OneOfBuilder) BuildDescriptor() (desc.Descriptor, error) {
	return doBuild(oob, BuilderOptions{})
}

func entryTypeName(fieldName string) string {
	return internal.InitCap(internal.JsonName(fieldName)) + "Entry"
}
