package bench

import "github.com/sourcegraph/sourcegraph/cmd/bitmask"

func main() {
	z, err := bitmask.NewZipFileSystem("")
	if err != nil {
		panic(err)
	}
}
