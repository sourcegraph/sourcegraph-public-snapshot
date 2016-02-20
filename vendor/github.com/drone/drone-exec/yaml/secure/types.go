package secure

import "strings"

type MapEqualSlice struct {
	parts map[string]string
}

func (s *MapEqualSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	s.parts = map[string]string{}
	err := unmarshal(&s.parts)
	if err == nil {
		return nil
	}

	var sliceType []string

	err = unmarshal(&sliceType)
	if err != nil {
		return err
	}

	for _, v := range sliceType {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			val := parts[1]
			s.parts[key] = val
		}
	}

	return nil
}

func (s *MapEqualSlice) Map() map[string]string {
	return s.parts
}

func (s MapEqualSlice) MarshalYAML() (interface{}, error) {
	return s.parts, nil
}
