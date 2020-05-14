package codeintel

import (
	"net/url"
	"strconv"
)

// queryValues is a convenience wrapper around url.Values that adds
// behaviors to set values of non-string types and optionally set
// values that may be a zero-value.
type queryValues struct {
	values url.Values
}

// newQueryValues creates a new queryValues.
func newQueryValues() queryValues {
	return queryValues{values: url.Values{}}
}

// Set adds the given name/string-value pairing to the underlying values.
func (qv queryValues) Set(name string, value string) {
	qv.values[name] = []string{value}
}

// SetInt adds the given name/int-value pairing to the underlying values.
func (qv queryValues) SetInt(name string, value int) {
	qv.Set(name, strconv.FormatInt(int64(value), 10))
}

// SetOptionalString adds the given name/string-value pairing to the underlying values.
// If the value is empty, the underlying values are not modified.
func (qv queryValues) SetOptionalString(name string, value string) {
	if value != "" {
		qv.Set(name, value)
	}
}

// SetOptionalInt adds the given name/int-value pairing to the underlying values.
// If the value is zero, the underlying values are not modified.
func (qv queryValues) SetOptionalInt(name string, value int) {
	if value != 0 {
		qv.SetInt(name, value)
	}
}

// SetOptionalBool adds the given name/bool-value pairing to the underlying values.
// If the value is false, the underlying values are not modified.
func (qv queryValues) SetOptionalBool(name string, value bool) {
	if value {
		qv.Set(name, "true")
	}
}

// Encode encodes the underlying values.
func (qv queryValues) Encode() string {
	return qv.values.Encode()
}
