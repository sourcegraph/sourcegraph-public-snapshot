package cloud

import (
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// MaxDuration is the maximum lease duration by which a deployment can be extended by, which is 4 days
const MaxDuration = time.Hour * 24 * 4

var leaseEphemeralCommand = cli.Command{
	Name:        "lease",
	Usage:       "Extend or reduce the lease expiry time of an ephemeral instance",
	UsageText:   "sg cloud lease [command options]",
	Description: "Extend or reduce the lease expiry time of an ephemeral instance. Once the expiry time of an ephemeral instance is reached it gets deleted and inaccessible",
	Action:      leaseCloudEphemeral,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "name",
			Usage:       "name of the ephemeral instance",
			DefaultText: "current branch name will be used",
		},
		&cli.DurationFlag{
			Name:  "extend",
			Usage: "extend the lease by the given duration value, where the value can be any parseable duration of hours (h) minutes (m) or seconds (s).  Example: --extend '1h' or --extend '50m30s' or --extend '24h50m'",
		},
		&cli.DurationFlag{
			Name:  "reduce",
			Usage: "reduce the the lease expiry time by the given duration value, where the value ca be any parseable duration of hours (h) minutes (m) or seconds (s). Example: --reduce '1h' or --reduce '50m30s' or --reduce '24h50m'",
		},
		&cli.BoolFlag{
			Name:  "expire-now",
			Usage: "sets the lease expiry time to now",
		},
	},
}

func printLeaseTimeDiff(oldTime, newTime time.Time) {
	std.Out.Write("Updating instance lease with the following values:")
	std.Out.WriteLine(output.Linef("", output.StyleRed, "- Lease time %s", oldTime.Format(time.RFC3339)))
	std.Out.WriteLine(output.Linef("", output.StyleGreen, "+ Lease time %s", newTime.Format(time.RFC3339)))
	std.Out.Write("\n")
}

func calcLeaseEnd(currentLeaseTime time.Time, extension, reduction time.Duration) (time.Time, error) {
	var leaseEndTime time.Time
	if extension == 0 && reduction == 0 {
		return leaseEndTime, errors.New("extension and reduction times cannot both be 0")
	}

	if extension > 0 {
		leaseEndTime = currentLeaseTime.Add(extension)
		return leaseEndTime, nil
	} else if extension < 0 {
		return leaseEndTime, errors.New("lease extension value should be a positive value")
	}
	if reduction > 0 {
		leaseEndTime = currentLeaseTime.Add(-reduction)
	} else if reduction < 0 {
		return leaseEndTime, errors.New("lease reduction value should be a positive value")
	}

	return leaseEndTime, nil
}

func leaseCloudEphemeral(ctx *cli.Context) error {
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

	// Do various checks before upgrading the instance
	if !inst.HasStatus(InstanceStatusCompleted) {
		std.Out.WriteWarningf("Cannot update lease time of instance with status %q - if this issue persists, please reach out to #discuss-dev-infra", inst.Status.Status)
		return ErrInstanceStatusNotComplete
	}
	if !inst.IsEphemeral() {
		std.Out.WriteWarningf("Cannot update lease time of non-ephemeral instance %q", name)
		return ErrNotEphemeralInstance
	}
	if inst.IsExpired() {
		std.Out.WriteWarningf(" Cannot update lease time of expired instance %q", name)
		return ErrExpiredInstance
	}

	// All the checks passed, we can try to extend the lease
	currentLeaseTime := inst.ExpiresAt
	var leaseEndTime time.Time
	if ctx.Bool("expire-now") {
		leaseEndTime = time.Now()
	} else if t, err := calcLeaseEnd(currentLeaseTime, ctx.Duration("extend"), ctx.Duration("reduce")); err != nil {
		return err
	} else {
		leaseEndTime = t
	}

	if leaseEndTime.Sub(currentLeaseTime) > MaxDuration {
		return errors.Newf("cannot update lease time by more than %s", MaxDuration)
	}

	pending = std.Out.Pending(output.Linef(CloudEmoji, output.StylePending, "Updating lease of instance %q", name))

	printLeaseTimeDiff(currentLeaseTime, leaseEndTime)
	inst, err = cloudClient.ExtendLease(ctx.Context, name, leaseEndTime)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleFailure, "failed to update lease"))
	}

	pending.Complete(output.Linef(CloudEmoji, output.StyleSuccess, "Lease of instance %q updated by %s", name, leaseEndTime.Sub(currentLeaseTime)))
	return newDefaultTerminalInstancePrinter().Print(inst)
}
