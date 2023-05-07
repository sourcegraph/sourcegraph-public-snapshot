//go:build arm64

package embeddings

import (
	"github.com/klauspost/cpuid/v2"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	simdEnabled   = env.MustGetBool("ENABLE_EMBEDDINGS_SEARCH_SIMD", false, "Enable SIMD dot product for embeddings search")
	hasDotProduct = cpuid.CPU.Supports(cpuid.ASIMD, cpuid.ASIMDDP)
	haveDotArch   = simdEnabled && hasDotProduct
)

func dotArch(a []int8, b []int8) int32 {
	if len(a) != len(b) {
		panic("mismatched lengths")
	}

	if len(a) == 0 {
		return 0
	}

	rem := len(a) % 16
	blockA := a[:len(a)-rem]
	blockB := b[:len(b)-rem]

	sum := int32(dotSIMD(blockA, blockB))

	for i := len(a) - rem; i < len(a); i++ {
		sum += int32(a[i]) * int32(b[i])
	}

	return sum
}

func dotSIMD(a, b []int8) int64
