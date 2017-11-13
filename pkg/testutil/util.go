package testutil

import (
	"reflect"
	"testing"
)

// Check checks if condition is true and errors with errMsg if false.
func Check(t *testing.T, condition bool, errMsg string) {
	if !condition {
		t.Error(errMsg)
	}
}

// CheckEq checks for equality *if* the expected value is non-zero.
func CheckEq(t *testing.T, expected, actual interface{}, errMsg string) {
	if expected == nil {
		return
	}
	if exp, ok := expected.(int); ok && exp == 0 {
		return
	}
	if exp, ok := expected.(string); ok && exp == "" {
		return
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v, but got %v (%s)", expected, actual, errMsg)
	}
}
