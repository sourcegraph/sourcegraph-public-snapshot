package util

import (
	"fmt"

	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// nameMessage names the given message descriptor, returning a copy with the
// Name pointer replaced with &newName.
func nameMessage(old *descriptor.DescriptorProto, newName string) *descriptor.DescriptorProto {
	cpy := *old
	cpy.Name = &newName
	return &cpy
}

// nameEnum names the given enum descriptor, returning a copy with the Name0
// pointer replaced with &newName.
func nameEnum(old *descriptor.EnumDescriptorProto, newName string) *descriptor.EnumDescriptorProto {
	cpy := *old
	cpy.Name = &newName
	return &cpy
}

// appendEnum is a helper function for appending two symbol paths together. It's
// not generic enough to be public (i.e. it can only work for the below use
// cases).
func appendElem(symbolPath, childName string) string {
	if symbolPath == "" {
		return childName + "."
	}
	return fmt.Sprintf("%s.%s", symbolPath, childName)
}

// AllMessages returnes a list of all the message type nodes in f, including
// nested ones.
func AllMessages(f *descriptor.FileDescriptorProto, swapNames bool) []*descriptor.DescriptorProto {
	var (
		all        []*descriptor.DescriptorProto
		walk       func(n *descriptor.DescriptorProto)
		symbolPath string
	)

	// Define the function that will perform the recursive walk of the AST nodes.
	walk = func(n *descriptor.DescriptorProto) {
		// The name of the type should only ever be a single element.
		if CountElem(n.GetName()) != 1 {
			panic("unexpected name elements")
		}

		for _, child := range n.NestedType {
			// Accumulate the node and swap the names of it, if desired.
			if swapNames {
				all = append(all, nameMessage(child, symbolPath+child.GetName()))
			} else {
				all = append(all, child)
			}

			symbolPath = appendElem(symbolPath, child.GetName()) // push parent type name
			walk(child)                                          // walk nested types, recursively
			symbolPath = TrimElem(symbolPath, 1)                 // pop parent type name
		}
	}

	for _, m := range f.MessageType {
		// Accumulate each root-level message type.
		all = append(all, m)

		symbolPath = appendElem(symbolPath, m.GetName()) // push parent type name
		walk(m)                                          // walk nested types
		symbolPath = TrimElem(symbolPath, 1)             // pop parent type name
	}
	return all
}

// AllEnums returnes a list of all the enum type nodes in f, including nested
// ones.
func AllEnums(f *descriptor.FileDescriptorProto, swapNames bool) []*descriptor.EnumDescriptorProto {
	var (
		all        []*descriptor.EnumDescriptorProto
		walk       func(n *descriptor.DescriptorProto)
		symbolPath string
	)

	// Define the function that will perform the recursive walk of the AST nodes.
	walk = func(n *descriptor.DescriptorProto) {
		// The name of the type should only ever be a single element.
		if CountElem(n.GetName()) != 1 {
			panic("unexpected name elements")
		}

		for _, child := range n.EnumType {
			// Accumulate the node, swapping the names of it if desired.
			if swapNames {
				child = nameEnum(child, symbolPath+child.GetName())
			}
			all = append(all, child)
		}

		// Walk the nested types for this message node, in case there are more child
		// enum types.
		for _, child := range n.NestedType {
			symbolPath = appendElem(symbolPath, child.GetName()) // push parent type name
			walk(child)                                          // walk nested types, recursively
			symbolPath = TrimElem(symbolPath, 1)                 // pop parent type name
		}
	}

	// Accumulate each root-level enum type.
	for _, e := range f.EnumType {
		all = append(all, e)
	}

	// Walk each root-level message type for nested enums.
	for _, m := range f.MessageType {
		symbolPath = appendElem(symbolPath, m.GetName()) // push parent type name
		walk(m)                                          // walk nested types, recursively
		symbolPath = TrimElem(symbolPath, 1)             // pop parent type name
	}
	return all
}
