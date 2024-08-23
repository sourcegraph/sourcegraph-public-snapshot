// Package builder contains a means of building and modifying proto descriptors
// programmatically. There are numerous factory methods to aid in constructing
// new descriptors as are there methods for converting existing descriptors into
// builders, for modification.
//
// # Factory Functions
//
// Builders are created using the numerous factory functions. Each type of
// descriptor has two kinds of factory functions for the corresponding type of
// builder.
//
// One accepts a descriptor (From*) and returns a copy of it as a builder. The
// descriptor can be manipulated and built to produce a new variant of the
// original. So this is useful for changing existing constructs.
//
// The other kind of factory function (New*) accepts basic arguments and returns
// a new, empty builder (other than the arguments required by that function).
// This second kind can be used to fabricate descriptors from scratch.
//
// Factory functions panic on bad input. Bad input includes invalid names (all
// identifiers must begin with letter or underscore and include only letters,
// numbers, and underscores), invalid types (for example, map field keys must be
// an integer type, bool, or string), invalid tag numbers (anything greater than
// 2^29-1, anything in the range 19,000->19,999, any non-positive number), and
// nil values for required fields (such as field types, RPC method types, and
// extendee type for extensions).
//
// # Auto-Assigning Tag Numbers and File Names
//
// The factory function for fields does not accept a tag number. This is because
// tags, for fields where none is set or is explicitly set to zero, are
// automatically assigned. The tags are assigned in the order the fields were
// added to a message builder. Fields in one-ofs are assigned after other fields
// that were added to the message before the one-of. Within a single one-of,
// fields are assigned tags in the order they were added to the one-of. Across
// one-ofs, fields are assigned in the order their one-of was added to a message
// builder. It is fine if some fields have tags and some do not. The assigned
// tags will start at one and proceed from there but will not conflict with tags
// that were explicitly assigned to fields.
//
// Similarly, when constructing a file builder, a name is accepted but can be
// blank. A blank name means that the file will be given a generated, unique
// name when the descriptor is built.
//
// Note that extensions *must* be given a tag number. Only non-extension fields
// can have their tags auto-assigned. If an extension is constructed with a zero
// tag (which is not valid), the factory function will panic.
//
// # Descriptor Hierarchy
//
// The hierarchy for builders is mutable. A descriptor builder, such as a field,
// can be moved from one message to another. When this is done, the field is
// unlinked from its previous location (so the message to which it previously
// belonged will no longer have any reference to such a field) and linked with
// its new parent. To instead *duplicate* a descriptor builder, its struct value
// can simply be copied. This allows for copying a descriptor from one parent to
// another, like so:
//
//	msg := builder.FromMessage(someMsgDesc)
//	field1 := msg.GetField("foo")
//	field2 := *field1 // field2 is now a copy
//	otherMsg.AddField(&field2)
//
// All descriptors have a link up the hierarchy to the file in which they were
// declared. However, it is *not* necessary to construct this full hierarchy
// with builders. One can create a message builder, for example, and then
// immediately build it to get the descriptor for that message. If it was never
// added to a file then the GetFile() method on the resulting descriptor returns
// a synthetic file that contains only the one message.
//
// Note, however, that this is not true for enum values, methods, and
// non-extension fields. These kinds of builders *must* be added to an enum, a
// service, or a message (respectively) before they can be "built" into a
// descriptor.
//
// When descriptors are created this way, they are created in the default (e.g.
// unnamed) package. In order to put descriptors into a proper package
// namespace, they must be added to a file that has the right package name.
//
// # Builder Pattern and Method Chaining
//
// Each descriptor has some special fields that can only be updated via a Set*
// method. They also all have some exported fields that can be updated by just
// assigning to the field. But even exported fields have accompanying Set*
// methods in order to support a typical method-chaining flow when building
// objects:
//
//	msg, err := builder.NewMessage("MyMessage").
//	    AddField(NewField("foo", FieldTypeScalar(descriptor.FieldDescriptorProto_TYPE_STRING)).
//	        SetDefaultValue("bar")).
//	    AddField(NewField("baz", FieldTypeScalar(descriptor.FieldDescriptorProto_TYPE_INT64)).
//	        SetLabel(descriptor.FieldDescriptorProto_LABEL_REPEATED).
//	        SetOptions(&descriptor.FieldOptions{Packed: proto.Bool(true)})).
//	    Build()
//
// So the various Set* methods all return the builder itself so that multiple
// fields may be set in a single invocation chain.
//
// The Set* operations which perform validations also have a TrySet* form which
// can return an error if validation fails. If the method-chaining Set* form is
// used with inputs that fail validation, the Set* method will panic.
//
// # Type References and Imported Types
//
// When defining fields whose type is a message or enum and when defining
// methods (whose request and response type are a message), the type can be set
// to an actual descriptor (e.g. a *desc.MessageDescriptor) or to a builder for
// the type (e.g. a *builder.MessageBuilder). Since Go does not allow method
// overloading, the naming convention is that types referring to descriptors are
// "imported types" (since their use will result in an import statement in the
// resulting file descriptor, to import the file in which the type was defined.)
//
// When referring to other builders, it is not necessary that the referenced
// types be in the same file. When building the descriptors, multiple file
// descriptors may be created, so that all referenced builders are themselves
// resolved into descriptors.
//
// However, it is illegal to create an import cycle. So if two messages, for
// example, refer to each other (message Foo has a field "bar" of type Bar, and
// message Bar has a field "foo" of type Foo), they must explicitly be assigned
// to the same file builder. If they are not assigned to any files, they will be
// assigned to synthetic files which would result in an import cycle (each file
// imports the other). And the same would be true if one or both files were
// explicitly assigned to a file, but not both to the same file.
//
// # Validations and Caveats
//
// Descriptors that are attained from a builder do not necessarily represent a
// valid construct in the proto source language. There are some validations
// enforced by protoc that are not enforced by builders, for example, ensuring
// that there are no namespace conflicts (e.g. file "foo.proto" declares an
// element named "pkg.bar" and so does a file that it imports). Because of this,
// it is possible for builders to wire up references in a way that the resulting
// descriptors are incorrect. This is mainly possible when file builders are
// used to create files with duplicate symbols and then cross-linked. It can
// also happen when a builder is linked to descriptors from more than one
// version of the same file.
//
// When constructing descriptors using builders, applications should not attempt
// to build invalid constructs. Even though there are many rules in the protobuf
// language that are not enforced, those rules that are enforced can result in
// panics when a violation is detected. Generally, builder methods that do not
// return errors (like those used for method chaining) will panic on bad input
// or when trying to mutate a proto into an invalid state.
//
// Several rules are enforced by the builders. Violating these rules will result
// in errors (or panics for factory functions and methods that do not return
// errors). These are the rules that are currently enforced:
//
//  1. Import cycles are not allowed. (See above for more details.)
//  2. Within a single file, symbols are not allowed to have naming conflicts.
//     This means that is not legal to create a message and an extension with
//     the same name in the same file.
//  3. Messages are not allowed to have multiple fields with the same tag. Note
//     that only non-extension fields are checked when using builders. So
//     builders will allow tag collisions for extensions. (Use caution.)
//  4. Map keys can only be integer types, booleans, and strings.
//  5. Fields cannot have tags in the special reserved range 19000-19999. Also
//     the maximum allowed tag value is 536870911 (2^29 - 1). Finally, fields
//     cannot have negative values.
//  6. Element names should include only underscore, letters, and numbers, and
//     must begin with an underscore or letter.
//  7. Files with a syntax of proto3 are not allowed to have required fields.
//  8. Files with a syntax of proto3 are not allowed to have messages that
//     define extension ranges.
//  9. Files with a syntax of proto3 are not allowed to use groups.
//  10. Files with a syntax of proto3 are not allowed to declare default values
//     for fields.
//  11. Extension fields must use tag numbers that are in an extension range
//     defined on the extended message.
//  12. Non-extension fields are not allowed to use tags that lie in a message's
//     extension ranges or reserved ranges.
//  13. Non-extension fields are not allowed to use names that the message has
//     marked as reserved.
//  14. Extension ranges and reserved ranges must not overlap.
//
// Validation rules that are *not* enforced by builders, and thus would be
// allowed and result in illegal constructs, include the following:
//
//  1. Names are supposed to be globally unique, even across multiple files
//     if multiple files are defined in the same package.
//  2. Multiple extensions for the same message cannot re-use tag numbers, even
//     across multiple files.
package builder
