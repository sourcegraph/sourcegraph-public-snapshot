package builder

import (
	"bytes"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

// Builder is the core interface implemented by all descriptor builders. It
// exposes some basic information about the descriptor hierarchy's structure.
//
// All Builders also have a Build() method, but that is not part of this
// interface because its return type varies with the type of descriptor that
// is built.
type Builder interface {
	// GetName returns this element's name. The name returned is a simple name,
	// not a qualified name.
	GetName() string

	// TrySetName attempts to set this element's name. If the rename cannot
	// proceed (e.g. this element's parent already has an element with that
	// name) then an error is returned.
	//
	// All builders also have a method named SetName that panics on error and
	// returns the builder itself (for method chaining). But that isn't defined
	// on this interface because its return type varies with the type of the
	// descriptor builder.
	TrySetName(newName string) error

	// GetParent returns this element's parent element. It returns nil if there
	// is no parent element. File builders never have parent elements.
	GetParent() Builder

	// GetFile returns this element's file. This returns nil if the element has
	// not yet been assigned to a file.
	GetFile() *FileBuilder

	// GetChildren returns all of this element's child elements. A file will
	// return all of its top-level messages, enums, extensions, and services. A
	// message will return all of its fields as well as nested messages, enums,
	// and extensions, etc. Children will generally be grouped by type and,
	// within a group, in the same order as the children were added to their
	// parent.
	GetChildren() []Builder

	// GetComments returns the comments for this element. If the element has no
	// comments then the returned struct will have all empty fields. Comments
	// can be added to the element by setting fields of the returned struct.
	//
	// All builders also have a SetComments method that modifies the comments
	// and returns the builder itself (for method chaining). But that isn't
	// defined on this interface because its return type varies with the type of
	// the descriptor builder.
	GetComments() *Comments

	// BuildDescriptor is a generic form of the Build method. Its declared
	// return type is general so that it can be included in this interface and
	// implemented by all concrete builder types.
	//
	// If the builder includes references to custom options, only those known to
	// the calling program (i.e. linked in and registered with the proto
	// package) can be correctly interpreted. If the builder references other
	// custom options, use BuilderOptions.Build instead.
	BuildDescriptor() (desc.Descriptor, error)

	// findChild returns the child builder with the given name or nil if this
	// builder has no such child.
	findChild(string) Builder

	// removeChild removes the given child builder from this element. If the
	// given element is not a child, it should do nothing.
	//
	// NOTE: It is this method's responsibility to call child.setParent(nil)
	// after removing references to the child from this element.
	removeChild(Builder)

	// renamedChild updates references by-name references to the given child and
	// validates its name. The given string is the child's old name. If the
	// rename can proceed, no error should be returned and any by-name
	// references to the old name should be removed.
	renamedChild(Builder, string) error

	// setParent simply updates the up-link (from child to parent) so that the
	// this element's parent is up-to-date. It does NOT try to remove references
	// from the parent to this child. (See doc for removeChild(Builder)).
	setParent(Builder)
}

// BuilderOptions includes additional options to use when building descriptors.
type BuilderOptions struct {
	// This registry provides definitions for custom options. If a builder
	// refers to an option that is not known by this registry, it can still be
	// interpreted if the extension is "known" to the calling program (i.e.
	// linked in and registered with the proto package).
	Extensions *dynamic.ExtensionRegistry

	// If this option is true, then all options referred to in builders must
	// be interpreted. That means that if an option is present that is neither
	// recognized by Extenions nor known to the calling program, trying to build
	// the descriptor will fail.
	RequireInterpretedOptions bool
}

// Build processes the given builder into a descriptor using these options.
// Using the builder's Build() or BuildDescriptor() method is equivalent to
// building with a zero-value BuilderOptions.
func (opts BuilderOptions) Build(b Builder) (desc.Descriptor, error) {
	return doBuild(b, opts)
}

// Comments represents the various comments that might be associated with a
// descriptor. These are equivalent to the various kinds of comments found in a
// *dpb.SourceCodeInfo_Location struct that protoc associates with elements in
// the parsed proto source file. This can be used to create or preserve comments
// (including documentation) for elements.
type Comments struct {
	LeadingDetachedComments []string
	LeadingComment          string
	TrailingComment         string
}

func setComments(c *Comments, loc *descriptorpb.SourceCodeInfo_Location) {
	c.LeadingDetachedComments = loc.GetLeadingDetachedComments()
	c.LeadingComment = loc.GetLeadingComments()
	c.TrailingComment = loc.GetTrailingComments()
}

func addCommentsTo(sourceInfo *descriptorpb.SourceCodeInfo, path []int32, c *Comments) {
	var lead, trail *string
	if c.LeadingComment != "" {
		lead = proto.String(c.LeadingComment)
	}
	if c.TrailingComment != "" {
		trail = proto.String(c.TrailingComment)
	}

	// we need defensive copies of the slices
	p := make([]int32, len(path))
	copy(p, path)

	var detached []string
	if len(c.LeadingDetachedComments) > 0 {
		detached = make([]string, len(c.LeadingDetachedComments))
		copy(detached, c.LeadingDetachedComments)
	}

	sourceInfo.Location = append(sourceInfo.Location, &descriptorpb.SourceCodeInfo_Location{
		LeadingDetachedComments: detached,
		LeadingComments:         lead,
		TrailingComments:        trail,
		Path:                    p,
		Span:                    []int32{0, 0, 0},
	})
}

/* NB: There are a few flows that need to maintain strong referential integrity
 * and perform symbol and/or number uniqueness checks. The way these flows are
 * implemented is described below. The actions generally involve two different
 * components: making local changes to an element and making corresponding
 * and/or related changes in a parent element. Below describes the separation of
 * responsibilities between the two.
 *
 *
 * RENAMING AN ELEMENT
 *
 * Renaming an element is initiated via Builder.TrySetName. Implementations
 * should do the following:
 *  1. Validate the new name using any local constraints and naming rules.
 *  2. If there are child elements whose names should be kept in sync in some
 *     way, rename them.
 *  3. Invoke baseBuilder.setName. This changes this element's name and then
 *     invokes Builder.renamedChild(child, oldName) to update any by-name
 *     references from the parent to the child.
 *  4. If step #3 failed, any other element names that were changed to keep
 *     them in sync (from step #2) should be reverted.
 *
 * A key part of this flow is how parents react to child elements being renamed.
 * This is done in Builder.renamedChild. Implementations should do the
 * following:
 *  1. Validate the name using any local constraints. (Often there are no new
 *     constraints and any checks already done by Builder.TrySetName should
 *     suffice.)
 *  2. If the parent element should be renamed to keep it in sync with the
 *     child's name, rename it.
 *  3. Register references to the element using the new name. A possible cause
 *     of error in this step is a uniqueness constraint, e.g. the element's new
 *     name collides with a sibling element's name.
 *  4. If step #3 failed and this element name was changed to keep it in sync
 *     (from step #2), it should be reverted.
 *  5. Finally, remove references to the element for the old name. This step
 *     should always succeed.
 *
 * Changing the tag number for a non-extension field has a similar flow since it
 * is also checked for uniqueness, to make sure the new tag number does not
 * conflict with another existing field.
 *
 * Note that TrySetName and renamedChild methods both can return an error, which
 * should indicate why the element could not be renamed (e.g. name is invalid,
 * new name conflicts with existing sibling names, etc).
 *
 *
 * MOVING/REMOVING AN ELEMENT
 *
 * When an element is added to a new parent but is already assigned to a parent,
 * it is "moved" to the new parent. This is done via "Add" methods on the parent
 * entity (for example, MessageBuilder.AddField). Implementations of such a
 * method should do the following:
 *  1. Register references to the element. A possible cause of failure in this
 *     step is that the new element collides with an existing child.
 *  2. Use the Unlink function to remove the element from any existing parent.
 *  3. Use Builder.setParent to link the child to its parent.
 *
 * The Unlink function, which removes an element from its parent if it has a
 * parent, relies on the parent's Builder.removeChild method. Implementations of
 * that method should do the following:
 *  1. Check that the element is actually a child. If not, return without doing
 *     anything.
 *  2. Remove all references to the child.
 *  3. Finally, this method must call Builder.setParent(nil) to clear the
 *     element's up-link so it no longer refers to the old parent.
 *
 * The "Add" methods typically have a "Try" form which can return an error. This
 * could happen if the new child is not legal to add (including, for example,
 * that its name collides with an existing child element).
 *
 * The removeChild and setParent methods, on the other hand, cannot return an
 * error and thus must always succeed.
 */

// baseBuilder is a struct that can be embedded into each Builder implementation
// and provides a kernel of builder-wiring support (to reduce boiler-plate in
// each implementation).
type baseBuilder struct {
	name     string
	parent   Builder
	comments Comments
}

func baseBuilderWithName(name string) baseBuilder {
	if err := checkName(name); err != nil {
		panic(err)
	}
	return baseBuilder{name: name}
}

func checkName(name string) error {
	for i, ch := range name {
		if i == 0 {
			if ch != '_' && (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') {
				return fmt.Errorf("name %q is invalid; It must start with an underscore or letter", name)
			}
		} else {
			if ch != '_' && (ch < '0' || ch > '9') && (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') {
				return fmt.Errorf("name %q contains invalid character %q; only underscores, letters, and numbers are allowed", name, string(ch))
			}
		}
	}
	return nil
}

// GetName returns the name of the element that will be built by this builder.
func (b *baseBuilder) GetName() string {
	return b.name
}

func (b *baseBuilder) setName(fullBuilder Builder, newName string) error {
	if newName == b.name {
		return nil // no change
	}
	if err := checkName(newName); err != nil {
		return err
	}
	oldName := b.name
	b.name = newName
	if b.parent != nil {
		if err := b.parent.renamedChild(fullBuilder, oldName); err != nil {
			// revert the rename on error
			b.name = oldName
			return err
		}
	}
	return nil
}

// GetParent returns the parent builder to which this builder has been added. If
// the builder has not been added to another, this returns nil.
//
// The parents of message builders will be file builders or other message
// builders. Same for the parents of extension field builders and enum builders.
// One-of builders and non-extension field builders will return a message
// builder. Method builders' parents are service builders; enum value builders'
// parents are enum builders. Finally, service builders will always return file
// builders as their parent.
func (b *baseBuilder) GetParent() Builder {
	return b.parent
}

func (b *baseBuilder) setParent(newParent Builder) {
	b.parent = newParent
}

// GetFile returns the file to which this builder is assigned. This examines the
// builder's parent, and its parent, and so on, until it reaches a file builder
// or nil.
//
// If the builder is not assigned to a file (even transitively), this method
// returns nil.
func (b *baseBuilder) GetFile() *FileBuilder {
	p := b.parent
	for p != nil {
		if fb, ok := p.(*FileBuilder); ok {
			return fb
		}
		p = p.GetParent()
	}
	return nil
}

// GetComments returns comments associated with the element that will be built
// by this builder.
func (b *baseBuilder) GetComments() *Comments {
	return &b.comments
}

// doBuild is a helper for implementing the Build() method that each builder
// exposes. It is used for all builders except for the root FileBuilder type.
func doBuild(b Builder, opts BuilderOptions) (desc.Descriptor, error) {
	fd, err := newResolver(opts).resolveElement(b, nil)
	if err != nil {
		return nil, err
	}
	if _, ok := b.(*FileBuilder); ok {
		return fd, nil
	}
	return fd.FindSymbol(GetFullyQualifiedName(b)), nil
}

func getFullyQualifiedName(b Builder, buf *bytes.Buffer) {
	if fb, ok := b.(*FileBuilder); ok {
		buf.WriteString(fb.Package)
	} else if b != nil {
		p := b.GetParent()
		if _, ok := p.(*FieldBuilder); ok {
			// field can be the parent of a message (if it's
			// the field's map entry or group type), but its
			// name is not part of message's fqn; so skip
			p = p.GetParent()
		}
		if _, ok := p.(*OneOfBuilder); ok {
			// one-of can be the parent of a field, but its
			// name is not part of field's fqn; so skip
			p = p.GetParent()
		}
		getFullyQualifiedName(p, buf)
		if buf.Len() > 0 {
			buf.WriteByte('.')
		}
		buf.WriteString(b.GetName())
	}
}

// GetFullyQualifiedName returns the given builder's fully-qualified name. This
// name is based on the parent elements the builder may be linked to, which
// provide context like package and (optional) enclosing message names.
func GetFullyQualifiedName(b Builder) string {
	var buf bytes.Buffer
	getFullyQualifiedName(b, &buf)
	return buf.String()
}

// Unlink removes the given builder from its parent. The parent will no longer
// refer to the builder and vice versa.
func Unlink(b Builder) {
	if p := b.GetParent(); p != nil {
		p.removeChild(b)
	}
}

// getRoot navigates up the hierarchy to find the root builder for the given
// instance.
func getRoot(b Builder) Builder {
	for {
		p := b.GetParent()
		if p == nil {
			return b
		}
		b = p
	}
}

// deleteBuilder will delete a descriptor builder with the given name from the
// given slice. The slice's elements can be any builder type. The parameter has
// type interface{} so it can accept []*MessageBuilder or []*FieldBuilder, for
// example. It returns a value of the same type with the named builder omitted.
func deleteBuilder(name string, descs interface{}) interface{} {
	rv := reflect.ValueOf(descs)
	for i := 0; i < rv.Len(); i++ {
		c := rv.Index(i).Interface().(Builder)
		if c.GetName() == name {
			head := rv.Slice(0, i)
			tail := rv.Slice(i+1, rv.Len())
			return reflect.AppendSlice(head, tail).Interface()
		}
	}
	return descs
}
