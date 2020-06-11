package indexer

import (
	"os"
	"path/filepath"
	"strings"
)

// expectedCloneTarSizes are the expected file paths and sizes relative
// to the tarfile root of ./testdata/clone.tar.
var expectedCloneTarSizes = map[string]int{
	filepath.Join("x", "a"): 671,
	filepath.Join("x", "b"): 582,
	filepath.Join("y", "c"): 534,
	filepath.Join("y", "d"): 380,
	filepath.Join("z", "e"): 539,
	filepath.Join("z", "f"): 433,
	filepath.Join("g"):      393,
}

func readFiles(root string) (map[string]int, error) {
	sizes := map[string]int{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && !strings.HasPrefix(info.Name(), ".") {
			sizes[path[len(root)+1:]] = int(info.Size())
		}

		return nil
	})

	return sizes, err
}
