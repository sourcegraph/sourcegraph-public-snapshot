package cloud

import (
	"net/url"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const devCloudEphemeralDashURL = "https://cloud-ops-dev.sgdev.org/dashboard/environments/dev"

var dashboardEphemeralCommand = cli.Command{
	Name:        "dashboard",
	Aliases:     []string{"dash"},
	Usage:       "Opens the ephemeral dashboard",
	Description: "Opens the ephemeral dashboard",
	Action:      showCloudEphemeralDash,
}

var opsDashboardEphemeralCommand = cli.Command{
	Name: "ops",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "name",
			Usage:       "name of the instance to update the lease expiry time for",
			DefaultText: "current branch name will be used",
		},
	},
	Usage:       "Opens the ops page for the ephemeral deployment",
	Description: "Opens the ops page for the ephemeral deployment",
	Action:      showCloudEphemeralOps,
}

func showCloudEphemeralOps(ctx *cli.Context) error {
	email, err := GetGCloudAccount(ctx.Context)
	if err != nil {
		writeGCloudErrorSuggestion()
		return err
	}

	cloudClient, err := NewClient(ctx.Context, email, APIEndpoint)
	if err != nil {
		return err
	}

	name := ctx.String("name")
	if name == "" {
		currentBranch, err := repo.GetCurrentBranch(ctx.Context)
		if err != nil {
			return errors.Wrap(err, "failed to determine current branch")
		}
		name = currentBranch
	}
	name = sanitizeInstanceName(name)

	cloudEmoji := "☁️"
	pending := std.Out.Pending(output.Linef(cloudEmoji, output.StylePending, "Getting details of ephemeral instance %q", name))

	// Check if the deployment already exists
	inst, err := cloudClient.GetInstance(ctx.Context, name)
	if err != nil {
		if errors.Is(err, ErrInstanceNotFound) {
			pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "Getting details of %q failed", name))
			return ErrInstanceNotFound
		}
	}
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleBold, "Ephemeral instances details for %q retrieved", name))

	pageURL, err := url.JoinPath(devCloudEphemeralDashURL, "instances", inst.ID)
	if err != nil {
		return err
	}
	std.Out.WriteNoticef("Opening ops page for %q", inst.Name)
	return open.URL(pageURL)
}

func showCloudEphemeralDash(ctx *cli.Context) error {
	std.Out.WriteNoticef("Opening cloud ephemeral dashboard")
	return open.URL(devCloudEphemeralDashURL)
}
