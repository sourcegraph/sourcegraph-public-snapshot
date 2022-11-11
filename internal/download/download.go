package download

import (
	"bytes"
	"context"
	"io"
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
func Executable(ctx context.Context, url string, path string, failOn404 bool) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Sometimes the release is available, but the binaries are not
	if resp.StatusCode == http.StatusNotFound {
		if failOn404 {
			return false, errors.Newf("%s not found", url)
		}
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, errors.Newf("downloading %s: status %d",
			url, resp.StatusCode)
	}

	content := &bytes.Buffer{}
	if n, err := content.ReadFrom(resp.Body); err != nil {
		return false, errors.Wrap(err, "reading response")
	} else if n == 0 {
		return false, errors.New("got empty response")
	}

	updated, err := fileutil.UpdateFileIfDifferent(path, content.Bytes())
	if err != nil {
		return false, errors.Wrapf(err, "saving to %q", path)
	}
	if updated {
		return true, exec.CommandContext(ctx, "chmod", "+x", path).Run()
	}

	return false, nil
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
	tmpDirName, err := os.MkdirTemp("", "archived-executable-download*")
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
			return filepath.Clean(path) == fileInArchive && !file.IsDir()
		},
	}
	if err := unpack.Tgz(resp.Body, tmpDirName, opts); err != nil {
		return errors.Wrap(err, "unpacking archive failed")
	}

	fileInArchivePath := filepath.Join(tmpDirName, fileInArchive)
	if ok, err := fileExists(fileInArchivePath); !ok || err != nil {
		return errors.Newf("expected %s to exist in extracted archive at %s, but does not", fileInArchivePath, tmpDirName)
	}

	if err := safeRename(fileInArchivePath, targetFile); err != nil {
		return err
	}

	return exec.CommandContext(ctx, "chmod", "+x", targetFile).Run()
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

// safeRename copies src into dst before finally removing src.
// This is needed because in some cause, the tmp folder is living
// on a different filesystem.
func safeRename(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	inStat, err := in.Stat()
	perm := inStat.Mode() & os.ModePerm
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if err != nil {
		closeErr := in.Close()
		return errors.Append(err, closeErr)
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	closeErr := in.Close()
	if err != nil {
		return errors.Append(err, closeErr)
	}
	return os.Remove(src)
}
