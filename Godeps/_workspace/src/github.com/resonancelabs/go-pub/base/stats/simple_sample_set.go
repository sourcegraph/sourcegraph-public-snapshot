// SimpleSampleSet
//
// A naive implementation for computing simple stats like min, max, average,
// and standard deviation on a set of float64 samples.
//
// NOTE: consider using the Distribution type instead, as it is (nearly) a
// superset of the functionality here.
//
package stats

import (
	"math"
	"sort"
)

type SimpleSampleStats struct {
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Sum      float64 `json:"sum"`
	Avg      float64 `json:"average"`
	Variance float64 `json:"variance"`
	StdDev   float64 `json:"standard_deviation"`
}

type SimpleSampleSet struct {
	SimpleSampleStats
	Values []float64 `json:"-"` // Do not export the values themselves
}

func NewSimpleSampleSet(estimatedSize int) *SimpleSampleSet {
	return &SimpleSampleSet{
		SimpleSampleStats: SimpleSampleStats{
			Min: math.Inf(+1),
			Max: math.Inf(-1),
		},
		Values: make([]float64, 0, estimatedSize),
	}
}

func (p *SimpleSampleSet) Add(s float64) {
	p.Sum += s
	p.Values = append(p.Values, s)
}

// Call when done calling Add() to compute the aggregate
// stats.
//
// Note: the implementation is idempotent.
//
func (p *SimpleSampleSet) Finish() {
	if len(p.Values) == 0 {
		return
	}

	sort.Float64Slice(p.Values).Sort()
	p.Min = p.Values[0]
	p.Max = p.Values[len(p.Values)-1]
	p.Avg = p.Sum / float64(len(p.Values))

	sum := 0.0
	for _, v := range p.Values {
		d := v - p.Avg
		sum += d * d
	}
	p.Variance = sum / float64(len(p.Values))
	p.StdDev = math.Sqrt(p.Variance)
}

// Return a copy of the stats only
func (p *SimpleSampleSet) Stats() *SimpleSampleStats {
	p.Finish()
	q := &SimpleSampleStats{}
	*q = p.SimpleSampleStats
	return q
}

// Creates a new object with the "fraction" outliers trimmed away.  The
// fraction is based on the number of samples, not percentiles.
//
// E.g. a "fraction" of 0.05 will remove lowest 2.5% values and highest
// 2.5% values.
func (p *SimpleSampleSet) TrimOutliers(fraction float64) *SimpleSampleSet {
	trim := int(float64(len(p.Values)) * (fraction / 2.0))
	subset := NewSimpleSampleSet(len(p.Values) - 2*trim)
	for i := trim; i < len(p.Values)-trim; i++ {
		subset.Add(p.Values[i])
	}
	subset.Finish()
	return subset
}
