package readers

import (
	"context"
	"io/ioutil"
	"os"
	"sort"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
	sqlitereader "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite"
)

// Migrate runs through each SQLite database on disk and opens a reader instance which will perform
// any necessary migrations to transform it to the newest schema. Because this may have a non-negligible
// cost cost some intersection of migrations and database size, we try to pay this cost up-front instead
// of being paid on-demand when the database is opened within the query path. This method does not block
// the startup of the bundle manager as it does not change the correctness of the service.
func Migrate(bundleDir string) error {
	paths, err := sqlitePaths(bundleDir)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return nil
	}

	log15.Info("Performing bundle migrations in background", "numBundles", len(paths))

	go func() {
		// After fetching the paths to convert, perform the migrations in a background
		// goroutine. This is just a best-effort migration and isn't a huge deal if it
		// happens to fail.

		for _, filename := range paths {
			log15.Debug("Migrating bundle", "filename", filename)

			if err := migrateDB(context.Background(), filename); err != nil {
				log15.Error("Failed to migrate bundle", "err", err, "filename", filename)
			}
		}

		log15.Info("Finished migrating bundles", "numBundles", len(paths))
	}()

	return nil
}

// migrateDB opens then immediately closes a reader instance for the given db filename.
func migrateDB(ctx context.Context, filename string) error {
	sqliteReader, err := sqlitereader.NewReader(ctx, filename)
	if err != nil {
		return err
	}

	if err := sqliteReader.Close(); err != nil {
		return err
	}

	return nil
}

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
