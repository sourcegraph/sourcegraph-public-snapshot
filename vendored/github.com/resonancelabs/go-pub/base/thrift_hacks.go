package base

import (
	"math"
)

// XXX(bhs): This is bad. https://issues.apache.org/jira/browse/THRIFT-2232
const ThriftMissingEnum int64 = math.MinInt32 - 1

// Return the dereferenced value of the float64 or return 0.0 if the pointer
// is nil.
func ThriftGetFloat64OrZero(v *float64) float64 {
	if v != nil {
		return *v
	} else {
		return 0.0
	}
}

// Return the dereferenced value of the string or return a empty string if
// the pointer is nil.
func ThriftGetStringOrEmpty(s *string) string {
	if s != nil {
		return *s
	} else {
		return ""
	}
}
