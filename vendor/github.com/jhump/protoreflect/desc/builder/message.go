package builder

import (
	"fmt"
	"sort"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/internal"
)

// MessageBuilder is a builder used to construct a desc.MessageDescriptor. A
// message builder can define nested messages, enums, and extensions in addition
// to defining the message's fields.
//
// Note that when building a descriptor from a MessageBuilder, not all protobuf
// validation rules are enforced. See the package documentation for more info.
//
// To create a new MessageBuilder, use NewMessage.
type MessageBuilder struct {
	baseBuilder

	Options         *descriptorpb.MessageOptions
	ExtensionRanges []*descriptorpb.DescriptorProto_ExtensionRange
	ReservedRanges  []*descriptorpb.DescriptorProto_ReservedRange
	ReservedNames   []string

	fieldsAndOneOfs  []Builder
	fieldTags        map[int32]*FieldBuilder
	nestedMessages   []*MessageBuilder
	nestedExtensions []*FieldBuilder
	nestedEnums      []*EnumBuilder
	symbols          map[string]Builder
}

// NewMessage creates a new MessageBuilder for a message with the given name.
// Since the new message has no parent element, it also has no package name
// (e.g. it is in the unnamed package, until it is assigned to a file builder
// that defines a package name).
func NewMessage(name string) *MessageBuilder {
	return &MessageBuilder{
		baseBuilder: baseBuilderWithName(name),
		fieldTags:   map[int32]*FieldBuilder{},
		symbols:     map[string]Builder{},
	}
}

// FromMessage returns a MessageBuilder that is effectively a copy of the given
// descriptor.
//
// Note that it is not just the given message that is copied but its entire
// file. So the caller can get the parent element of the returned builder and
// the result would be a builder that is effectively a copy of the message
// descriptor's parent.
//
// This means that message builders created from descriptors do not need to be
// explicitly assigned to a file in order to preserve the original message's
// package name.
func FromMessage(md *desc.MessageDescriptor) (*MessageBuilder, error) {
	if fb, err := FromFile(md.GetFile()); err != nil {
		return nil, err
	} else if mb, ok := fb.findFullyQualifiedElement(md.GetFullyQualifiedName()).(*MessageBuilder); ok {
		return mb, nil
	} else {
		return nil, fmt.Errorf("could not find message %s after converting file %q to builder", md.GetFullyQualifiedName(), md.GetFile().GetName())
	}
}

func fromMessage(md *desc.MessageDescriptor,
	localMessages map[*desc.MessageDescriptor]*MessageBuilder,
	localEnums map[*desc.EnumDescriptor]*EnumBuilder) (*MessageBuilder, error) {

	mb := NewMessage(md.GetName())
	mb.Options = md.GetMessageOptions()
	mb.ExtensionRanges = md.AsDescriptorProto().GetExtensionRange()
	mb.ReservedRanges = md.AsDescriptorProto().GetReservedRange()
	mb.ReservedNames = md.AsDescriptorProto().GetReservedName()
	setComments(&mb.comments, md.GetSourceInfo())

	localMessages[md] = mb

	oneOfs := make([]*OneOfBuilder, len(md.GetOneOfs()))
	for i, ood := range md.GetOneOfs() {
		if ood.IsSynthetic() {
			continue
		}
		if oob, err := fromOneOf(ood); err != nil {
			return nil, err
		} else {
			oneOfs[i] = oob
		}
	}

	for _, fld := range md.GetFields() {
		if fld.GetOneOf() != nil && !fld.GetOneOf().IsSynthetic() {
			// add one-ofs in the order of their first constituent field
			oob := oneOfs[fld.AsFieldDescriptorProto().GetOneofIndex()]
			if oob != nil {
				oneOfs[fld.AsFieldDescriptorProto().GetOneofIndex()] = nil
				if err := mb.TryAddOneOf(oob); err != nil {
					return nil, err
				}
			}
			continue
		}
		if flb, err := fromField(fld); err != nil {
			return nil, err
		} else if err := mb.TryAddField(flb); err != nil {
			return nil, err
		}
	}

	for _, nmd := range md.GetNestedMessageTypes() {
		if nmb, err := fromMessage(nmd, localMessages, localEnums); err != nil {
			return nil, err
		} else if err := mb.TryAddNestedMessage(nmb); err != nil {
			return nil, err
		}
	}
	for _, ed := range md.GetNestedEnumTypes() {
		if eb, err := fromEnum(ed, localEnums); err != nil {
			return nil, err
		} else if err := mb.TryAddNestedEnum(eb); err != nil {
			return nil, err
		}
	}
	for _, exd := range md.GetNestedExtensions() {
		if exb, err := fromField(exd); err != nil {
			return nil, err
		} else if err := mb.TryAddNestedExtension(exb); err != nil {
			return nil, err
		}
	}

	return mb, nil
}

// SetName changes this message's name, returning the message builder for method
// chaining. If the given new name is not valid (e.g. TrySetName would have
// returned an error) then this method will panic.
func (mb *MessageBuilder) SetName(newName string) *MessageBuilder {
	if err := mb.TrySetName(newName); err != nil {
		panic(err)
	}
	return mb
}

// TrySetName changes this message's name. It will return an error if the given
// new name is not a valid protobuf identifier or if the parent builder already
// has an element with the given name.
//
// If the message is a map or group type whose parent is the corresponding map
// or group field, the parent field's enclosing message is checked for elements
// with a conflicting name. Despite the fact that these message types are
// modeled as children of their associated field builder, in the protobuf IDL
// they are actually all defined in the enclosing message's namespace.
func (mb *MessageBuilder) TrySetName(newName string) error {
	if p, ok := mb.parent.(*FieldBuilder); ok && p.fieldType.fieldType != descriptorpb.FieldDescriptorProto_TYPE_GROUP {
		return fmt.Errorf("cannot change name of map entry %s; change name of field instead", GetFullyQualifiedName(mb))
	}
	return mb.trySetNameInternal(newName)
}

func (mb *MessageBuilder) trySetNameInternal(newName string) error {
	return mb.baseBuilder.setName(mb, newName)
}

func (mb *MessageBuilder) setNameInternal(newName string) {
	if err := mb.trySetNameInternal(newName); err != nil {
		panic(err)
	}
}

// SetComments sets the comments associated with the message. This method
// returns the message builder, for method chaining.
func (mb *MessageBuilder) SetComments(c Comments) *MessageBuilder {
	mb.comments = c
	return mb
}

// GetChildren returns any builders assigned to this message builder. These will
// include the message's fields and one-ofs as well as any nested messages,
// extensions, and enums.
func (mb *MessageBuilder) GetChildren() []Builder {
	ch := append([]Builder(nil), mb.fieldsAndOneOfs...)
	for _, nmb := range mb.nestedMessages {
		ch = append(ch, nmb)
	}
	for _, exb := range mb.nestedExtensions {
		ch = append(ch, exb)
	}
	for _, eb := range mb.nestedEnums {
		ch = append(ch, eb)
	}
	return ch
}

func (mb *MessageBuilder) findChild(name string) Builder {
	return mb.symbols[name]
}

func (mb *MessageBuilder) removeChild(b Builder) {
	if p, ok := b.GetParent().(*MessageBuilder); !ok || p != mb {
		return
	}

	switch b := b.(type) {
	case *FieldBuilder:
		if b.IsExtension() {
			mb.nestedExtensions = deleteBuilder(b.GetName(), mb.nestedExtensions).([]*FieldBuilder)
		} else {
			mb.fieldsAndOneOfs = deleteBuilder(b.GetName(), mb.fieldsAndOneOfs).([]Builder)
			delete(mb.fieldTags, b.GetNumber())
			if b.msgType != nil {
				delete(mb.symbols, b.msgType.GetName())
			}
		}
	case *OneOfBuilder:
		mb.fieldsAndOneOfs = deleteBuilder(b.GetName(), mb.fieldsAndOneOfs).([]Builder)
		for _, flb := range b.choices {
			delete(mb.symbols, flb.GetName())
			delete(mb.fieldTags, flb.GetNumber())
		}
	case *MessageBuilder:
		mb.nestedMessages = deleteBuilder(b.GetName(), mb.nestedMessages).([]*MessageBuilder)
	case *EnumBuilder:
		mb.nestedEnums = deleteBuilder(b.GetName(), mb.nestedEnums).([]*EnumBuilder)
	}
	delete(mb.symbols, b.GetName())
	b.setParent(nil)
}

func (mb *MessageBuilder) renamedChild(b Builder, oldName string) error {
	if p, ok := b.GetParent().(*MessageBuilder); !ok || p != mb {
		return nil
	}

	if err := mb.addSymbol(b); err != nil {
		return err
	}
	delete(mb.symbols, oldName)
	return nil
}

func (mb *MessageBuilder) addSymbol(b Builder) error {
	if ex, ok := mb.symbols[b.GetName()]; ok {
		return fmt.Errorf("message %s already contains element (%T) named %q", GetFullyQualifiedName(mb), ex, b.GetName())
	}
	mb.symbols[b.GetName()] = b
	return nil
}

func (mb *MessageBuilder) addTag(flb *FieldBuilder) error {
	if flb.number == 0 {
		return nil
	}
	if ex, ok := mb.fieldTags[flb.GetNumber()]; ok {
		return fmt.Errorf("message %s already contains field with tag %d: %s", GetFullyQualifiedName(mb), flb.GetNumber(), ex.GetName())
	}
	mb.fieldTags[flb.GetNumber()] = flb
	return nil
}

func (mb *MessageBuilder) registerField(flb *FieldBuilder) error {
	if err := mb.addSymbol(flb); err != nil {
		return err
	}
	if err := mb.addTag(flb); err != nil {
		delete(mb.symbols, flb.GetName())
		return err
	}
	if flb.msgType != nil {
		if err := mb.addSymbol(flb.msgType); err != nil {
			delete(mb.symbols, flb.GetName())
			delete(mb.fieldTags, flb.GetNumber())
			return err
		}
	}
	return nil
}

// GetField returns the field with the given name. If no such field exists in
// the message, nil is returned. The field does not have to be an immediate
// child of this message but could instead be an indirect child via a one-of.
func (mb *MessageBuilder) GetField(name string) *FieldBuilder {
	b := mb.symbols[name]
	if flb, ok := b.(*FieldBuilder); ok && !flb.IsExtension() {
		return flb
	} else {
		return nil
	}
}

// RemoveField removes the field with the given name. If no such field exists in
// the message, this is a no-op. If the field is part of a one-of, the one-of
// remains assigned to this message and the field is removed from it. This
// returns the message builder, for method chaining.
func (mb *MessageBuilder) RemoveField(name string) *MessageBuilder {
	mb.TryRemoveField(name)
	return mb
}

// TryRemoveField removes the field with the given name and returns false if the
// message has no such field. If the field is part of a one-of, the one-of
// remains assigned to this message and the field is removed from it.
func (mb *MessageBuilder) TryRemoveField(name string) bool {
	b := mb.symbols[name]
	if flb, ok := b.(*FieldBuilder); ok && !flb.IsExtension() {
		// parent could be mb, but could also be a one-of
		flb.GetParent().removeChild(flb)
		return true
	}
	return false
}

// AddField adds the given field to this message. If an error prevents the field
// from being added, this method panics. If the given field is an extension,
// this method panics. This returns the message builder, for method chaining.
func (mb *MessageBuilder) AddField(flb *FieldBuilder) *MessageBuilder {
	if err := mb.TryAddField(flb); err != nil {
		panic(err)
	}
	return mb
}

// TryAddField adds the given field to this message, returning any error that
// prevents the field from being added (such as a name collision with another
// element already added to the message). An error is returned if the given
// field is an extension field.
func (mb *MessageBuilder) TryAddField(flb *FieldBuilder) error {
	if flb.IsExtension() {
		return fmt.Errorf("field %s is an extension, not a regular field", flb.GetName())
	}
	// If we are moving field from a one-of that belongs to this message
	// directly to this message, we have to use different order of operations
	// to prevent failure (otherwise, it looks like it's being added twice).
	// (We do similar if moving the other direction, from message to a one-of
	// that is already assigned to same message.)
	needToUnlinkFirst := mb.isPresentButNotChild(flb)
	if needToUnlinkFirst {
		Unlink(flb)
		mb.registerField(flb)
	} else {
		if err := mb.registerField(flb); err != nil {
			return err
		}
		Unlink(flb)
	}
	flb.setParent(mb)
	mb.fieldsAndOneOfs = append(mb.fieldsAndOneOfs, flb)
	return nil
}

// GetOneOf returns the one-of with the given name. If no such one-of exists in
// the message, nil is returned.
func (mb *MessageBuilder) GetOneOf(name string) *OneOfBuilder {
	b := mb.symbols[name]
	if oob, ok := b.(*OneOfBuilder); ok {
		return oob
	} else {
		return nil
	}
}

// RemoveOneOf removes the one-of with the given name. If no such one-of exists
// in the message, this is a no-op. This returns the message builder, for method
// chaining.
func (mb *MessageBuilder) RemoveOneOf(name string) *MessageBuilder {
	mb.TryRemoveOneOf(name)
	return mb
}

// TryRemoveOneOf removes the one-of with the given name and returns false if
// the message has no such one-of.
func (mb *MessageBuilder) TryRemoveOneOf(name string) bool {
	b := mb.symbols[name]
	if oob, ok := b.(*OneOfBuilder); ok {
		mb.removeChild(oob)
		return true
	}
	return false
}

// AddOneOf adds the given one-of to this message. If an error prevents the
// one-of from being added, this method panics. This returns the message
// builder, for method chaining.
func (mb *MessageBuilder) AddOneOf(oob *OneOfBuilder) *MessageBuilder {
	if err := mb.TryAddOneOf(oob); err != nil {
		panic(err)
	}
	return mb
}

// TryAddOneOf adds the given one-of to this message, returning any error that
// prevents the one-of from being added (such as a name collision with another
// element already added to the message).
func (mb *MessageBuilder) TryAddOneOf(oob *OneOfBuilder) error {
	if err := mb.addSymbol(oob); err != nil {
		return err
	}
	// add nested fields to symbol and tag map
	for i, flb := range oob.choices {
		if err := mb.registerField(flb); err != nil {
			// must undo all additions we've made so far
			delete(mb.symbols, oob.GetName())
			for i > 1 {
				i--
				flb := oob.choices[i]
				delete(mb.symbols, flb.GetName())
				delete(mb.fieldTags, flb.GetNumber())
			}
			return err
		}
	}
	Unlink(oob)
	oob.setParent(mb)
	mb.fieldsAndOneOfs = append(mb.fieldsAndOneOfs, oob)
	return nil
}

// GetNestedMessage returns the nested message with the given name. If no such
// message exists, nil is returned. The named message must be in this message's
// scope. If the message is nested more deeply, this will return nil. This means
// the message must be a direct child of this message or a child of one of this
// message's fields (e.g. the group type for a group field or a map entry for a
// map field).
func (mb *MessageBuilder) GetNestedMessage(name string) *MessageBuilder {
	b := mb.symbols[name]
	if nmb, ok := b.(*MessageBuilder); ok {
		return nmb
	} else {
		return nil
	}
}

// RemoveNestedMessage removes the nested message with the given name. If no
// such message exists, this is a no-op. This returns the message builder, for
// method chaining.
func (mb *MessageBuilder) RemoveNestedMessage(name string) *MessageBuilder {
	mb.TryRemoveNestedMessage(name)
	return mb
}

// TryRemoveNestedMessage removes the nested message with the given name and
// returns false if this message has no nested message with that name. If the
// named message is a child of a field (e.g. the group type for a group field or
// the map entry for a map field), it is removed from that field and thus
// removed from this message's scope.
func (mb *MessageBuilder) TryRemoveNestedMessage(name string) bool {
	b := mb.symbols[name]
	if nmb, ok := b.(*MessageBuilder); ok {
		// parent could be mb, but could also be a field (if the message
		// is the field's group or map entry type)
		nmb.GetParent().removeChild(nmb)
		return true
	}
	return false
}

// AddNestedMessage adds the given message as a nested child of this message. If
// an error prevents the message from being added, this method panics. This
// returns the message builder, for method chaining.
func (mb *MessageBuilder) AddNestedMessage(nmb *MessageBuilder) *MessageBuilder {
	if err := mb.TryAddNestedMessage(nmb); err != nil {
		panic(err)
	}
	return mb
}

// TryAddNestedMessage adds the given message as a nested child of this message,
// returning any error that prevents the message from being added (such as a
// name collision with another element already added to the message).
func (mb *MessageBuilder) TryAddNestedMessage(nmb *MessageBuilder) error {
	// If we are moving nested message from field (map entry or group type)
	// directly to this message, we have to use different order of operations
	// to prevent failure (otherwise, it looks like it's being added twice).
	// (We don't need to do similar for the other direction, because that isn't
	// possible: you can't add messages to a field, they can only be constructed
	// that way using NewGroupField or NewMapField.)
	needToUnlinkFirst := mb.isPresentButNotChild(nmb)
	if needToUnlinkFirst {
		Unlink(nmb)
		_ = mb.addSymbol(nmb)
	} else {
		if err := mb.addSymbol(nmb); err != nil {
			return err
		}
		Unlink(nmb)
	}
	nmb.setParent(mb)
	mb.nestedMessages = append(mb.nestedMessages, nmb)
	return nil
}

func (mb *MessageBuilder) isPresentButNotChild(b Builder) bool {
	if p, ok := b.GetParent().(*MessageBuilder); ok && p == mb {
		// it's a child
		return false
	}
	return mb.symbols[b.GetName()] == b
}

// GetNestedExtension returns the nested extension with the given name. If no
// such extension exists, nil is returned. The named extension must be in this
// message's scope. If the extension is nested more deeply, this will return
// nil. This means the extension must be a direct child of this message.
func (mb *MessageBuilder) GetNestedExtension(name string) *FieldBuilder {
	b := mb.symbols[name]
	if exb, ok := b.(*FieldBuilder); ok && exb.IsExtension() {
		return exb
	} else {
		return nil
	}
}

// RemoveNestedExtension removes the nested extension with the given name. If no
// such extension exists, this is a no-op. This returns the message builder, for
// method chaining.
func (mb *MessageBuilder) RemoveNestedExtension(name string) *MessageBuilder {
	mb.TryRemoveNestedExtension(name)
	return mb
}

// TryRemoveNestedExtension removes the nested extension with the given name and
// returns false if this message has no nested extension with that name.
func (mb *MessageBuilder) TryRemoveNestedExtension(name string) bool {
	b := mb.symbols[name]
	if exb, ok := b.(*FieldBuilder); ok && exb.IsExtension() {
		mb.removeChild(exb)
		return true
	}
	return false
}

// AddNestedExtension adds the given extension as a nested child of this
// message. If an error prevents the extension from being added, this method
// panics. This returns the message builder, for method chaining.
func (mb *MessageBuilder) AddNestedExtension(exb *FieldBuilder) *MessageBuilder {
	if err := mb.TryAddNestedExtension(exb); err != nil {
		panic(err)
	}
	return mb
}

// TryAddNestedExtension adds the given extension as a nested child of this
// message, returning any error that prevents the extension from being added
// (such as a name collision with another element already added to the message).
func (mb *MessageBuilder) TryAddNestedExtension(exb *FieldBuilder) error {
	if !exb.IsExtension() {
		return fmt.Errorf("field %s is not an extension", exb.GetName())
	}
	if err := mb.addSymbol(exb); err != nil {
		return err
	}
	Unlink(exb)
	exb.setParent(mb)
	mb.nestedExtensions = append(mb.nestedExtensions, exb)
	return nil
}

// GetNestedEnum returns the nested enum with the given name. If no such enum
// exists, nil is returned. The named enum must be in this message's scope. If
// the enum is nested more deeply, this will return nil. This means the enum
// must be a direct child of this message.
func (mb *MessageBuilder) GetNestedEnum(name string) *EnumBuilder {
	b := mb.symbols[name]
	if eb, ok := b.(*EnumBuilder); ok {
		return eb
	} else {
		return nil
	}
}

// RemoveNestedEnum removes the nested enum with the given name. If no such enum
// exists, this is a no-op. This returns the message builder, for method
// chaining.
func (mb *MessageBuilder) RemoveNestedEnum(name string) *MessageBuilder {
	mb.TryRemoveNestedEnum(name)
	return mb
}

// TryRemoveNestedEnum removes the nested enum with the given name and returns
// false if this message has no nested enum with that name.
func (mb *MessageBuilder) TryRemoveNestedEnum(name string) bool {
	b := mb.symbols[name]
	if eb, ok := b.(*EnumBuilder); ok {
		mb.removeChild(eb)
		return true
	}
	return false
}

// AddNestedEnum adds the given enum as a nested child of this message. If an
// error prevents the enum from being added, this method panics. This returns
// the message builder, for method chaining.
func (mb *MessageBuilder) AddNestedEnum(eb *EnumBuilder) *MessageBuilder {
	if err := mb.TryAddNestedEnum(eb); err != nil {
		panic(err)
	}
	return mb
}

// TryAddNestedEnum adds the given enum as a nested child of this message,
// returning any error that prevents the enum from being added (such as a name
// collision with another element already added to the message).
func (mb *MessageBuilder) TryAddNestedEnum(eb *EnumBuilder) error {
	if err := mb.addSymbol(eb); err != nil {
		return err
	}
	Unlink(eb)
	eb.setParent(mb)
	mb.nestedEnums = append(mb.nestedEnums, eb)
	return nil
}

// SetOptions sets the message options for this message and returns the message,
// for method chaining.
func (mb *MessageBuilder) SetOptions(options *descriptorpb.MessageOptions) *MessageBuilder {
	mb.Options = options
	return mb
}

// AddExtensionRange adds the given extension range to this message. The range
// is inclusive of both the start and end, just like defining a range in proto
// IDL source. This returns the message, for method chaining.
func (mb *MessageBuilder) AddExtensionRange(start, end int32) *MessageBuilder {
	return mb.AddExtensionRangeWithOptions(start, end, nil)
}

// AddExtensionRangeWithOptions adds the given extension range to this message.
// The range is inclusive of both the start and end, just like defining a range
// in proto IDL source. This returns the message, for method chaining.
func (mb *MessageBuilder) AddExtensionRangeWithOptions(start, end int32, options *descriptorpb.ExtensionRangeOptions) *MessageBuilder {
	er := &descriptorpb.DescriptorProto_ExtensionRange{
		Start:   proto.Int32(start),
		End:     proto.Int32(end + 1),
		Options: options,
	}
	mb.ExtensionRanges = append(mb.ExtensionRanges, er)
	return mb
}

// SetExtensionRanges replaces all of this message's extension ranges with the
// given slice of ranges. Unlike AddExtensionRange and unlike the way ranges are
// defined in proto IDL source, a DescriptorProto_ExtensionRange struct treats
// the end of the range as *exclusive*. So the range is inclusive of the start
// but exclusive of the end. This returns the message, for method chaining.
func (mb *MessageBuilder) SetExtensionRanges(ranges []*descriptorpb.DescriptorProto_ExtensionRange) *MessageBuilder {
	mb.ExtensionRanges = ranges
	return mb
}

// AddReservedRange adds the given reserved range to this message. The range is
// inclusive of both the start and end, just like defining a range in proto IDL
// source. This returns the message, for method chaining.
func (mb *MessageBuilder) AddReservedRange(start, end int32) *MessageBuilder {
	rr := &descriptorpb.DescriptorProto_ReservedRange{
		Start: proto.Int32(start),
		End:   proto.Int32(end + 1),
	}
	mb.ReservedRanges = append(mb.ReservedRanges, rr)
	return mb
}

// SetReservedRanges replaces all of this message's reserved ranges with the
// given slice of ranges. Unlike AddReservedRange and unlike the way ranges are
// defined in proto IDL source, a DescriptorProto_ReservedRange struct treats
// the end of the range as *exclusive* (so it would be the value defined in the
// IDL plus one). So the range is inclusive of the start but exclusive of the
// end. This returns the message, for method chaining.
func (mb *MessageBuilder) SetReservedRanges(ranges []*descriptorpb.DescriptorProto_ReservedRange) *MessageBuilder {
	mb.ReservedRanges = ranges
	return mb
}

// AddReservedName adds the given name to the list of reserved field names for
// this message. This returns the message, for method chaining.
func (mb *MessageBuilder) AddReservedName(name string) *MessageBuilder {
	mb.ReservedNames = append(mb.ReservedNames, name)
	return mb
}

// SetReservedNames replaces all of this message's reserved field names with the
// given slice of names. This returns the message, for method chaining.
func (mb *MessageBuilder) SetReservedNames(names []string) *MessageBuilder {
	mb.ReservedNames = names
	return mb
}

func (mb *MessageBuilder) buildProto(path []int32, sourceInfo *descriptorpb.SourceCodeInfo) (*descriptorpb.DescriptorProto, error) {
	addCommentsTo(sourceInfo, path, &mb.comments)

	var needTagsAssigned []*descriptorpb.FieldDescriptorProto
	nestedMessages := make([]*descriptorpb.DescriptorProto, 0, len(mb.nestedMessages))
	oneOfCount := 0
	for _, b := range mb.fieldsAndOneOfs {
		if _, ok := b.(*OneOfBuilder); ok {
			oneOfCount++
		}
	}

	fields := make([]*descriptorpb.FieldDescriptorProto, 0, len(mb.fieldsAndOneOfs)-oneOfCount)
	oneOfs := make([]*descriptorpb.OneofDescriptorProto, 0, oneOfCount)

	addField := func(flb *FieldBuilder, fld *descriptorpb.FieldDescriptorProto) error {
		fields = append(fields, fld)
		if flb.number == 0 {
			needTagsAssigned = append(needTagsAssigned, fld)
		}
		if flb.msgType != nil {
			nmpath := append(path, internal.Message_nestedMessagesTag, int32(len(nestedMessages)))
			if entry, err := flb.msgType.buildProto(nmpath, sourceInfo); err != nil {
				return err
			} else {
				nestedMessages = append(nestedMessages, entry)
			}
		}
		return nil
	}

	for _, b := range mb.fieldsAndOneOfs {
		if flb, ok := b.(*FieldBuilder); ok {
			fldpath := append(path, internal.Message_fieldsTag, int32(len(fields)))
			fld, err := flb.buildProto(fldpath, sourceInfo, mb.Options.GetMessageSetWireFormat())
			if err != nil {
				return nil, err
			}
			if err := addField(flb, fld); err != nil {
				return nil, err
			}
		} else {
			oopath := append(path, internal.Message_oneOfsTag, int32(len(oneOfs)))
			oob := b.(*OneOfBuilder)
			oobIndex := len(oneOfs)
			ood, err := oob.buildProto(oopath, sourceInfo)
			if err != nil {
				return nil, err
			}
			oneOfs = append(oneOfs, ood)
			for _, flb := range oob.choices {
				path := append(path, internal.Message_fieldsTag, int32(len(fields)))
				fld, err := flb.buildProto(path, sourceInfo, mb.Options.GetMessageSetWireFormat())
				if err != nil {
					return nil, err
				}
				fld.OneofIndex = proto.Int32(int32(oobIndex))
				if err := addField(flb, fld); err != nil {
					return nil, err
				}
			}
		}
	}

	if len(needTagsAssigned) > 0 {
		tags := make([]int, len(fields)-len(needTagsAssigned))
		tagsIndex := 0
		for _, fld := range fields {
			tag := fld.GetNumber()
			if tag != 0 {
				tags[tagsIndex] = int(tag)
				tagsIndex++
			}
		}
		sort.Ints(tags)
		t := 1
		for len(needTagsAssigned) > 0 {
			for len(tags) > 0 && t == tags[0] {
				t++
				tags = tags[1:]
			}
			needTagsAssigned[0].Number = proto.Int32(int32(t))
			needTagsAssigned = needTagsAssigned[1:]
			t++
		}
	}

	for _, nmb := range mb.nestedMessages {
		path := append(path, internal.Message_nestedMessagesTag, int32(len(nestedMessages)))
		if nmd, err := nmb.buildProto(path, sourceInfo); err != nil {
			return nil, err
		} else {
			nestedMessages = append(nestedMessages, nmd)
		}
	}

	nestedExtensions := make([]*descriptorpb.FieldDescriptorProto, 0, len(mb.nestedExtensions))
	for _, exb := range mb.nestedExtensions {
		path := append(path, internal.Message_extensionsTag, int32(len(nestedExtensions)))
		if exd, err := exb.buildProto(path, sourceInfo, isExtendeeMessageSet(exb)); err != nil {
			return nil, err
		} else {
			nestedExtensions = append(nestedExtensions, exd)
		}
	}

	nestedEnums := make([]*descriptorpb.EnumDescriptorProto, 0, len(mb.nestedEnums))
	for _, eb := range mb.nestedEnums {
		path := append(path, internal.Message_enumsTag, int32(len(nestedEnums)))
		if ed, err := eb.buildProto(path, sourceInfo); err != nil {
			return nil, err
		} else {
			nestedEnums = append(nestedEnums, ed)
		}
	}

	md := &descriptorpb.DescriptorProto{
		Name:           proto.String(mb.name),
		Options:        mb.Options,
		Field:          fields,
		OneofDecl:      oneOfs,
		NestedType:     nestedMessages,
		EnumType:       nestedEnums,
		Extension:      nestedExtensions,
		ExtensionRange: mb.ExtensionRanges,
		ReservedName:   mb.ReservedNames,
		ReservedRange:  mb.ReservedRanges,
	}

	if mb.GetFile().IsProto3 {
		internal.ProcessProto3OptionalFields(md, nil)
	}

	return md, nil
}

// Build constructs a message descriptor based on the contents of this message
// builder. If there are any problems constructing the descriptor, including
// resolving symbols referenced by the builder or failing to meet certain
// validation rules, an error is returned.
func (mb *MessageBuilder) Build() (*desc.MessageDescriptor, error) {
	md, err := mb.BuildDescriptor()
	if err != nil {
		return nil, err
	}
	return md.(*desc.MessageDescriptor), nil
}

// BuildDescriptor constructs a message descriptor based on the contents of this
// message builder. Most usages will prefer Build() instead, whose return type
// is a concrete descriptor type. This method is present to satisfy the Builder
// interface.
func (mb *MessageBuilder) BuildDescriptor() (desc.Descriptor, error) {
	return doBuild(mb, BuilderOptions{})
}
