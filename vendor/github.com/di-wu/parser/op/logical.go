package op

// Not (!) represents a negation of the Value. This should not consume data.
// e.g. Not{'a'} should check if the first rune is not an 'a'.
type Not struct {
	Value interface{}
}

// Ensure (&) represents a positive lookup of the Value. This should not consume
// data. e.g. Ensure{"abc"} should check if the strings is present.
type Ensure struct {
	Value interface{}
}

// And (&&) represents a sequence of values.
type And []interface{}

// Or (||) represents a sequence of alternative values. This is an ordered list,
// if a valid match is found it wil not try the remaining values.
type Or []interface{}

// XOr represents a sequence of exclusive alternative values. Only one of the
// values van be valid. It can contain only one valid match.
type XOr []interface{}
