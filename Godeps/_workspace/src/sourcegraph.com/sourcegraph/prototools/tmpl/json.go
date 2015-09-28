package tmpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"

	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// jsonFieldTypeZero converts a simple (string, number, bool) field type to it's
// equivilent zero-value Go type. If the field type is not simple ok == false
/// is returned.
func jsonFieldTypeZero(f descriptor.FieldDescriptorProto_Type) (v interface{}, ok bool) {
	switch f {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return 0.0, true
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return 0.0, true
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		return 0.0, true
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		return 0.0, true
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		return 0.0, true
	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		return 0.0, true
	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		return 0.0, true
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		return 0.0, true
	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		return 0.0, true
	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return 0.0, true
	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		return 0.0, true
	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		return 0.0, true
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return false, true
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return "", true
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return []byte{}, true
	default:
		return nil, false
	}
}

// jsonMessage converts a protobuf message into it's JSON representation with
// links inside it (it's actually HTML).
func (f *tmplFuncs) jsonMessage(m *descriptor.DescriptorProto) (template.HTML, error) {
	var (
		// swap is literally strings to swap out in the marshaled JSON text.
		swap = make(map[string]string)

		// items is a map of field name to equivilent Go zero value.
		items = make(map[string]interface{})
	)
	for _, field := range m.Field {
		repeated := field.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
		val, ok := jsonFieldTypeZero(field.GetType())
		if ok {
			// It's a simple data type.
			if repeated {
				k := fmt.Sprintf("{{%s}}", field.GetName())
				items[field.GetName()] = k

				// Marshal the Go zero-value to get a JSON equivilent.
				newText, err := json.Marshal(val)
				if err != nil {
					return "", err
				}

				// Later on swap out the JSON string with the array style text.
				k = fmt.Sprintf(`"%s"`, k)
				swap[k] = fmt.Sprintf("[%s, ...]", newText)
				continue
			}
			items[field.GetName()] = val
			continue
		}

		// It's a more complex type:
		switch field.GetType() {
		// TODO(slimsag): determine their proper JSON representation
		//case descriptor.FieldDescriptorProto_TYPE_ENUM:
		//case descriptor.FieldDescriptorProto_TYPE_GROUP:

		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			sym := field.GetTypeName()
			k := fmt.Sprintf("{{%s}}", sym)
			items[field.GetName()] = k

			// Later on swap out the JSON string with the array or object style
			// text plus a link to the message type.
			k = fmt.Sprintf(`"%s"`, k)
			if repeated {
				swap[k] = fmt.Sprintf(`[<a href="%s">%s</a>, ...]`, f.urlToType(sym), f.cleanType(sym))
			} else {
				swap[k] = fmt.Sprintf(`{<a href="%s">%s</a>}`, f.urlToType(sym), f.cleanType(sym))
			}
		}
	}

	// Encode the JSON data.
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return "", err
	}

	// Perform string swapping.
	for old, new := range swap {
		data = bytes.Replace(data, []byte(old), []byte(new), -1)
	}
	return template.HTML(data), err
}
