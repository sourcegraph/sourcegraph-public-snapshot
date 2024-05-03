package cloud

import (
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const MaxDuration = time.Hour * 24 * 4

var ExtendLeaseEphemeralCommand = cli.Command{
	Name:        "extend-lease",
	ArgsUsage:   "sg cloud extend-lease [--name instance-name] <duration>",
	Description: "extend the lease of the instance for the given duration (max 4 days)",
	Action:      wipAction(extendLeaseCloudEphemeral),
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "name",
			DefaultText: "name of the instance to extend lease for",
		},
	},
}

func printLeaseTimeDiff(oldTime, newTime time.Time) {
	std.Out.Write("Updating instance lease with the following values:")
	std.Out.WriteLine(output.Linef("", output.StyleRed, "- Lease time %s", oldTime.Format(time.RFC3339)))
	std.Out.WriteLine(output.Linef("", output.StyleGreen, "+ Lease time %s", newTime.Format(time.RFC3339)))
	std.Out.Write("\n")
}

func extendLeaseCloudEphemeral(ctx *cli.Context) error {
	if ctx.Args().Len() != 1 {
		return errors.New("no duration provided")
	}
	duration, err := time.ParseDuration(ctx.Args().First())
	if err != nil {
		return errors.Wrapf(err, "failed to parse duration %q", ctx.Args().First())
	}

	if duration > MaxDuration {
		std.Out.Writef("duration %s is greater than max duration %s, setting to max duration", duration, MaxDuration)
		duration = MaxDuration
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

	pending = std.Out.Pending(output.Linef(CloudEmoji, output.StylePending, "Extending lease of instance %q by %s", name, duration))

	current, err := inst.GetExpiry()
	if err != nil {
		errors.Wrap(err, "failed to get instance lease expiry")
	}

	leaseEndTime := current.Add(duration)
	printLeaseTimeDiff(*current, leaseEndTime)
	inst, err = cloudClient.ExtendLease(ctx.Context, name, leaseEndTime)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "failed to extend lease"))
	}

	pending.Complete(output.Linef(CloudEmoji, output.StyleSuccess, "Lease of instance %q extended by %s", name, duration))
	newDefaultTerminalInstancePrinter().Print(inst)
	return nil
}
