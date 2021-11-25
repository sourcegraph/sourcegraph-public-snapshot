package batches

import "encoding/json"

type branches []string

func (b *branches) UnmarshalJSON(data []byte) error {
	// This matches the behaviour of the YAML unmarshaller, which we have less
	// ability to control in the null case.
	if string(data) == "null" {
		*b = nil
		return nil
	}

	// Branches may be either a string or an array of strings, so we'll try it
	// both ways.
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*b = []string{s}
		return nil
	}

	var ss []string
	if err := json.Unmarshal(data, &ss); err != nil {
		return err
	}

	*b = ss
	return nil
}

func (b *branches) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Branches may be either a string or an array of strings, so we'll try it
	// both ways.
	var s string
	if err := unmarshal(&s); err == nil {
		*b = []string{s}
		return nil
	}

	var ss []string
	if err := unmarshal(&ss); err != nil {
		return err
	}

	*b = ss
	return nil
}
