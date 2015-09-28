package fsync

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestSync(t *testing.T) {
	// create test directory and chdir to it
	dir, err := ioutil.TempDir(os.TempDir(), "fsync_test")
	check(err)
	check(os.Chdir(dir))

	// create test files and directories
	check(os.MkdirAll("src/a", 0755))
	check(ioutil.WriteFile("src/a/b", []byte("file b"), 0644))
	check(ioutil.WriteFile("src/c", []byte("file c"), 0644))
	// set times in the past to make sure times are synced, not accidentally
	// the same
	tt := time.Now().Add(-1 * time.Hour)
	check(os.Chtimes("src/a", tt, tt))
	check(os.Chtimes("src/a/b", tt, tt))
	check(os.Chtimes("src/c", tt, tt))

	// create Syncer
	s := NewSyncer()

	// sync
	check(s.SyncTo("dst", "src/a", "src/c"))

	// check results
	testDirContents("dst", 2, t)
	testDirContents("dst/a", 1, t)
	testFile("dst/a/b", []byte("file b"), t)
	testFile("dst/c", []byte("file c"), t)
	testPerms("dst", 0755, t)
	testPerms("dst/a", 0755, t)
	testPerms("dst/a/b", 0644, t)
	testPerms("dst/c", 0644, t)
	testModTime("dst", getModTime("src"), t)
	testModTime("dst/a", getModTime("src/a"), t)
	testModTime("dst/a/b", getModTime("src/a/b"), t)
	testModTime("dst/c", getModTime("src/c"), t)

	// modify src
	check(ioutil.WriteFile("src/a/b", []byte("file b changed"), 0644))
	check(os.Chmod("src/a", 0775))

	// sync
	check(s.Sync("dst", "src"))

	// check results
	testFile("dst/a/b", []byte("file b changed"), t)
	testPerms("dst/a", 0775, t)
	testModTime("dst/a", getModTime("src/a"), t)
	testModTime("dst/a/b", getModTime("src/a/b"), t)
	testModTime("dst/c", getModTime("src/c"), t)

	// remove c
	check(os.Remove("src/c"))

	// sync
	check(s.Sync("dst", "src"))

	// check results; c should still exist
	testDirContents("dst", 2, t)
	testExistence("dst/c", true, t)

	// sync
	s.Delete = true
	check(s.Sync("dst", "src"))

	// check results; c should no longer exist
	testDirContents("dst", 1, t)
	testExistence("dst/c", false, t)

	s.Delete = false
	if err = s.Sync("dst", "src/a/b"); err == nil {
		t.Errorf("expecting ErrFileOverDir, got nothing.\n")
	} else if err != nil && err != ErrFileOverDir {
		panic(err)
	}
}

func testFile(name string, b []byte, t *testing.T) {
	testExistence(name, true, t)
	c, err := ioutil.ReadFile(name)
	check(err)
	if !bytes.Equal(b, c) {
		t.Errorf("content of file \"%s\" is:\n%s\nexpected:\n%s\n",
			name, c, b)
	}
}

func testExistence(name string, e bool, t *testing.T) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		if e {
			t.Errorf("file \"%s\" does not exist.\n", name)
		}
	} else if err != nil {
		panic(err)
	} else {
		if !e {
			t.Errorf("file \"%s\" exists.\n", name)
		}
	}
}

func testDirContents(name string, count int, t *testing.T) {
	files, err := ioutil.ReadDir(name)
	check(err)
	if len(files) != count {
		t.Errorf("directory \"%s\" has %d children, shoud have %d.\n",
			name, len(files), count)
	}
}

func testPerms(name string, p os.FileMode, t *testing.T) {
	info, err := os.Stat(name)
	check(err)
	if info.Mode().Perm() != p {
		t.Errorf("permissions for \"%s\" is %v, should be %v.\n",
			name, info.Mode().Perm(), p)
	}
}

func testModTime(name string, m time.Time, t *testing.T) {
	info, err := os.Stat(name)
	check(err)
	if !info.ModTime().Equal(m) {
		t.Errorf("modification time for \"%s\" is %v, should be %v.\n",
			name, info.ModTime(), m)
	}
}

func getModTime(name string) time.Time {
	info, err := os.Stat(name)
	check(err)
	return info.ModTime()
}
