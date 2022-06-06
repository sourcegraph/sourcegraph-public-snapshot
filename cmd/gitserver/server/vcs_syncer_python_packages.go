package server

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/pypi"
	"github.com/sourcegraph/sourcegraph/internal/unpack"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/schema"
)

func assertPythonParsesPlaceholder() *reposource.PythonDependency {
	placeholder, err := reposource.ParsePythonDependency("sourcegraph.com/placeholder@v0.0.0")
	if err != nil {
		panic(fmt.Sprintf("expected placeholder dependency to parse but got %v", err))
	}

	return placeholder
}

func NewPythonPackagesSyncer(
	connection *schema.PythonPackagesConnection,
	svc *dependencies.Service,
	client *pypi.Client,
) VCSSyncer {
	placeholder := assertPythonParsesPlaceholder()

	return &vcsDependenciesSyncer{
		logger:      log.Scoped("vcs syncer", "vcsDependenciesSyncer implements the VCSSyncer interface for dependency repos"),
		typ:         "python_packages",
		scheme:      dependencies.PythonPackagesScheme,
		placeholder: placeholder,
		svc:         svc,
		configDeps:  connection.Dependencies,
		source:      &pythonPackagesSyncer{client: client},
	}
}

// pythonPackagesSyncer implements dependenciesSource
type pythonPackagesSyncer struct {
	client *pypi.Client
}

func (pythonPackagesSyncer) ParseDependency(dep string) (reposource.PackageDependency, error) {
	return reposource.ParsePythonDependency(dep)
}

func (pythonPackagesSyncer) ParseDependencyFromRepoName(repoName string) (reposource.PackageDependency, error) {
	return reposource.ParsePythonDependencyFromRepoName(repoName)
}

func (s *pythonPackagesSyncer) Get(ctx context.Context, name, version string) (reposource.PackageDependency, error) {
	f, err := s.client.Version(ctx, name, version)
	if err != nil {
		return nil, err
	}
	dep := reposource.NewPythonDependency(name, version)
	dep.PackageURL = f.URL
	return dep, nil
}

func (s *pythonPackagesSyncer) Download(ctx context.Context, dir string, dep reposource.PackageDependency) error {
	packageURL := dep.(*reposource.PythonDependency).PackageURL

	pkg, err := s.client.Download(ctx, packageURL)
	if err != nil {
		return errors.Wrap(err, "download")
	}

	if err = unpackPythonPackage(pkg, packageURL, dir); err != nil {
		return errors.Wrap(err, "failed to unzip python module")
	}

	return nil
}

// unpackPythonPackages unpacks the given python package archive into workDir, skipping any
// files that aren't valid or that are potentially malicious. It detects the kind of archive
// and compression used with the given packageURL.
func unpackPythonPackage(pkg []byte, packageURL, workDir string) error {
	logger := log.Scoped("unpackPythonPackages", "unpackPythonPackages unpacks the given python package archive into workDir")
	u, err := url.Parse(packageURL)
	if err != nil {
		return errors.Wrap(err, "bad python package URL")
	}

	filename := path.Base(u.Path)

	r := bytes.NewReader(pkg)
	opts := unpack.Opts{
		SkipInvalid: true,
		Filter: func(path string, file fs.FileInfo) bool {
			size := file.Size()

			const sizeLimit = 15 * 1024 * 1024
			slogger := logger.With(
				log.String("path", file.Name()),
				log.Int64("size", size),
				log.Float64("limit", sizeLimit),
			)
			if size >= sizeLimit {
				slogger.Warn("skipping large file in npm package")
				return false
			}

			_, malicious := isPotentiallyMaliciousFilepathInArchive(path, workDir)
			return !malicious
		},
	}

	switch {
	case strings.HasSuffix(filename, ".tar.gz"), strings.HasSuffix(filename, ".tgz"):
		err = unpack.Tgz(r, workDir, opts)
	case strings.HasSuffix(filename, ".whl"), strings.HasSuffix(filename, ".zip"):
		err = unpack.Zip(r, int64(len(pkg)), workDir, opts)
	case strings.HasSuffix(filename, ".tar"):
		err = unpack.Tar(r, workDir, opts)
	default:
		return errors.Errorf("unsupported python package type %q", filename)
	}

	if err != nil {
		return err
	}

	return stripSingleOutermostDirectory(workDir)
}
