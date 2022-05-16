package overridable

import "encoding/json"

// Bool represents a bool value that can be modified on a per-repo basis.
type Bool struct {
	rules rules
}

// FromBool creates a Bool representing a static, scalar value.
func FromBool(b bool) Bool {
	return Bool{
		rules: rules{simpleRule(b)},
	}
}

// Value returns the bool value for the given repository.
func (b *Bool) Value(name string) bool {
	v := b.rules.Match(name)
	if v == nil {
		return false
	}
	return v.(bool)
}

// MarshalJSON encodes the Bool overridable to a json representation.
func (b Bool) MarshalJSON() ([]byte, error) {
	if len(b.rules) == 0 {
		return []byte("false"), nil
	}
	return json.Marshal(b.rules)
}

// UnmarshalJSON unmarshalls a JSON value into a Bool.
func (b *Bool) UnmarshalJSON(data []byte) error {
	var all bool
	if err := json.Unmarshal(data, &all); err == nil {
		*b = Bool{rules: rules{simpleRule(all)}}
		return nil
	}

	var c complex
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}

	return b.rules.hydrateFromComplex(c)
}

// UnmarshalYAML unmarshalls a YAML value into a Bool.
func (b *Bool) UnmarshalYAML(unmarshal func(any) error) error {
	var all bool
	if err := unmarshal(&all); err == nil {
		*b = Bool{rules: rules{simpleRule(all)}}
		return nil
	}

	var c complex
	if err := unmarshal(&c); err != nil {
		return err
	}

	return b.rules.hydrateFromComplex(c)
}

// Equal tests two Bools for equality, used in cmp.
func (b Bool) Equal(other Bool) bool {
	return b.rules.Equal(other.rules)
}
