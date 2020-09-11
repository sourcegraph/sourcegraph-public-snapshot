package readers

import (
	"io/ioutil"
	"os"
	"sort"
	"strconv"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-bundle-manager/internal/paths"
)

// sqlitePaths returns the paths of all SQLite files currently on disk ordered by file size
// (largest first). We order the files this way as we want to do expensive migrations in the
// background rather than on the query path and larger files take longer.
func sqlitePaths(bundleDir string) ([]string, error) {
	fileInfos, err := ioutil.ReadDir(paths.DBsDir(bundleDir))
	if err != nil {
		return nil, err
	}

	sizeByPath := map[string]int64{}
	for _, fileInfo := range fileInfos {
		id, err := strconv.Atoi(fileInfo.Name())
		if err != nil {
			continue
		}

		filename := paths.SQLiteDBFilename(bundleDir, int64(id))
		stat, err := os.Stat(filename)
		if err != nil {
			continue
		}

		sizeByPath[filename] = stat.Size()
	}

	// Construct a slice of the map keys
	paths := make([]string, 0, len(sizeByPath))
	for path := range sizeByPath {
		paths = append(paths, path)
	}

	// Order by descending size
	sort.Slice(paths, func(i, j int) bool {
		return sizeByPath[paths[j]] < sizeByPath[paths[i]]
	})

	return paths, nil
}
