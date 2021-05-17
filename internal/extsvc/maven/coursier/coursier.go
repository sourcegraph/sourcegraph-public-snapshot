package coursier

import (
	"bytes"
	"context"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func ListAllGroupsForPrefix(ctx context.Context, repository, prefix string) ([]string, error) {
	// TODO: codeintel make this work properly
	return runCoursierCommand(ctx, repository, "complete", prefix)
}

func ListArtifactIDs(ctx context.Context, repository, groupID string) ([]string, error) {
	return runCoursierCommand(ctx, repository, "complete", groupID+":")
}

func ListVersions(ctx context.Context, repository, groupID, artifactID string) ([]string, error) {
	return runCoursierCommand(ctx, repository, "complete", groupID+":"+artifactID+":")
}

// TODO: codeintel, we are manually constructing URLs to pass to 'cs get' rather than just calling
// 'cs fetch --sources groupID:artifactID'. this is because roux couldn't figure out how to get
// coursier to only fetch the sources we're interested rather than pulling down the entire
// dependency tree. the TODO is to figure out:
//   - if the URL construction works universally
//   - if we do actually need to pull down the deps
//   - if we can get fetch to not pull down the deps
func FetchVersions(ctx context.Context, repository, groupID, artifactID string, versions []string) ([]string, error) {
	var urls []string
	for _, version := range versions {
		filename := strings.Join([]string{artifactID, version, "sources.jar"}, "-")
		url := strings.Join([]string{
			repository,
			strings.ReplaceAll(groupID, ".", "/"),
			artifactID,
			version,
			filename,
		}, "/")
		urls = append(urls, url)
	}

	return runCoursierCommand(ctx, repository, append([]string{"get"}, urls...)...)
}

func Exists(ctx context.Context, repository, groupID, artifactID string) (bool, error) {
	versions, err := ListVersions(ctx, repository, groupID, artifactID)
	return len(versions) > 0, err
}

func runCoursierCommand(ctx context.Context, repository string, args ...string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "coursier", append([]string{"-r", repository}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, errors.Wrap(err, stderr.String())
	}

	return strings.Split(string(stdout.String()), "\n"), nil
}
