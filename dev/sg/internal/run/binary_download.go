package run

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/unpack"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Download(bin DownloadBinary) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	targetFile := filepath.Join(repoRoot, bin.TargetFile)

	// Skip download if file already exists
	if ok, _ := fileExists(targetFile); ok {
		return nil
	}

	url, err := renderField("url", bin.URL)
	if err != nil {
		return errors.Wrapf(err, "failed to render field %q as template", "url")
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Newf("GET %s failed, status code is not OK: %d", url, resp.StatusCode)
	}
	defer resp.Body.Close()

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
			return path == bin.FileInArchive && !file.IsDir()
		},
	}
	if err := unpack.Tgz(resp.Body, tmpDirName, opts); err != nil {
		return errors.Wrap(err, "unpacking archive failed")
	}

	fileInArchive := filepath.Join(tmpDirName, bin.FileInArchive)
	if ok, err := fileExists(fileInArchive); !ok || err != nil {
		return errors.Newf("expected %s to exist in extracted archive at %s, but does not", bin.FileInArchive, tmpDirName)
	}

	if err := os.Rename(fileInArchive, targetFile); err != nil {
		return err
	}

	return nil
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

func renderField(name, content string) (string, error) {
	t, err := template.New(fmt.Sprintf("field-%s", name)).Parse(content)
	if err != nil {
		return "", err
	}

	var rendered bytes.Buffer
	if err := t.Execute(&rendered, templatedFieldsData); err != nil {
		return "", err
	}

	return rendered.String(), nil
}

var templatedFieldsData = map[string]string{
	"GOARCH": runtime.GOARCH,
	"GOOS":   runtime.GOOS,
	// Some package use "mac" instead of "darwin" but leave the the other OS
	// names the same.
	"GOOS_WITH_MAC": strings.ReplaceAll(runtime.GOOS, "darwin", "mac"),
}
