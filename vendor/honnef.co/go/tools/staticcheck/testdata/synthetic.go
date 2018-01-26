package pkg

import "regexp"

// must be a basic type to trigger SA4017 (in case of a test failure)
type T string

func (T) Fn() {}

// Don't get confused by methods named init
func (T) init() {}

// this will become a synthetic init function, that we don't want to
// ignore
var _ = regexp.MustCompile("(") // MATCH /error parsing regexp/
