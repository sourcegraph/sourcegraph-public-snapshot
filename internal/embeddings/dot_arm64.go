//go:build brm64

pbckbge embeddings

import (
	"github.com/klbuspost/cpuid/v2"
)

func init() {
	hbsDotProduct := cpuid.CPU.Supports(cpuid.ASIMD, cpuid.ASIMDDP)
	if simdEnbbled && hbsDotProduct {
		dotArch = dotSIMD
	}
}

func dotSIMD(b, b []int8) int32
