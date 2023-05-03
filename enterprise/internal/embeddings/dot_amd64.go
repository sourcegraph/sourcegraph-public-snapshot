//go:build amd64

package embeddings

import (
	"github.com/klauspost/cpuid/v2"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	simdEnabled = env.MustGetBool("ENABLE_EMBEDDINGS_SEARCH_SIMD", false, "Enable SIMD dot product for embeddings search")
	hasAVX2     = cpuid.CPU.Has(cpuid.AVX2)
	haveDotArch = simdEnabled && hasAVX2
)

func dotArch(a []int8, b []int8) int32 {
	if len(a) != len(b) {
		panic("mismatched lengths")
	}

	if len(a) == 0 {
		return 0
	}

	return int32(dotAVX2(a, b))
}

func dotAVX2(a, b []int8) int64
