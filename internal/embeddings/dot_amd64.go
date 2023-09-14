//go:build amd64

package embeddings

import (
	"github.com/klauspost/cpuid/v2"
)

func init() {
	hasAVX2 := cpuid.CPU.Has(cpuid.AVX2)
	hasVNNI := cpuid.CPU.Supports(
		cpuid.AVX512F,    // required by VPXORQ, VPSUBD, VPBROADCASTD
		cpuid.AVX512BW,   // required by VMOVDQU8, VADDB, VPSRLDQ
		cpuid.AVX512VNNI, // required by VPDPBUSD
	)

	if simdEnabled && hasVNNI {
		dotArch = dotVNNI
	} else if simdEnabled && hasAVX2 {
		dotArch = dotAVX2
	}
}

func dotAVX2(a, b []int8) int32

func dotVNNI(a, b []int8) int32
