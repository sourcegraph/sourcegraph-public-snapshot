package builder

import (
	"fmt"
	"sort"
	"strings"
	"sync/atomic"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/internal"
	"github.com/jhump/protoreflect/dynamic"
)

var uniqueFileCounter uint64

func uniqueFileName() string {
	i := atomic.AddUint64(&uniqueFileCounter, 1)
	return fmt.Sprintf("{generated-file-%04x}.proto", i)
}

func makeUnique(name string, existingNames map[string]struct{}) string {
	i := 1
	n := name
	for {
		if _, ok := existingNames[n]; !ok {
			return n
		}
		n = fmt.Sprintf("%s(%d)", name, i)
		i++
	}
}

// FileBuilder is a builder used to construct a desc.FileDescriptor. This is the
// root of the hierarchy. All other descriptors belong to a file, and thus all
// other builders also belong to a file.
//
// If a builder is *not* associated with a file, the resulting descriptor will
// be associated with a synthesized file that contains only the built descriptor
// and its ancestors. This means that such descriptors will have no associated
// package name.
//
// To create a new FileBuilder, use NewFile.
type FileBuilder struct {
	name string

	IsProto3 bool
	Package  string
	Options  *descriptorpb.FileOptions

	comments        Comments
	SyntaxComments  Comments
	PackageComments Comments

	messages   []*MessageBuilder
	extensions []*FieldBuilder
	enums      []*EnumBuilder
	services   []*ServiceBuilder
	symbols    map[string]Builder

	origExts        *dynamic.ExtensionRegistry
	explicitDeps    map[*FileBuilder]struct{}
	explicitImports map[*desc.FileDescriptor]struct{}
}

// NewFile creates a new FileBuilder for a file with the given name. The
// name can be blank, which indicates a unique name should be generated for it.
func NewFile(name string) *FileBuilder {
	return &FileBuilder{
		name:    name,
		symbols: map[string]Builder{},
	}
}

// FromFile returns a FileBuilder that is effectively a copy of the given
// descriptor. Note that builders do not retain full source code info, even if
// the given descriptor included it. Instead, comments are extracted from the
// given descriptor's source info (if present) and, when built, the resulting
// descriptor will have just the comment info (no location information).
func FromFile(fd *desc.FileDescriptor) (*FileBuilder, error) {
	fb := NewFile(fd.GetName())
	fb.IsProto3 = fd.IsProto3()
	fb.Package = fd.GetPackage()
	fb.Options = fd.GetFileOptions()
	setComments(&fb.comments, fd.GetSourceInfo())

	// find syntax and package comments, too
	for _, loc := range fd.AsFileDescriptorProto().GetSourceCodeInfo().GetLocation() {
		if len(loc.Path) == 1 {
			if loc.Path[0] == internal.File_syntaxTag {
				setComments(&fb.SyntaxComments, loc)
			} else if loc.Path[0] == internal.File_packageTag {
				setComments(&fb.PackageComments, loc)
			}
		}
	}

	// add imports explicitly
	for _, dep := range fd.GetDependencies() {
		fb.AddImportedDependency(dep)
		fb.addExtensionsFromImport(dep)
	}

	localMessages := map[*desc.MessageDescriptor]*MessageBuilder{}
	localEnums := map[*desc.EnumDescriptor]*EnumBuilder{}

	for _, md := range fd.GetMessageTypes() {
		if mb, err := fromMessage(md, localMessages, localEnums); err != nil {
			return nil, err
		} else if err := fb.TryAddMessage(mb); err != nil {
			return nil, err
		}
	}
	for _, ed := range fd.GetEnumTypes() {
		if eb, err := fromEnum(ed, localEnums); err != nil {
			return nil, err
		} else if err := fb.TryAddEnum(eb); err != nil {
			return nil, err
		}
	}
	for _, exd := range fd.GetExtensions() {
		if exb, err := fromField(exd); err != nil {
			return nil, err
		} else if err := fb.TryAddExtension(exb); err != nil {
			return nil, err
		}
	}
	for _, sd := range fd.GetServices() {
		if sb, err := fromService(sd); err != nil {
			return nil, err
		} else if err := fb.TryAddService(sb); err != nil {
			return nil, err
		}
	}

	// we've converted everything, so now we update all foreign type references
	// to be local type references if possible
	for _, mb := range fb.messages {
		updateLocalRefsInMessage(mb, localMessages, localEnums)
	}
	for _, exb := range fb.extensions {
		updateLocalRefsInField(exb, localMessages, localEnums)
	}
	for _, sb := range fb.services {
		for _, mtb := range sb.methods {
			updateLocalRefsInRpcType(mtb.ReqType, localMessages)
			updateLocalRefsInRpcType(mtb.RespType, localMessages)
		}
	}

	return fb, nil
}

func updateLocalRefsInMessage(mb *MessageBuilder, localMessages map[*desc.MessageDescriptor]*MessageBuilder, localEnums map[*desc.EnumDescriptor]*EnumBuilder) {
	for _, b := range mb.fieldsAndOneOfs {
		if flb, ok := b.(*FieldBuilder); ok {
			updateLocalRefsInField(flb, localMessages, localEnums)
		} else {
			oob := b.(*OneOfBuilder)
			for _, flb := range oob.choices {
				updateLocalRefsInField(flb, localMessages, localEnums)
			}
		}
	}
	for _, nmb := range mb.nestedMessages {
		updateLocalRefsInMessage(nmb, localMessages, localEnums)
	}
	for _, exb := range mb.nestedExtensions {
		updateLocalRefsInField(exb, localMessages, localEnums)
	}
}

func updateLocalRefsInField(flb *FieldBuilder, localMessages map[*desc.MessageDescriptor]*MessageBuilder, localEnums map[*desc.EnumDescriptor]*EnumBuilder) {
	if flb.fieldType.foreignMsgType != nil {
		if mb, ok := localMessages[flb.fieldType.foreignMsgType]; ok {
			flb.fieldType.foreignMsgType = nil
			flb.fieldType.localMsgType = mb
		}
	}
	if flb.fieldType.foreignEnumType != nil {
		if eb, ok := localEnums[flb.fieldType.foreignEnumType]; ok {
			flb.fieldType.foreignEnumType = nil
			flb.fieldType.localEnumType = eb
		}
	}
	if flb.foreignExtendee != nil {
		if mb, ok := localMessages[flb.foreignExtendee]; ok {
			flb.foreignExtendee = nil
			flb.localExtendee = mb
		}
	}
	if flb.msgType != nil {
		updateLocalRefsInMessage(flb.msgType, localMessages, localEnums)
	}
}

func updateLocalRefsInRpcType(rpcType *RpcType, localMessages map[*desc.MessageDescriptor]*MessageBuilder) {
	if rpcType.foreignType != nil {
		if mb, ok := localMessages[rpcType.foreignType]; ok {
			rpcType.foreignType = nil
			rpcType.localType = mb
		}
	}
}

// GetName returns the name of the file. It may include relative path
// information, too.
func (fb *FileBuilder) GetName() string {
	return fb.name
}

// SetName changes this file's name, returning the file builder for method
// chaining.
func (fb *FileBuilder) SetName(newName string) *FileBuilder {
	fb.name = newName
	return fb
}

// TrySetName changes this file's name. It always returns nil since renaming
// a file cannot fail. (It is specified to return error to satisfy the Builder
// interface.)
func (fb *FileBuilder) TrySetName(newName string) error {
	fb.name = newName
	return nil
}

// GetParent always returns nil since files are the roots of builder
// hierarchies.
func (fb *FileBuilder) GetParent() Builder {
	return nil
}

func (fb *FileBuilder) setParent(parent Builder) {
	if parent != nil {
		panic("files cannot have parent elements")
	}
}

// GetComments returns comments associated with the file itself and not any
// particular element therein. (Note that such a comment will not be rendered by
// the protoprint package.)
func (fb *FileBuilder) GetComments() *Comments {
	return &fb.comments
}

// SetComments sets the comments associated with the file itself, not any
// particular element therein. (Note that such a comment will not be rendered by
// the protoprint package.) This method returns the file, for method chaining.
func (fb *FileBuilder) SetComments(c Comments) *FileBuilder {
	fb.comments = c
	return fb
}

// SetSyntaxComments sets the comments associated with the syntax declaration
// element (which, if present, is required to be the first element in a proto
// file). This method returns the file, for method chaining.
func (fb *FileBuilder) SetSyntaxComments(c Comments) *FileBuilder {
	fb.SyntaxComments = c
	return fb
}

// SetPackageComments sets the comments associated with the package declaration
// element. (This comment will not be rendered if the file's declared package is
// empty.) This method returns the file, for method chaining.
func (fb *FileBuilder) SetPackageComments(c Comments) *FileBuilder {
	fb.PackageComments = c
	return fb
}

// GetFile implements the Builder interface and always returns this file.
func (fb *FileBuilder) GetFile() *FileBuilder {
	return fb
}

// GetChildren returns builders for all nested elements, including all top-level
// messages, enums, extensions, and services.
func (fb *FileBuilder) GetChildren() []Builder {
	var ch []Builder
	for _, mb := range fb.messages {
		ch = append(ch, mb)
	}
	for _, exb := range fb.extensions {
		ch = append(ch, exb)
	}
	for _, eb := range fb.enums {
		ch = append(ch, eb)
	}
	for _, sb := range fb.services {
		ch = append(ch, sb)
	}
	return ch
}

func (fb *FileBuilder) findChild(name string) Builder {
	return fb.symbols[name]
}

func (fb *FileBuilder) removeChild(b Builder) {
	if p, ok := b.GetParent().(*FileBuilder); !ok || p != fb {
		return
	}

	switch b.(type) {
	case *MessageBuilder:
		fb.messages = deleteBuilder(b.GetName(), fb.messages).([]*MessageBuilder)
	case *FieldBuilder:
		fb.extensions = deleteBuilder(b.GetName(), fb.extensions).([]*FieldBuilder)
	case *EnumBuilder:
		fb.enums = deleteBuilder(b.GetName(), fb.enums).([]*EnumBuilder)
	case *ServiceBuilder:
		fb.services = deleteBuilder(b.GetName(), fb.services).([]*ServiceBuilder)
	}
	delete(fb.symbols, b.GetName())
	b.setParent(nil)
}

func (fb *FileBuilder) renamedChild(b Builder, oldName string) error {
	if p, ok := b.GetParent().(*FileBuilder); !ok || p != fb {
		return nil
	}

	if err := fb.addSymbol(b); err != nil {
		return err
	}
	delete(fb.symbols, oldName)
	return nil
}

func (fb *FileBuilder) addSymbol(b Builder) error {
	if ex, ok := fb.symbols[b.GetName()]; ok {
		return fmt.Errorf("file %q already contains element (%T) named %q", fb.GetName(), ex, b.GetName())
	}
	fb.symbols[b.GetName()] = b
	return nil
}

func (fb *FileBuilder) findFullyQualifiedElement(fqn string) Builder {
	if fb.Package != "" {
		if !strings.HasPrefix(fqn, fb.Package+".") {
			return nil
		}
		fqn = fqn[len(fb.Package)+1:]
	}
	names := strings.Split(fqn, ".")
	var b Builder = fb
	for b != nil && len(names) > 0 {
		b = b.findChild(names[0])
		names = names[1:]
	}
	return b
}

// GetMessage returns the top-level message with the given name. If no such
// message exists in the file, nil is returned.
func (fb *FileBuilder) GetMessage(name string) *MessageBuilder {
	b := fb.symbols[name]
	if mb, ok := b.(*MessageBuilder); ok {
		return mb
	} else {
		return nil
	}
}

// RemoveMessage removes the top-level message with the given name. If no such
// message exists in the file, this is a no-op. This returns the file builder,
// for method chaining.
func (fb *FileBuilder) RemoveMessage(name string) *FileBuilder {
	fb.TryRemoveMessage(name)
	return fb
}

// TryRemoveMessage removes the top-level message with the given name and
// returns false if the file has no such message.
func (fb *FileBuilder) TryRemoveMessage(name string) bool {
	b := fb.symbols[name]
	if mb, ok := b.(*MessageBuilder); ok {
		fb.removeChild(mb)
		return true
	}
	return false
}

// AddMessage adds the given message to this file. If an error prevents the
// message from being added, this method panics. This returns the file builder,
// for method chaining.
func (fb *FileBuilder) AddMessage(mb *MessageBuilder) *FileBuilder {
	if err := fb.TryAddMessage(mb); err != nil {
		panic(err)
	}
	return fb
}

// TryAddMessage adds the given message to this file, returning any error that
// prevents the message from being added (such as a name collision with another
// element already added to the file).
func (fb *FileBuilder) TryAddMessage(mb *MessageBuilder) error {
	if err := fb.addSymbol(mb); err != nil {
		return err
	}
	Unlink(mb)
	mb.setParent(fb)
	fb.messages = append(fb.messages, mb)
	return nil
}

// GetExtension returns the top-level extension with the given name. If no such
// extension exists in the file, nil is returned.
func (fb *FileBuilder) GetExtension(name string) *FieldBuilder {
	b := fb.symbols[name]
	if exb, ok := b.(*FieldBuilder); ok {
		return exb
	} else {
		return nil
	}
}

// RemoveExtension removes the top-level extension with the given name. If no
// such extension exists in the file, this is a no-op. This returns the file
// builder, for method chaining.
func (fb *FileBuilder) RemoveExtension(name string) *FileBuilder {
	fb.TryRemoveExtension(name)
	return fb
}

// TryRemoveExtension removes the top-level extension with the given name and
// returns false if the file has no such extension.
func (fb *FileBuilder) TryRemoveExtension(name string) bool {
	b := fb.symbols[name]
	if exb, ok := b.(*FieldBuilder); ok {
		fb.removeChild(exb)
		return true
	}
	return false
}

// AddExtension adds the given extension to this file. If an error prevents the
// extension from being added, this method panics. This returns the file
// builder, for method chaining.
func (fb *FileBuilder) AddExtension(exb *FieldBuilder) *FileBuilder {
	if err := fb.TryAddExtension(exb); err != nil {
		panic(err)
	}
	return fb
}

// TryAddExtension adds the given extension to this file, returning any error
// that prevents the extension from being added (such as a name collision with
// another element already added to the file).
func (fb *FileBuilder) TryAddExtension(exb *FieldBuilder) error {
	if !exb.IsExtension() {
		return fmt.Errorf("field %s is not an extension", exb.GetName())
	}
	if err := fb.addSymbol(exb); err != nil {
		return err
	}
	Unlink(exb)
	exb.setParent(fb)
	fb.extensions = append(fb.extensions, exb)
	return nil
}

// GetEnum returns the top-level enum with the given name. If no such enum
// exists in the file, nil is returned.
func (fb *FileBuilder) GetEnum(name string) *EnumBuilder {
	b := fb.symbols[name]
	if eb, ok := b.(*EnumBuilder); ok {
		return eb
	} else {
		return nil
	}
}

// RemoveEnum removes the top-level enum with the given name. If no such enum
// exists in the file, this is a no-op. This returns the file builder, for
// method chaining.
func (fb *FileBuilder) RemoveEnum(name string) *FileBuilder {
	fb.TryRemoveEnum(name)
	return fb
}

// TryRemoveEnum removes the top-level enum with the given name and returns
// false if the file has no such enum.
func (fb *FileBuilder) TryRemoveEnum(name string) bool {
	b := fb.symbols[name]
	if eb, ok := b.(*EnumBuilder); ok {
		fb.removeChild(eb)
		return true
	}
	return false
}

// AddEnum adds the given enum to this file. If an error prevents the enum from
// being added, this method panics. This returns the file builder, for method
// chaining.
func (fb *FileBuilder) AddEnum(eb *EnumBuilder) *FileBuilder {
	if err := fb.TryAddEnum(eb); err != nil {
		panic(err)
	}
	return fb
}

// TryAddEnum adds the given enum to this file, returning any error that
// prevents the enum from being added (such as a name collision with another
// element already added to the file).
func (fb *FileBuilder) TryAddEnum(eb *EnumBuilder) error {
	if err := fb.addSymbol(eb); err != nil {
		return err
	}
	Unlink(eb)
	eb.setParent(fb)
	fb.enums = append(fb.enums, eb)
	return nil
}

// GetService returns the top-level service with the given name. If no such
// service exists in the file, nil is returned.
func (fb *FileBuilder) GetService(name string) *ServiceBuilder {
	b := fb.symbols[name]
	if sb, ok := b.(*ServiceBuilder); ok {
		return sb
	} else {
		return nil
	}
}

// RemoveService removes the top-level service with the given name. If no such
// service exists in the file, this is a no-op. This returns the file builder,
// for method chaining.
func (fb *FileBuilder) RemoveService(name string) *FileBuilder {
	fb.TryRemoveService(name)
	return fb
}

// TryRemoveService removes the top-level service with the given name and
// returns false if the file has no such service.
func (fb *FileBuilder) TryRemoveService(name string) bool {
	b := fb.symbols[name]
	if sb, ok := b.(*ServiceBuilder); ok {
		fb.removeChild(sb)
		return true
	}
	return false
}

// AddService adds the given service to this file. If an error prevents the
// service from being added, this method panics. This returns the file builder,
// for method chaining.
func (fb *FileBuilder) AddService(sb *ServiceBuilder) *FileBuilder {
	if err := fb.TryAddService(sb); err != nil {
		panic(err)
	}
	return fb
}

// TryAddService adds the given service to this file, returning any error that
// prevents the service from being added (such as a name collision with another
// element already added to the file).
func (fb *FileBuilder) TryAddService(sb *ServiceBuilder) error {
	if err := fb.addSymbol(sb); err != nil {
		return err
	}
	Unlink(sb)
	sb.setParent(fb)
	fb.services = append(fb.services, sb)
	return nil
}

func (fb *FileBuilder) addExtensionsFromImport(dep *desc.FileDescriptor) {
	if fb.origExts == nil {
		fb.origExts = &dynamic.ExtensionRegistry{}
	}
	fb.origExts.AddExtensionsFromFile(dep)
	// we also add any extensions from this dependency's "public" imports since
	// they are also visible to the importing file
	for _, publicDep := range dep.GetPublicDependencies() {
		fb.addExtensionsFromImport(publicDep)
	}
}

// AddDependency adds the given file as an explicit import. Normally,
// dependencies can be inferred during the build process by finding the files
// for all referenced types (such as message and enum types used in this file).
// However, this does not work for custom options, which must be known in order
// to be interpretable. And they aren't known unless an explicit import is added
// for the file that contains the custom options.
//
// Knowledge of custom options can also be provided by using BuilderOptions with
// an ExtensionRegistry, when building the file.
func (fb *FileBuilder) AddDependency(dep *FileBuilder) *FileBuilder {
	if fb.explicitDeps == nil {
		fb.explicitDeps = map[*FileBuilder]struct{}{}
	}
	fb.explicitDeps[dep] = struct{}{}
	return fb
}

// AddImportedDependency adds the given file as an explicit import. Normally,
// dependencies can be inferred during the build process by finding the files
// for all referenced types (such as message and enum types used in this file).
// However, this does not work for custom options, which must be known in order
// to be interpretable. And they aren't known unless an explicit import is added
// for the file that contains the custom options.
//
// Knowledge of custom options can also be provided by using BuilderOptions with
// an ExtensionRegistry, when building the file.
func (fb *FileBuilder) AddImportedDependency(dep *desc.FileDescriptor) *FileBuilder {
	if fb.explicitImports == nil {
		fb.explicitImports = map[*desc.FileDescriptor]struct{}{}
	}
	fb.explicitImports[dep] = struct{}{}
	return fb
}

// PruneUnusedDependencies removes all imports that are not actually used in the
// file. Note that this undoes any calls to AddDependency or AddImportedDependency
// which means that custom options may be missing from the resulting built
// descriptor unless BuilderOptions are used that include an ExtensionRegistry with
// knowledge of all custom options.
//
// When FromFile is used to create a FileBuilder from an existing descriptor, all
// imports are usually preserved in any subsequent built descriptor. But this method
// can be used to remove imports from the original file, like if mutations are made
// to the file's contents such that not all imports are needed anymore. When FromFile
// is used, any custom options present in the original descriptor will be correctly
// retained. If the file is mutated such that new custom options are added to the file,
// they may be missing unless AddImportedDependency is called after pruning OR
// BuilderOptions are used that include an ExtensionRegistry with knowledge of the
// new custom options.
func (fb *FileBuilder) PruneUnusedDependencies() *FileBuilder {
	fb.explicitImports = nil
	fb.explicitDeps = nil
	return fb
}

// SetOptions sets the file options for this file and returns the file, for
// method chaining.
func (fb *FileBuilder) SetOptions(options *descriptorpb.FileOptions) *FileBuilder {
	fb.Options = options
	return fb
}

// SetPackageName sets the name of the package for this file and returns the
// file, for method chaining.
func (fb *FileBuilder) SetPackageName(pkg string) *FileBuilder {
	fb.Package = pkg
	return fb
}

// SetProto3 sets whether this file is declared to use "proto3" syntax or not
// and returns the file, for method chaining.
func (fb *FileBuilder) SetProto3(isProto3 bool) *FileBuilder {
	fb.IsProto3 = isProto3
	return fb
}

func (fb *FileBuilder) buildProto(deps []*desc.FileDescriptor) (*descriptorpb.FileDescriptorProto, error) {
	name := fb.name
	if name == "" {
		name = uniqueFileName()
	}
	var syntax *string
	if fb.IsProto3 {
		syntax = proto.String("proto3")
	}
	var pkg *string
	if fb.Package != "" {
		pkg = proto.String(fb.Package)
	}

	path := make([]int32, 0, 10)
	sourceInfo := descriptorpb.SourceCodeInfo{}
	addCommentsTo(&sourceInfo, path, &fb.comments)
	addCommentsTo(&sourceInfo, append(path, internal.File_syntaxTag), &fb.SyntaxComments)
	addCommentsTo(&sourceInfo, append(path, internal.File_packageTag), &fb.PackageComments)

	imports := make([]string, 0, len(deps))
	for _, dep := range deps {
		imports = append(imports, dep.GetName())
	}
	sort.Strings(imports)

	messages := make([]*descriptorpb.DescriptorProto, 0, len(fb.messages))
	for _, mb := range fb.messages {
		path := append(path, internal.File_messagesTag, int32(len(messages)))
		if md, err := mb.buildProto(path, &sourceInfo); err != nil {
			return nil, err
		} else {
			messages = append(messages, md)
		}
	}

	enums := make([]*descriptorpb.EnumDescriptorProto, 0, len(fb.enums))
	for _, eb := range fb.enums {
		path := append(path, internal.File_enumsTag, int32(len(enums)))
		if ed, err := eb.buildProto(path, &sourceInfo); err != nil {
			return nil, err
		} else {
			enums = append(enums, ed)
		}
	}

	extensions := make([]*descriptorpb.FieldDescriptorProto, 0, len(fb.extensions))
	for _, exb := range fb.extensions {
		path := append(path, internal.File_extensionsTag, int32(len(extensions)))
		if exd, err := exb.buildProto(path, &sourceInfo, isExtendeeMessageSet(exb)); err != nil {
			return nil, err
		} else {
			extensions = append(extensions, exd)
		}
	}

	services := make([]*descriptorpb.ServiceDescriptorProto, 0, len(fb.services))
	for _, sb := range fb.services {
		path := append(path, internal.File_servicesTag, int32(len(services)))
		if sd, err := sb.buildProto(path, &sourceInfo); err != nil {
			return nil, err
		} else {
			services = append(services, sd)
		}
	}

	return &descriptorpb.FileDescriptorProto{
		Name:           proto.String(name),
		Package:        pkg,
		Dependency:     imports,
		Options:        fb.Options,
		Syntax:         syntax,
		MessageType:    messages,
		EnumType:       enums,
		Extension:      extensions,
		Service:        services,
		SourceCodeInfo: &sourceInfo,
	}, nil
}

func isExtendeeMessageSet(flb *FieldBuilder) bool {
	if flb.localExtendee != nil {
		return flb.localExtendee.Options.GetMessageSetWireFormat()
	}
	return flb.foreignExtendee.GetMessageOptions().GetMessageSetWireFormat()
}

// Build constructs a file descriptor based on the contents of this file
// builder. If there are any problems constructing the descriptor, including
// resolving symbols referenced by the builder or failing to meet certain
// validation rules, an error is returned.
func (fb *FileBuilder) Build() (*desc.FileDescriptor, error) {
	fd, err := fb.BuildDescriptor()
	if err != nil {
		return nil, err
	}
	return fd.(*desc.FileDescriptor), nil
}

// BuildDescriptor constructs a file descriptor based on the contents of this
// file builder. Most usages will prefer Build() instead, whose return type is a
// concrete descriptor type. This method is present to satisfy the Builder
// interface.
func (fb *FileBuilder) BuildDescriptor() (desc.Descriptor, error) {
	return doBuild(fb, BuilderOptions{})
}
