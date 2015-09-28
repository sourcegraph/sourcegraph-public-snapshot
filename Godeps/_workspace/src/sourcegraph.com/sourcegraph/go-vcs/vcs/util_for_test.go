package vcs_test

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

var logTmpDirs = flag.Bool("logtmpdirs", false, "log the temporary directories used by each test for inspection/debugging")

// baseTempDir is the parent directory for all temporary directories
// used by tests. Before each test run, all of its subdirectories are
// removed.
var baseTempDir = filepath.Join(os.TempDir(), "go-vcs-test")

func init() {
	// Remove and recreate baseTempDir.
	// if err := os.RemoveAll(baseTempDir); err != nil {
	// 	log.Fatal(err)
	// }
	if err := os.MkdirAll(baseTempDir, 0700); err != nil {
		log.Fatal(err)
	}
}

// makeTmpDir creates a temporary directory underneath baseTempDir.
func makeTmpDir(t testing.TB, suffix string) string {
	dir, err := ioutil.TempDir(baseTempDir, suffix)
	if err != nil {
		t.Fatal(err)
	}
	if *logTmpDirs {
		t.Logf("Using temp dir %s", dir)
	}
	return dir
}

func asJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
