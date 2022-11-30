package honey

import (
	"bytes"
	"encoding/json"
)

// wrapper type for interface{} slice that marshals to a plain string
// in which values are comma separated and strings are unquoted aka
// []string{"asdf", "fdsa"} would render as the JSON string "asdf, fdsa".
type sliceWrapper []any

func (s sliceWrapper) MarshalJSON() ([]byte, error) {
	if len(s) == 0 {
		return nil, nil
	}

	var b bytes.Buffer

	for _, val := range (s)[:len(s)-1] {
		out, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		if out[0] == '"' {
			out = out[1 : len(out)-1]
		}
		b.Write(out)
		b.Write([]byte(", "))
	}

	out, err := json.Marshal(s[len(s)-1])
	if err != nil {
		return nil, err
	}
	if out[0] == '"' {
		out = out[1 : len(out)-1]
	}
	b.Write(out)

	return json.Marshal(b.String())
}
