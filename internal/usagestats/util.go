package usagestats

import "github.com/sourcegraph/sourcegraph/lib/pointers"

func int32Ptr(v int) *int32 {
	return pointers.Ptr(int32(v))
}
