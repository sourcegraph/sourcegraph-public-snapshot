package migrator

import (
	"io/ioutil"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/reader"
	jsonserializer "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/serializer/json"
)

type Migrator struct {
	bundleDir string
}

func New(bundleDir string) *Migrator {
	return &Migrator{
		bundleDir: bundleDir,
	}
}

func (m *Migrator) Run() {
	infos, err := ioutil.ReadDir(paths.DBsDir(m.bundleDir))
	if err != nil {
		panic(err.Error())
	}

	for _, info := range infos {
		id, err := strconv.Atoi(info.Name())
		if err != nil {
			continue
		}

		filename := paths.SQLiteDBFilename(m.bundleDir, int64(id))

		//
		// TODO - need to do some sort of caching
		// TODO - need to early-out the things we know won't be migrated
		//

		sqliteReader, err := reader.NewSQLiteReader(filename, jsonserializer.New())
		if err != nil {
			panic(err.Error())
		}
		if err := sqliteReader.Close(); err != nil {
			panic(err.Error())
		}
	}
}
