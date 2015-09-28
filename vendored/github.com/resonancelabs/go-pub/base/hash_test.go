package base

import (
	"fmt"
	"math/rand"
	"testing"
)

// Embarrasingly bad
func TestHashSimple(t *testing.T) {
	rand := rand.NewSource(0)
	for i := 0; i < 10000; i++ {
		// Shouldn't get the same result for swapped args
		a, b := rand.Int63(), rand.Int63()
		if HashInts64(a, b) == HashInts64(b, a) {
			t.Error(fmt.Sprintf("hash of (%v,%v) == (%v,%v)", a, b, b, a))
		}

		// 1 bit differences should not hash the same
		if HashInts64(a, 0) == HashInts64(a, 1) ||
			HashInts64(a, 0) == HashInts64(1, a) ||
			HashInts64(0, a) == HashInts64(1, a) ||
			HashInts64(0, a) == HashInts64(a, 1) {
			t.Error("Expected %v,1/1,%v to not be same as %v,0")
		}
	}
}
