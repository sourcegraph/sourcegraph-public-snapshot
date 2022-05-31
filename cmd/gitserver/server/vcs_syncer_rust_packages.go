package server

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/crates"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/unpack"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func assertRustParsesPlaceholder() *reposource.RustDependency {
	placeholder, err := reposource.ParseRustDependency("sourcegraph.com/placeholder@0.0.0")
	if err != nil {
		panic(fmt.Sprintf("expected placeholder dependency to parse but got %v", err))
	}

	return placeholder
}

func NewRustPackagesSyncer(
	connection *schema.RustPackagesConnection,
	svc *dependencies.Service,
	client *crates.Client,
) VCSSyncer {
	placeholder := assertRustParsesPlaceholder()

	return &vcsDependenciesSyncer{
		typ:         "rust_packages",
		scheme:      dependencies.RustPackagesScheme,
		placeholder: placeholder,
		svc:         svc,
		configDeps:  connection.Dependencies,
		source:      &rustDependencySource{client: client},
	}
}

// pythonPackagesSyncer implements dependenciesSource
type rustDependencySource struct {
	client *crates.Client
}

func (rustDependencySource) ParseDependency(dep string) (reposource.PackageDependency, error) {
	return reposource.ParseRustDependency(dep)
}

func (rustDependencySource) ParseDependencyFromRepoName(repoName string) (reposource.PackageDependency, error) {
	return reposource.ParseRustDependencyFromRepoName(repoName)
}

func (s *rustDependencySource) Get(ctx context.Context, name, version string) (reposource.PackageDependency, error) {
	// f, err := s.client.Version(ctx, name, version)
	// if err != nil {
	// 	return nil, err
	// }
	dep := reposource.NewRustDependency(name, version)
	return dep, nil
}

func (s *rustDependencySource) Download(ctx context.Context, dir string, dep reposource.PackageDependency) error {
	// packageURL := dep.(*reposource.RustDependency).PackageURL
	packageURL := fmt.Sprintf("https://crates.io/api/v1/crates/%s/%s/%s", string(dep.PackageSyntax()), dep.PackageVersion(), "download")

	pkg, err := s.client.Download(ctx, packageURL)
	if err != nil {
		return errors.Wrap(err, "download")
	}

	// TODO: we could add `.sourcegraph/repo.json` here with more information,
	// to be used by rust analyzer
	if err = unpackRustPackage(pkg, dir); err != nil {
		return errors.Wrap(err, "failed to unzip rust module")
	}

	return nil
}

// unpackRustPackages unpacks the given rust package archive into workDir, skipping any
// files that aren't valid or that are potentially malicious.
func unpackRustPackage(pkg []byte, workDir string) error {
	r := bytes.NewReader(pkg)
	opts := unpack.Opts{
		SkipInvalid: true,
		Filter: func(path string, file fs.FileInfo) bool {
			size := file.Size()

			const sizeLimit = 15 * 1024 * 1024
			if size >= sizeLimit {
				log15.Warn("skipping large file in cargo package",
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

	if err := unpack.Tgz(r, workDir, opts); err != nil {
		return err
	}

	return stripSingleOutermostDirectory(workDir)
}
