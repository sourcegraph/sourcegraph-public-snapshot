package util

import "regexp"

// RegexpFlag is a wrapper around *regexp.Regexp that implements go-flags.Unmarshaler and go-flags.Marshaler,
// allowing it to be used as a flag.
//
// Empty string flag value maps to a nil *regexp.Regexp.
type RegexpFlag struct {
	*regexp.Regexp // This need to be unexported, otherwise go-flags will set it to a non-nil value.
}

// UnmarshalFlag unmarshals a string value representation to the flag value.
func (rf *RegexpFlag) UnmarshalFlag(value string) error {
	if value == "" {
		rf.Regexp = nil
		return nil
	}
	r, err := regexp.Compile(value)
	if err != nil {
		return err
	}
	rf.Regexp = r
	return nil
}

// MarshalFlag marshals a flag value to its string representation.
func (rf RegexpFlag) MarshalFlag() (string, error) {
	if rf.Regexp == nil {
		return "", nil
	}
	return rf.Regexp.String(), nil
}
