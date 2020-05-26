package jsonlines

import (
	"encoding/json"
	"strconv"
)

type ID string

func (id *ID) UnmarshalJSON(raw []byte) error {
	if raw[0] == '"' {
		var v string
		if err := json.Unmarshal(raw, &v); err != nil {
			return err
		}

		*id = ID(v)
		return nil
	}

	var v int64
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}

	*id = ID(strconv.FormatInt(v, 10))
	return nil
}
