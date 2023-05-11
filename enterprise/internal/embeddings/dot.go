package embeddings

import "github.com/sourcegraph/sourcegraph/internal/env"

var (
	simdEnabled = env.MustGetBool("ENABLE_EMBEDDINGS_SEARCH_SIMD", false, "Enable SIMD dot product for embeddings search")

	// dotArch is a dot product function that may have architecture-specific
	// optimizations. Its inputs are expected to be the same length, which must
	// be a multiple of 64, which is least common multiple of all the implementations.
	dotArch func([]int8, []int8) int32 = dotPortable
)

// Dot computes the dot product of the two vectors:
// sum_i(a_i * b_i) for i in [0, len(a)).
//
// Precondition: len(a) == len(b)
func Dot(a, b []int8) int32 {
	if len(a) != len(b) {
		panic("mismatched lengths")
	}

	if len(a) == 0 {
		return 0
	}

	rem := len(a) % 64
	blockA := a[:len(a)-rem]
	blockB := b[:len(b)-rem]

	sum := dotArch(blockA, blockB)

	for i := len(a) - rem; i < len(a); i++ {
		sum += int32(a[i]) * int32(b[i])
	}

	return sum
}

func dotPortable(a, b []int8) int32 {
	similarity := int32(0)

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
		similarity += (m0 + m1 + m2 + m3)
	}

	return similarity
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
