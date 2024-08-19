// This file was ported from https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/jsonEdit.ts,
// which is licensed as follows:
//
// Copyright (c) Microsoft Corporation. All rights reserved. Licensed under the MIT License.

package jsonx

import (
	"encoding/json"
	"errors"
	"fmt"
)

// An Edit represents an edit to a JSON document.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/jsonFormatter.ts#L24
type Edit struct {
	Offset  int    // the character offset where the edit begins
	Length  int    // the character length of the region to replace with the content
	Content string // the content to insert into the document
}

// ComputePropertyEdit returns the edits necessary to set the value at the specified
// key path to the value. If value is nil, the property's value is set to JSON null;
// use ComputePropertyRemoval to obtain the edits necessary to remove a property. If
// value is a json.RawMessage, it is treated as an opaque value to insert (which
// means it can contain comments, trailing commas, etc.).
//
// If the insertionIndex is non-nil, it is called to determine the index at which to
// insert the value (given the existing properties, in order).
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/jsonEdit.ts#L14
func ComputePropertyEdit(text string, path Path, value interface{}, insertionIndex func(properties []string) int, options FormatOptions) ([]Edit, []ParseErrorCode, error) {
	if value == nil {
		value = json.RawMessage("null") // otherwise would remove property
	}
	return computePropertyEdit(text, path, value, insertionIndex, options)
}

// ComputePropertyRemoval returns the edits necessary to remove the property at the
// specified key path.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/jsonEdit.ts#L10
func ComputePropertyRemoval(text string, path Path, options FormatOptions) ([]Edit, []ParseErrorCode, error) {
	return computePropertyEdit(text, path, nil, nil, options)
}

func computePropertyEdit(text string, path Path, valueObj interface{}, insertionIndex func(properties []string) int, options FormatOptions) ([]Edit, []ParseErrorCode, error) {
	// Tolerate errors in value if it's json.RawMessage.
	var value string
	if v, ok := valueObj.(json.RawMessage); ok {
		value = string(v)
	} else {
		data, err := json.Marshal(valueObj)
		if err != nil {
			return nil, nil, err
		}
		value = string(data)
	}

	root, parseErrorCodes := ParseTree(text, ParseOptions{Comments: true, TrailingCommas: true})

	var parent *Node

	var lastSegment Segment
	for len(path) > 0 {
		lastSegment = path[len(path)-1]
		path = path[:len(path)-1]

		parent = FindNodeAtLocation(root, path)
		if parent == nil && valueObj != nil {
			if lastSegment.IsProperty {
				key, err := json.Marshal(lastSegment.Property)
				if err != nil {
					return nil, nil, err
				}
				value = "{" + string(key) + ":" + value + "}"
			} else {
				value = "[" + value + "]"
			}
		} else {
			break
		}
	}

	if parent == nil {
		// empty document
		if valueObj == nil { // delete
			return nil, nil, errors.New("can't delete in empty document")
		}
		edit := Edit{Content: value}
		if root != nil {
			edit.Offset = root.Offset
			edit.Length = root.Length
		}
		edits, err := FormatEdit(text, edit, options)
		return edits, parseErrorCodes, err
	} else if parent.Type == Object && lastSegment.IsProperty {
		indexOf := func(slice []*Node, candidateElement *Node) int {
			for i, e := range slice {
				if e == candidateElement {
					return i
				}
			}
			return -1
		}

		existing := FindNodeAtLocation(parent, Path{lastSegment})
		if existing != nil {
			if valueObj == nil { // delete
				propertyIndex := indexOf(parent.Children, existing.Parent)
				var removeBegin int
				removeEnd := existing.Parent.Offset + existing.Parent.Length
				if propertyIndex > 0 {
					// remove the comma of the previous node
					previous := parent.Children[propertyIndex-1]
					removeBegin = previous.Offset + previous.Length
				} else {
					removeBegin = parent.Offset + 1
					if len(parent.Children) > 1 {
						// remove the comma of the next node
						next := parent.Children[1]
						removeEnd = next.Offset
					} else {
						// remove trailing comma after this node, if any
						removeEnd = parent.Offset + parent.Length - 1
					}
				}
				edits, err := FormatEdit(text, Edit{Offset: removeBegin, Length: removeEnd - removeBegin, Content: ""}, options)
				return edits, parseErrorCodes, err
			}

			// set value of existing property
			edits, err := FormatEdit(text, Edit{Offset: existing.Offset, Length: existing.Length, Content: value}, options)
			return edits, parseErrorCodes, err
		}

		if valueObj == nil { // delete
			return nil, parseErrorCodes, nil // property does not exist, nothing to do
		}

		propNameData, err := json.Marshal(lastSegment.Property)
		if err != nil {
			return nil, nil, err
		}
		newProperty := string(propNameData) + ": " + value

		var index int
		if insertionIndex != nil {
			index = insertionIndex(ObjectPropertyNames(*parent))
		} else {
			index = len(parent.Children)
		}

		var edit Edit
		if index > 0 {
			previous := parent.Children[index-1]
			edit = Edit{Offset: previous.Offset + previous.Length, Length: 0, Content: "," + newProperty}
		} else if len(parent.Children) == 0 {
			edit = Edit{Offset: parent.Offset + 1, Length: 0, Content: newProperty}
		} else {
			edit = Edit{Offset: parent.Offset + 1, Length: 0, Content: newProperty + ","}
		}

		edits, err := FormatEdit(text, edit, options)
		return edits, parseErrorCodes, err
	} else if parent.Type == Array && !lastSegment.IsProperty {
		insertIndex := lastSegment
		if insertIndex.Index == -1 {
			// Insert
			var edit Edit
			if len(parent.Children) == 0 {
				edit = Edit{Offset: parent.Offset + 1, Length: 0, Content: value}
			} else {
				previous := parent.Children[len(parent.Children)-1]
				edit = Edit{Offset: previous.Offset + previous.Length, Length: 0, Content: "," + value}
			}
			edits, err := FormatEdit(text, edit, options)
			return edits, parseErrorCodes, err
		}

		if valueObj == nil && len(parent.Children) >= 0 {
			// Removal
			removalIndex := lastSegment.Index
			toRemove := parent.Children[removalIndex]
			var edit Edit
			if len(parent.Children) == 1 {
				// only item
				edit = Edit{Offset: parent.Offset + 1, Length: parent.Length - 2, Content: ""}
			} else if len(parent.Children)-1 == removalIndex {
				// last item
				previous := parent.Children[removalIndex-1]
				offset := previous.Offset + previous.Length
				parentEndOffset := parent.Offset + parent.Length
				edit = Edit{Offset: offset, Length: parentEndOffset - 2 - offset, Content: ""}
			} else {
				edit = Edit{Offset: toRemove.Offset, Length: parent.Children[removalIndex+1].Offset - toRemove.Offset, Content: ""}
			}
			edits, err := FormatEdit(text, edit, options)
			return edits, parseErrorCodes, err
		}

		// Modify
		editIndex := lastSegment.Index
		toEdit := parent.Children[editIndex]
		edit := Edit{Offset: toEdit.Offset, Length: toEdit.Length, Content: value}
		edits, err := FormatEdit(text, edit, options)
		return edits, parseErrorCodes, err
	}

	var noun string
	if lastSegment.IsProperty {
		noun = "property"
	} else {
		noun = "index"
	}
	return nil, nil, fmt.Errorf("can't add %s to parent of type %s", noun, parent.Type)
}

// ApplyEdits applies the edits to the JSON document and returns the edited
// document. The edits must be ordered and within the bounds of the document.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/jsonFormatter.ts#L34
func ApplyEdits(text string, edits ...Edit) (string, error) {
	chars := []rune(text)
	lastEditOffset := len(chars)
	for i := len(edits) - 1; i >= 0; i-- {
		edit := edits[i]
		if edit.Offset < 0 || edit.Length < 0 && edit.Offset+edit.Length > len(chars) {
			return "", fmt.Errorf("edit out of bounds: offset %d, length %d, doc length %d", edit.Offset, edit.Length, len(chars))
		}
		if lastEditOffset < edit.Offset+edit.Length {
			return "", fmt.Errorf("edit out of order: edit end offset %d exceeds next edit offset %d", edit.Offset+edit.Length, lastEditOffset)
		}
		lastEditOffset = edit.Offset
		chars = []rune(string(chars[:edit.Offset]) + edit.Content + string(chars[edit.Offset+edit.Length:]))
	}
	return string(chars), nil
}

// FormatEdit returns the edits necessary to perform the original edit for maintaining the
// formatting of the JSON document.
//
// Source: https://github.com/Microsoft/vscode/blob/c0bc1ace7ca3ce2d6b1aeb2bde9d1bb0f4b4bae6/src/vs/base/common/jsonEdit.ts#L122
func FormatEdit(text string, edit Edit, options FormatOptions) ([]Edit, error) {
	// apply the edit
	newText, err := ApplyEdits(text, edit)
	if err != nil {
		return nil, err
	}

	// format the new text
	begin := edit.Offset
	end := edit.Offset + len([]rune(edit.Content))
	edits := FormatRange(newText, begin, end-begin, options)

	// apply the formatting edits and track the begin and end offsets of the changes
	for i := len(edits) - 1; i >= 0; i-- {
		edit := edits[i]
		newText, err = ApplyEdits(newText, edit)
		if err != nil {
			return nil, err
		}
		if edit.Offset < begin {
			begin = edit.Offset
		}
		if edit.Offset+edit.Length > end {
			end = edit.Offset + edit.Length
		}
		end += len([]rune(edit.Content)) - edit.Length
	}

	// create a single edit with all changes
	editLength := len([]rune(text)) - (len([]rune(newText)) - end) - begin
	return []Edit{{Offset: begin, Length: editLength, Content: string(([]rune(newText))[begin:end])}}, nil
}
