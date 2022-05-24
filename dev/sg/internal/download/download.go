package download

import (
	"bytes"
	"context"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/unpack"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Executable downloads a binary from the given URL, updates the given path if different, and
// makes the downloaded file executable.
func Executable(url string, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Newf("downloading %s: status %d",
			url, resp.StatusCode)
	}

	content := &bytes.Buffer{}
	content.ReadFrom(resp.Body)

	updated, err := fileutil.UpdateFileIfDifferent(path, content.Bytes())
	if err != nil {
		return errors.Wrapf(err, "saving to %q", path)
	}
	if updated {
		return exec.Command("chmod", "+x", path).Run()
	}

	return nil
}

// ArchivedExecutable downloads an executable that's in an archive and extracts
// it.
func ArchivedExecutable(ctx context.Context, url, targetFile, fileInArchive string) error {
	if ok, _ := fileExists(targetFile); ok {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Newf("GET %s failed, status code is not OK: %d", url, resp.StatusCode)
	}
	defer resp.Body.Close()

	// Create a temp directory to unarchive files to
	tmpDirName, err := os.MkdirTemp("", "sg-binary-download*")
	if err != nil {
		return err
	}
	defer func() {
		// Clean up the temporary directory but ignore any possible errors
		_ = os.Remove(tmpDirName)
	}()

	// Only extract the file that we want
	opts := unpack.Opts{
		Filter: func(path string, file fs.FileInfo) bool {
			return path == fileInArchive && !file.IsDir()
		},
	}
	if err := unpack.Tgz(resp.Body, tmpDirName, opts); err != nil {
		return errors.Wrap(err, "unpacking archive failed")
	}

	fileInArchivePath := filepath.Join(tmpDirName, fileInArchive)
	if ok, err := fileExists(fileInArchivePath); !ok || err != nil {
		return errors.Newf("expected %s to exist in extracted archive at %s, but does not", fileInArchivePath, tmpDirName)
	}

	if err := os.Rename(fileInArchivePath, targetFile); err != nil {
		return err
	}

	return exec.Command("chmod", "+x", targetFile).Run()
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
