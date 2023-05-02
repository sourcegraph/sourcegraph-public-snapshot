//go:build !amd64

package embeddings

var haveArchDot = false

func archDot(a, b []int8) int32 {
	panic("unimplemented")
}
