package server

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/crates"
	"github.com/sourcegraph/sourcegraph/internal/unpack"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func assertRustParsesPlaceholder() *reposource.RustVersionedPackage {
	placeholder, err := reposource.ParseRustVersionedPackage("sourcegraph.com/placeholder@0.0.0")
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

	return &vcsPackagesSyncer{
		logger:      log.Scoped("RustPackagesSyncer", "sync Rust packages"),
		typ:         "rust_packages",
		scheme:      dependencies.RustPackagesScheme,
		placeholder: placeholder,
		svc:         svc,
		configDeps:  connection.Dependencies,
		source:      &rustDependencySource{client: client},
	}
}

type rustDependencySource struct {
	client *crates.Client
}

func (rustDependencySource) ParseVersionedPackageFromNameAndVersion(name reposource.PackageName, version string) (reposource.VersionedPackage, error) {
	return reposource.ParseRustVersionedPackage(string(name) + "@" + version)
}

func (rustDependencySource) ParseVersionedPackageFromConfiguration(dep string) (reposource.VersionedPackage, error) {
	return reposource.ParseRustVersionedPackage(dep)
}

func (rustDependencySource) ParsePackageFromName(name reposource.PackageName) (reposource.Package, error) {
	return reposource.ParseRustPackageFromName(name)

}
func (rustDependencySource) ParsePackageFromRepoName(repoName api.RepoName) (reposource.Package, error) {
	return reposource.ParseRustPackageFromRepoName(repoName)
}

func (s *rustDependencySource) Download(ctx context.Context, dir string, dep reposource.VersionedPackage) error {
	packageURL := fmt.Sprintf("https://static.crates.io/crates/%s/%s-%s.crate", dep.PackageSyntax(), dep.PackageSyntax(), dep.PackageVersion())

	pkg, err := s.client.Get(ctx, packageURL)
	if err != nil {
		return errors.Wrapf(err, "error downloading crate with URL '%s'", packageURL)
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
		SkipInvalid:    true,
		SkipDuplicates: true,
		Filter: func(path string, file fs.FileInfo) bool {
			size := file.Size()

			const sizeLimit = 15 * 1024 * 1024
			if size >= sizeLimit {
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
