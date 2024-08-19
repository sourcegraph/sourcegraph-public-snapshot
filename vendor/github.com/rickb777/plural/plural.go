package plural

import (
	"fmt"
	"strings"
)

// Case is the inner element of this API and describes one case. When the number to be described
// matches the number here, the corresponding format string will be used. If the format string
// includes '%', then fmt.Sprintf will be used. Otherwise the format string will be returned verbatim.
type Case struct {
	Number int
	Format string
}

// Plurals provides a list of plural cases in the order they will be searched.
// For plurals of continuous ranges (e.g. weight), the cases must be in ascending number order.
// For plurals of discrete ranges (i.e. integers), the cases can be in any order you require,
// but will conventionally be in ascending number order.
// If no match is found, the last case will be used.
type Plurals []Case

// Format searches through the plural cases for the first match. If none is found, the last
// case is used. The value passed in can be any number type, or pointer to a number type, except
// complex numbers are not supported. The value will be converted to an int in order to
// find the first case that matches.
// The only possible error arises if value has a type that is not numeric.
// It panics if 'plurals' is empty.
func (plurals Plurals) Format(value interface{}) (string, error) {
	switch x := value.(type) {
	case int:
		return plurals.FormatInt(x), nil
	case int8:
		return plurals.FormatInt(int(x)), nil
	case int16:
		return plurals.FormatInt(int(x)), nil
	case int32:
		return plurals.FormatInt(int(x)), nil
	case int64:
		return plurals.FormatInt(int(x)), nil
	case uint8:
		return plurals.FormatInt(int(x)), nil
	case uint16:
		return plurals.FormatInt(int(x)), nil
	case uint32:
		return plurals.FormatInt(int(x)), nil
	case uint64:
		return plurals.FormatInt(int(x)), nil
	case float32:
		return plurals.FormatFloat(x), nil
	case float64:
		return plurals.FormatFloat(float32(x)), nil

	case *int:
		return plurals.FormatInt(*x), nil
	case *int8:
		return plurals.FormatInt(int(*x)), nil
	case *int16:
		return plurals.FormatInt(int(*x)), nil
	case *int32:
		return plurals.FormatInt(int(*x)), nil
	case *int64:
		return plurals.FormatInt(int(*x)), nil
	case *uint:
		return plurals.FormatInt(int(*x)), nil
	case *uint8:
		return plurals.FormatInt(int(*x)), nil
	case *uint16:
		return plurals.FormatInt(int(*x)), nil
	case *uint32:
		return plurals.FormatInt(int(*x)), nil
	case *uint64:
		return plurals.FormatInt(int(*x)), nil
	case *float32:
		return plurals.FormatFloat(*x), nil
	case *float64:
		return plurals.FormatFloat(float32(*x)), nil

	case nil:
		return "", fmt.Errorf("Unexpected nil value for %s", plurals)
	default:
		return "", fmt.Errorf("Unexpected type %T for %v", x, value)
	}
}

// FormatInt expresses an int in plural form. It panics if 'plurals' is empty.
func (plurals Plurals) FormatInt(value int) string {
	for _, c := range plurals {
		if value == c.Number {
			return c.FormatInt(value)
		}
	}
	c := plurals[len(plurals)-1]
	return c.FormatInt(value)
}

// FormatFloat expresses a float32 in plural form. It panics if 'plurals' is empty.
func (plurals Plurals) FormatFloat(value float32) string {
	for _, c := range plurals {
		if value <= float32(c.Number) {
			return c.FormatFloat(value)
		}
	}
	c := plurals[len(plurals)-1]
	return c.FormatFloat(value)
}

// FormatInt renders a specific case with a given value.
func (c Case) FormatInt(value int) string {
	if strings.IndexByte(c.Format, '%') < 0 {
		return c.Format
	}
	return fmt.Sprintf(c.Format, value)
}

// FormatFloat renders a specific case with a given value.
func (c Case) FormatFloat(value float32) string {
	if strings.IndexByte(c.Format, '%') < 0 {
		return c.Format
	}
	return fmt.Sprintf(c.Format, value)
}

//-------------------------------------------------------------------------------------------------

// String implements io.Stringer.
func (plurals Plurals) String() string {
	ss := make([]string, 0, len(plurals))
	for _, c := range plurals {
		ss = append(ss, c.String())
	}
	return fmt.Sprintf("Plurals(%s)", strings.Join(ss, ", "))
}

// String implements io.Stringer.
func (c Case) String() string {
	return fmt.Sprintf("{%v -> %q}", c.Number, c.Format)
}

//-------------------------------------------------------------------------------------------------

// ByOrdinal constructs a simple set of cases using small ordinals (0, 1, 2, 3 etc), which is a
// common requirement. It is an alias for FromZero.
func ByOrdinal(zeroth string, rest ...string) Plurals {
	return FromZero(zeroth, rest...)
}

// FromZero constructs a simple set of cases using small ordinals (0, 1, 2, 3 etc), which is a
// common requirement. It prevents creation of a Plurals list that is empty, which would be invalid.
//
// The 'zeroth' string becomes Case{0, first}. The rest are appended similarly. Notice that the
// counting starts from zero.
//
// So
//
//   FromZero("nothing", "%v thing", "%v things")
//
// is simply a shorthand for
//
//   Plurals{Case{0, "nothing"}, Case{1, "%v thing"}, Case{2, "%v things"}}
//
// which would also be valid but a little more verbose.
//
// This helper function is less flexible than constructing Plurals directly, but covers many common
// situations.
func FromZero(zeroth string, rest ...string) Plurals {
	p := make(Plurals, 0, len(rest)+1)
	p = append(p, Case{0, zeroth})
	for i, c := range rest {
		p = append(p, Case{i+1, c})
	}
	return p
}

// FromOne constructs a simple set of cases using small positive numbers (1, 2, 3 etc), which is a
// common requirement. It prevents creation of a Plurals list that is empty, which would be invalid.
//
// The 'first' string becomes Case{1, first}. The rest are appended similarly. Notice that the
// counting starts from one.
//
// So
//
//   FromOne("%v thing", "%v things")
//
// is simply a shorthand for
//
//   Plurals{Case{1, "%v thing"}, Case{2, "%v things"}}
//
// which would also be valid but a little more verbose.
//
// Note the behaviour of formatting when the count is zero. As a consequence of Format evaluating
// the cases in order, FromOne(...).FormatInt(0) will pick the last case you provide, not the first.
//
// This helper function is less flexible than constructing Plurals directly, but covers many common
// situations.
func FromOne(first string, rest ...string) Plurals {
	p := make(Plurals, 0, len(rest)+1)
	p = append(p, Case{1, first})
	for i, c := range rest {
		p = append(p, Case{i+2, c})
	}
	return p
}
