//go:build arm64

package embeddings

import (
	"github.com/klauspost/cpuid/v2"
)

func init() {
	hasDotProduct := cpuid.CPU.Supports(cpuid.ASIMD, cpuid.ASIMDDP)
	if simdEnabled && hasDotProduct {
		dotArch = dotSIMD
	}
}

func dotSIMD(a, b []int8) int32
