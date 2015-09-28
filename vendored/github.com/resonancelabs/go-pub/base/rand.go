package base

import (
	"math/rand"

	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/base/imath"
)

var gRandomFloat64Pool []float64

func init() {
	const (
		kArbitraryRngSeed = 220129
		kCount            = 4096
	)

	rng := rand.New(rand.NewSource(kArbitraryRngSeed))

	orderedList := make([]float64, kCount)
	for i := range orderedList {
		orderedList[i] = (float64(i) + rng.Float64()) / kCount
	}

	shuffle := rng.Perm(kCount)
	gRandomFloat64Pool = make([]float64, kCount)
	for i := range gRandomFloat64Pool {
		gRandomFloat64Pool[i] = orderedList[shuffle[i]]
	}
}

// Returns a consistent random float64 for the given seed with a normalized
// distribution in the range [0.0,1.0).
//
// Designed to be more convenient than creating a new seeded rand.Rand every time
// a consistent seeded float is needed, at the expense of less mathematically random
// behavior (since the pool is fixed).  I.e. don't use this for cryptography, use
// it when you want a simple pseudo-random float for a given seed.
//
func RandFloatWithSeed(seed int) float64 {
	return gRandomFloat64Pool[imath.Abs(seed)%len(gRandomFloat64Pool)]
}
