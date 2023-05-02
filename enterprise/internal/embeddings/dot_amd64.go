//go:build amd64

package embeddings

import (
	"github.com/klauspost/cpuid/v2"
)

var haveArchDot = cpuid.CPU.Has(cpuid.AVX2)

func archDot(a []int8, b []int8) int32 {
	if len(a) != len(b) {
		panic("mismatched lengths")
	}

	if len(a) == 0 {
		return 0
	}

	return int32(avx2Dot(a, b))
}

func avx2Dot(a, b []int8) int64
