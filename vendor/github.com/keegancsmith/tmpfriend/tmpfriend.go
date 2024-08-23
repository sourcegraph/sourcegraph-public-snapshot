package tmpfriend

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
)

var dirRe = regexp.MustCompile(`^tmpfriend-([0-9]+)-.*$`)

// SetupOrNOOP is the expected way to use tmpfriend. It is a wrapper around
// RootTempDir and IsTmpFriendDir. It will always return a cleanup function,
// and will do a NOOP under error.
//
//   cleanup := tmpfriend.SetupOrNOOP()
//   defer cleanup()
//   ...
func SetupOrNOOP() func() {
	if !IsTmpFriendDir("") {
		cleanup, err := RootTempDir("")
		if err == nil {
			return cleanup
		}
	}
	return func() {}
}

// RootTempDir creates a new TMPDIR tied to this process, as well as cleaning
// up TMPDIRs from defunct processes.
//
// TMPDIR is used by both ioutil.TempDir and os.TempDir. The callback returned
// will cleanup TMPDIR and restore its old value. Note this function is not
// safe to use concurently since modifying the environment is shared by the
// whole process.
//
// To use put code like this somewhere like `func main()`
//
//   if !IsTmpFriendDir("") {
//     f, err := tmpfriend.RootTempDir("")
//     if err != nil {
//       return err
//     }
//     defer f()
//   }
//   ...
func RootTempDir(rootDir string) (func(), error) {
	if rootDir == "" {
		rootDir = os.TempDir()
	}
	go clean(rootDir)
	prefix := fmt.Sprintf("tmpfriend-%d-", os.Getpid())
	dir, err := ioutil.TempDir(rootDir, prefix)
	if err != nil {
		return nil, err
	}
	old := os.Getenv("TMPDIR")
	os.Setenv("TMPDIR", dir)
	return func() {
		if os.Getenv("TMPDIR") == dir {
			os.Setenv("TMPDIR", old)
		}
		os.RemoveAll(dir)
	}, nil
}

// IsTmpFriendDir returns true if we are in a tmpfriend dir.
func IsTmpFriendDir(rootDir string) bool {
	if rootDir == "" {
		rootDir = os.TempDir()
	}
	return dirRe.MatchString(filepath.Base(rootDir))
}

func clean(rootDir string) {
	list, err := ioutil.ReadDir(rootDir)
	if err != nil {
		log.Printf("tmpfriend: Failed to read dir %s: %s", rootDir, err)
		return
	}
	for _, d := range list {
		m := dirRe.FindStringSubmatch(d.Name())
		if len(m) == 0 {
			continue
		}
		pid, err := strconv.Atoi(m[1])
		if err != nil || processIsRunning(pid) {
			continue
		}
		dir := filepath.Join(rootDir, d.Name())
		log.Printf("tmpfriend: Removing %s", dir)
		os.RemoveAll(dir)
	}
}

func processIsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err == nil {
		return process.Signal(syscall.Signal(0)) == nil
	}
	return false
}
