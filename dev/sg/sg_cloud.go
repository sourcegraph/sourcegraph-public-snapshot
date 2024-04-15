package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/ci/images"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/cloud"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

var ErrUserCancelled = errors.New("user cancelled")

var cloudCommand = &cli.Command{
	Name:  "cloud",
	Usage: "Install and work with Sourcegraph Cloud tools",
	Description: `Learn more about Sourcegraph Cloud:

- Product: https://sourcegraph.com/docs/cloud
- Handbook: https://handbook.sourcegraph.com/departments/cloud/
`,
	Category: category.Company,
	Subcommands: []*cli.Command{
		{
			Name:        "install",
			Usage:       "Install or upgrade local `mi2` CLI (for Cloud V2)",
			Description: "To learn more about Cloud V2, see https://handbook.sourcegraph.com/departments/cloud/technical-docs/v2.0/",
			Action: func(c *cli.Context) error {
				if err := installCloudCLI(c.Context); err != nil {
					return err
				}
				if err := checkGKEAuthPlugin(); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:        "deploy",
			Usage:       "sg could deploy --branch <branch> --tag <tag>",
			Description: "Deploy the specified branch or tag to an ephemeral Sourcegraph Cloud environment",
			Action:      deployCloudEphemeral,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name: "branch",
				},
				&cli.StringFlag{
					Name: "tag",
				},
			},
		},
	},
}

func determineVersion(build *buildkite.Build, tag string) string {
	return images.BranchImageTag(
		time.Now(),
		pointers.DerefZero(build.Commit),
		pointers.DerefZero(build.Number),
		pointers.DerefZero(build.Branch),
		tag,
	)
}

func ensureBranchIsSyncd(ctx context.Context, currRepo *repo.GitRepo) error {
	if ok, err := currRepo.IsOutOfSync(ctx); err != nil {
		return err
	} else if ok {
		return nil
	}

	var answer string
	oneOf := func(value string, i ...string) bool {
		for _, item := range i {
			if value == item {
				return true
			}
		}
		return false
	}
	ok, err := std.PromptAndScan(std.Out, fmt.Sprintf("Commit %q on branch %q does not exist remotely. Do you want to push it to origin? (yes/no)", currRepo.Ref, currRepo.Branch), &answer)
	if err != nil {
		return err
	}
	if !ok || !oneOf(answer, "yes", "y") {
		return ErrUserCancelled
	}

	std.Out.WriteNoticef("Pushing commit %q to origin/\n", currRepo.Ref, currRepo.Branch)
	if err := currRepo.Sync(ctx); err != nil {
		return err
	}

	// if we pushed we wait a little bit otherwise follow up actions might not trigger properly
	time.Sleep(3 * time.Second)
	return nil
}

func deployCloudEphemeral(ctx *cli.Context) error {
	branch := ctx.String("branch")
	currentBranch, err := repo.GetCurrentBranch(ctx.Context)
	if err != nil {
		return errors.Wrap(err, "failed to determine current branch")
	}
	// if the tag is set - we should prefer it over the branch
	tag := ctx.String("tag")

	if branch == "" && tag == "" {
		branch = currentBranch
	} else if branch != currentBranch {
		// we are not on the intended branch so we create a cloud-ephemeral branch so that we don't interfere with the branch specified
		branch = fmt.Sprintf("cloud-ephemeral/%s", strings.ReplaceAll(branch, "/", "-"))
		std.Out.Writef("currently not on %q branch - using %q as branch\n", currentBranch, branch)
	}

	currRepo, err := repo.NewWithBranch(ctx.Context, branch)
	err = ensureBranchIsSyncd(ctx.Context, currRepo)
	if err != nil {
		return errors.Wrap(err, "failed to ensure current commit can be built")
	}

	// Check that branch has been pushed
	// offer to push branch
	//
	// 1. kick of a build so that we can get the images
	// 2. Once the build is kicked off we will need the build number so taht we can generate the version locally
	std.Out.WriteNoticef("Starting build for %q on commit %q\n", currRepo.Branch, currRepo.Ref)
	client, err := bk.NewClient(ctx.Context, std.Out)
	if err != nil {
		return err
	}
	build, err := client.TriggerBuild(ctx.Context, "sourcegraph", currRepo.Branch, currRepo.Ref, bk.WithEnvVar("CLOUD_EPHEMERAL", "true"))
	if err != nil {
		return err
	}
	std.Out.WriteSuccessf("Started build %d. Build progress can be viewed at %s\n", pointers.DerefZero(build.Number), pointers.DerefZero(build.WebURL))

	version := determineVersion(build, tag)
	std.Out.Writef("Starting cloud ephemeral deployment for version %q\n", version)

	panic("stop")
	cloudClient, err := cloud.NewClient(ctx.Context, cloud.APIEndpoint)
	if err != nil {
		return err
	}

	inst, err := cloudClient.ListInstances(ctx.Context)
	if err != nil {
		return err
	}

	std.Out.Writef("Found %d instances\n", len(inst))
	// 3. Once we have the version we can kick off the cloud deploy so that it can start provisioning the environment

	return nil
}

func checkGKEAuthPlugin() error {
	const executable = "gke-gcloud-auth-plugin"
	existingPath, err := exec.LookPath(executable)
	if err != nil {
		return errors.Wrapf(err, "gke-gcloud-auth-plugin not found on path, run `brew info google-cloud-sdk` for instructions OR \n"+
			"run `gcloud components install gke-gcloud-auth-plugin` to install it manually")
	}
	std.Out.WriteNoticef("Using gcloud auth plugin at %q", existingPath)
	return nil
}

func installCloudCLI(ctx context.Context) error {
	const executable = "mi2"

	// Ensure gh is installed
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return errors.Wrap(err, "GitHub CLI (https://cli.github.com/) is required for installation")
	}
	std.Out.Writef("Using GitHub CLI at %q", ghPath)

	// Use the same directory as sg, since we add that to path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	locationDir, err := sgInstallDir(homeDir)
	if err != nil {
		return err
	}

	// Remove existing install if there is one
	if existingPath, err := exec.LookPath(executable); err == nil {
		// If this mi2 installation is installed elsewhere, remove it to
		// avoid conflicts
		if !strings.HasPrefix(existingPath, locationDir) {
			std.Out.WriteNoticef("Removing existing installation at of %q at %q",
				executable, existingPath)
			_ = os.Remove(existingPath)
		}
	}

	version, err := run.Cmd(ctx, ghPath, "version").Run().String()
	if err != nil {
		return errors.Wrap(err, "get gh version")
	}
	std.Out.WriteNoticef("Using GitHub CLI version %q", strings.Split(version, "\n")[0])

	start := time.Now()
	pending := std.Out.Pending(output.Styledf(output.StylePending, "Downloading %q to %q... (hang tight, this might take a while!)",
		executable, locationDir))

	const tempExecutable = "mi2_tmp"
	tempInstallPath := filepath.Join(locationDir, tempExecutable)
	finalInstallPath := filepath.Join(locationDir, executable)
	_ = os.Remove(tempInstallPath)
	// Get release
	if err := run.Cmd(ctx,
		ghPath, " release download -R github.com/sourcegraph/controller",
		"--pattern", fmt.Sprintf("mi2_%s_%s", runtime.GOOS, runtime.GOARCH),
		"--output", tempInstallPath).
		Run().Wait(); err != nil {
		pending.Close()
		return errors.Wrap(err, "download mi2")
	}
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess,
		"Download complete! (elapsed: %s)",
		time.Since(start).String()))

	// Move binary to final destination
	if err := os.Rename(tempInstallPath, finalInstallPath); err != nil {
		return errors.Wrap(err, "move mi2 to final path")
	}

	// Make binary executable
	if err := os.Chmod(finalInstallPath, 0755); err != nil {
		return errors.Wrap(err, "make mi2 executable")
	}

	std.Out.WriteSuccessf("%q successfully installed!", executable)
	return nil
}
