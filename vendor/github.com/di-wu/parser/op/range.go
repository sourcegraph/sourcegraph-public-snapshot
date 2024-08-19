package op

// Range ([]) represents a range of repeated values.
// e.g. Range{1, -1, 'a'} expects at least one rune and consumes all following
// matching runes until it encounters another one.
type Range struct {
	// Min indicates the lower bound of the range. Values less than 0 will get
	// interpreted as 0.
	Min int
	// Max indicates the lower bound of the range. Values less than Min will be
	// set equal to Min. On exception: -1 indicates that there is no upper bound.
	Max int
	// Value to check.
	Value interface{}
}

// Min returns the range '[min:['.
func Min(min int, i interface{}) Range {
	return MinMax(min, -1, i)
}

// MinZero returns the range '[0:['.
func MinZero(i interface{}) Range {
	return Min(0, i)
}

// MinOne returns the range '[1:['.
func MinOne(i interface{}) Range {
	return Min(1, i)
}

// MinMax returns the range '[min:max]'.
func MinMax(min, max int, i interface{}) Range {
	if min < 0 {
		min = 0
	}
	if max != -1 && max < min {
		max = min
	}
	return Range{
		Min:   min,
		Max:   max,
		Value: i,
	}
}

// Optional returns the range '[0:1]'. It represents an optional value.
func Optional(i interface{}) Range {
	return MinMax(0, 1, i)
}

// Repeat returns the range '[c:c]'. It checks for a specific number of values.
// The count needs to be greater or equal to 1.
func Repeat(c int, i interface{}) Range {
	if c < 1 {
		c = 1
	}
	return Range{
		Min:   c,
		Max:   c,
		Value: i,
	}
}
