package campaigns

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
)

type WorkspaceCreator struct {
	dir    string
	client api.Client

	deleteZips bool
}

func (wc *WorkspaceCreator) Create(ctx context.Context, repo *graphql.Repository) (string, error) {
	path := localRepositoryZipArchivePath(wc.dir, repo)

	exists, err := fileExists(path)
	if err != nil {
		return "", err
	}

	if !exists {
		err = fetchRepositoryArchive(ctx, wc.client, repo, path)
		if err != nil {
			// If the context got cancelled, or we ran out of disk space, or
			// ... while we were downloading the file, we remove the partially
			// downloaded file.
			os.Remove(path)

			return "", errors.Wrap(err, "fetching ZIP archive")
		}
	}

	if wc.deleteZips {
		defer os.Remove(path)
	}

	prefix := "workspace-" + repo.Slug()
	workspace, err := unzipToTempDir(ctx, path, wc.dir, prefix)
	if err != nil {
		return "", errors.Wrap(err, "unzipping the ZIP archive")
	}

	return workspace, nil
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

func unzipToTempDir(ctx context.Context, zipFile, tempDir, tempFilePrefix string) (string, error) {
	volumeDir, err := ioutil.TempDir(tempDir, tempFilePrefix)
	if err != nil {
		return "", err
	}

	if err := os.Chmod(volumeDir, 0777); err != nil {
		return "", err
	}

	return volumeDir, unzip(zipFile, volumeDir)
}

func fetchRepositoryArchive(ctx context.Context, client api.Client, repo *graphql.Repository, dest string) error {
	req, err := client.NewHTTPRequest(ctx, "GET", repositoryZipArchivePath(repo), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/zip")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to fetch archive (HTTP %d from %s)", resp.StatusCode, req.URL.String())
	}

	// Unlike the mkdirAll() calls elsewhere in this file, this is only giving
	// us a temporary place on the filesystem to keep the archive. Since it's
	// never mounted into the containers being run, we can keep these
	// directories 0700 without issue.
	if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	return nil
}

func repositoryZipArchivePath(repo *graphql.Repository) string {
	return path.Join("", repo.Name+"@"+repo.DefaultBranch.Name, "-", "raw")
}

func localRepositoryZipArchivePath(dir string, repo *graphql.Repository) string {
	ref := repo.DefaultBranch.Target.OID
	return filepath.Join(dir, fmt.Sprintf("%s-%s.zip", repo.Slug(), ref))
}

func unzip(zipFile, dest string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	outputBase := filepath.Clean(dest) + string(os.PathSeparator)

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: https://snyk.io/research/zip-slip-vulnerability#go
		if !strings.HasPrefix(fpath, outputBase) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			if err := mkdirAll(dest, f.Name, 0777); err != nil {
				return err
			}
			continue
		}

		if err := mkdirAll(dest, filepath.Dir(f.Name), 0777); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		// Since the container might not run as the same user, we need to ensure
		// that the file is globally writable. If the execute bit is normally
		// set on the zipped up file, let's ensure we propagate that to the
		// group and other permission bits too.
		if f.Mode()&0111 != 0 {
			outFile.Chmod(0777)
		} else {
			outFile.Chmod(0666)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		cerr := outFile.Close()
		// Now we have safely closed everything that needs it, and can check errors
		if err != nil {
			return errors.Wrapf(err, "copying %q failed", f.Name)
		}
		if cerr != nil {
			return errors.Wrap(err, "closing output file failed")
		}

	}

	return nil
}

// Technically, this is a misnomer, since it might be a socket or block special,
// but errPathExistsAsNonDir is just ugly for an internal type.
type errPathExistsAsFile string

var _ error = errPathExistsAsFile("")

func (e errPathExistsAsFile) Error() string {
	return fmt.Sprintf("path already exists, but not as a directory: %s", string(e))
}

// mkdirAll is essentially os.MkdirAll(filepath.Join(base, path), perm), but
// applies the given permission regardless of the user's umask.
func mkdirAll(base, path string, perm os.FileMode) error {
	abs := filepath.Join(base, path)

	// Create the directory if it doesn't exist.
	st, err := os.Stat(abs)
	if err != nil {
		// It's expected that we'll get an error if the directory doesn't exist,
		// so let's check that it's of the type we expect.
		if !os.IsNotExist(err) {
			return err
		}

		// Now we're clear to create the directory.
		if err := os.MkdirAll(abs, perm); err != nil {
			return err
		}
	} else if !st.IsDir() {
		// The file/socket/whatever exists, but it's not a directory. That's
		// definitely going to be an issue.
		return errPathExistsAsFile(abs)
	}

	// If os.MkdirAll() was invoked earlier, then the permissions it set were
	// subject to the umask. Let's walk the directories we may or may not have
	// created and ensure their permissions look how we want.
	return ensureAll(base, path, perm)
}

// ensureAll ensures that all directories under path have the expected
// permissions.
func ensureAll(base, path string, perm os.FileMode) error {
	var errs *multierror.Error

	// In plain English: for each directory in the path parameter, we should
	// chmod that path to the permissions that are expected.
	acc := []string{base}
	for _, element := range strings.Split(path, string(os.PathSeparator)) {
		acc = append(acc, element)
		if err := os.Chmod(filepath.Join(acc...), perm); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}
