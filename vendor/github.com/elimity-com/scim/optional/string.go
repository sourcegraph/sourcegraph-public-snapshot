package optional

// String represents an optional string value.
type String struct {
	value   string
	present bool
}

// NewString returns an optional string with given value.
func NewString(value string) String {
	return String{
		value:   value,
		present: true,
	}
}

// Present returns whether it contains a value or not.
func (s String) Present() bool {
	return s.present
}

// Value returns the value of the optional string.
func (s String) Value() string {
	return s.value
}
