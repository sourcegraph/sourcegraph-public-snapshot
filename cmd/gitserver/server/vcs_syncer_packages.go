package server

import (
	"context"
	"os/exec"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func runCommandInDirectory(ctx context.Context, cmd *exec.Cmd, workingDirectory string, dependency reposource.PackageDependency) (string, error) {
	gitName := dependency.PackageManagerSyntax() + " authors"
	gitEmail := "code-intel@sourcegraph.com"
	cmd.Dir = workingDirectory
	cmd.Env = append(cmd.Env, "EMAIL="+gitEmail)
	cmd.Env = append(cmd.Env, "GIT_AUTHOR_NAME="+gitName)
	cmd.Env = append(cmd.Env, "GIT_AUTHOR_EMAIL="+gitEmail)
	cmd.Env = append(cmd.Env, "GIT_AUTHOR_DATE="+stableGitCommitDate)
	cmd.Env = append(cmd.Env, "GIT_COMMITTER_NAME="+gitName)
	cmd.Env = append(cmd.Env, "GIT_COMMITTER_EMAIL="+gitEmail)
	cmd.Env = append(cmd.Env, "GIT_COMMITTER_DATE="+stableGitCommitDate)
	output, err := runWith(ctx, cmd, false, nil)
	if err != nil {
		return "", errors.Wrapf(err, "command %s failed with output %s", cmd.Args, string(output))
	}
	return string(output), nil
}

func isPotentiallyMaliciousFilepathInArchive(filepath, destinationDir string) (outputPath string, _ bool) {
	if strings.HasPrefix(filepath, ".git/") {
		// For security reasons, don't unzip files under the `.git/`
		// directory. See https://github.com/sourcegraph/security-issues/issues/163
		return "", true
	}
	if strings.HasSuffix(filepath, "/") {
		// Skip directory entries. Directory entries must end
		// with a forward slash (even on Windows) according to
		// `file.Name` docstring.
		return "", true
	}
	if strings.HasPrefix(filepath, "/") {
		// Skip absolute paths. While they are extracted relative to `destination`,
		// they should be unimportant. Related issue https://github.com/golang/go/issues/48085#issuecomment-912659635
		return "", true
	}
	cleanedOutputPath := path.Join(destinationDir, filepath)
	if !strings.HasPrefix(cleanedOutputPath, destinationDir) {
		// For security reasons, skip file if it's not a child
		// of the target directory. See "Zip Slip Vulnerability".
		return "", true
	}
	return cleanedOutputPath, false
}

// [NOTE: LSIF-config-json]
//
// For JVM languages, when we create a fake Git repository from a Maven module
// we also add a lsif-java.json file to the repository. However, we don't create
// an analogous lsif-node.json for JavaScript/TypeScript. Here's why:
//
// 1. A specific JDK version is needed to correctly index the code. This JDK
//    version needs to be specified when launching lsif-java. So if we wanted
//    to determine the JDK version at auto-indexing time (instead of at upload
//    time), we'd need to have a separate tool that ran before lsif-java.
//
//    This doesn't apply to JS/TS because there is no clear source of truth for
//    the version of the runtime (Node etc.) that is needed.
//
// 2. The lsif-java.json file indicates whether the repo contains the sources of
//    the JDK or whether it's a regular Maven artifact. For the JDK, lsif-java
//    has a special case to emit "export" monikers instead of "import".
//
//    This doesn't apply to JS/S because there is no special NPM module
//    analogous to the JDK.
//
// 3. The lsif-java.json file is used as a marker file to enable inference
//    of the auto-indexing configuration for a JVM package repo
//    (e.g. from Maven Central). Since JVM source repos (e.g. from GitHub) lack
//    this file, the auto-indexing configuration is not inferred.
//
//    This doesn't apply to JS/TS because auto-indexing configuration is
//    inferred from the package.json file for both source repos and package
//    repos.
