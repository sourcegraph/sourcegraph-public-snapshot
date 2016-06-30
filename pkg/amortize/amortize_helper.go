package amortize

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// ShouldAmortize is a non-deterministic function that returns true 'a' out of
// every 'b' times on average. ShouldAmortize panics if a > b, or if either a
// or b is not positive.
func ShouldAmortize(a, b int) bool {
	if a <= 0 || b <= 0 {
		panic("a and b must be positive")
	}
	if a > b {
		panic("a must be <= b")
	}

	bigN, err := rand.Int(rand.Reader, big.NewInt(int64(b)))
	if err != nil {
		panic(fmt.Sprintf("error when trying to generate a random number: %s", err))
	}
	return bigN.Int64() < int64(a)
}
