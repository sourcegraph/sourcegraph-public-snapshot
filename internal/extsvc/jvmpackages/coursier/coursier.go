package coursier

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/schema"
)

var CoursierBinary = "coursier"

func ListArtifactIDs(ctx context.Context, config *schema.JVMPackagesConnection, groupID string) ([]string, error) {
	return runCoursierCommand(ctx, config, "complete", groupID+":")
}

func ListVersions(ctx context.Context, config *schema.JVMPackagesConnection, groupID, artifactID string) ([]string, error) {
	return runCoursierCommand(ctx, config, "complete", groupID+":"+artifactID+":")
}

func FetchSources(ctx context.Context, config *schema.JVMPackagesConnection, dependency reposource.MavenDependency) ([]string, error) {
	if dependency.IsJdk() {
		output, err := runCoursierCommand(
			ctx,
			config,
			"java-home", "--jvm",
			dependency.Version,
		)
		if err != nil {
			return nil, err
		}
		for _, outputPath := range output {
			for _, srcPath := range []string{
				path.Join(outputPath, "src.zip"),
				path.Join(outputPath, "lib", "src.zip"),
			} {
				stat, err := os.Stat(srcPath)
				if !os.IsNotExist(err) && stat.Mode().IsRegular() {
					return []string{srcPath}, nil
				}
			}
		}
		return nil, errors.Errorf("failed to find src.zip for JVM dependency %s", dependency)
	}
	return runCoursierCommand(
		ctx,
		config,
		"fetch", "--intransitive",
		dependency.CoursierSyntax(),
		"--classifier", "sources",
	)
}

func FetchByteCode(ctx context.Context, config *schema.JVMPackagesConnection, dependency reposource.MavenDependency) ([]string, error) {
	return runCoursierCommand(
		ctx,
		config,
		"fetch", "--intransitive",
		dependency.CoursierSyntax(),
	)
}

func Exists(ctx context.Context, config *schema.JVMPackagesConnection, dependency reposource.MavenDependency) (bool, error) {
	if dependency.IsJdk() {
		lines, err := runCoursierCommand(
			ctx,
			config,
			"java-home",
			"--jvm",
			dependency.Version,
		)
		return len(lines) > 0, err
	}
	versions, err := runCoursierCommand(
		ctx,
		config,
		"complete",
		dependency.CoursierSyntax(),
	)
	return len(versions) > 0, err
}

func runCoursierCommand(ctx context.Context, config *schema.JVMPackagesConnection, args ...string) ([]string, error) {
	cmd := exec.CommandContext(ctx, CoursierBinary, args...)
	if config.Maven.Credentials != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("COURSIER_CREDENTIALS=%v", config.Maven.Credentials))
	}
	if len(config.Maven.Repositories) > 0 {
		cmd.Env = append(
			cmd.Env,
			fmt.Sprintf("COURSIER_REPOSITORIES=%v", strings.Join(config.Maven.Repositories, "|")),
		)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "coursier command %q failed with stderr %q and stdout %q", cmd, stderr, &stdout)
	}

	return strings.Split(strings.TrimSpace(stdout.String()), "\n"), nil
}
