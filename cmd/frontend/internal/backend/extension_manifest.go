package backend

import "encoding/json"

// ValidateExtensionManifest validates a JSON extension manifest for syntax.
//
// TODO(sqs): Also validate it against the JSON Schema.
func ValidateExtensionManifest(text string) error {
	var o interface{}
	return json.Unmarshal([]byte(text), &o)
}
