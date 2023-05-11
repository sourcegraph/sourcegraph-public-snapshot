//go:build amd64

package embeddings

import (
	"github.com/klauspost/cpuid/v2"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	simdEnabled = env.MustGetBool("ENABLE_EMBEDDINGS_SEARCH_SIMD", false, "Enable SIMD dot product for embeddings search")
	hasAVX2     = cpuid.CPU.Has(cpuid.AVX2)
	hasAVX512   = cpuid.CPU.Supports(
		cpuid.AVX512F,    // VPXORQ, VPSUBD, VPBROADCASTD
		cpuid.AVX512BW,   // VMOVDQU8, VADDB, VPSRLDQ
		cpuid.AVX512VNNI, // VPDPBUSD
	)
	haveDotArch = simdEnabled && (hasAVX512 || hasAVX2)
)

func dotArch(a []int8, b []int8) int32 {
	if len(a) != len(b) {
		panic("mismatched lengths")
	}

	if len(a) == 0 {
		return 0
	}

	if hasAVX512 {
		rem := len(a) % 64
		blockA := a[:len(a)-rem]
		blockB := b[:len(b)-rem]

		sum := int32(dotAVX512(blockA, blockB))

		for i := len(a) - rem; i < len(a); i++ {
			sum += int32(a[i]) * int32(b[i])
		}

		return sum
	} else {
		rem := len(a) % 16
		blockA := a[:len(a)-rem]
		blockB := b[:len(b)-rem]

		sum := int32(dotAVX2(blockA, blockB))

		for i := len(a) - rem; i < len(a); i++ {
			sum += int32(a[i]) * int32(b[i])
		}

		return sum
	}
}

func dotAVX2(a, b []int8) int64

func dotAVX512(a, b []int8) int64
