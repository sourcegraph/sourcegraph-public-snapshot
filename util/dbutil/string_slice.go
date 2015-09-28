package dbutil

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

type StringSlice struct {
	Slice []string
}

func (s *StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}

	inner := make([]string, len(s.Slice))
	for i, elem := range s.Slice {
		if strings.TrimSpace(elem) == "" || strings.Contains(elem, `"`) {
			inner[i] = strconv.Quote(elem)
		} else {
			inner[i] = elem
		}
	}
	return []byte("{" + strings.Join(inner, ",") + "}"), nil
}

func (s *StringSlice) Scan(v interface{}) error {
	if data, ok := v.([]byte); ok {
		interior := strings.Trim(string(data), "{}")
		if interior != "" {
			rawElems := strings.Split(interior, ",")
			s.Slice = make([]string, len(rawElems))
			for r, raw := range rawElems {
				if elem, err := strconv.Unquote(raw); err == nil {
					s.Slice[r] = elem
				} else {
					s.Slice[r] = raw
				}
			}
		} else {
			s.Slice = []string{}
		}
		return nil
	}
	return fmt.Errorf("%T.Scan failed: %v", s, v)
}

func NewSlice(goslice []string) *StringSlice {
	return &StringSlice{Slice: goslice}
}
