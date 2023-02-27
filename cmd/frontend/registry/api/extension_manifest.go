package api

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
)

// extensionManifest implements the GraphQL type ExtensionManifest.
type extensionManifest struct {
	raw string
}

// NewExtensionManifest creates a new resolver for the GraphQL type ExtensionManifest with the given
// raw contents of an extension manifest.
func NewExtensionManifest(raw *string) graphqlbackend.ExtensionManifest {
	if raw == nil {
		return nil
	}
	return &extensionManifest{raw: *raw}
}

func (r *extensionManifest) Raw() string { return r.raw }

func (r *extensionManifest) JSONFields(args *struct{ Fields []string }) graphqlbackend.JSONValue {
	return jsonValueWithFields(r.raw, args.Fields)
}

func jsonValueWithFields(jsoncStr string, fields []string) graphqlbackend.JSONValue {
	o := map[string]json.RawMessage{}
	jsonc.Unmarshal(jsoncStr, &o) // ignore error and treat as empty object

	keepField := func(field string) bool {
		for _, f := range fields {
			if f == field {
				return true
			}
		}
		return false
	}
	for field := range o {
		if !keepField(field) {
			delete(o, field)
		}
	}
	return graphqlbackend.JSONValue{Value: o}
}
