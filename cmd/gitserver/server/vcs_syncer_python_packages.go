package server

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"path"
	"strings"

	"github.com/inconshreveable/log15"

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
	placeholder, err := reposource.ParsePythonDependency("sourcegraph.com/placeholder@v0.0.0")
	if err != nil {
		panic(fmt.Sprintf("expected placeholder dependency to parse but got %v", err))
	}

	return &vcsDependenciesSyncer{
		typ:         "python_packages",
		scheme:      dependencies.PythonPackagesScheme,
		placeholder: placeholder,
		svc:         svc,
		configDeps:  connection.Dependencies,
		source:      &pythonPackagesSyncer{client: client},
	}
}

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
		return errors.Wrap(err, "failed to unzip go module")
	}

	return nil
}

// unpackPythonPackages unpacks the given python package archive into workDir, skipping any
// files that aren't valid or that are potentially malicious. It detects the kind of archive
// and compression used with the given packageURL.
func unpackPythonPackage(pkg []byte, packageURL, workDir string) error {
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
			if size >= sizeLimit {
				log15.Warn("skipping large file in npm package",
					"path", file.Name(),
					"size", size,
					"limit", sizeLimit,
				)
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
