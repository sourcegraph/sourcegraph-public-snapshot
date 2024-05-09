package cloud

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var DeleteEphemeralCommand = cli.Command{
	Name:        "delete",
	Usage:       "sg could delete <name/slug>",
	Description: "delete ephemeral cloud instance identified either by the current branch or provided as a cli arg",
	Action:      deleteCloudEphemeral,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "name or slug of the cloud ephemeral instance to delete",
		},
	},
}

func deleteCloudEphemeral(ctx *cli.Context) error {
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
	name = sanitizeInstanceName(name)

	var answ string
	_, err = std.PromptAndScan(std.Out, fmt.Sprintf("Are you sure you want to delete ephemeral instance %q? (yes/no)", name), &answ)
	if err != nil {
		return err
	}

	if oneOfEquals(answ, "no", "n") {
		return ErrUserCancelled
	}

	pending := std.Out.Pending(output.Linef(CloudEmoji, output.StylePending, "Deleting ephemeral instance %q", name))
	err = cloudClient.DeleteInstance(ctx.Context, name)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "deleting of %q failed", name))
		return err
	}

	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleBold, "Ephemeral instance %q deleted", name))
	return nil
}
