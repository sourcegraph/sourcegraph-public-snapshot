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

func Download(bin DownloadArchive, lookupEnv func(string) string) error {
	// Treat fields as templats and render them
	targetFile, err := renderField("targetFile", bin.TargetFile, lookupEnv)
	if err != nil {
		return errors.Wrapf(err, "failed to render field %q as template", "targetFile")
	}

	fileInArchive, err := renderField("fileInArchive", bin.FileInArchive, lookupEnv)
	if err != nil {
		return errors.Wrapf(err, "failed to render field %q as template", "fileInArchive")
	}

	url, err := renderField("url", bin.URL, lookupEnv)
	if err != nil {
		return errors.Wrapf(err, "failed to render field %q as template", "targetFile")
	}

	// Skip download if file already exists
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	targetFilePath := filepath.Join(repoRoot, targetFile)
	if ok, _ := fileExists(targetFilePath); ok {
		return nil
	}

	resp, err := http.Get(url)
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
		// _ = os.Remove(tmpDirName)
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

	fileInArchivePath := filepath.Join(tmpDirName, fileInArchive)
	if ok, err := fileExists(fileInArchivePath); !ok || err != nil {
		return errors.Newf("expected %s to exist in extracted archive at %s, but does not", fileInArchivePath, tmpDirName)
	}

	if err := os.Rename(fileInArchivePath, targetFilePath); err != nil {
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

func renderField(name, content string, lookupEnv func(string) string) (string, error) {
	t, err := template.
		New(fmt.Sprintf("field-%s", name)).
		Funcs(template.FuncMap{"getEnv": lookupEnv}).
		Parse(content)
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
