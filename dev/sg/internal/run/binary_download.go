package run

import (
	"bytes"
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
	urlTmpl, err := template.New("url-tmpl").Parse(bin.URL)
	if err != nil {
		return err
	}

	var renderedUrl bytes.Buffer
	err = urlTmpl.Execute(&renderedUrl, map[string]string{
		"GOARCH":        runtime.GOARCH,
		"GOOS_WITH_MAC": strings.ReplaceAll(runtime.GOOS, "darwin", "mac"),
	})
	if err != nil {
		return err
	}

	resp, err := http.Get(renderedUrl.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Newf("not okay man")
	}

	tmpDirName, err := os.MkdirTemp("", "sg-binary-download*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmpDirName) }()

	if err := unpack.Tgz(resp.Body, tmpDirName, unpack.Opts{}); err != nil {
		return errors.Wrap(err, "unpacking failed")
	}

	fileInArchive := filepath.Join(tmpDirName, bin.FileInArchive)
	if ok, err := fileExists(fileInArchive); !ok || err != nil {
		return errors.Newf("expected %s to exist in extracted archive at %s, but does not", bin.FileInArchive, tmpDirName)
	}

	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	targetFile := filepath.Join(repoRoot, bin.TargetFile)
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
