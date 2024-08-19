package common

import (
	"encoding/json"
	"reflect"
)

// MultilineString stores multi-line text as strings or string arrays in the raw JSON.
// TODO(optimization): check if we can scan into [][]byte directly
type MultilineString []string

func (s *MultilineString) UnmarshalJSON(data []byte) error {
	var nums interface{}
	err := json.Unmarshal(data, &nums)
	if err != nil {
		return err
	}

	items := reflect.ValueOf(nums)
	switch items.Kind() {
	case reflect.String:
		*s = append(*s, items.String())

	case reflect.Slice:
		*s = make(MultilineString, 0, items.Len())
		for i := 0; i < items.Len(); i++ {
			item := items.Index(i)
			switch item.Kind() {
			case reflect.String:
				*s = append(*s, item.String())
			case reflect.Interface:
				*s = append(*s, item.Interface().(string))
			}
		}
	}
	return nil
}

// Text concatenates all lines in a multiline string into a single byte slice.
func (s MultilineString) Text() (txt []byte) {
	for _, line := range s {
		txt = append(txt, line...)
	}
	return
}
