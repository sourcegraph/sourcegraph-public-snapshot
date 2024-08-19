package builder

import (
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/internal"
	"github.com/jhump/protoreflect/dynamic"
)

type dependencies struct {
	descs map[*desc.FileDescriptor]struct{}
	res   protoregistry.Types
}

func newDependencies() *dependencies {
	return &dependencies{
		descs: map[*desc.FileDescriptor]struct{}{},
	}
}

func (d *dependencies) add(fd *desc.FileDescriptor) {
	if _, ok := d.descs[fd]; ok {
		// already added
		return
	}
	d.descs[fd] = struct{}{}
	internal.RegisterExtensionsForFile(&d.res, fd.UnwrapFile())
}

// dependencyResolver is the work-horse for converting a tree of builders into a
// tree of descriptors. It scans a root (usually a file builder) and recursively
// resolves all dependencies (references to builders in other trees as well as
// references to other already-built descriptors). The result of resolution is a
// file descriptor (or an error).
type dependencyResolver struct {
	resolvedRoots map[Builder]*desc.FileDescriptor
	seen          map[Builder]struct{}
	opts          BuilderOptions
}

func newResolver(opts BuilderOptions) *dependencyResolver {
	return &dependencyResolver{
		resolvedRoots: map[Builder]*desc.FileDescriptor{},
		seen:          map[Builder]struct{}{},
		opts:          opts,
	}
}

func (r *dependencyResolver) resolveElement(b Builder, seen []Builder) (*desc.FileDescriptor, error) {
	b = getRoot(b)

	if fd, ok := r.resolvedRoots[b]; ok {
		return fd, nil
	}

	for _, s := range seen {
		if s == b {
			names := make([]string, len(seen)+1)
			for i, s := range seen {
				names[i] = s.GetName()
			}
			names[len(seen)] = b.GetName()
			return nil, fmt.Errorf("descriptors have cyclic dependency: %s", strings.Join(names, " ->  "))
		}
	}
	seen = append(seen, b)

	var fd *desc.FileDescriptor
	var err error
	switch b := b.(type) {
	case *FileBuilder:
		fd, err = r.resolveFile(b, b, seen)
	default:
		fd, err = r.resolveSyntheticFile(b, seen)
	}
	if err != nil {
		return nil, err
	}
	r.resolvedRoots[b] = fd
	return fd, nil
}

func (r *dependencyResolver) resolveFile(fb *FileBuilder, root Builder, seen []Builder) (*desc.FileDescriptor, error) {
	deps := newDependencies()
	// add explicit imports first
	for fd := range fb.explicitImports {
		deps.add(fd)
	}
	for dep := range fb.explicitDeps {
		if dep == fb {
			// ignore erroneous self references
			continue
		}
		fd, err := r.resolveElement(dep, seen)
		if err != nil {
			return nil, err
		}
		deps.add(fd)
	}
	// now accumulate implicit dependencies based on other types referenced
	for _, mb := range fb.messages {
		if err := r.resolveTypesInMessage(root, seen, deps, mb); err != nil {
			return nil, err
		}
	}
	for _, exb := range fb.extensions {
		if err := r.resolveTypesInExtension(root, seen, deps, exb); err != nil {
			return nil, err
		}
	}
	for _, sb := range fb.services {
		if err := r.resolveTypesInService(root, seen, deps, sb); err != nil {
			return nil, err
		}
	}

	// finally, resolve custom options (which may refer to deps already
	// computed above)
	if err := r.resolveTypesInFileOptions(root, deps, fb); err != nil {
		return nil, err
	}

	depSlice := make([]*desc.FileDescriptor, 0, len(deps.descs))
	depMap := make(map[string]*desc.FileDescriptor, len(deps.descs))
	for dep := range deps.descs {
		isDuplicate, err := isDuplicateDependency(dep, depMap)
		if err != nil {
			return nil, err
		}
		if !isDuplicate {
			depSlice = append(depSlice, dep)
		}
	}

	fp, err := fb.buildProto(depSlice)
	if err != nil {
		return nil, err
	}

	// make sure this file name doesn't collide with any of its dependencies
	fileNames := map[string]struct{}{}
	for _, d := range depSlice {
		addFileNames(d, fileNames)
		fileNames[d.GetName()] = struct{}{}
	}
	unique := makeUnique(fp.GetName(), fileNames)
	if unique != fp.GetName() {
		fp.Name = proto.String(unique)
	}

	return desc.CreateFileDescriptor(fp, depSlice...)
}

// isDuplicateDependency checks for duplicate descriptors
func isDuplicateDependency(dep *desc.FileDescriptor, depMap map[string]*desc.FileDescriptor) (bool, error) {
	if _, exists := depMap[dep.GetName()]; !exists {
		depMap[dep.GetName()] = dep
		return false, nil
	}
	prevFDP := depMap[dep.GetName()].AsFileDescriptorProto()
	depFDP := dep.AsFileDescriptorProto()

	// temporarily reset Source Code Fields as builders does not have SourceCodeInfo
	defer setSourceCodeInfo(prevFDP, nil)()
	defer setSourceCodeInfo(depFDP, nil)()

	if !proto.Equal(prevFDP, depFDP) {
		return true, fmt.Errorf("multiple versions of descriptors found with same file name: %s", dep.GetName())
	}
	return true, nil
}

func setSourceCodeInfo(fdp *descriptorpb.FileDescriptorProto, info *descriptorpb.SourceCodeInfo) (reset func()) {
	prevSourceCodeInfo := fdp.SourceCodeInfo
	fdp.SourceCodeInfo = info
	return func() { fdp.SourceCodeInfo = prevSourceCodeInfo }
}

func addFileNames(fd *desc.FileDescriptor, files map[string]struct{}) {
	if _, ok := files[fd.GetName()]; ok {
		// already added
		return
	}
	files[fd.GetName()] = struct{}{}
	for _, d := range fd.GetDependencies() {
		addFileNames(d, files)
	}
}

func (r *dependencyResolver) resolveSyntheticFile(b Builder, seen []Builder) (*desc.FileDescriptor, error) {
	// find ancestor to temporarily attach to new file
	curr := b
	for curr.GetParent() != nil {
		curr = curr.GetParent()
	}
	f := NewFile("")
	switch curr := curr.(type) {
	case *MessageBuilder:
		f.messages = append(f.messages, curr)
	case *EnumBuilder:
		f.enums = append(f.enums, curr)
	case *ServiceBuilder:
		f.services = append(f.services, curr)
	case *FieldBuilder:
		if curr.IsExtension() {
			f.extensions = append(f.extensions, curr)
		} else {
			panic("field must be added to message before calling Build()")
		}
	case *OneOfBuilder:
		if _, ok := b.(*OneOfBuilder); ok {
			panic("one-of must be added to message before calling Build()")
		} else {
			// b was a child of one-of which means it must have been a field
			panic("field must be added to message before calling Build()")
		}
	case *MethodBuilder:
		panic("method must be added to service before calling Build()")
	case *EnumValueBuilder:
		panic("enum value must be added to enum before calling Build()")
	default:
		panic(fmt.Sprintf("Unrecognized kind of builder: %T", b))
	}
	curr.setParent(f)

	// don't forget to reset when done
	defer func() {
		curr.setParent(nil)
	}()

	return r.resolveFile(f, b, seen)
}

func (r *dependencyResolver) resolveTypesInMessage(root Builder, seen []Builder, deps *dependencies, mb *MessageBuilder) error {
	for _, b := range mb.fieldsAndOneOfs {
		if flb, ok := b.(*FieldBuilder); ok {
			if err := r.resolveTypesInField(root, seen, deps, flb); err != nil {
				return err
			}
		} else {
			oob := b.(*OneOfBuilder)
			for _, flb := range oob.choices {
				if err := r.resolveTypesInField(root, seen, deps, flb); err != nil {
					return err
				}
			}
		}
	}
	for _, nmb := range mb.nestedMessages {
		if err := r.resolveTypesInMessage(root, seen, deps, nmb); err != nil {
			return err
		}
	}
	for _, exb := range mb.nestedExtensions {
		if err := r.resolveTypesInExtension(root, seen, deps, exb); err != nil {
			return err
		}
	}
	return nil
}

func (r *dependencyResolver) resolveTypesInExtension(root Builder, seen []Builder, deps *dependencies, exb *FieldBuilder) error {
	if err := r.resolveTypesInField(root, seen, deps, exb); err != nil {
		return err
	}
	if exb.foreignExtendee != nil {
		deps.add(exb.foreignExtendee.GetFile())
	} else if err := r.resolveType(root, seen, exb.localExtendee, deps); err != nil {
		return err
	}
	return nil
}

func (r *dependencyResolver) resolveTypesInService(root Builder, seen []Builder, deps *dependencies, sb *ServiceBuilder) error {
	for _, mtb := range sb.methods {
		if err := r.resolveRpcType(root, seen, mtb.ReqType, deps); err != nil {
			return err
		}
		if err := r.resolveRpcType(root, seen, mtb.RespType, deps); err != nil {
			return err
		}
	}
	return nil
}

func (r *dependencyResolver) resolveRpcType(root Builder, seen []Builder, t *RpcType, deps *dependencies) error {
	if t.foreignType != nil {
		deps.add(t.foreignType.GetFile())
	} else {
		return r.resolveType(root, seen, t.localType, deps)
	}
	return nil
}

func (r *dependencyResolver) resolveTypesInField(root Builder, seen []Builder, deps *dependencies, flb *FieldBuilder) error {
	if flb.fieldType.foreignMsgType != nil {
		deps.add(flb.fieldType.foreignMsgType.GetFile())
	} else if flb.fieldType.foreignEnumType != nil {
		deps.add(flb.fieldType.foreignEnumType.GetFile())
	} else if flb.fieldType.localMsgType != nil {
		if flb.fieldType.localMsgType == flb.msgType {
			return r.resolveTypesInMessage(root, seen, deps, flb.msgType)
		} else {
			return r.resolveType(root, seen, flb.fieldType.localMsgType, deps)
		}
	} else if flb.fieldType.localEnumType != nil {
		return r.resolveType(root, seen, flb.fieldType.localEnumType, deps)
	}
	return nil
}

func (r *dependencyResolver) resolveType(root Builder, seen []Builder, typeBuilder Builder, deps *dependencies) error {
	otherRoot := getRoot(typeBuilder)
	if root == otherRoot {
		// local reference, so it will get resolved when we finish resolving this root
		return nil
	}
	fd, err := r.resolveElement(otherRoot, seen)
	if err != nil {
		return err
	}
	deps.add(fd)
	return nil
}

func (r *dependencyResolver) resolveTypesInFileOptions(root Builder, deps *dependencies, fb *FileBuilder) error {
	for _, mb := range fb.messages {
		if err := r.resolveTypesInMessageOptions(root, fb.origExts, deps, mb); err != nil {
			return err
		}
	}
	for _, eb := range fb.enums {
		if err := r.resolveTypesInEnumOptions(root, fb.origExts, deps, eb); err != nil {
			return err
		}
	}
	for _, exb := range fb.extensions {
		if err := r.resolveTypesInOptions(root, fb.origExts, deps, exb.Options); err != nil {
			return err
		}
	}
	for _, sb := range fb.services {
		for _, mtb := range sb.methods {
			if err := r.resolveTypesInOptions(root, fb.origExts, deps, mtb.Options); err != nil {
				return err
			}
		}
		if err := r.resolveTypesInOptions(root, fb.origExts, deps, sb.Options); err != nil {
			return err
		}
	}
	return r.resolveTypesInOptions(root, fb.origExts, deps, fb.Options)
}

func (r *dependencyResolver) resolveTypesInMessageOptions(root Builder, fileExts *dynamic.ExtensionRegistry, deps *dependencies, mb *MessageBuilder) error {
	for _, b := range mb.fieldsAndOneOfs {
		if flb, ok := b.(*FieldBuilder); ok {
			if err := r.resolveTypesInOptions(root, fileExts, deps, flb.Options); err != nil {
				return err
			}
		} else {
			oob := b.(*OneOfBuilder)
			for _, flb := range oob.choices {
				if err := r.resolveTypesInOptions(root, fileExts, deps, flb.Options); err != nil {
					return err
				}
			}
			if err := r.resolveTypesInOptions(root, fileExts, deps, oob.Options); err != nil {
				return err
			}
		}
	}
	for _, extr := range mb.ExtensionRanges {
		if err := r.resolveTypesInOptions(root, fileExts, deps, extr.Options); err != nil {
			return err
		}
	}
	for _, eb := range mb.nestedEnums {
		if err := r.resolveTypesInEnumOptions(root, fileExts, deps, eb); err != nil {
			return err
		}
	}
	for _, nmb := range mb.nestedMessages {
		if err := r.resolveTypesInMessageOptions(root, fileExts, deps, nmb); err != nil {
			return err
		}
	}
	for _, exb := range mb.nestedExtensions {
		if err := r.resolveTypesInOptions(root, fileExts, deps, exb.Options); err != nil {
			return err
		}
	}
	if err := r.resolveTypesInOptions(root, fileExts, deps, mb.Options); err != nil {
		return err
	}
	return nil
}

func (r *dependencyResolver) resolveTypesInEnumOptions(root Builder, fileExts *dynamic.ExtensionRegistry, deps *dependencies, eb *EnumBuilder) error {
	for _, evb := range eb.values {
		if err := r.resolveTypesInOptions(root, fileExts, deps, evb.Options); err != nil {
			return err
		}
	}
	if err := r.resolveTypesInOptions(root, fileExts, deps, eb.Options); err != nil {
		return err
	}
	return nil
}

func (r *dependencyResolver) resolveTypesInOptions(root Builder, fileExts *dynamic.ExtensionRegistry, deps *dependencies, opts proto.Message) error {
	// nothing to see if opts is nil
	if opts == nil {
		return nil
	}
	if rv := reflect.ValueOf(opts); rv.Kind() == reflect.Ptr && rv.IsNil() {
		return nil
	}

	ref := opts.ProtoReflect()
	tags := map[int32]protoreflect.ExtensionType{}
	proto.RangeExtensions(opts, func(xt protoreflect.ExtensionType, _ interface{}) bool {
		num := int32(xt.TypeDescriptor().Number())
		tags[num] = xt
		return true
	})

	unk := ref.GetUnknown()
	for len(unk) > 0 {
		v, n := protowire.ConsumeVarint(unk)
		if n < 0 {
			break
		}
		unk = unk[n:]

		num, t := protowire.DecodeTag(v)
		if _, ok := tags[int32(num)]; !ok {
			tags[int32(num)] = nil
		}

		switch t {
		case protowire.VarintType:
			_, n = protowire.ConsumeVarint(unk)
		case protowire.Fixed64Type:
			_, n = protowire.ConsumeFixed64(unk)
		case protowire.BytesType:
			_, n = protowire.ConsumeBytes(unk)
		case protowire.StartGroupType:
			_, n = protowire.ConsumeGroup(num, unk)
		case protowire.EndGroupType:
			// invalid encoding
			break
		case protowire.Fixed32Type:
			_, n = protowire.ConsumeFixed32(unk)
		}
		if n < 0 {
			break
		}
		unk = unk[n:]
	}

	msgName := string(proto.MessageName(opts))
	for tag, xt := range tags {
		// see if known dependencies have this option
		if _, err := deps.res.FindExtensionByNumber(protoreflect.FullName(msgName), protoreflect.FieldNumber(tag)); err == nil {
			// yep! nothing else to do
			continue
		}
		// see if this extension is defined in *this* builder
		if findExtension(root, msgName, tag) {
			// yep!
			continue
		}
		// see if configured extension registry knows about it
		if extd := r.opts.Extensions.FindExtension(string(msgName), int32(tag)); extd != nil {
			// extension registry recognized it!
			deps.add(extd.GetFile())
			continue
		}
		// see if given file extensions knows about it
		if fileExts != nil {
			extd := fileExts.FindExtension(msgName, tag)
			if extd != nil {
				// file extensions recognized it!
				deps.add(extd.GetFile())
				continue
			}
		}

		if xt != nil {
			// known extension? add its file to builder's deps
			fd, err := desc.WrapFile(xt.TypeDescriptor().ParentFile())
			if err != nil {
				return err
			}
			deps.add(fd)
			continue
		}

		if r.opts.RequireInterpretedOptions {
			// we require options to be interpreted but are not able to!
			return fmt.Errorf("could not interpret custom option for %s, tag %d", msgName, tag)
		}
	}
	return nil
}

func findExtension(b Builder, messageName string, extTag int32) bool {
	if fb, ok := b.(*FileBuilder); ok && findExtensionInFile(fb, messageName, extTag) {
		return true
	}
	if mb, ok := b.(*MessageBuilder); ok && findExtensionInMessage(mb, messageName, extTag) {
		return true
	}
	return false
}

func findExtensionInFile(fb *FileBuilder, messageName string, extTag int32) bool {
	for _, extb := range fb.extensions {
		if extb.GetExtendeeTypeName() == messageName && extb.number == extTag {
			return true
		}
	}
	for _, mb := range fb.messages {
		if findExtensionInMessage(mb, messageName, extTag) {
			return true
		}
	}
	return false
}

func findExtensionInMessage(mb *MessageBuilder, messageName string, extTag int32) bool {
	for _, extb := range mb.nestedExtensions {
		if extb.GetExtendeeTypeName() == messageName && extb.number == extTag {
			return true
		}
	}
	for _, mb := range mb.nestedMessages {
		if findExtensionInMessage(mb, messageName, extTag) {
			return true
		}
	}
	return false
}
