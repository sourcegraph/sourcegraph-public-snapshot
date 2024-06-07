package pbt

import (
	"fmt"
	"golang.org/x/exp/rand"
	"math"
	"slices"

	"pgregory.net/rapid"
)

type WithProbability[T any] struct {
	Chance float64
	Value  T
}

type GeneratorChoice[T any] WithProbability[*rapid.Generator[T]]

const epsilon = 0.00000001

// WithProbabilities runs the generators based on the
// provided fixed probability distribution. The sum
// of probabilities should be 1.0.
//
// The rng should be seeded using rapid.Uint64() for
// maintaining determinism.
func WithProbabilities[T any](rng *rand.Rand, possibilities []GeneratorChoice[T]) *rapid.Generator[T] {
	total := 0.0
	cumulativeSums := []float64{}
	for i, p := range possibilities {
		if math.IsNaN(p.Chance) {
			panic(fmt.Sprintf("found NaN probability at index %d", i))
		}
		if p.Chance < 0 {
			panic(fmt.Sprintf("found negative probability at index %d: %v", i, p.Chance))
		}
		if p.Value == nil {
			panic(fmt.Sprintf("found nil generator at index: %d; did you accidentally use nil instead of rapid.Just[T](nil)?", i))
		}
		total = total + p.Chance
		cumulativeSums = append(cumulativeSums, total)
	}
	if math.Abs(total-1.0) > epsilon {
		panic(fmt.Sprintf("total probability should be about 1.0, but got %f", total))
	}
	return rapid.Custom(func(t *rapid.T) T {
		val := rng.Float64() // between 0 and 1.0 by default
		i, _ := slices.BinarySearch(cumulativeSums, val)
		return possibilities[i].Value.Draw(t, "")
	})
}

// Bool returns true with probability trueChance.
//
// The rng should be seeded using rapid.Uint64() for
// maintaining determinism.
func Bool(rng *rand.Rand, trueChance float64) *rapid.Generator[bool] {
	return WithProbabilities[bool](rng, []GeneratorChoice[bool]{
		{trueChance, rapid.Just(true)},
		{1.0 - trueChance, rapid.Just(false)},
	})
}
