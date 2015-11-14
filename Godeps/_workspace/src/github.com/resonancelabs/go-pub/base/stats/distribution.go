package stats

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/resonancelabs/go-pub/base"
	"github.com/resonancelabs/go-pub/base/fmath"

	"github.com/golang/glog"
)

const kExemplarSamplingPeriodMicros = 5 * base.MICROS_PER_SECOND
const kM2Epsilon = 0.0000001

type BucketExemplar struct {
	TimeMicros base.Micros
	SampleVal  float64
	Payload    interface{}
}

func (e *BucketExemplar) String() string {
	return fmt.Sprintf("BucketExemplar(TimeMicros=%v, SampleVal=%v, Payload=%+v)", e.TimeMicros.ToTime(), e.SampleVal, e.Payload)
}

// Distribution that supports positively valued float64 samples.
//
// Bucket max/right boundary "R" in terms of growth factor "g" and bucket index "i":
//
//   R = (g^(i+1) - 1) / (g - 1)
//
// Bucket index "i" in terms of growth factor "g" and sample value "x":
//
//   i = int((ln(1 + x*(g - 1)) / ln(g)) - 1)
//
// ... except that for g=1.0, we use width=1.0 buckets across the board.
// (XXX: allow for a custom bucket width in the g=1.0 case)
type Distribution struct {
	mean, m2 float64
	count    int64

	// Must be >= 1.0:
	bucketGrowthFactor float64
	maxSampleValue     float64

	// bucketCounts[len(bucketMaxes)] is for overflow. (Negative samples /
	// underflows are not allowed)
	bucketCounts    []int64
	bucketMaxes     []float64
	bucketExemplars []*BucketExemplar
}

// Commented out as the 1 bucket case is useful in some circumstances, but
// unfortunately a subset of the Distribution methods will quietly fail if
// called in this case.  The code should likely be refactored as distinct
// types to ensure invalid methods are not callable in such cases.
//
/*func NewDistribution() *Distribution {
	return &Distribution{
		// Special casing:
		bucketGrowthFactor: 0,
		maxSampleValue:     0,
		bucketCounts:       []int64{0},
		bucketMaxes:        []float64{0},
		bucketExemplars:    []*BucketExemplar{nil},
	}
}*/

func NewLinearDistribution(maxSampleValue float64) *Distribution {
	return NewDistributionWithBuckets(1, maxSampleValue)
}

func (d *Distribution) ApproxBytes(bytesPerExemplar int) int64 {
	return int64(5*8 + // POD fields
		len(d.bucketCounts)*8 +
		len(d.bucketMaxes)*8 +
		len(d.bucketExemplars)*bytesPerExemplar)
}

// Create a new distribution for *positive* float64 samples.  Negative
// samples are not supported.
//
// maxSampleValue
//		If samples above this value are added to the distribution, the
//		class will still behave "correctly" but at reduced numerical
//		accuracy.
//
func NewDistributionWithBuckets(bucketGrowthFactor float64, maxSampleValue float64) *Distribution {
	if bucketGrowthFactor < 1 {
		panic("bucket widths shrinking")
	}
	// +1 to compensate for int() truncation, +2 to include the overflow bucket.
	numBuckets := 2 + bucketIndexForSampleAndGrowthFactor(maxSampleValue, bucketGrowthFactor)
	bucketMaxes := make([]float64, numBuckets)
	if bucketGrowthFactor == 1 {
		for i := 0; i < numBuckets-1; i++ {
			bucketMaxes[i] = float64(i + 1)
		}
	} else {
		for i := 0; i < numBuckets-1; i++ {
			// See the formula in the Distribution type comment.
			bucketMaxes[i] = (math.Pow(bucketGrowthFactor, float64(i+1.0)) - 1.0) / (bucketGrowthFactor - 1.0)
		}
	}
	bucketMaxes[numBuckets-1] = math.MaxFloat64
	bucketCounts := make([]int64, len(bucketMaxes))
	bucketExemplars := make([]*BucketExemplar, len(bucketMaxes))
	return &Distribution{
		bucketGrowthFactor: bucketGrowthFactor,
		maxSampleValue:     maxSampleValue,
		bucketCounts:       bucketCounts,
		bucketMaxes:        bucketMaxes,
		bucketExemplars:    bucketExemplars,
	}
}

func (p *Distribution) NewEmptyDistributionWithMatchingBuckets() *Distribution {
	return NewDistributionWithBuckets(p.bucketGrowthFactor, p.maxSampleValue)
}

func NewDistributionFromData(
	bucketGrowthFactor float64, maxSampleValue float64,
	sum int64,
	count int64,
	stdDev float64,
	bucketCounts []int64,
	exemplars []*BucketExemplar) *Distribution {

	rval := NewDistributionWithBuckets(bucketGrowthFactor, maxSampleValue)
	rval.count = count
	if rval.count > 0 {
		rval.mean = float64(sum) / float64(count)
		rval.m2 = stdDev * stdDev * float64(count-1)
	}
	// bucketMaxes are taken care of; however, we do need to update the counts. Per ToThrift(), runs of 0-count buckets are encoding as their negative length.
	i := 0
	for _, val := range bucketCounts {
		if val < 0 {
			// Zero-length buckets; skip forward.
			i += int(-val)
		} else {
			rval.bucketCounts[i] = val
			// Advance by one for positive values, of course.
			i++
		}
	}
	for _, ex := range exemplars {
		exemplarIdx := rval.BucketIndexForSample(ex.SampleVal)
		rval.bucketExemplars[exemplarIdx] = ex
	}
	return rval
}

func (p *Distribution) Clone() *Distribution {
	numBuckets := len(p.bucketCounts)
	bucketMaxes := make([]float64, numBuckets)
	bucketCounts := make([]int64, numBuckets)
	bucketExemplars := make([]*BucketExemplar, numBuckets)
	copy(bucketMaxes, p.bucketMaxes)
	copy(bucketCounts, p.bucketCounts)
	copy(bucketExemplars, p.bucketExemplars)

	if p.m2 < 0 {
		glog.Warningf("p.m2=%v; should never happen! p=%v", p.m2, p)
	}

	return &Distribution{
		mean:               p.mean,
		m2:                 p.m2,
		count:              p.count,
		bucketGrowthFactor: p.bucketGrowthFactor,
		maxSampleValue:     p.maxSampleValue,
		bucketCounts:       bucketCounts,
		bucketMaxes:        bucketMaxes,
		bucketExemplars:    bucketExemplars,
	}
}

// Scale returns a new distribution that is computed by applying the given
// scale factor to counts in the input distribution.
func (p *Distribution) ScaleCounts(factor float64) *Distribution {
	numBuckets := len(p.bucketCounts)
	bucketMaxes := make([]float64, numBuckets)
	bucketCounts := make([]int64, numBuckets)
	bucketExemplars := make([]*BucketExemplar, numBuckets)
	copy(bucketMaxes, p.bucketMaxes)
	copy(bucketExemplars, p.bucketExemplars)
	totalCount := int64(0)
	for i, c := range p.bucketCounts {
		bucketCounts[i] = int64(fmath.Round64(float64(c) * factor))
		totalCount += bucketCounts[i]
		if bucketCounts[i] == 0 {
			bucketExemplars[i] = nil
		}
	}

	if p.m2 < 0 {
		glog.Warningf("p.m2=%v; should never happen! p=%v", p.m2, p)
	}

	return &Distribution{
		mean: p.mean,
		// TODO: not sure that scaling m2 linearly makes sense, but seems like
		// not a terrible thing to do.
		m2:                 p.m2 * factor,
		count:              totalCount,
		bucketGrowthFactor: p.bucketGrowthFactor,
		maxSampleValue:     p.maxSampleValue,
		bucketCounts:       bucketCounts,
		bucketMaxes:        bucketMaxes,
		bucketExemplars:    bucketExemplars,
	}
}

// This is not well-defined unless all samples ever added to `rhs` have also
// been added to the subject Distribution.
func (p *Distribution) Subtract(rhs *Distribution) (deltaDist *Distribution) {
	if p.bucketGrowthFactor != rhs.bucketGrowthFactor {
		panic(fmt.Errorf("inconsistent bucket growth factors"))
	}
	if len(p.bucketCounts) != len(rhs.bucketCounts) {
		panic(fmt.Errorf("inconsistent distribution sizes"))
	}
	numBuckets := len(p.bucketCounts)
	bucketMaxes := make([]float64, numBuckets)
	bucketCounts := make([]int64, numBuckets)
	bucketExemplars := make([]*BucketExemplar, numBuckets)

	copy(bucketMaxes, p.bucketMaxes)
	for i, lhsCount := range p.bucketCounts {
		bucketCounts[i] = lhsCount - rhs.bucketCounts[i]
	}
	for i, lhsExemplar := range p.bucketExemplars {
		// Ignore the case where the LHS is nil; there will be nothing to
		// subtract.
		if lhsExemplar != nil {
			// If both lhs and rhs have the same exact exemplar, replace it
			// with a nil in the delta distribution (as it's not new).
			// Otherwise, keep the lhs version.
			if lhsExemplar == rhs.bucketExemplars[i] {
				bucketExemplars[i] = nil
			} else {
				bucketExemplars[i] = lhsExemplar
			}
		}
	}

	deltaCount := p.count - rhs.count
	deltaMean := float64(0)
	if deltaCount > 0 {
		deltaMean = (p.mean*float64(p.count) - rhs.mean*float64(rhs.count)) / float64(deltaCount)
	}

	// NOTE(bhs): We are trying to compute our incremental "variance numerator"
	// m2 for the delta distribution. The math involved here is not
	// particularly sophisticated but easy [at least for me] to screw up.
	// Rather than try to write a latex-style comment, I took a picture of my
	// notebook. Here it is:
	//
	//   https://www.dropbox.com/s/v19t8tpv4sszl38/delta_distribution_m2_term_derivation.jpg?dl=0
	//
	// For more on the m2 term, read about the online algorithm for variance here:
	//
	//   http://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Online_algorithm
	//
	deltaM2 := p.m2 // in case `countRhs == 0` below
	{
		countLhs := float64(p.count)
		countRhs := float64(rhs.count)
		countDelta := countLhs - countRhs

		if countDelta > 0 && countRhs > 0 {
			if countDelta == 1 {
				// By definition of the m2 term, deltaM2 is 0 in this case;
				// without this conditional, numerical stability issues
				// sometimes give us an ever-so-slightly negative deltaM2 which
				// triggers the Warning below.
				deltaM2 = 0
			} else {
				totalLhs := countLhs * p.mean
				totalRhs := countRhs * rhs.mean
				totalDelta := totalLhs - totalRhs

				sumSquaresRhs := rhs.m2 + totalRhs*totalRhs/countRhs
				deltaM2 = p.m2 + (totalRhs*totalRhs+totalDelta*totalDelta+2*totalRhs*totalDelta)/countLhs - sumSquaresRhs - (totalDelta * totalDelta / countDelta)
				if deltaM2 < 0 {
					if deltaM2 < -kM2Epsilon {
						glog.Warningf(
							"deltaM2=%v; should never happen. p.m2=%v, totalRhs=%v, totalDelta=%v, countLhs=%v, sumSquaresRhs=%v, countDelta=%v",
							deltaM2, p.m2, totalRhs, totalDelta, countLhs, sumSquaresRhs, countDelta)
					}
					// Better safe than sorry / NaN:
					deltaM2 = 0
				}
			}
		}
	}

	return &Distribution{
		mean:               deltaMean,
		m2:                 deltaM2,
		count:              deltaCount,
		bucketGrowthFactor: p.bucketGrowthFactor,
		maxSampleValue:     p.maxSampleValue,
		bucketCounts:       bucketCounts,
		bucketMaxes:        bucketMaxes,
		bucketExemplars:    bucketExemplars,
	}
}

// Merges/Adds the two input distributions into a third (and potentially losing
// information about some exemplars in the process; the youngest exemplars are
// retained when there's a choice to make).
func DistributionMerge(a, b *Distribution) *Distribution {
	if a.bucketGrowthFactor != b.bucketGrowthFactor {
		panic(fmt.Errorf("inconsistent bucket growth factors"))
	}
	if len(a.bucketCounts) != len(b.bucketCounts) {
		panic(fmt.Errorf("inconsistent distribution sizes"))
	}
	numBuckets := len(a.bucketCounts)
	bucketMaxes := make([]float64, numBuckets)
	bucketCounts := make([]int64, numBuckets)
	bucketExemplars := make([]*BucketExemplar, numBuckets)

	copy(bucketMaxes, a.bucketMaxes)
	for i, aCount := range a.bucketCounts {
		bucketCounts[i] = aCount + b.bucketCounts[i]
	}
	// Choose the youngest available exemplar for each bucket.
	for i, aExemplar := range a.bucketExemplars {
		bExemplar := b.bucketExemplars[i]
		switch {
		case aExemplar == nil && bExemplar == nil:
			continue
		case aExemplar != nil && bExemplar == nil:
			// Use the 'a' Exemplar (it's the only one).
			bucketExemplars[i] = aExemplar
		case aExemplar == nil && bExemplar != nil:
			// Use the 'b' Exemplar (it's the only one).
			bucketExemplars[i] = bExemplar
		case aExemplar != nil && bExemplar != nil:
			// Use whichever Exemplar is younger.
			if aExemplar.TimeMicros > bExemplar.TimeMicros {
				bucketExemplars[i] = aExemplar
			} else {
				bucketExemplars[i] = bExemplar
			}
		}
	}

	totalCount := a.count + b.count
	totalMean := float64(0)
	if totalCount > 0 {
		totalMean = (a.mean*float64(a.count) + b.mean*float64(b.count)) / float64(totalCount)
	}

	// NOTE(bhs): See the comment for deltaM2 in Distribution.Subtract above.
	// This side of things is markedly easier so there's no need for a separate
	// dropbox picture of my scribblings.
	//
	// For more on the m2 term, read about the online algorithm for variance here:
	//
	//   http://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Online_algorithm
	//
	var totalM2 float64
	if a.count == 0 {
		totalM2 = b.m2
	} else if b.count == 0 {
		totalM2 = a.m2
	} else {
		countA := float64(a.count)
		countB := float64(b.count)
		countTotal := countA + countB

		if countTotal > 0 {
			totalA := countA * a.mean
			totalB := countB * b.mean

			sumSquaresA := a.m2 + totalA*totalA/countA
			sumSquaresB := b.m2 + totalB*totalB/countB
			totalM2 = sumSquaresA + sumSquaresB - (totalA*totalA+totalB*totalB+2*totalA*totalB)/countTotal
			if totalM2 < 0 {
				if totalM2 < -kM2Epsilon {
					glog.Warningf(
						"totalM2=%v; should never happen. sumSquaresA=%v, sumSquaresB=%v, totalA=%v, totalB=%v, countTotal=%v",
						totalM2, sumSquaresA, sumSquaresB, totalA, totalB, countTotal)
				}
				// Better safe than sorry / NaN:
				totalM2 = 0
			}
		}
	}

	return &Distribution{
		mean:               totalMean,
		m2:                 totalM2,
		count:              totalCount,
		bucketGrowthFactor: a.bucketGrowthFactor,
		maxSampleValue:     a.maxSampleValue,
		bucketCounts:       bucketCounts,
		bucketMaxes:        bucketMaxes,
		bucketExemplars:    bucketExemplars,
	}
}

func (p *Distribution) String() string {
	return fmt.Sprintf("[mean %f, stddev %f (%v)]", p.Mean(), p.StandardDeviation(), p.count)
}

func (p *Distribution) bucketString(b int, idx int, maxCount int64) string {
	count := p.bucketCounts[b]
	return fmt.Sprintf("%s", colorString(idx, strings.Repeat("#", 1+int((100*count)/maxCount))))
}

func (p *Distribution) BucketMin(b int) float64 {
	if b == 0 {
		return 0
	}
	return p.bucketMaxes[b-1]
}

func (p *Distribution) BucketMax(b int) float64 {
	return p.bucketMaxes[b]
}

func (p *Distribution) NumBuckets() int {
	return len(p.bucketCounts) - 1
}

func (p *Distribution) BucketCount(b int) int64 {
	return p.bucketCounts[b]
}

func (p *Distribution) BucketGrowthFactor() float64 {
	return p.bucketGrowthFactor
}

func (p *Distribution) DebugExemplars() string {
	rval := make([]string, len(p.bucketExemplars)+2)
	rval[0] = "|"
	rval[len(rval)-1] = "|"
	for i, be := range p.bucketExemplars {
		if be == nil {
			rval[i+1] = " "
		} else {
			rval[i+1] = "."
		}
	}
	return strings.Join(rval, "")
}

func (p *Distribution) Exemplar(b int) (valid bool, ex *BucketExemplar) {
	// TODO: range checking throughout for bucket index params
	e := p.bucketExemplars[b]
	return e != nil, e
}

func (p *Distribution) SampleCount() int64 {
	return p.count
}

// Returns the number of samples in this Distribution that are greater than or
// "near" (i.e., in the same bucket) as sampleVal.
func (p *Distribution) SampleCountGreaterOrNear(sample float64) int64 {
	rval := int64(0)
	for i := p.BucketIndexForSample(sample); i < len(p.bucketCounts); i++ {
		rval += p.bucketCounts[i]
	}
	return rval
}

func (p *Distribution) Mean() float64 {
	return p.mean
}

func (p *Distribution) Total() float64 {
	return p.mean * float64(p.SampleCount())
}

func (p *Distribution) StandardDeviation() float64 {
	if p.count > 1 {
		return math.Sqrt(p.m2 / float64(p.count-1))
	}
	return 0
}

func (p *Distribution) MaxSampleValue() float64 {
	return p.maxSampleValue
}

func bucketIndexForSampleAndGrowthFactor(sample, growthFactor float64) int {
	// The inverted test will also catch NaNs
	if !(sample >= 0) {
		panic("negative and NaN sample values not supported!")
	}
	if sample < 1 {
		return 0
	}
	if growthFactor == 1 {
		return int(sample)
	}
	// See the formula in the Distribution type comment.
	return int(math.Log(1+sample*(growthFactor-1)) / math.Log(growthFactor))
}

func (p *Distribution) BucketIndexForSample(sample float64) int {
	idx := bucketIndexForSampleAndGrowthFactor(sample, p.bucketGrowthFactor)
	if idx >= len(p.bucketCounts) {
		// the overflow bucket:
		idx = len(p.bucketCounts) - 1
	}
	return idx
}

func (p *Distribution) AddSample(sample float64) {
	p.AddSampleWithExemplar(sample, nil)
}

// payloadFunc() is only invoked if the sample is selected as an exemplar
// (i.e., it need not be ultra-cheap).
func (p *Distribution) AddSampleWithExemplar(sample float64, payload interface{}) {
	// There's nothing useful that can be done with a NaN sample
	if math.IsNaN(sample) {
		return
	}

	// Deal with the distribution-wide summary statistics first.
	//
	// See Knuth TAOCP vol 2, 3rd edition, page 232 or
	// http://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Online_algorithm
	p.count += 1
	delta := (sample - p.mean)
	p.mean += delta / float64(p.count)
	p.m2 += delta * (sample - p.mean)

	// Adjust bucket counts.
	if p.bucketGrowthFactor == 0 {
		// Only the overflow bucket.
		p.bucketCounts[0] += 1
		return
	}

	bucketIdx := p.BucketIndexForSample(sample)
	p.bucketCounts[bucketIdx] += 1

	// Handle exemplars.
	if payload != nil {
		now := base.NowMicros()
		if p.bucketExemplars[bucketIdx] == nil ||
			now-kExemplarSamplingPeriodMicros > p.bucketExemplars[bucketIdx].TimeMicros {
			p.bucketExemplars[bucketIdx] = &BucketExemplar{
				TimeMicros: now,
				SampleVal:  sample,
				Payload:    payload,
			}
		}
	}
}

// Returns the percentile of `sample` in the range [0.0, 1.0].
func (p *Distribution) PercentileForSample(sample float64) float64 {
	// Subtract the overflow bucket from the denominator.
	denominator := float64(p.count - p.bucketCounts[len(p.bucketCounts)-1])
	if denominator == 0 {
		return 0 // the percentile is undefined when we have no bucketed samples.
	}

	// Count the number of entries in the distribution less than the entry for
	// `sample`, linearly interpolating in `sample`s bucket.
	//
	// XXX: the linear interpolation is obviously sketchy... we could at least
	// make the interpolation follow the exponential curve of the bucket
	// boundaries.
	numerator := float64(0)
	for i := 0; i < len(p.bucketCounts)-1; i++ {
		if sample < p.bucketMaxes[i] {
			// Assume that samples in this bucket are distributed linearly, and
			// accumulate the linear fraction of this final bucket for the
			// eventual percentile numerator.
			finalBucketMin := float64(0)
			if i > 0 {
				finalBucketMin = p.bucketMaxes[i-1]
			}
			finalBucketMax := p.bucketMaxes[i]
			finalBucketFrac := (sample - finalBucketMin) / (finalBucketMax - finalBucketMin)
			numerator += finalBucketFrac * float64(p.bucketCounts[i])
			break
		}
		numerator += float64(p.bucketCounts[i])
	}

	return numerator / denominator
}

// Returns a sample value that would lie at `percentile` (in range [0.0, 1.0]).
func (p *Distribution) SampleForPercentile(percentile float64) float64 {
	// Subtract the overflow bucket from the denominator.
	denominator := float64(p.count - p.bucketCounts[len(p.bucketCounts)-1])
	if denominator == 0 {
		return 0 // the sample is undefined when we have no bucketed samples.
	}
	accumGoal := denominator * percentile
	currAccum := float64(0)
	for i := 0; i < len(p.bucketCounts)-1; i++ {
		prevAccum := currAccum
		currAccum += float64(p.bucketCounts[i])
		if currAccum > accumGoal {
			prevPctile := prevAccum / denominator
			currPctile := currAccum / denominator
			// TODO: this shouldn't be linear since we're using exponential
			// bucket widths.
			currBucketScaleFactor := (percentile - prevPctile) / (currPctile - prevPctile)

			// Estimate the sample value by looking at the bucket maxes on
			// either side of bucket `i`.
			leftBucketMax := float64(0)
			if i > 0 {
				leftBucketMax = p.bucketMaxes[i-1]
			}
			rightBucketMax := p.bucketMaxes[i]
			return leftBucketMax + (rightBucketMax-leftBucketMax)*currBucketScaleFactor
		}
	}
	// Fallback: return the max sample value for this distribution.
	return p.bucketMaxes[len(p.bucketMaxes)-1]
}

func (p *Distribution) FullString() string {
	lines := []string{p.String(), ""}
	maxCount := int64(0)
	lastNonemptyIndex := 0
	for i, count := range p.bucketCounts {
		if count > maxCount {
			maxCount = count
		}
		if count > 0 && i < len(p.bucketCounts)-1 {
			lastNonemptyIndex = i
		}
	}
	if maxCount > 0 {
		for i := 0; i <= lastNonemptyIndex; i++ {
			lines = append(lines, fmt.Sprintf("%8.2f - %8.2f: %v",
				p.BucketMin(i),
				p.BucketMax(i),
				p.bucketString(i, 8, maxCount)))
		}
		if p.bucketCounts[len(p.bucketCounts)-1] > 0 {
			lines = append(lines, fmt.Sprintf("%19v: %v", "overflow",
				p.bucketString(len(p.bucketCounts)-1, 8, maxCount)))
		}
	}

	return strings.Join(lines, "\n")
}

func colorString(idx int, s string) string {
	return fmt.Sprintf("\033[1;3%vm%v\033[0m", idx%7+1, s)
}

func RenderDistributionMap(dmap map[string]*Distribution) string {
	// XXX: verify that all distributions have the same bucket counts, etc.

	lines := []string{}
	sortedKeys := make([]string, len(dmap))
	i := 0
	for name, _ := range dmap {
		sortedKeys[i] = name
		i++
	}
	sort.Strings(sortedKeys)
	for i, key := range sortedKeys {
		lines = append(lines, fmt.Sprintf("%12s: %v", key, colorString(i, strings.Repeat("#", 10))))
	}
	lines = append(lines, "")
	maxCount := int64(0)
	lastNonemptyIndex := 0
	arbitraryDist := dmap[sortedKeys[0]]
	for i := 0; i < len(arbitraryDist.bucketCounts); i++ {
		totalCount := int64(0)
		for _, dist := range dmap {
			totalCount += dist.bucketCounts[i]
		}
		if totalCount > maxCount {
			maxCount = totalCount
		}
		if totalCount > 0 && i < len(arbitraryDist.bucketCounts)-1 {
			lastNonemptyIndex = i
		}
	}

	if maxCount > 0 {
		for i := 0; i <= lastNonemptyIndex; i++ {
			linePfx := fmt.Sprintf("%8.2f - %8.2f: ", arbitraryDist.BucketMin(i), arbitraryDist.BucketMax(i))
			segments := []string{}
			for c, key := range sortedKeys {
				dist := dmap[key]
				segments = append(segments, dist.bucketString(i, c, maxCount))
				lines = append(lines, fmt.Sprintf("%v%v", linePfx, segments[len(segments)-1]))
			}
			// lines = append(lines, fmt.Sprintf("%v%v", linePfx, strings.Join(segments, "")))
		}
		segments := []string{}
		overflowTotal := int64(0)
		for c, key := range sortedKeys {
			dist := dmap[key]
			segments = append(segments, dist.bucketString(len(dist.bucketCounts)-1, c, maxCount))
			overflowTotal += dist.bucketCounts[len(dist.bucketCounts)-1]
		}
		if overflowTotal > 0 {
			lines = append(lines, fmt.Sprintf("%19v: %v", "overflow", strings.Join(segments, "")))
		}
	}

	return strings.Join(lines, "\n")
}
