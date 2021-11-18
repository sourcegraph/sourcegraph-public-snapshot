package server

import (
	"context"
	"github.com/cockroachdb/errors"
	"os/exec"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
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
