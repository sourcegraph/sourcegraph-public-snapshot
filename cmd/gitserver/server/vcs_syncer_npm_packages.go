package server

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npmpackages/npm"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	placeholderNPMDependency = reposource.NPMDependency{
		NPMPackage: func() reposource.NPMPackage {
			pkg, err := reposource.NewNPMPackage("sourcegraph", "placeholder")
			if err != nil {
				panic(fmt.Sprintf("expected placeholder package to parse but got %v", err))
			}
			return *pkg
		}(),
		Version: "1.0.0",
	}
)

type NPMPackagesSyncer struct {
	Config  *schema.NPMPackagesConnection
	DBStore repos.NPMPackagesRepoStore
}

var _ VCSSyncer = &NPMPackagesSyncer{}

func (s *NPMPackagesSyncer) Type() string {
	return "npm_packages"
}

// IsCloneable checks to see if the VCS remote URL is cloneable. Any non-nil
// error indicates there is a problem.
func (s *NPMPackagesSyncer) IsCloneable(ctx context.Context, remoteURL *vcs.URL) error {
	dependencies, err := s.packageDependencies(ctx, remoteURL.Path)
	if err != nil {
		return err
	}

	for _, dependency := range dependencies {
		if err := npm.Exists(ctx, s.Config, dependency); err != nil {
			return err
		}
	}
	return nil
}

// Similar to CloneCommand for JVMPackagesSyncer; it handles cloning itself
// instead of returning a command that does the cloning.
func (s *NPMPackagesSyncer) CloneCommand(ctx context.Context, remoteURL *vcs.URL, bareGitDirectory string) (*exec.Cmd, error) {
	err := os.MkdirAll(bareGitDirectory, 0755)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "git", "--bare", "init")
	if _, err := runCommandInDirectory(ctx, cmd, bareGitDirectory, placeholderNPMDependency); err != nil {
		return nil, err
	}

	// The Fetch method is responsible for cleaning up temporary directories.
	if err := s.Fetch(ctx, remoteURL, GitDir(bareGitDirectory)); err != nil {
		return nil, errors.Wrapf(err, "failed to fetch repo for %s", remoteURL)
	}

	// no-op command to satisfy VCSSyncer interface, see docstring for more details.
	return exec.CommandContext(ctx, "git", "--version"), nil
}

// Fetch adds git tags for newly added dependency versions and removes git tags
// for deleted versions.
func (s *NPMPackagesSyncer) Fetch(ctx context.Context, remoteURL *vcs.URL, dir GitDir) error {
	dependencies, err := s.packageDependencies(ctx, remoteURL.Path)
	if err != nil {
		return err
	}

	out, err := runCommandInDirectory(ctx, exec.CommandContext(ctx, "git", "tag"), string(dir), placeholderNPMDependency)
	if err != nil {
		return err
	}

	tags := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		if len(line) == 0 {
			continue
		}
		tags[line] = true
	}

	for i, dependency := range dependencies {
		if tags[dependency.GitTagFromVersion()] {
			continue
		}
		// the gitPushDependencyTag method is responsible for cleaning up temporary directories.
		if err := s.gitPushDependencyTag(ctx, string(dir), dependency, i == 0); err != nil {
			return errors.Wrapf(err, "error pushing dependency %q", dependency.PackageManagerSyntax())
		}
	}

	dependencyTags := make(map[string]struct{}, len(dependencies))
	for _, dependency := range dependencies {
		dependencyTags[dependency.GitTagFromVersion()] = struct{}{}
	}

	for tag := range tags {
		if _, isDependencyTag := dependencyTags[tag]; !isDependencyTag {
			cmd := exec.CommandContext(ctx, "git", "tag", "-d", tag)
			if _, err := runCommandInDirectory(ctx, cmd, string(dir), placeholderNPMDependency); err != nil {
				log15.Error("Failed to delete git tag", "error", err, "tag", tag)
				continue
			}
		}
	}

	return nil
}

// RemoteShowCommand returns the command to be executed for showing remote.
func (s *NPMPackagesSyncer) RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error) {
	return exec.CommandContext(ctx, "git", "remote", "show", "./"), nil
}

// packageDependencies returns the list of NPM dependencies that belong to the
// given URL path. The returned package dependencies are sorted in descending
// semver order (newest first).
//
// For example, if the URL path represents pkg@1, and our configuration has
// [otherPkg@1, pkg@2, pkg@3], we will return [pkg@3, pkg@2].
func (s *NPMPackagesSyncer) packageDependencies(ctx context.Context, repoUrlPath string) (matchingDependencies []reposource.NPMDependency, err error) {
	repoPackage, err := reposource.ParseNPMPackageFromRepoURL(repoUrlPath)
	if err != nil {
		return nil, err
	}

	var (
		timedout []reposource.NPMDependency
	)
	for _, configDependencyString := range s.npmDependencies() {
		if repoPackage.MatchesDependencyString(configDependencyString) {
			depPtr, err := reposource.ParseNPMDependency(configDependencyString)
			if err != nil {
				return nil, err
			}
			configDependency := *depPtr

			if err := npm.Exists(ctx, s.Config, configDependency); err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					timedout = append(timedout, configDependency)
					continue
				} else {
					return nil, err
				}
			}
			matchingDependencies = append(matchingDependencies, configDependency)
			// Silently ignore non-existent dependencies because
			// they are already logged out in the `GetRepo` method
			// in internal/repos/jvm_packages.go.
		}
	}
	if len(timedout) > 0 {
		log15.Warn("non-zero number of timed-out npm invocations", "count", len(timedout), "dependencies", timedout)
	}
	var totalConfigMatched = len(matchingDependencies)

	parsedPackage, err := reposource.ParseNPMPackageFromRepoURL(repoUrlPath)
	if err != nil {
		return nil, err
	}
	dbDeps, err := s.DBStore.GetNPMDependencyRepos(ctx, dbstore.GetNPMDependencyReposOpts{
		ArtifactName: parsedPackage.PackageSyntax(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get npm dependencies from dbStore")
	}

	for _, dbDep := range dbDeps {
		parsedDbPackage, err := reposource.ParseNPMPackageFromPackageSyntax(dbDep.Package)
		if err != nil {
			log15.Error("failed to parse npm package", "package", dbDep.Package, "message", err)
			continue
		}
		if *repoPackage == *parsedDbPackage {
			matchingDependencies = append(matchingDependencies, reposource.NPMDependency{
				NPMPackage: *parsedDbPackage,
				Version:    dbDep.Version,
			})
		}
	}
	var totalDBMatched = len(matchingDependencies) - totalConfigMatched

	if len(matchingDependencies) == 0 {
		return nil, errors.Errorf("no NPM dependencies for URL path %s", repoUrlPath)
	}

	log15.Info("fetched npm artifact for repo path", "repoPath", repoUrlPath,
		"totalDB", totalDBMatched, "totalConfig", totalConfigMatched)
	reposource.SortNPMDependencies(matchingDependencies)
	return matchingDependencies, nil
}

func (s *NPMPackagesSyncer) npmDependencies() []string {
	if s.Config == nil || s.Config.Dependencies == nil {
		return nil
	}
	return s.Config.Dependencies
}

// gitPushDependencyTag pushes a git tag to the given bareGitDirectory path. The
// tag points to a commit that adds all sources of given dependency. When
// isLatestVersion is true, the HEAD of the bare git directory will also be
// updated to point to the same commit as the git tag.
func (s *NPMPackagesSyncer) gitPushDependencyTag(ctx context.Context, bareGitDirectory string, dependency reposource.NPMDependency, isLatestVersion bool) error {
	tmpDirectory, err := os.MkdirTemp("", "npm-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDirectory)

	sourceCodePath, err := npm.FetchSources(ctx, s.Config, dependency)
	if err != nil {
		return err
	}
	defer os.Remove(sourceCodePath)

	cmd := exec.CommandContext(ctx, "git", "init")
	if _, err := runCommandInDirectory(ctx, cmd, tmpDirectory, dependency); err != nil {
		return err
	}

	err = s.commitTgz(ctx, dependency, tmpDirectory, sourceCodePath, s.Config)
	if err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "git", "remote", "add", "origin", bareGitDirectory)
	if _, err := runCommandInDirectory(ctx, cmd, tmpDirectory, dependency); err != nil {
		return err
	}

	// Use --no-verify for security reasons. See https://github.com/sourcegraph/sourcegraph/pull/23399
	cmd = exec.CommandContext(ctx, "git", "push", "--no-verify", "--force", "origin", "--tags")
	if _, err := runCommandInDirectory(ctx, cmd, tmpDirectory, dependency); err != nil {
		return err
	}

	if isLatestVersion {
		defaultBranch, err := runCommandInDirectory(ctx, exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD"), tmpDirectory, dependency)
		if err != nil {
			return err
		}
		// Use --no-verify for security reasons. See https://github.com/sourcegraph/sourcegraph/pull/23399
		cmd = exec.CommandContext(ctx, "git", "push", "--no-verify", "--force", "origin", strings.TrimSpace(defaultBranch)+":latest", dependency.GitTagFromVersion())
		if _, err := runCommandInDirectory(ctx, cmd, tmpDirectory, dependency); err != nil {
			return err
		}
	}

	return nil
}

// commitTgz creates a git commit in the given working directory that adds all
// the file contents of the given tgz file.
func (s *NPMPackagesSyncer) commitTgz(ctx context.Context, dependency reposource.NPMDependency,
	workingDirectory, sourceCodeTgzPath string, connection *schema.NPMPackagesConnection) error {
	if err := decompressTgz(sourceCodeTgzPath, workingDirectory); err != nil {
		return errors.Wrapf(err, "failed to decompress gzipped tarball for %s to %v", dependency.PackageManagerSyntax(), sourceCodeTgzPath)
	}

	// See [NOTE: LSIF-config-json] for why we don't create a JSON file here
	// like we do for Java.

	cmd := exec.CommandContext(ctx, "git", "add", ".")
	if _, err := runCommandInDirectory(ctx, cmd, workingDirectory, dependency); err != nil {
		return err
	}

	// Use --no-verify for security reasons. See https://github.com/sourcegraph/sourcegraph/pull/23399
	cmd = exec.CommandContext(ctx, "git", "commit", "--no-verify",
		"-m", dependency.PackageManagerSyntax(), "--date", stableGitCommitDate)
	if _, err := runCommandInDirectory(ctx, cmd, workingDirectory, dependency); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "git", "tag",
		"-m", dependency.PackageManagerSyntax(), dependency.GitTagFromVersion())
	if _, err := runCommandInDirectory(ctx, cmd, workingDirectory, dependency); err != nil {
		return err
	}

	return nil
}

// withTgz is a helper function to handling IO-related actions
// so that the action argument can focus on reading the tarball.
func withTgz(tgzPath string, action func(*tar.Reader) error) (err error) {
	ioReader, err := os.Open(tgzPath)
	errMsg := "unable to decompress tgz file with package source"
	if err != nil {
		return errors.Wrap(err, errMsg)
	}
	gzipReader, err := gzip.NewReader(ioReader)
	defer func() {
		errClose := gzipReader.Close()
		if err != nil {
			err = errClose
		}
	}()
	if err != nil {
		return errors.Wrap(err, errMsg)
	}
	tarReader := tar.NewReader(gzipReader)

	return action(tarReader)
}

// Decompress a tarball at tgzPath, putting the files under destination.
//
// Additionally, if all the files in the tarball have paths of the form
// dir/<blah> for the same directory 'dir', the 'dir' will be stripped.
func decompressTgz(tgzPath, destination string) (err error) {
	// [NOTE: npm-strip-outermost-directory]
	// In practice, NPM tarballs seem to contain a superfluous directory which
	// contains the files. For example, if you extract react's tarball,
	// all files will be under a package/ directory, and if you extract
	// @types/lodash's files, all files are under lodash/.
	//
	// However, this additional directory has no meaning. Moreover, it makes
	// the UX slightly worse, as when you navigate to a repo, you would see
	// that it contains just 1 folder, and you'd need to click again to drill
	// down further. So we strip the superfluous directory if we detect one.
	//
	// https://github.com/sourcegraph/sourcegraph/pull/28057#issuecomment-987890718
	var superfluousDirName *string = nil
	commonSuperfluousDirectory := true
	err = withTgz(tgzPath, func(reader *tar.Reader) (err error) {
		for {
			header, err := reader.Next()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return errors.Wrapf(err, "failed to read tar file %s", tgzPath)
			}
			switch header.Typeflag {
			case tar.TypeReg:
				components := strings.SplitN(header.Name, string(os.PathSeparator), 2)
				if len(components) != 2 {
					commonSuperfluousDirectory = false
					return nil
				}
				outermostDir := components[0]
				if superfluousDirName == nil {
					superfluousDirName = &outermostDir
				} else if *superfluousDirName != outermostDir {
					commonSuperfluousDirectory = false
					return nil
				}
			default:
				continue
			}
		}
	})
	if err != nil {
		return err
	}
	if !commonSuperfluousDirectory {
		log15.Warn("found npm tarball which doesn't have all files in one top-level directory", "tarball", path.Base(tgzPath))
	}

	return withTgz(tgzPath, func(tarReader *tar.Reader) (err error) {
		destinationDir := strings.TrimSuffix(destination, string(os.PathSeparator)) + string(os.PathSeparator)
		count := 0
		tarballFileLimit := 10000
		for count < tarballFileLimit {
			header, err := tarReader.Next()
			if err == io.EOF {
				return nil
			}
			name := header.Name
			if commonSuperfluousDirectory {
				name = strings.SplitN(name, string(os.PathSeparator), 2)[1]
			}
			cleanedOutputPath, isPotentiallyMalicious := isPotentiallyMaliciousFilepathInArchive(name, destinationDir)
			if isPotentiallyMalicious {
				continue
			}
			switch header.Typeflag {
			case tar.TypeDir:
				continue // We will create directories later; empty directories don't matter for git.
			case tar.TypeReg:
				err = copyTarFileEntry(header, tarReader, cleanedOutputPath)
				if err != nil {
					return err
				}
				count++
			default:
				return errors.Errorf("unrecognized type of header %+v in tarball %+v", header.Typeflag, path.Base(tgzPath))
			}
		}
		return errors.Errorf("number of files in tarball %s exceeded limit (10000)", path.Base(tgzPath))
	})
}

func copyTarFileEntry(header *tar.Header, tarReader *tar.Reader, outputPath string) (err error) {
	if header.Size < 0 {
		return errors.Errorf("corrupt tar header with negative size %d bytes for %s",
			header.Size, path.Base(outputPath))
	}
	// For reference, "pathological" code like SQLite's amalgamation file is
	// about 7.9 MiB. So a 15 MiB limit seems good enough.
	const sizeLimitMiB = 15
	if header.Size >= (sizeLimitMiB * 1024 * 1024) {
		return errors.Errorf("file size for %s (%d bytes) exceeded limit (%d MiB)",
			path.Base(outputPath), header.Size, sizeLimitMiB)
	}
	if err = os.MkdirAll(path.Dir(outputPath), 0700); err != nil {
		return err
	}
	outputFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		errClose := outputFile.Close()
		if err != nil {
			err = errClose
		}
	}()
	_, err = io.CopyN(outputFile, tarReader, header.Size)
	return err
}
