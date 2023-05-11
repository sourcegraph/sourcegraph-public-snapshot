//go:build amd64

package embeddings

import (
	"github.com/klauspost/cpuid/v2"
)

func init() {
	hasAVX2 := cpuid.CPU.Has(cpuid.AVX2)
	hasAVX512 := cpuid.CPU.Supports(
		cpuid.AVX512F,    // VPXORQ, VPSUBD, VPBROADCASTD
		cpuid.AVX512BW,   // VMOVDQU8, VADDB, VPSRLDQ
		cpuid.AVX512VNNI, // VPDPBUSD
	)

	if simdEnabled && hasAVX512 {
		dotArch = dotAVX2
	} else if simdEnabled && hasAVX2 {
		dotArch = dotAVX2
	}
}

func dotAVX2(a, b []int8) int32

func dotAVX512(a, b []int8) int32
