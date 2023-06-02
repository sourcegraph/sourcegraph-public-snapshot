package embeddings

import "github.com/sourcegraph/sourcegraph/internal/env"

var (
	simdEnabled = env.MustGetBool("ENABLE_EMBEDDINGS_SEARCH_SIMD", true, "Enable SIMD dot product for embeddings search")

	// dotArch is a dot product function that is architecture-optimized. Its
	// inputs must be of equal length and that length must be a multiple of 64.
	// The default implementation will work on all architectures, but it may
	// be overridden if a more efficient method is available.
	dotArch = func(a, b []int8) int32 {
		sum := int32(0)

		count := len(a)
		if count > len(b) {
			// Do this ahead of time so the compiler doesn't need to bounds check
			// every time we index into b.
			panic("mismatched vector lengths")
		}

		for i := 0; i+3 < count; i += 4 {
			m0 := int32(a[i]) * int32(b[i])
			m1 := int32(a[i+1]) * int32(b[i+1])
			m2 := int32(a[i+2]) * int32(b[i+2])
			m3 := int32(a[i+3]) * int32(b[i+3])
			sum += (m0 + m1 + m2 + m3)
		}

		return sum
	}
)

// Dot computes the dot product of the two vectors:
// sum_i(a_i * b_i) for i in [0, len(a)).
//
// Precondition: len(a) == len(b)
func Dot(a, b []int8) int32 {
	if len(a) != len(b) {
		panic("mismatched lengths")
	}

	// dotArch requires 64-byte chunks.
	rem := len(a) % 64
	blockA := a[:len(a)-rem]
	blockB := b[:len(b)-rem]

	sum := dotArch(blockA, blockB)

	// add the remaining elements separately
	for i := len(a) - rem; i < len(a); i++ {
		sum += int32(a[i]) * int32(b[i])
	}

	return sum
}

func DotFloat32(row []float32, query []float32) float32 {
	similarity := float32(0)

	count := len(row)
	if count > len(query) {
		// Do this ahead of time so the compiler doesn't need to bounds check
		// every time we index into query.
		panic("mismatched vector lengths")
	}

	i := 0
	for ; i+3 < count; i += 4 {
		m0 := row[i] * query[i]
		m1 := row[i+1] * query[i+1]
		m2 := row[i+2] * query[i+2]
		m3 := row[i+3] * query[i+3]
		similarity += (m0 + m1 + m2 + m3)
	}

	for ; i < count; i++ {
		similarity += row[i] * query[i]
	}

	return similarity
}
