package cloud

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var GetEphemeralCommand = cli.Command{
	Name:        "status",
	Usage:       "sg could status",
	Description: "get the status of the ephemeral cloud instance for this branch or instance with the provided slug",
	Action:      getCloudEphemeral,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "name",
			Usage: "name of the instance to get the status of",
		},
	},
}

func getCloudEphemeral(ctx *cli.Context) error {
	// while we work on this command we print a notice and ask to continue
	if err := printWIPNotice(ctx); err != nil {
		return err
	}
	email, err := GetGCloudAccount(ctx.Context)
	if err != nil {
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

	cloudEmoji := "☁️"
	pending := std.Out.Pending(output.Linef(cloudEmoji, output.StylePending, "Getting status of ephemeral instance %q", name))
	inst, err := cloudClient.GetInstance(ctx.Context, name)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "getting status of %q failed", name))
		return err
	}

	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleWhiteOnPurple, "Details of ephemeral instance %q:\n%s", inst))
	return nil
}
