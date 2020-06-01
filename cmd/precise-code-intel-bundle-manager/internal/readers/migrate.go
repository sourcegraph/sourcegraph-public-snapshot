package readers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
	sqlitereader "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite"
)

func Migrate(bundleDir string) error {
	paths, err := sqlitePaths(bundleDir)
	if err != nil {
		return err
	}

	ch := make(chan string, len(paths))
	for _, path := range paths {
		ch <- path
	}
	close(ch)

	var wg sync.WaitGroup

	for i := 0; i < 64; i++ { // TODO - configure
		wg.Add(1)

		go func() {
			defer wg.Done()

			for filename := range ch {
				fmt.Printf(".\n")
				if err := migrateDB(context.Background(), filename); err != nil {
					log15.Error("Failed to migrate database", "err", err, "filename", filename)
				}
			}
		}()
	}

	fmt.Printf("WAITING\n")
	wg.Wait()
	fmt.Printf("DONE\n")
	return nil
}

func migrateDB(ctx context.Context, filename string) error {
	// TODO - needs to be locked against cache
	sqliteReader, err := sqlitereader.NewReader(ctx, filename)
	if err != nil {
		return err
	}

	if err := sqliteReader.Close(); err != nil {
		return err
	}

	return nil
}

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

	paths := make([]string, 0, len(sizeByPath))
	for path := range sizeByPath {
		paths = append(paths, path)
	}

	sort.Slice(paths, func(i, j int) bool {
		return sizeByPath[paths[i]] < sizeByPath[paths[j]]
	})

	return paths, nil
}
