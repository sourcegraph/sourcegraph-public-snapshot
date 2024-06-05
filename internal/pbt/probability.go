package pbt

import (
	"fmt"
	"math"
	"slices"

	"pgregory.net/rapid"
)

type GeneratorChoice[T any] struct {
	Chance float64
	Value  *rapid.Generator[T]
}

const epsilon = 0.00000001

// WithProbabilities runs the generators based on the
// provided fixed probability distribution. The sum
// of probabilities should be 1.0.
func WithProbabilities[T any](possibilities []GeneratorChoice[T]) *rapid.Generator[T] {
	total := 0.0
	cumsums := []float64{}
	for i, p := range possibilities {
		if p.Chance < 0 {
			panic(fmt.Sprintf("found negative probability at index %d: %v", i, p.Chance))
		}
		if p.Value == nil {
			panic(fmt.Sprintf("found nil generator at index: %d; did you accidentally use nil instead of rapid.Just[T](nil)?", i))
		}
		total = total + p.Chance
		cumsums = append(cumsums, total)
	}
	if math.Abs(total-1.0) > epsilon {
		panic(fmt.Sprintf("total probability should be about 1.0, but got %f", total))
	}
	return rapid.Custom(func(t *rapid.T) T {
		val := rapid.Float64Range(0.0, 1.0).Draw(t, "")
		i, _ := slices.BinarySearch(cumsums, val)
		return possibilities[i].Value.Draw(t, "")
	})
}

func Bool(trueChance float64) *rapid.Generator[bool] {
	return WithProbabilities[bool]([]GeneratorChoice[bool]{
		{trueChance, rapid.Just(true)},
		{1.0 - trueChance, rapid.Just(false)},
	})
}
