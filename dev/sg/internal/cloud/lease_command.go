package cloud

import (
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const MaxDuration = time.Hour * 24 * 4

var LeaseEphemeralCommand = cli.Command{
	Name:        "lease",
	ArgsUsage:   "sg cloud lease [--name instance-name] <duration>",
	Description: "extend the lease of the instance for the given duration (max 4 days)",
	Action:      wipAction(leaseCloudEphemeral),
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "name",
			DefaultText: "name of the instance to extend lease for",
		},
		&cli.DurationFlag{
			Name:        "extend",
			DefaultText: "the duration to extend the lease by (max 4 days = 96h)",
		},
		&cli.DurationFlag{
			Name:        "reduce",
			DefaultText: "the duration to reduce the lease by - if the lease time is reduced to be in the passed the instance will be deleted!",
		},
	},
}

func printLeaseTimeDiff(oldTime, newTime time.Time) {
	std.Out.Write("Updating instance lease with the following values:")
	std.Out.WriteLine(output.Linef("", output.StyleRed, "- Lease time %s", oldTime.Format(time.RFC3339)))
	std.Out.WriteLine(output.Linef("", output.StyleGreen, "+ Lease time %s", newTime.Format(time.RFC3339)))
	std.Out.Write("\n")
}

func leaseCloudEphemeral(ctx *cli.Context) error {
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
		n, err := inferInstanceNameFromBranch(ctx.Context)
		if err != nil {
			return err
		}
		name = n
	}

	pending := std.Out.Pending(output.Linef(CloudEmoji, output.StylePending, "Getting instance with name %q", name))
	inst, err := cloudClient.GetInstance(ctx.Context, name)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "failed to get instance"))
		return err
	}
	pending.Complete(output.Linef(CloudEmoji, output.StyleSuccess, "Fetched instance with name %q", name))

	if !inst.IsEphemeral() {
		std.Out.WriteWarningf("Cannot update lease time of non-ephemeral instance %q", name)
		return ErrNotEphemeralInstance
	}

	if ctx.Duration("extend") == 0 && ctx.Duration("reduce") == 0 {
		return errors.New("must specify a duration for either --extend or --reduce")
	}

	pending = std.Out.Pending(output.Linef(CloudEmoji, output.StylePending, "Updating lease of instance %q", name))

	currentLeaseTime := inst.ExpiresAt

	var leaseEndTime time.Time
	if ctx.Duration("extend") > 0 {
		leaseEndTime.Add(ctx.Duration("extend"))
	}
	if ctx.Duration("reduce") > 0 {
		leaseEndTime = currentLeaseTime.Add(-ctx.Duration("reduce"))
	}

	printLeaseTimeDiff(currentLeaseTime, leaseEndTime)
	inst, err = cloudClient.ExtendLease(ctx.Context, name, leaseEndTime)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "failed to update lease"))
	}

	pending.Complete(output.Linef(CloudEmoji, output.StyleSuccess, "Lease of instance %q updated by %s", name, currentLeaseTime.Sub(leaseEndTime)))
	return newDefaultTerminalInstancePrinter().Print(inst)
}
