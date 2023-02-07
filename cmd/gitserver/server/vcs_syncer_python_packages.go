package server

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"net/url"
	"path"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/pypi"
	"github.com/sourcegraph/sourcegraph/internal/unpack"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func NewPythonPackagesSyncer(
	connection *schema.PythonPackagesConnection,
	svc *dependencies.Service,
	client *pypi.Client,
) VCSSyncer {
	return &vcsPackagesSyncer{
		logger:      log.Scoped("PythonPackagesSyncer", "sync Python packages"),
		typ:         "python_packages",
		scheme:      dependencies.PythonPackagesScheme,
		placeholder: reposource.ParseVersionedPackage("sourcegraph.com/placeholder@v0.0.0"),
		svc:         svc,
		configDeps:  connection.Dependencies,
		source:      &pythonPackagesSyncer{client: client},
	}
}

// pythonPackagesSyncer implements packagesSource
type pythonPackagesSyncer struct {
	client *pypi.Client
}

func (pythonPackagesSyncer) ParseVersionedPackageFromNameAndVersion(name reposource.PackageName, version string) (reposource.VersionedPackage, error) {
	return reposource.ParseVersionedPackage(string(name) + "==" + version), nil
}

func (pythonPackagesSyncer) ParseVersionedPackageFromConfiguration(dep string) (reposource.VersionedPackage, error) {
	return reposource.ParseVersionedPackage(dep), nil
}

func (pythonPackagesSyncer) ParsePackageFromName(name reposource.PackageName) (reposource.Package, error) {
	return reposource.ParsePythonPackageFromName(name), nil
}

func (pythonPackagesSyncer) ParsePackageFromRepoName(repoName api.RepoName) (reposource.Package, error) {
	return reposource.ParsePythonPackageFromRepoName(repoName)
}

func (s *pythonPackagesSyncer) Download(ctx context.Context, dir string, dep reposource.VersionedPackage) error {
	pythonDep := dep.(*reposource.PythonVersionedPackage)
	pypiFile, err := s.client.Version(ctx, pythonDep.Name, pythonDep.Version)
	if err != nil {
		return err
	}
	packageURL := pypiFile.URL
	pkgData, err := s.client.Download(ctx, packageURL)
	if err != nil {
		return errors.Wrap(err, "download")
	}
	defer pkgData.Close()

	if err = unpackPythonPackage(pkgData, packageURL, dir); err != nil {
		return errors.Wrap(err, "failed to unzip python module")
	}

	return nil
}

// unpackPythonPackages unpacks the given python package archive into workDir, skipping any
// files that aren't valid or that are potentially malicious. It detects the kind of archive
// and compression used with the given packageURL.
func unpackPythonPackage(pkg io.Reader, packageURL, workDir string) error {
	logger := log.Scoped("unpackPythonPackages", "unpackPythonPackages unpacks the given python package archive into workDir")
	u, err := url.Parse(packageURL)
	if err != nil {
		return errors.Wrap(err, "bad python package URL")
	}

	filename := path.Base(u.Path)

	opts := unpack.Opts{
		SkipInvalid: true,
		Filter: func(path string, file fs.FileInfo) bool {
			size := file.Size()

			const sizeLimit = 15 * 1024 * 1024
			if size >= sizeLimit {
				logger.With(
					log.String("path", file.Name()),
					log.Int64("size", size),
					log.Float64("limit", sizeLimit),
				).Warn("skipping large file in python package")
				return false
			}

			malicious := isPotentiallyMaliciousFilepathInArchive(path, workDir)
			return !malicious
		},
	}

	switch {
	case strings.HasSuffix(filename, ".tar.gz"), strings.HasSuffix(filename, ".tgz"):
		err = unpack.Tgz(pkg, workDir, opts)
	case strings.HasSuffix(filename, ".whl"), strings.HasSuffix(filename, ".zip"):
		var pkgBytes []byte
		pkgBytes, err = io.ReadAll(pkg)
		if err != nil {
			break
		}
		err = unpack.Zip(bytes.NewReader(pkgBytes), int64(len(pkgBytes)), workDir, opts)
	case strings.HasSuffix(filename, ".tar"):
		err = unpack.Tar(pkg, workDir, opts)
	default:
		return errors.Errorf("unsupported python package type %q", filename)
	}

	if err != nil {
		return err
	}

	return stripSingleOutermostDirectory(workDir)
}
