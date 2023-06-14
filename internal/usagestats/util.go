package usagestats

import "github.com/sourcegraph/sourcegraph/internal/pointers"

func int32Ptr(v int) *int32 {
	return pointers.Ptr(int32(v))
}
