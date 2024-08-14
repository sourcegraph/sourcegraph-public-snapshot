package download

import (
	"os"
	"path"
	"testing"
)

func TestSafeRename(t *testing.T) {
	t.Run("general from local file to target dir", func(t *testing.T) {
		fd, err := os.Create("safe-1.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			os.Remove(fd.Name())
		}()

		tmpDir, err := os.MkdirTemp("", "safe-rename")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			os.RemoveAll(tmpDir)
		}()

		dst := path.Join(tmpDir, "safe-2.txt")

		if err := safeRename(fd.Name(), dst); err != nil {
			t.Errorf("failed to safeRename %q to %q", fd.Name(), dst)
		}

		if exists, err := fileExists(dst); err != nil {
			t.Fatal(err)
		} else if !exists {
			t.Errorf("expected %q to exist after safeRename", dst)
		}
	})
	t.Run("destination dir does not exist and gets created", func(t *testing.T) {
		fd, err := os.Create("safe-1.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			os.Remove(fd.Name())
		}()

		tmpDir, err := os.MkdirTemp("", "safe-rename")
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			os.RemoveAll(tmpDir)
		}()

		dst := path.Join(tmpDir, "i-do-not-exist", "safe-2.txt")

		if err := safeRename(fd.Name(), dst); err != nil {
			t.Errorf("failed to safeRename %q to %q", fd.Name(), dst)
		}

		if exists, err := fileExists(dst); err != nil {
			t.Fatal(err)
		} else if !exists {
			t.Errorf("expected %q to exist after safeRename", dst)
		}
	})

}
