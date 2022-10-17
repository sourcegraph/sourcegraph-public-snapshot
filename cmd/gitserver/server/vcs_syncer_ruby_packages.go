package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/rubygems"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/unpack"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func NewRubyPackagesSyncer(
	connection *schema.RubyPackagesConnection,
	svc *dependencies.Service,
	client *rubygems.Client,
) VCSSyncer {

	repositoryURL := connection.Repository
	if repositoryURL == "" {
		repositoryURL = "https://rubygems.org/"
	}
	return &vcsPackagesSyncer{
		logger:      log.Scoped("RubyPackagesSyncer", "sync Ruby packages"),
		typ:         "ruby_packages",
		scheme:      dependencies.RubyPackagesScheme,
		placeholder: reposource.NewRubyVersionedPackage("shopify_api", "12.0.0"),
		svc:         svc,
		configDeps:  connection.Dependencies,
		source:      &rubyDependencySource{repositoryURL: repositoryURL, client: client},
	}
}

type rubyDependencySource struct {
	repositoryURL string
	client        *rubygems.Client
}

func (rubyDependencySource) ParseVersionedPackageFromNameAndVersion(name reposource.PackageName, version string) (reposource.VersionedPackage, error) {
	return reposource.ParseRubyVersionedPackage(string(name) + "@" + version)
}

func (rubyDependencySource) ParseVersionedPackageFromConfiguration(dep string) (reposource.VersionedPackage, error) {
	return reposource.ParseRubyVersionedPackage(dep)
}

func (rubyDependencySource) ParsePackageFromName(name reposource.PackageName) (reposource.Package, error) {
	return reposource.ParseRubyPackageFromName(name)

}
func (rubyDependencySource) ParsePackageFromRepoName(repoName api.RepoName) (reposource.Package, error) {
	return reposource.ParseRubyPackageFromRepoName(repoName)
}

func (s *rubyDependencySource) Download(ctx context.Context, dir string, dep reposource.VersionedPackage) error {
	packageURL := fmt.Sprintf("%s/gems/%s-%s.gem", strings.TrimSuffix(s.repositoryURL, "/"), dep.PackageSyntax(), dep.PackageVersion())

	pkg, err := s.client.Get(ctx, packageURL)
	if err != nil {
		return errors.Wrapf(err, "error downloading RubyGem with URL '%s'", packageURL)
	}

	if err = unpackRubyPackage(packageURL, pkg, dir); err != nil {
		return errors.Wrapf(err, "failed to unzip ruby module from URL %s", packageURL)
	}

	return nil
}

func unpackRubyPackage(packageURL string, pkg []byte, workDir string) error {
	r := bytes.NewReader(pkg)
	opts := unpack.Opts{
		SkipInvalid:    true,
		SkipDuplicates: true,
		Filter: func(path string, file fs.FileInfo) bool {
			return path == "data.tar.gz" || path == "metadata.gz"
		},
	}

	tmpDir, err := os.MkdirTemp("", "rubygems")
	if err != nil {
		return errors.Wrap(err, "failed to create a temporary directory")
	}
	defer os.RemoveAll(tmpDir)

	if err := unpack.Tar(r, tmpDir, opts); err != nil {
		return errors.Wrapf(err, "failed to tar downloaded bytes from URL %s", packageURL)
	}

	err = unpackRubyDataTarGz(packageURL, filepath.Join(tmpDir, "data.tar.gz"), workDir)
	if err != nil {
		return err
	}
	metadata, err := os.ReadFile(filepath.Join(tmpDir, "metadata.gz"))
	if err != nil {
		return err
	}
	metadataReader, err := gzip.NewReader(bytes.NewReader(metadata))
	if err != nil {
		return err
	}
	metadataBytes, err := ioutil.ReadAll(metadataReader)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(workDir, "rubygems-metadata.yml"), metadataBytes, 0644)
}

// unpackRubyDataTarGz unpacks the given `data.tar.gz` from a downloaded RubyGem.
func unpackRubyDataTarGz(packageURL, path string, workDir string) error {
	r, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "failed to read file from downloaded URL %s", packageURL)
	}
	defer r.Close()
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
