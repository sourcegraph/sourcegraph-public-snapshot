//go:build amd64

package embeddings

import (
	"unsafe"

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

	return int32(avx2Dot(
		uintptr(unsafe.Pointer(&a[0])),
		uintptr(unsafe.Pointer(&b[0])),
		int64(len(a)),
	))
}

func avx2Dot(a, b uintptr, n int64) int64
