//go:build !amd64

package embeddings

var haveArchDot = false

func dotArch(a, b []int8) int32 {
	panic("unimplemented")
}
