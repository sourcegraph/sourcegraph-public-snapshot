package base

import (
	. "testing"
)

// Only returns whether the call succeeded or panicked.
func tryCall(lenBytes int) bool {
	success := true
	defer func() {
		if r := recover(); r != nil {
			success = false
		}
	}()
	SecureRandomBase32(lenBytes)
	return true
}

func TestSecureRandomBase32(t *T) {
	successCases := []int{5, 10, 50}
	failureCases := []int{1, 6}
	for _, c := range successCases {
		if !tryCall(c) {
			t.Errorf("lenBytes==%v should succeed", c)
		}
	}
	for _, c := range failureCases {
		if tryCall(c) {
			t.Errorf("lenBytes==%v should fail", c)
		}
	}
}
