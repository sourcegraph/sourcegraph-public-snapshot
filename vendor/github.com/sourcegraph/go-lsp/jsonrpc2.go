package lsp

import (
	"encoding/json"
	"strconv"
)

// ID represents a JSON-RPC 2.0 request ID, which may be either a
// string or number (or null, which is unsupported).
type ID struct {
	// At most one of Num or Str may be nonzero. If both are zero
	// valued, then IsNum specifies which field's value is to be used
	// as the ID.
	Num uint64
	Str string

	// IsString controls whether the Num or Str field's value should be
	// used as the ID, when both are zero valued. It must always be
	// set to true if the request ID is a string.
	IsString bool
}

func (id ID) String() string {
	if id.IsString {
		return strconv.Quote(id.Str)
	}
	return strconv.FormatUint(id.Num, 10)
}

// MarshalJSON implements json.Marshaler.
func (id ID) MarshalJSON() ([]byte, error) {
	if id.IsString {
		return json.Marshal(id.Str)
	}
	return json.Marshal(id.Num)
}

// UnmarshalJSON implements json.Unmarshaler.
func (id *ID) UnmarshalJSON(data []byte) error {
	// Support both uint64 and string IDs.
	var v uint64
	if err := json.Unmarshal(data, &v); err == nil {
		*id = ID{Num: v}
		return nil
	}
	var v2 string
	if err := json.Unmarshal(data, &v2); err != nil {
		return err
	}
	*id = ID{Str: v2, IsString: true}
	return nil
}
