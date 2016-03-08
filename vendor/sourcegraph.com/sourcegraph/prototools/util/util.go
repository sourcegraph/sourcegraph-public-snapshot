// Package util implements utilities for building protoc compiler plugins.
package util // import "sourcegraph.com/sourcegraph/prototools/util"

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"strings"

	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

// ParseParams parses the comma-separated command-line parameters passed to the
// generator by protoc via r.GetParameters. Returned is a map of key=value
// parameters with whitespace preserved.
func ParseParams(r *plugin.CodeGeneratorRequest) map[string]string {
	// Split the parameter string and initialize the map.
	split := strings.Split(r.GetParameter(), ",")
	param := make(map[string]string, len(split))

	// Map the parameters.
	for _, p := range split {
		eq := strings.Split(p, "=")
		if len(eq) == 1 {
			param[strings.TrimSpace(eq[0])] = ""
			continue
		}
		val := strings.TrimSpace(eq[1])
		param[strings.TrimSpace(eq[0])] = val
	}
	return param
}

// FieldTypeName returns the protobuf-syntax name for the given field type. It
// panics on errors (e.g. zero value).
func FieldTypeName(f *descriptor.FieldDescriptorProto_Type) string {
	switch *f {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return "double"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return "float"
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		return "int64"
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		return "uint64"
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		return "int32"
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		return "fixed64"
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		return "fixed32"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		return "group"
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return "message"
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return "bytes"
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		return "uint32"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		return "enum"
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		return "sfixed32"
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return "sfixed64"
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		return "sint32"
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		return "sint64"
	default:
		panic("FieldTypeName: unknown field type")
	}
}

// IsFullyQualified tells if the given symbol path is fully-qualified or not (i.e.
// starts with a period).
func IsFullyQualified(symbolPath string) bool {
	return symbolPath[0] == '.'
}

// TrimElem returns the given symbol path with at max N elements trimmed off the
// left (outermost) side.
//
//  TrimElem("a.b.c", 1) == "b.c"
//  TrimElem(".a.b.c", 1) == "b.c"
//  TrimElem(".a.b.c", -1) == ".a.b"
//
// Extreme cases won't panic, either:
//
//  TrimElem("a.b.c", 1000) == ""
//  TrimElem(".a.b.c", 1000) == ""
//  TrimElem("a.b.c", -1000) == ""
//  TrimElem(".a.b.c", -1000) == ""
//
func TrimElem(symbolPath string, n int) string {
	if n == 0 {
		return symbolPath
	}
	if n > 0 {
		// Trimming from the left side. All paths here will become relative if n > 0
		// as we're trimming the root (left side) off.
		symbolPath = strings.TrimPrefix(symbolPath, ".")

		// Avoid indexing panic if we don't actually have N elements.
		split := strings.Split(symbolPath, ".")
		if len(split) < n {
			n = len(split)
		}
		return strings.Join(split[n:], ".")
	}

	// Trimming from the right side, then.
	split := strings.Split(symbolPath, ".")

	// Avoid indexing panic if we don't actually have N elements.
	if -n > len(split) {
		n = -len(split)
	}
	return strings.Join(split[:len(split)+n], ".")
}

// CountElem returns the number of elements that the symbol path contains.
//
//  CountElem("a.b.c") == 3
//  CountElem(".a.b.c") == 3
//  CountElem("a.b.c.d") == 4
//  CountElem("a") == 1
//  CountElem(".") == 0
//  CountElem("") == 0
//
func CountElem(symbolPath string) int {
	// Don't care about fully-qualified dot prefix.
	symbolPath = strings.TrimPrefix(symbolPath, ".")
	count := 0
	for _, s := range strings.Split(symbolPath, ".") {
		if len(s) > 0 {
			count++
		}
	}
	return count
}

// PackageName returns the package name of the given file, which is either the
// result of f.GetPackage (a package set explicitly by the user) or the name of
// the file.
func PackageName(f *descriptor.FileDescriptorProto) string {
	// Check for an explicit package name given by the user in a protobuf file as
	//
	//  package foo;
	//
	if pkg := f.GetPackage(); len(pkg) > 0 {
		return pkg
	}
	// Otherwise use the name of the file (note: not filepath.Base because
	// protobuf only speaks in unix path terms).
	pkg := path.Base(f.GetName())
	return strings.TrimSuffix(pkg, path.Ext(pkg))
}

// ReadJSONFile opens and unmarshals the JSON dump file from the protoc-gen-json
// plugin, returning any error that occurs.
func ReadJSONFile(path string) (*plugin.CodeGeneratorRequest, error) {
	// Read the file.
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal the request.
	r := &plugin.CodeGeneratorRequest{}
	err = json.Unmarshal(data, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
