package graphqlbackend

import (
	"encoding/json"
	"sort"
)

func toKeyValueList(m map[string]interface{}) []keyValue {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	kv := make([]keyValue, len(keys))
	for i, k := range keys {
		kv[i] = keyValue{k, opaqueValue{m[k]}}
	}
	return kv
}

type keyValue struct {
	key   string
	value opaqueValue
}

func (kv keyValue) Key() string        { return kv.key }
func (kv keyValue) Value() opaqueValue { return kv.value }

// opaqueValue implements the OpaqueValue scalar type. In GraphQL queries, it is
// represented as the JSON value corresponding to the Go value.
type opaqueValue struct{ value interface{} }

func (opaqueValue) ImplementsGraphQLType(name string) bool {
	return name == "OpaqueValue"
}

func (v opaqueValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *opaqueValue) UnmarshalGraphQL(input interface{}) error {
	*v = opaqueValue{value: input}
	return nil
}
