//go:build !amd64 && !arm64

package embeddings

var haveDotArch = false

func dotArch(a, b []int8) int32 {
	panic("unimplemented")
}
