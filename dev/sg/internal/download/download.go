package download

import (
	"bytes"
	"net/http"
	"os/exec"

	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Exeuctable downloads a binary from the given URL, updates the given path if different, and
// makes the downloaded file executable.
func Exeuctable(url string, path string) error {
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
