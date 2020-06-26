package lsif

import (
	"fmt"
	"strings"

	simdjson "github.com/minio/simdjson-go"
)

// unmarshalElement populates the target pointer by reading the keys of the object value under
// the given iterator. This is done by first reading the common fields, then switching on those
// values to parse the remaining payload.
func unmarshalElement(iter *simdjson.Iter) (element Element, _ error) {
	if err := unmarshalCommon(iter, &element); err != nil {
		return Element{}, err
	}

	if err := unmarshalElementPayload(iter, element.Type, element.Label, &element.Payload); err != nil {
		return Element{}, err
	}

	return element, nil
}

// unmarshalCommon populates the target pointer by reading the keys of the object value under
// the given iterator with the following shape.
//
//   {
//     "id": string | int,
//     "type": string,
//     "label": string,
//   }
func unmarshalCommon(iter *simdjson.Iter, element *Element) error {
	var temp simdjson.Object
	obj, err := iter.Object(&temp)
	if err != nil {
		return err
	}

	for {
		var iter simdjson.Iter
		name, t, err := obj.NextElement(&iter)
		if err != nil {
			return err
		}
		if t == simdjson.TypeNone {
			break
		}

		switch name {
		case "id":
			err = unmarshalStringOrInt(&iter, &element.ID)
		case "type":
			err = unmarshalString(&iter, &element.Type)
		case "label":
			err = unmarshalString(&iter, &element.Label)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// unmarshalElementPayload populates the target pointer by reading the keys of the object value
// under the given iterator interpreted as an edge or vertex (indicated by the given type).
func unmarshalElementPayload(iter *simdjson.Iter, t, label string, target *interface{}) error {
	var temp simdjson.Object
	obj, err := iter.Object(&temp)
	if err != nil {
		return err
	}

	switch t {
	case "edge":
		var edge Edge
		err = unmarshalEdge(obj, &edge)
		*target = edge

	case "vertex":
		err = unmarshalVertexPayload(obj, label, target)
	}

	return err
}

// unmarshalVertexPayload populates the target pointer by reading the keys of the given object
// value with the shape determined by the given label.
func unmarshalVertexPayload(obj *simdjson.Object, label string, target *interface{}) (err error) {
	switch label {
	case "metaData":
		var metaData MetaData
		err = unmarshalMetaData(obj, &metaData)
		*target = metaData

	case "document":
		document := newDocument()
		err = unmarshalDocument(obj, &document)
		*target = document

	case "range":
		r := newRange()
		err = unmarshalRange(obj, &r)
		*target = r

	case "hoverResult":
		var hover string
		err = unmarshalHover(obj, &hover)
		*target = hover

	case "moniker":
		var moniker Moniker
		err = unmarshalMoniker(obj, &moniker)
		*target = moniker

	case "packageInformation":
		var packageInformation PackageInformation
		err = unmarshalPackageInformation(obj, &packageInformation)
		*target = packageInformation

	case "diagnosticResult":
		var diagnosticResult DiagnosticResult
		err = unmarshalDiagnosticResult(obj, &diagnosticResult)
		*target = diagnosticResult
	}

	return err
}

// unmarshalMetaData populates the target pointer by reading the keys of the given object with the
// following shape.
//
//   {
//     "version": string,
//     "projectRoot": string
//   }
func unmarshalMetaData(obj *simdjson.Object, metaData *MetaData) error {
	for {
		var iter simdjson.Iter
		name, t, err := obj.NextElement(&iter)
		if err != nil {
			return err
		}
		if t == simdjson.TypeNone {
			break
		}

		switch name {
		case "version":
			err = unmarshalString(&iter, &metaData.Version)
		case "projectRoot":
			err = unmarshalString(&iter, &metaData.ProjectRoot)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// unmarshalDocument populates the target pointer by reading the keys of the given object with the
// following shape.
//
//   {
//     "uri": string
//   }
func unmarshalDocument(obj *simdjson.Object, document *Document) error {
	for {
		var iter simdjson.Iter
		name, t, err := obj.NextElement(&iter)
		if err != nil {
			return err
		}
		if t == simdjson.TypeNone {
			break
		}

		if name == "uri" {
			if err := unmarshalString(&iter, &document.URI); err != nil {
				return err
			}
		}
	}

	return nil
}

// unmarshalRange populates the target pointer by reading the keys of the given object with the
// following shape.
//
//   {
//     "start": {
//       "line": int,
//       "character": int
//     },
//     "end": {
//       "line": int,
//       "character": int
//     }
//   }
func unmarshalRange(obj *simdjson.Object, r *Range) error {
	return unmarshalStartAndEndFromObject(obj, &r.StartLine, &r.StartCharacter, &r.EndLine, &r.EndCharacter)
}

// unmarshalHover populates the target pointer by reading the keys of the given object with the
// following shape. See unmarshalHoverParts for the shape of the hover parts values.
//
//   {
//     "result": {
//       "contents": {
//         hover parts
//       }
//     }
//   }

func unmarshalHover(obj *simdjson.Object, target *string) error {
	var tempElem simdjson.Element
	resultElem := obj.FindKey("result", &tempElem)
	if resultElem == nil {
		return fmt.Errorf("missing key 'result'")
	}
	resultIter := &resultElem.Iter

	var tempObj simdjson.Object
	resultObj, err := resultIter.Object(&tempObj)
	if err != nil {
		return err
	}

	contentsElem := resultObj.FindKey("contents", resultElem)
	if contentsElem == nil {
		return fmt.Errorf("missing key 'contents'")
	}
	iter := &contentsElem.Iter

	hover, err := unmarshalHoverParts(iter)
	if err != nil {
		return err
	}
	*target = hover
	return nil
}

// unmarshalHoverParts returns the a slice of hover text values by interpreting the value under the given
// iterator. This value is expected to be either a single object or an array of objects parseable by the
// function unmarshalHoverPart. If there are multiple objects, they will be parsed by unmarshalHoverPart
// and concatenated.
func unmarshalHoverParts(iter *simdjson.Iter) (string, error) {
	if iter.Type() != simdjson.TypeArray {
		return unmarshalHoverPart(iter)
	}

	var temp simdjson.Array
	arr, err := iter.Array(&temp)
	if err != nil {
		return "", err
	}
	arrIter := arr.Iter()

	var parts []string
	for arrIter.Advance() != simdjson.TypeNone {
		part, err := unmarshalHoverPart(&arrIter)
		if err != nil {
			return "", err
		}

		parts = append(parts, part)
	}

	return strings.Join(parts, "\n\n---\n\n"), nil
}

// unmarshalHoverPart returns the hover text as a string by interpreting the value under the given iterator.
// This value is expected to be either a bare string, or an object with the following shape.
//
//   {
//     "language": string,
//     "value": string
//   }
//
// This object will be returned as a single string (the raw value) or a code fence (if language is supplied).
func unmarshalHoverPart(iter *simdjson.Iter) (string, error) {
	if iter.Type() == simdjson.TypeString {
		return iter.String()
	}

	var temp simdjson.Object
	obj, err := iter.Object(&temp)
	if err != nil {
		return "", err
	}

	var language, value string

	for {
		var iter simdjson.Iter
		name, t, err := obj.NextElement(&iter)
		if err != nil {
			return "", err
		}
		if t == simdjson.TypeNone {
			break
		}

		switch name {
		case "language":
			err = unmarshalString(&iter, &language)
		case "value":
			err = unmarshalString(&iter, &value)
		}
		if err != nil {
			return "", err
		}
	}

	if language != "" {
		value = fmt.Sprintf("```%s\n%s\n```", language, value)
	}
	return strings.TrimSpace(value), nil
}

// unmarshalMoniker populates the target pointer by reading the keys of the given object with the
// following shape.
//
//   {
//     "kind": string,
//     "scheme": string,
//     "identifier": string
//   }
func unmarshalMoniker(obj *simdjson.Object, moniker *Moniker) error {
	for {
		var iter simdjson.Iter
		name, t, err := obj.NextElement(&iter)
		if err != nil {
			return err
		}
		if t == simdjson.TypeNone {
			break
		}

		switch name {
		case "kind":
			err = unmarshalString(&iter, &moniker.Kind)
		case "scheme":
			err = unmarshalString(&iter, &moniker.Scheme)
		case "identifier":
			err = unmarshalString(&iter, &moniker.Identifier)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// unmarshalPackageInformation populates the target pointer by reading the keys of the given object with
// the following shape.
//
//   {
//     "name": string,
//     "version": string
//   }
func unmarshalPackageInformation(obj *simdjson.Object, packageInformation *PackageInformation) error {
	for {
		var iter simdjson.Iter
		name, t, err := obj.NextElement(&iter)
		if err != nil {
			return err
		}
		if t == simdjson.TypeNone {
			break
		}

		switch name {
		case "name":
			err = unmarshalString(&iter, &packageInformation.Name)
		case "version":
			err = unmarshalString(&iter, &packageInformation.Version)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// unmarshalEdge populates the target pointer by reading the keys of the given object with the following
// shape.
//
//   {
//     "outV": string | int,
//     "inV": string | int,
//     "inVs": [string | int, ...],
//     "document": string | int
//   }
func unmarshalEdge(obj *simdjson.Object, edge *Edge) error {
	for {
		var iter simdjson.Iter
		name, t, err := obj.NextElement(&iter)
		if err != nil {
			return err
		}
		if t == simdjson.TypeNone {
			break
		}

		switch name {
		case "outV":
			err = unmarshalStringOrInt(&iter, &edge.OutV)
		case "inV":
			err = unmarshalStringOrInt(&iter, &edge.InV)
		case "inVs":
			err = unmarshalStringOrIntArray(&iter, &edge.InVs)
		case "document":
			err = unmarshalStringOrInt(&iter, &edge.Document)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// unmarshalDiagnosticResult populates the target pointer by reading the keys of the given object with
// the following shape. See unmarshalDiagnostic for the shape of diagnostic values.
//
//   {
//     "result": [diagnostic, ...]
//   }
func unmarshalDiagnosticResult(obj *simdjson.Object, diagnosticResult *DiagnosticResult) error {
	var tempElem simdjson.Element
	elem := obj.FindKey("result", &tempElem)
	if elem == nil {
		return fmt.Errorf("missing key 'result'")
	}

	iter := elem.Iter

	var tempArray simdjson.Array
	arr, err := iter.Array(&tempArray)
	if err != nil {
		return err
	}
	arrIter := arr.Iter()

	for arrIter.Advance() != simdjson.TypeNone {
		var diagnostic Diagnostic
		if err := unmarshalDiagnostic(&arrIter, &diagnostic); err != nil {
			return err
		}

		diagnosticResult.Result = append(diagnosticResult.Result, diagnostic)
	}

	return nil
}

// unmarshalDiagnostic populates the target pointer by reading the keys of the given object under the
// given iterator with the following shape.
//
//   {
//     "severity": int,
//     "code": string | int,
//     "message": string,
//     "source": string,
//     "range": {
//       "start": {
//         "line": int,
//         "character": int
//       },
//       "end": {
//         "line": int,
//         "character": int
//       }
//     }
//   }
func unmarshalDiagnostic(iter *simdjson.Iter, diagnostic *Diagnostic) error {
	var temp simdjson.Object
	obj, err := iter.Object(&temp)
	if err != nil {
		return err
	}

	for {
		var iter simdjson.Iter
		name, t, err := obj.NextElement(&iter)
		if err != nil {
			return err
		}
		if t == simdjson.TypeNone {
			break
		}

		switch name {
		case "severity":
			err = unmarshalInt(&iter, &diagnostic.Severity)
		case "code":
			err = unmarshalStringOrInt(&iter, &diagnostic.Code)
		case "message":
			err = unmarshalString(&iter, &diagnostic.Message)
		case "source":
			err = unmarshalString(&iter, &diagnostic.Source)
		case "range":
			var temp simdjson.Object
			obj, err := iter.Object(&temp)
			if err != nil {
				return err
			}

			return unmarshalStartAndEndFromObject(
				obj,
				&diagnostic.StartLine,
				&diagnostic.StartCharacter,
				&diagnostic.EndLine,
				&diagnostic.EndCharacter,
			)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// unmarshalStartAndEndFromObject populates the target pointers by reading the keys of the given object
// with the following shape.
//
//   {
//     "start": { "line": int, "character": int },
//     "end": { "line": int, "character": int }
//   }
func unmarshalStartAndEndFromObject(obj *simdjson.Object, startLine, startCharacter, endLine, endCharacter *int) error {
	for {
		var iter simdjson.Iter
		name, t, err := obj.NextElement(&iter)
		if err != nil {
			return err
		}
		if t == simdjson.TypeNone {
			break
		}

		switch name {
		case "start":
			err = unmarshalPosition(&iter, startLine, startCharacter)
		case "end":
			err = unmarshalPosition(&iter, endLine, endCharacter)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// unmarshalPosition populates the target pointers by reading the keys of the given object with the
// following shape.
//
//   {
//     "line": int,
//     "character": int
//   }
func unmarshalPosition(iter *simdjson.Iter, line, character *int) error {
	var temp simdjson.Object
	obj, err := iter.Object(&temp)
	if err != nil {
		return err
	}

	for {
		var iter simdjson.Iter
		name, t, err := obj.NextElement(&iter)
		if err != nil {
			return err
		}
		if t == simdjson.TypeNone {
			break
		}

		switch name {
		case "line":
			err = unmarshalInt(&iter, line)
		case "character":
			err = unmarshalInt(&iter, character)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// unmarshalString populates the target pointer with the string value under the given iterator.
func unmarshalString(iter *simdjson.Iter, target *string) error {
	v, err := iter.String()
	if err != nil {
		return err
	}

	*target = v
	return nil
}

// unmarshalInt populates the target pointer with the integer value under the given iterator.
func unmarshalInt(iter *simdjson.Iter, target *int) error {
	v, err := iter.Int()
	if err != nil {
		return err
	}

	*target = int(v)
	return nil
}

// unmarshalStringOrInt populates the target pointer with the string or integer value under the given
// iterator.
func unmarshalStringOrInt(iter *simdjson.Iter, target *string) error {
	if iter.Type() == simdjson.TypeInt {
		var v int
		if err := unmarshalInt(iter, &v); err != nil {
			return err
		}

		*target = fmt.Sprintf("%d", v)
		return nil
	}

	return unmarshalString(iter, target)
}

// unmarshalStringOrIntArray reads the array value under the given iterator and appends each string or
// int element to the target slice.
func unmarshalStringOrIntArray(iter *simdjson.Iter, target *[]string) error {
	var temp simdjson.Array
	arr, err := iter.Array(&temp)
	if err != nil {
		return err
	}
	arrIter := arr.Iter()

	for arrIter.Advance() != simdjson.TypeNone {
		var v string
		if err := unmarshalStringOrInt(&arrIter, &v); err != nil {
			return err
		}

		*target = append(*target, v)
	}

	return nil
}
