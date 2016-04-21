package load

import (
	"math/rand"

	"github.com/dgryski/go-discreterand"
	"github.com/tsenart/vegeta/lib"
)

// WeightedTarget has a Weight multiplier affecting its probability of being
// Targetted
type WeightedTarget struct {
	Target vegeta.Target
	Weight float64
}

// NewWeightedTargeter will return a Targeter which randomly selects a Target
// with probability Weight / sum(tgts.Weight)
func NewWeightedTargeter(tgts ...WeightedTarget) vegeta.Targeter {
	var sum float64
	for i := range tgts {
		sum += tgts[i].Weight
	}
	probs := make([]float64, len(tgts))
	for i := range tgts {
		probs[i] = tgts[i].Weight / sum
	}
	// We want a rand.Source that is safe for concurrent
	// access. rand.NewSource is.
	r := discreterand.NewAlias(probs, rand.NewSource(rand.Int63()))
	return func(tgt *vegeta.Target) error {
		if tgt == nil {
			return vegeta.ErrNilTarget
		}
		*tgt = tgts[r.Next()].Target
		return nil
	}
}
