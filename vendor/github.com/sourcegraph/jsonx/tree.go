// This file was ported from https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts,
// which is licensed as follows:
//
// Copyright (c) Microsoft Corporation. All rights reserved. Licensed under the MIT License.

package jsonx

import (
	"encoding/json"
	"fmt"
)

// NodeType is the type of a JSON node.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L605
type NodeType int

// JSON node types
const (
	Object NodeType = iota
	Array
	Property
	String
	Number
	Boolean
	Null
)

// Node represents a node in a JSON document's parse tree.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L616
type Node struct {
	Type         NodeType    // the node's type
	Value        interface{} // the node's value
	Offset       int         // character offset of the node's starting position in the document
	Length       int         // the length (in characters) of the node
	ColumnOffset int         // character offset of the property's separator
	Parent       *Node       // the node's parent or nil if this is the root node
	Children     []*Node     // the node's children
}

// A Segment is a component of a JSON key path. It is either an object
// property name or an array index.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L626
type Segment struct {
	// IsProperty indicates the type of segment (true for property, false
	// for index). Because the zero values "" and 0 are valid values for the
	// Property and Index fields, this field is necessary to distinguish.
	IsProperty bool

	Property string // an object property name
	Index    int    // an array index
}

// MarshalJSON implements json.Marshaler.
func (s Segment) MarshalJSON() ([]byte, error) {
	if s.IsProperty {
		return json.Marshal(s.Property)
	}
	return json.Marshal(s.Index)
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Segment) UnmarshalJSON(data []byte) error {
	var t Segment
	var target interface{}
	if len(data) > 0 && data[0] == '"' {
		t.IsProperty = true
		target = &t.Property
	} else {
		target = &t.Index
	}
	if err := json.Unmarshal(data, target); err != nil {
		return err
	}
	*s = t
	return nil
}

// A Path is a JSON key path, which describes a path from an ancestor
// node in a JSON document's parse tree to one of its descendents.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L627
type Path []Segment

// PropertyPath returns a Path consisting of the specified property names.
func PropertyPath(properties ...string) Path {
	segments := make([]Segment, len(properties))
	for i, p := range properties {
		segments[i].IsProperty = true
		segments[i].Property = p
	}
	return Path(segments)
}

// MakePath returns a Path consisting of the specified components,
// each of which may be either a string (which is treated as a property segment)
// or an int (which is treated as an array index). Any other type causes it
// to panic.
func MakePath(components ...interface{}) Path {
	segments := make([]Segment, len(components))
	for i, c := range components {
		switch v := c.(type) {
		case string:
			segments[i].IsProperty = true
			segments[i].Property = v
		case int:
			segments[i].Index = v
		default:
			panic(fmt.Sprintf("unexpected path component type: %T", c))
		}
	}
	return Path(segments)
}

// ParseTree parses the given text and returns a tree representation the JSON content. On
// invalid input, the parser tries to be as fault tolerant as possible, but still return a result.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L688
func ParseTree(text string, options ParseOptions) (root *Node, errors []ParseErrorCode) {
	currentParent := &Node{Type: Array, Offset: -1, Length: -1} // artificial root

	ensurePropertyComplete := func(endOffset int) {
		if currentParent.Type == Property {
			currentParent.Length = endOffset - currentParent.Offset
			currentParent = currentParent.Parent
		}
	}

	onValue := func(valueNode *Node) *Node {
		currentParent.Children = append(currentParent.Children, valueNode)
		return valueNode
	}

	visitor := Visitor{
		OnObjectBegin: func(offset, length int) {
			currentParent = onValue(&Node{Type: Object, Offset: offset, Length: -1, Parent: currentParent})
		},
		OnObjectProperty: func(name string, offset, length int) {
			currentParent = onValue(&Node{Type: Property, Offset: offset, Length: -1, Parent: currentParent})
			currentParent.Children = append(currentParent.Children, &Node{Type: String, Value: name, Offset: offset, Length: length, Parent: currentParent})
		},
		OnObjectEnd: func(offset, length int) {
			currentParent.Length = offset + length - currentParent.Offset
			currentParent = currentParent.Parent
			ensurePropertyComplete(offset + length)
		},
		OnArrayBegin: func(offset, length int) {
			currentParent = onValue(&Node{Type: Array, Offset: offset, Length: -1, Parent: currentParent})
		},
		OnArrayEnd: func(offset, length int) {
			currentParent.Length = offset + length - currentParent.Offset
			currentParent = currentParent.Parent
			ensurePropertyComplete(offset + length)
		},
		OnLiteralValue: func(value interface{}, offset, length int) {
			onValue(&Node{Type: literalNodeType(value), Offset: offset, Length: length, Parent: currentParent, Value: value})
			ensurePropertyComplete(offset + length)
		},
		OnSeparator: func(sep rune, offset, length int) {
			if currentParent.Type == Property {
				if sep == ':' {
					currentParent.ColumnOffset = offset
				} else if sep == ',' {
					ensurePropertyComplete(offset)
				}
			}
		},
		OnError: func(errorCode ParseErrorCode, offset, length int) {
			errors = append(errors, errorCode)
		},
	}
	Walk(text, options, visitor)

	if len(currentParent.Children) > 0 {
		root = currentParent.Children[0]
		root.Parent = nil
	}
	return root, errors
}

func literalNodeType(value interface{}) NodeType {
	switch value.(type) {
	case bool:
		return Boolean

	case json.Number:
		return Number
	case int:
		return Number
	case int8:
		return Number
	case int16:
		return Number
	case int32:
		return Number
	case int64:
		return Number
	case uint:
		return Number
	case uint8:
		return Number
	case uint16:
		return Number
	case uint32:
		return Number
	case uint64:
		return Number

	case string:
		return String
	default:
		return Null
	}
}

// FindNodeAtLocation returns the node with the given key path under the JSON document parse
// tree root. If no such node exists, it returns nil.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L750
func FindNodeAtLocation(root *Node, path Path) *Node {
	if root == nil {
		return nil
	}
	node := root
	for _, segment := range path {
		if segment.IsProperty {
			if node.Type != Object {
				return nil
			}
			found := false
			for _, propertyNode := range node.Children {
				if propertyNode.Children[0].Value.(string) == segment.Property {
					node = propertyNode.Children[1]
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		} else {
			index := segment.Index
			if node.Type != Array || index < 0 || index >= len(node.Children) {
				return nil
			}
			node = node.Children[index]
		}
	}
	return node
}

// NodeValue returns the JSON parse tree node's value.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/json.ts#L782
func NodeValue(node Node) interface{} {
	switch node.Type {
	case Array:
		array := make([]interface{}, len(node.Children))
		for i, child := range node.Children {
			array[i] = NodeValue(*child)
		}
		return array

	case Object:
		object := make(map[string]interface{}, len(node.Children))
		for _, prop := range node.Children {
			object[prop.Children[0].Value.(string)] = NodeValue(*prop.Children[1])
		}
		return object

	default:
		return node.Value
	}
}

// ObjectPropertyNames returns property names of the JSON object represented
// by the specified JSON parse tree node.
func ObjectPropertyNames(node Node) []string {
	if node.Type != Object {
		panic("node.Type != Object")
	}
	props := make([]string, len(node.Children))
	for i, prop := range node.Children {
		props[i] = prop.Children[0].Value.(string)
	}
	return props
}
