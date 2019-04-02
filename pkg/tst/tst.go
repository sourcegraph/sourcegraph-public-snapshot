// Package tst defines convenience functions for tests
package tst

import (
	"reflect"
	"testing"
)

func CheckStrEqual(t *testing.T, exp, actual, errMsg string) {
	if exp != actual {
		t.Errorf("%s: (exp != actual) %q != %q", errMsg, exp, actual)
	}
}

func CheckDeepEqual(t *testing.T, exp, actual interface{}, errMsg string) {
	if !reflect.DeepEqual(exp, actual) {
		t.Errorf("%s: (exp != actual) %+v != %+v", errMsg, exp, actual)
	}
}
