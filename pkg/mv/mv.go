// Package mv implements functionality ala the unix mv command
package mv

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/termie/go-shutil"
)

// Atomic attempts to do an atomic move. It first tries with an
// os.Rename. Failing that it falls back to a recursive copy to a tmp file on
// the same Filesystem, followed by an os.Rename.
func Atomic(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}
	if _, ok := err.(*os.LinkError); !ok {
		return err
	}
	// We had a LinkError, which likely means we couldn't rename across
	// filesystem boundaries. Instead we can do a copy + rename
	return move(src, dst)
}

func move(src, dst string) error {
	tmp, err := ioutil.TempDir(filepath.Dir(dst), ".tmp-mvatomic")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	tmp = filepath.Join(tmp, filepath.Base(dst))
	err = shutil.CopyTree(src, tmp, nil)
	if err != nil {
		return err
	}
	err = os.Rename(tmp, dst)
	if err == nil {
		os.RemoveAll(src)
	}
	return err
}
