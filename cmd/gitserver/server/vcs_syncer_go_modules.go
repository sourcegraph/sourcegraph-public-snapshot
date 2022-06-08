package server

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"

	"golang.org/x/mod/module"
	modzip "golang.org/x/mod/zip"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gomodproxy"
	"github.com/sourcegraph/sourcegraph/internal/unpack"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/schema"
)

func NewGoModulesSyncer(
	connection *schema.GoModulesConnection,
	svc *dependencies.Service,
	client *gomodproxy.Client,
) VCSSyncer {
	placeholder, err := reposource.ParseGoDependency("sourcegraph.com/placeholder@v0.0.0")
	if err != nil {
		panic(fmt.Sprintf("expected placeholder dependency to parse but got %v", err))
	}

	return &vcsDependenciesSyncer{
		logger:      log.Scoped("vcs syncer", "vcsDependenciesSyncer implements the VCSSyncer interface for dependency repos"),
		typ:         "go_modules",
		scheme:      dependencies.GoModulesScheme,
		placeholder: placeholder,
		svc:         svc,
		configDeps:  connection.Dependencies,
		source:      &goModulesSyncer{client: client},
	}
}

type goModulesSyncer struct {
	client *gomodproxy.Client
}

func (goModulesSyncer) ParseDependency(dep string) (reposource.PackageDependency, error) {
	return reposource.ParseGoDependency(dep)
}

func (goModulesSyncer) ParseDependencyFromRepoName(repoName string) (reposource.PackageDependency, error) {
	return reposource.ParseGoDependencyFromRepoName(repoName)
}

func (s *goModulesSyncer) Get(ctx context.Context, name, version string) (reposource.PackageDependency, error) {
	mod, err := s.client.GetVersion(ctx, name, version)
	if err != nil {
		return nil, err
	}
	return reposource.NewGoDependency(*mod), nil
}

func (s *goModulesSyncer) Download(ctx context.Context, dir string, dep reposource.PackageDependency) error {
	zipBytes, err := s.client.GetZip(ctx, dep.PackageSyntax(), dep.PackageVersion())
	if err != nil {
		return errors.Wrap(err, "get zip")
	}

	mod := dep.(*reposource.GoDependency).Module
	if err = unzip(mod, zipBytes, dir); err != nil {
		return errors.Wrap(err, "failed to unzip go module")
	}

	return nil
}

// unzip the given go module zip into workDir, skipping any files that aren't
// valid according to modzip.CheckZip or that are potentially malicious.
func unzip(mod module.Version, zipBytes []byte, workDir string) error {
	zipFile := path.Join(workDir, "mod.zip")
	err := os.WriteFile(zipFile, zipBytes, 0666)
	if err != nil {
		return errors.Wrapf(err, "failed to create go module zip file %q", zipFile)
	}

	files, err := modzip.CheckZip(mod, zipFile)
	if err != nil && len(files.Valid) == 0 {
		return errors.Wrapf(err, "failed to check go module zip %q", zipFile)
	}

	if err = os.RemoveAll(zipFile); err != nil {
		return errors.Wrapf(err, "failed to remove module zip file %q", zipFile)
	}

	if len(files.Valid) == 0 {
		return nil
	}

	valid := make(map[string]struct{}, len(files.Valid))
	for _, f := range files.Valid {
		valid[f] = struct{}{}
	}

	br := bytes.NewReader(zipBytes)
	err = unpack.Zip(br, int64(br.Len()), workDir, unpack.Opts{
		SkipInvalid: true,
		Filter: func(path string, file fs.FileInfo) bool {
			_, malicious := isPotentiallyMaliciousFilepathInArchive(path, workDir)
			_, ok := valid[path]
			return ok && !malicious
		},
	})

	if err != nil {
		return err
	}

	// All files in module zips are prefixed by prefix below, but we don't want
	// those nested directories in our actual repository, so we move all the files up
	// with the below renames.
	tmpDir := workDir + ".tmp"

	// mv $workDir $tmpDir
	err = os.Rename(workDir, tmpDir)
	if err != nil {
		return err
	}

	// mv $tmpDir/$(basename $prefix) $workDir
	prefix := fmt.Sprintf("%s@%s/", mod.Path, mod.Version)
	return os.Rename(path.Join(tmpDir, prefix), workDir)
}
