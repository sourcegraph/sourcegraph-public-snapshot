package pbt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
	"pgregory.net/rapid"
)

func TestWithProbability(t *testing.T) {
	chanceMap := map[int]float64{
		2: 0.5,
		5: 0.5,
	}
	choices := []GeneratorChoice[int]{}
	for i, f := range chanceMap {
		choices = append(choices, GeneratorChoice[int]{
			Chance: f,
			Value:  rapid.Just(i),
		})
	}
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64().Draw(t, "")
		gen := WithProbabilities[int](rand.New(rand.NewSource(seed)), choices)
		values := rapid.SliceOfN(gen, 1000, 1000).Draw(t, "")
		counts := map[int]int{}
		for inputValue, _ := range chanceMap {
			counts[inputValue] = 0
		}
		for _, v := range values {
			counts[v] += 1
		}
		countsWithProbability := []WithProbability[int]{}
		for v, count := range counts {
			expectedChance, ok := chanceMap[v]
			require.True(t, ok, "value in slice was not recorded in input distribution: %d", v)
			countsWithProbability = append(countsWithProbability, WithProbability[int]{
				Chance: expectedChance,
				Value:  count,
			})
		}
		// Be much more lax than the usual 0.05 significance level since we'll be running
		// hundreds of tests, so it's quite possible we see some failure.
		// There is a 0.01% chance of this test failing. This number can be reduced
		// further if this test ends up being too flaky.
		testPearsonChiSquared(t, countsWithProbability, 0.0001)
	})
}

func testPearsonChiSquared(t require.TestingT, counts []WithProbability[int], significanceLevel float64) {
	chiSquared := pearsonChiSquared(counts)
	limit := distuv.ChiSquared{K: float64(len(counts) - 1), Src: rand.NewSource(0)}.Quantile(1 - significanceLevel)
	if chiSquared >= limit {
		fmt.Printf("dof: %d, limit: %f\n", len(counts)-1, limit)
	}
	require.Less(t, chiSquared, limit)
}

func pearsonChiSquared(counts []WithProbability[int]) float64 {
	//https://en.wikipedia.org/wiki/Pearson%27s_chi-squared_test
	sum := 0.0
	totalCount := 0
	for _, count := range counts {
		totalCount += count.Value
	}
	for _, count := range counts {
		if count.Chance < epsilon {
			panic("Expected non-zero probability")
		}
		expectedCount := float64(totalCount) * count.Chance
		observedCount := count.Value
		countDifference := float64(observedCount) - expectedCount
		sum += countDifference * (countDifference / expectedCount)
	}
	return sum
}
