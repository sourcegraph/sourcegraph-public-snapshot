package store

import (
	"flag"
	"io/ioutil"
	"log"

	"sourcegraph.com/sourcegraph/rwvfs"
)

var fsType = flag.String("test.fs", "map", "vfs type to use for tests (map|os)")

func newTestFS() rwvfs.WalkableFileSystem {
	switch *fsType {
	case "map":
		fs := rwvfs.Map(map[string]string{})
		return rwvfs.Walkable(rwvfs.Sub(fs, "/testdata"))
	case "os":
		tmpDir, err := ioutil.TempDir("", "srclib-test")
		if err != nil {
			log.Fatal(err)
		}
		fs := rwvfs.OS(tmpDir)
		setCreateParentDirs(fs)
		return rwvfs.Walkable(fs)
	default:
		log.Fatalf("unrecognized -test.fs option: %q", *fsType)
		panic("unreachable")
	}
}
