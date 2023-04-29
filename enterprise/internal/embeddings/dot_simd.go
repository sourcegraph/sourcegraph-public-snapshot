//go:build amd64 && cgo

package embeddings

// #cgo CFLAGS: -mavx2 -O2 -fno-stack-protector
/*
#include <x86intrin.h>
#include <stdint.h>

int32_t dot(const int8_t *a, const int8_t *b, int64_t length) {
	int64_t partition = length - length % 16;

    __m256i sum = _mm256_setzero_si256();
    for (int64_t i = 0; i < partition; i += 16) {
        __m128i a_vec = _mm_loadu_si128((const __m128i *) (a + i));
        __m128i b_vec = _mm_loadu_si128((const __m128i *) (b + i));
		__m256i a_vec_16 = _mm256_cvtepi8_epi16(a_vec);
        __m256i b_vec_16 = _mm256_cvtepi8_epi16(b_vec);
        __m256i prod = _mm256_madd_epi16(a_vec_16, b_vec_16);
        sum = _mm256_add_epi32(sum, prod);
    }

    int32_t results[8];
    _mm256_store_si256((__m256i*)results, sum);

	int32_t result = 0;
    for (int i = 0; i < 8; ++i) {
		result += results[i];
    }

	for (int64_t i = partition; i < length; ++i) {
		result += (int32_t)(a[i]) * (int32_t)(b[i]);
	}

    return result;
}
*/
import "C"

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

	return int32(C.dot(
		(*C.int8_t)(unsafe.Pointer(&a[0])),
		(*C.int8_t)(unsafe.Pointer(&b[0])),
		C.int64_t(int64(len(a))),
	))
}
