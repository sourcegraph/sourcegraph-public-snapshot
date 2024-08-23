package utils

import "strconv"

// Bucketize will put the value into the correct bucket.
// It is expected that the buckets are already sorted in increasing order and non-empty.
func Bucketize(value uint, buckets []uint) string {
	for _, bucketValue := range buckets {
		if value <= bucketValue {
			return strconv.Itoa(int(bucketValue))
		}
	}
	return ">" + strconv.Itoa(int(buckets[len(buckets)-1]))
}

// LinearBuckets returns an evenly distributed range of buckets in the closed interval
// [min...max]. The min and max count toward the bucket count since they are included
// in the range.
func LinearBuckets(min, max float64, count int) []float64 {
	var buckets []float64

	width := (max - min) / float64(count-1)

	for i := min; i <= max; i += width {
		buckets = append(buckets, i)
	}

	return buckets
}
