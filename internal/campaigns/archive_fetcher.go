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
		if err := fetchRepositoryArchive(ctx, wc.client, repo, path); err != nil {
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
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
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
