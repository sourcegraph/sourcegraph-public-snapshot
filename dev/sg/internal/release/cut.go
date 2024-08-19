package release

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/grafana/regexp"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/internal/execute"
)

func cutReleaseBranch(cctx *cli.Context) error {
	p := std.Out.Pending(output.Styled(output.StylePending, "Checking for GitHub CLI..."))
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		p.Destroy()
		return errors.Wrap(err, "GitHub CLI (https://cli.github.com/) is required for installation")
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Using GitHub CLI at %q", ghPath))

	// Check that main has been pulled to the local branch
	defaultBranch := cctx.String("branch")
	defaultGitRepoBranch := repo.NewGitRepo(defaultBranch, defaultBranch)
	p = std.Out.Pending(output.Styled(output.StylePending, "Checking if the default branch exists locally..."))
	if err := defaultGitRepoBranch.Checkout(cctx.Context); err != nil {
		p.Destroy()
		return errors.Wrapf(err, "checking out %q", defaultBranch)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Default branch exists locally"))
	p = std.Out.Pending(output.Styled(output.StylePending, "Checking if the default branch is up to date with remote ..."))
	if _, err := defaultGitRepoBranch.FetchOrigin(cctx.Context); err != nil {
		p.Destroy()
		return errors.Wrapf(err, "fetching origin for %q", defaultBranch)
	}
	if ok, err := defaultGitRepoBranch.IsOutOfSync(cctx.Context); err != nil {
		p.Destroy()
		return errors.Wrapf(err, "checking if %q branch is up to date with remote", defaultBranch)
	} else if ok {
		p.Destroy()
		return errors.Newf("local branch %q is not up to date with remote", defaultBranch)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Local branch is up to date with remote"))

	// Normalize the version string, to prevent issues where this was given with the wrong convention
	// which requires a full rebuild.
	version := fmt.Sprintf("v%s", strings.TrimPrefix(cctx.String("version"), "v"))
	if !regexp.MustCompile("\\d+\\.\\d+\\.0$").MatchString(version) {
		return errors.Newf("invalid version input %q, must be of the form X.Y.0", version)
	}
	releaseBranch := strings.TrimPrefix(strings.Replace(version, ".0", ".x", 1), "v")

	// Ensure release branch conforms to release process policy
	if !regexp.MustCompile("\\d+\\.\\d+\\.x$").MatchString(releaseBranch) {
		return errors.Newf("invalid branch name %q, must be of the form X.Y.x", releaseBranch)
	}

	// Check that the branch doesn't exist locally
	releaseGitRepoBranch := repo.NewGitRepo(releaseBranch, releaseBranch)
	p = std.Out.Pending(output.Styled(output.StylePending, "Checking if the release branch exists locally ..."))
	if ok, err := releaseGitRepoBranch.HasLocalBranch(cctx.Context); err != nil {
		p.Destroy()
		return errors.Wrapf(err, "checking if %q branch exists locally", releaseBranch)
	} else if ok {
		p.Destroy()
		return errors.Newf("branch %q exists locally", releaseBranch)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Release branch %q does not exist locally", releaseBranch))

	// Check the remote doesn't have the new release branch already
	p = std.Out.Pending(output.Styled(output.StylePending, "Checking if the release branch exists in remote ..."))
	if ok, err := releaseGitRepoBranch.HasRemoteBranch(cctx.Context); err != nil {
		p.Destroy()
		return errors.Wrapf(err, "checking if %q branch exists in remote repo", releaseBranch)
	} else if ok {
		p.Destroy()
		return errors.Newf("release branch %q exists in remote repo", releaseBranch)
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Release branch %q does not exist in remote", releaseBranch))

	// Checkout and push release branch
	err = releaseGitRepoBranch.CheckoutNewBranch(cctx.Context)
	if err != nil {
		return errors.Wrapf(err, "could not checkout release branch %q", releaseBranch)
	}
	_, err = releaseGitRepoBranch.Push(cctx.Context)
	if err != nil {
		return errors.Wrapf(err, "could not push release branch %q", releaseBranch)
	}

	// generate max string const and migration graph
	p = std.Out.Pending(output.Styled(output.StylePending, "Increasing max version number const"))
	err = replaceMaxVersion(cctx.Context, version)
	if err != nil {
		p.Destroy()
		return err
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Max version number increased"))

	p = std.Out.Pending(output.Styled(output.StylePending, "Generating stitched migration graph"))
	err = genStitchMigrationGraph(cctx.Context)
	if err != nil {
		p.Destroy()
		return err
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Stitched migration graph generated"))

	// commit changes
	_, err = releaseGitRepoBranch.Add(cctx.Context, "internal/database/migration/shared/data/cmd/generator/consts.go")
	if err != nil {
		return errors.Wrap(err, "could not add consts.go to staged changes")
	}
	_, err = releaseGitRepoBranch.Add(cctx.Context, "internal/database/migration/shared/data/stitched-migration-graph.json")
	if err != nil {
		return errors.Wrap(err, "could not add stitched-migration-graph.json to staged changes")
	}
	_, err = releaseGitRepoBranch.Commit(cctx.Context, "chore: update max version")
	if err != nil {
		return errors.Wrap(err, "could not commit staged changes")
	}
	_, err = releaseGitRepoBranch.Push(cctx.Context)
	if err != nil {
		return errors.Wrap(err, "could not push staged changes")
	}

	// generate new stitched migration archive
	err = genStitchMigrationArchive(cctx.Context, version)
	if err != nil {
		return err
	}

	// Create backport label
	p = std.Out.Pending(output.Styled(output.StylePending, "Creating backport label..."))
	if _, err := execute.GH(
		cctx.Context,
		"label",
		"create",
		fmt.Sprintf("backport %s", releaseBranch),
		"-d",
		fmt.Sprintf("label used to backport PRs to the %s release branch", releaseBranch),
	); err != nil {
		p.Destroy()
		return errors.Wrap(err, "failed to create backport label")
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Backport label created"))

	return nil
}

func replaceMaxVersion(ctx context.Context, newVersion string) error {
	p := std.Out.Pending(output.Styled(output.StylePending, "Checking for comby CLI..."))
	combyPath, err := exec.LookPath("comby")
	if err != nil {
		p.Destroy()
		return errors.Wrap(err, "Comby (https://comby.dev/) is required for installation")
	}
	p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Using Comby at %q", combyPath))
	noVNewVersion := strings.TrimPrefix(newVersion, "v")
	err = run.Cmd(ctx, "comby", "-in-place", "\"const maxVersionString = :[1]\"", fmt.Sprintf("\"const maxVersionString = \\\"%s\\\"\"", noVNewVersion), "internal/database/migration/shared/data/cmd/generator/consts.go").Run().Wait()
	if err != nil {
		return errors.Wrap(err, "Could not run comby to change maxVersionString")
	}
	return nil
}

func genStitchMigrationGraph(ctx context.Context) error {
	err := run.Cmd(ctx, "sg", "bazel", "run", "//internal/database/migration/shared:write_stitched_migration_graph").Run().Wait()
	if err != nil {
		return errors.Wrap(err, "Could not run stitch migration generator")
	}
	return nil
}

// TODO set timestamp during archive generation for cache
func genStitchMigrationArchive(ctx context.Context, newVersion string) error {
	err := run.Cmd(ctx, "git", "archive", "--format=tar.gz", "HEAD", "migrations", "-o", fmt.Sprintf("migrations-%s.tar.gz", newVersion)).Run().Wait()
	if err != nil {
		return errors.Wrap(err, "Could not create git archive")
	}
	err = run.Cmd(ctx, "CLOUDSDK_CORE_PROJECT=\"sourcegraph-ci\"", "gsutil", "cp", fmt.Sprintf("migrations-%s", newVersion), "gs://schemas-migrations/migrations/").Run().Wait()
	if err != nil {
		return errors.Wrap(err, "Could not push git archive to GCS")
	}
	return nil
}
