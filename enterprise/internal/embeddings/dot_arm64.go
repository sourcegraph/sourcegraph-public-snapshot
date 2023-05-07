//go:build arm64

package embeddings

import (
	"github.com/klauspost/cpuid/v2"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	simdEnabled   = env.MustGetBool("ENABLE_EMBEDDINGS_SEARCH_SIMD", false, "Enable SIMD dot product for embeddings search")
	hasDotProduct = cpuid.CPU.Has(cpuid.ASIMDDP)
	haveDotArch   = simdEnabled && hasDotProduct
)

func dotArch(a []int8, b []int8) int32 {
	la := len(a)
	lb := len(b)

	if la != lb {
		panic("mismatched lengths")
	}

	if la == 0 {
		return 0
	}

	rem := la % 16
	blockA := a[:la-rem]
	blockB := b[:lb-rem]

	sum := int32(dotSIMD(blockA, blockB))

	for i := la - rem; i < la; i++ {
		sum += int32(a[i]) * int32(b[i])
	}

	return sum
}

func dotSIMD(a, b []int8) int64
