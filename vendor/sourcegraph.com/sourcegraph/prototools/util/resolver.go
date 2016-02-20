package util

import (
	"reflect"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// ASTNode is a type from the descriptor package.
type ASTNode interface{}

// ASTNamedNode is a type from the descriptor package that can return its name
// string. These include (but are not limited to):
//
//  *descriptor.DescriptorProto
//  *descriptor.EnumDescriptorProto
//  *descriptor.ServiceDescriptorProto
//  *descriptor.FieldDescriptorProto
//
type ASTNamedNode interface {
	GetName() string
}

// search searches below the given AST node for an message, enum, service, or
// extension with the given symbol path. Valid input types are:
//
//  ASTNamedNode
//  []ASTNamedNode
//  *descriptor.FileDescriptorProto
//
func search(a ASTNode, symbolPath string) ASTNode {
	// Handle the slice types.
	rv := reflect.ValueOf(a)
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			if found := search(rv.Index(i).Interface(), symbolPath); found != nil {
				return found
			}
		}
		return nil
	}

	switch n := a.(type) {
	case *descriptor.FileDescriptorProto:
		// Note: It's important for this case to be above the ASTNamedNode case below
		// (which only intends to match messages, enums, services and extensions) as
		// FileDescriptorProto also implements the ASTNamedNode interface.

		// Search all the top-level messages, enums, services and extensions in the
		// file.
		if v := search(n.MessageType, symbolPath); v != nil {
			return v
		}
		if v := search(n.EnumType, symbolPath); v != nil {
			return v
		}
		if v := search(n.Service, symbolPath); v != nil {
			return v
		}
		if v := search(n.Extension, symbolPath); v != nil {
			return v
		}
		return nil

	case ASTNamedNode:
		// Check if it's this named node.
		if symbolPath == n.GetName() {
			return a
		}

		// If the symbol path contains only one element (e.g. "Option") then we
		// cannot search below this node any further, so this search failed.
		if CountElem(symbolPath) == 1 {
			return nil
		}

		// If it's a message type, it can have nested type declarations and
		// enumerations, so we need to search those.
		msg, ok := a.(*descriptor.DescriptorProto)
		if !ok {
			// It's an enum, service, or extension which don't have nested types.
			return nil
		}

		// Trim one element off the symbol path, e.g. "Type.Sub" -> "Sub".
		symbolPath = TrimElem(symbolPath, 1)

		// Search nested types.
		if v := search(msg.NestedType, symbolPath); v != nil {
			return v
		}
		// Search enum types.
		return search(msg.EnumType, symbolPath)

	default:
		panic("unexpected type")
	}
}

// Resolver handles the resolution of symbol names to their respective files (it
// answers the question "which file was this symbol defined in?").
type Resolver struct {
	f []*descriptor.FileDescriptorProto
}

// ResolveFile resolves the file that the given symbol is declared inside of, or
// nil if it is not. It is short-handed for:
//
//  _, file := r.Resolve(symbolPath)
//
func (r *Resolver) ResolveFile(symbolPath string, relative ASTNode) *descriptor.FileDescriptorProto {
	_, file := r.Resolve(symbolPath, relative)
	return file
}

// ResolveSymbol resolves the symbol with the given path, or nil of it cannot be
// resolved. It is short-handed for:
//
//  node, _ := r.Resolve(symbolPath)
//
func (r *Resolver) ResolveSymbol(symbolPath string, relative ASTNode) ASTNode {
	node, _ := r.Resolve(symbolPath, relative)
	return node
}

// Resolve resolves the named symbol into its actual AST node and the file that
// node is inside of. Example symbolPath strings are:
//
//  Sym
//  pkg.Sym
//  foo.bar.pkg.Sym
//  .foo.bar.pkg.Sym
//
// Relative symbol paths like:
//
//  Sym
//  pkg.Sym
//  foo.bar.Pkg.Sym
//
// Are resolved according to the protobuf language doc:
//
//  "Packages and Name Resolution" - https://developers.google.com/protocol-buffers/docs/proto#packages
//
// As all relative symbol paths in protobuf follow C++ style scoping rules, the
// path can only be resolved reliably whilst knowing the AST node that
// resolution is relative to. If the relative node is nil and the symbol path is
// not fully-qualified, a panic will occur.
//
// For example in the pseudo-code:
//
//  package pkg;
//
//  message Foo {
//      ...
//  }
//
//  message Bar {
//      message Foo {
//          ...
//      }
//
//      Foo this = 1;
//  }
//
// Resolution of the message field pkg.Bar.this must be done *relative* to the
// AST node for pkg.Bar, because pkg.Bar.thi sis of type pkg.Bar.Foo, not
// pkg.Foo.
//
// TODO(slimsag): relative symbol path resolution is not yet implemented and as
// such will always panic.
func (r *Resolver) Resolve(symbolPath string, relative ASTNode) (ASTNode, *descriptor.FileDescriptorProto) {
	if !IsFullyQualified(symbolPath) {
		panic("resolution of relative (non-fully-qualified) symbol paths is not implemented")
	}
	symbolPath = strings.TrimPrefix(symbolPath, ".")

	// Determine the package that symbolPath is part of, considering multiple
	// matches like these:
	//
	//  symbolPath="foo.bar.pkg.Sym" && GetPackage() == "foo"
	//  symbolPath="foo.bar.pkg.Sym" && GetPackage() == "foo.bar"
	//
	for _, f := range r.f {
		pkg := PackageName(f)
		if len(pkg) == 0 {
			panic("no package name")
		}
		if !strings.HasPrefix(symbolPath, pkg) {
			continue // not a match
		}

		// Trim the package prefix off the symbol, for example:
		//
		//  symbolPath = "world.buildings.Options"
		//  pkg = "world.buildings"
		//  CountElem(pkg) == 2
		//  TrimElem(symbolPath, CountElem(pkg)) == "Options"
		//
		// Use a temporary variable as we don't want it to affect the next iteration
		// of this loop.
		tmp := TrimElem(symbolPath, CountElem(pkg))

		// Search for the symbol inside this file.
		if n := search(ASTNode(f), tmp); n != nil {
			return n, f
		}
	}
	return nil, nil
}

// NewResolver returns a new symbol resolver for the given files.
func NewResolver(f []*descriptor.FileDescriptorProto) *Resolver {
	return &Resolver{f: f}
}
