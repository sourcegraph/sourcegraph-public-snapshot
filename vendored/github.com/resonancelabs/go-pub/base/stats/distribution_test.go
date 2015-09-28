package stats

import (
	"math"
	"math/rand"
	"reflect"
	"runtime"
	. "testing"

	"github.com/golang/glog"
)

const kEpsilon = 0.0000001

func assertEqual(t *T, actual, expected interface{}) {

	actualType := reflect.TypeOf(actual)
	expectedType := reflect.TypeOf(expected)

	// Convert just to avoid trivial int -> int64 type mismatches. This is test
	// code after all.
	expectedCmp := expected
	if actualType != expectedType {
		expectedCmp = reflect.ValueOf(expected).Convert(actualType).Interface()
	}
	if !reflect.DeepEqual(actual, expectedCmp) {
		_, _, line, _ := runtime.Caller(1)
		t.Errorf("(line %d) Equality test failed : found=%v, expected=%v", line, actual, expected)
	}
}

func assertEquiv(t *T, a, b interface{}, epsilon float64) {

	float64Type := reflect.TypeOf(float64(0.0))
	fa := reflect.ValueOf(a).Convert(float64Type).Interface().(float64)
	fb := reflect.ValueOf(b).Convert(float64Type).Interface().(float64)

	if math.Abs(fb-fa) > epsilon {
		_, _, line, _ := runtime.Caller(1)
		t.Errorf("(line %d) Equality test failed : found=%v, expected=%v", line, a, b)
	}
}

func TestBasics(t *T) {
	dist := NewDistributionWithBuckets(1.0, 99.0)
	for i := 0; i < 100; i++ {
		dist.AddSample(float64(i))
	}

	assertEqual(t, dist.SampleCount(), 100)
	assertEquiv(t, dist.StandardDeviation(), 29.0114919, 1e-6)
	assertEqual(t, dist.Mean(), 49.5)

	// Expect percentiles to be accurate to the nearest percent
	const percentileEpsilon = 1e-2
	assertEquiv(t, dist.PercentileForSample(49.5), .500, percentileEpsilon)
	assertEquiv(t, dist.PercentileForSample(0), .005, percentileEpsilon)
	assertEquiv(t, dist.PercentileForSample(99), .995, percentileEpsilon)
}

func TestPercentiles(t *T) {
	const oneMinuteInMs = 60000
	latenciesMs := []float64{8.8, 10.1, 8.1, 9.4, 6.7, 40.3, 7.8, 8.3, 9.0, 120, 6.6, 8.7, 8.8, 8.3, 8.6}
	dist := NewDistributionWithBuckets(1.0, oneMinuteInMs)

	// Loop over the data to ensure the accuracy stays intact as more data is added
	for i := 0; i < 1000; i++ {
		for _, value := range latenciesMs {
			dist.AddSample(value)
		}

		assertEqual(t, dist.SampleCount(), (i+1)*len(latenciesMs))
		assertEquiv(t, dist.StandardDeviation(), 29.41168, 1.0)
		assertEquiv(t, dist.Mean(), 17.96667, 1e-5)

		// Expect percentiles to be accurate to 2 percent
		const percentileEpsilon = 0.02
		assertEquiv(t, dist.PercentileForSample(8), .200, percentileEpsilon)
		assertEquiv(t, dist.PercentileForSample(6), .0, percentileEpsilon)
		assertEquiv(t, dist.PercentileForSample(100), .933, percentileEpsilon)
		assertEquiv(t, dist.PercentileForSample(40.3), .900, percentileEpsilon)
	}
}

func TestSubtractionVariance(t *T) {
	// The algebra behind the deltaM2 local variable is "in-house" and not
	// entirely trusted. Worth writing a simple test to verify!
	//
	// The subroutine adds two batches of samples with different distributions,
	// taking a snapshot after the first batch. It then subtracts the
	// distribution from the first batch from the total and expects to get the
	// variance from the second batch.
	type batchSpec struct {
		N      int
		Mean   float64
		StdDev float64
	}
	verifyDeltaVariance := func(bs1, bs2 batchSpec) {
		// Distribution freaks out about negative samples.
		safeSamp := func(mean, stddev float64) float64 {
			samp := mean + stddev*rand.NormFloat64()
			if samp < 0 {
				return 0
			}
			return samp
		}
		dAll := NewDistributionWithBuckets(1.25, 60000)
		for i := 0; i < bs1.N; i++ {
			samp := safeSamp(bs1.Mean, bs1.StdDev)
			dAll.AddSample(samp)
		}
		d1 := dAll.Clone()

		d2Expected := NewDistributionWithBuckets(1.25, 60000)
		for i := 0; i < bs2.N; i++ {
			samp := safeSamp(bs2.Mean, bs2.StdDev)
			dAll.AddSample(samp)
			d2Expected.AddSample(samp)
		}
		d2 := dAll.Subtract(d1)

		expectedStdDev := d2Expected.StandardDeviation()
		actualStdDev := d2.StandardDeviation()
		absDiff := math.Abs(expectedStdDev - actualStdDev)
		glog.Infof("|%v - %v| = %v", expectedStdDev, actualStdDev, absDiff)
		if absDiff > kEpsilon {
			t.Errorf("Delta Distribution variance differs by more than epsilon; %v vs %v (|diff|=%v)", expectedStdDev, actualStdDev, absDiff)
		}
	}

	// Small N.
	verifyDeltaVariance(
		batchSpec{10, 100, 5},
		batchSpec{10, 300, 15})

	// Big N.
	verifyDeltaVariance(
		batchSpec{10000, 1000, 5},
		batchSpec{10000, 3000, 15})

	// Different N (part I)
	verifyDeltaVariance(
		batchSpec{10, 1000, 5},
		batchSpec{10000, 3000, 15})

	// Different N (part II)
	verifyDeltaVariance(
		batchSpec{10000, 1000, 5},
		batchSpec{10, 3000, 15})

	// Identical distributions.
	verifyDeltaVariance(
		batchSpec{10000, 1000, 5},
		batchSpec{10000, 1000, 5})
}

func TestMergeVariance(t *T) {
	// The algebra behind the totalM2 local variable is "in-house" and not
	// entirely trusted. Worth writing a simple test to verify!
	//
	// The subroutine adds two batches of samples with different distributions,
	// taking a snapshot after the first batch. It then subtracts the
	// distribution from the first batch from the total and expects to get the
	// variance from the second batch.
	type batchSpec struct {
		N      int
		Mean   float64
		StdDev float64
	}
	verifyMergedVariance := func(bs1, bs2 batchSpec) {
		// Distribution freaks out about negative samples.
		safeSamp := func(mean, stddev float64) float64 {
			samp := mean + stddev*rand.NormFloat64()
			if samp < 0 {
				return 0
			}
			return samp
		}
		d1 := NewDistributionWithBuckets(1.25, 60000)
		dAllExpected := NewDistributionWithBuckets(1.25, 60000)
		for i := 0; i < bs1.N; i++ {
			samp := safeSamp(bs1.Mean, bs1.StdDev)
			d1.AddSample(samp)
			dAllExpected.AddSample(samp)
		}

		d2 := NewDistributionWithBuckets(1.25, 60000)
		for i := 0; i < bs2.N; i++ {
			samp := safeSamp(bs2.Mean, bs2.StdDev)
			d2.AddSample(samp)
			dAllExpected.AddSample(samp)
		}
		dAll := DistributionMerge(d1, d2)

		expectedStdDev := dAllExpected.StandardDeviation()
		actualStdDev := dAll.StandardDeviation()
		absDiff := math.Abs(expectedStdDev - actualStdDev)
		glog.Infof("|%v - %v| = %v", expectedStdDev, actualStdDev, absDiff)
		if absDiff > kEpsilon {
			t.Errorf("Merged Distribution variance differs by more than epsilon; %v vs %v (|diff|=%v)", expectedStdDev, actualStdDev, absDiff)
		}
	}

	// Small N.
	verifyMergedVariance(
		batchSpec{10, 100, 5},
		batchSpec{10, 300, 15})

	// Big N.
	verifyMergedVariance(
		batchSpec{10000, 1000, 5},
		batchSpec{10000, 3000, 15})

	// Different N (part I)
	verifyMergedVariance(
		batchSpec{10, 1000, 5},
		batchSpec{10000, 3000, 15})

	// Different N (part II)
	verifyMergedVariance(
		batchSpec{10000, 1000, 5},
		batchSpec{10, 3000, 15})

	// Identical distributions.
	verifyMergedVariance(
		batchSpec{10000, 1000, 5},
		batchSpec{10000, 1000, 5})
}
