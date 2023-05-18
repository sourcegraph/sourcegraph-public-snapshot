package v1

import (
	"strings"
	"unicode/utf8"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protopath"
	"google.golang.org/protobuf/reflect/protorange"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// findNonUTF8StringFields returns a list of field names that contain invalid UTF-8 strings
// in the given proto message.
//
// Example: ["author", "attachments[1].key_value_attachment.data["key2"]`]
func findNonUTF8StringFields(m proto.Message) ([]string, error) {
	var fields []string
	err := protorange.Range(m.ProtoReflect(), func(p protopath.Values) error {
		last := p.Index(-1)
		s, ok := last.Value.Interface().(string)
		if ok && !utf8.ValidString(s) {
			fieldName := p.Path[1:].String()
			fields = append(fields, strings.TrimPrefix(fieldName, "."))
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "iterating over proto message")
	}

	return fields, nil
}
