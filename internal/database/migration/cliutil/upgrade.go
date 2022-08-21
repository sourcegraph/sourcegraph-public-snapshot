package cliutil

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Upgrade(
	logger log.Logger,
	commandName string,
	runnerFactory RunnerFactoryWithSchemas,
	outFactory OutputFactory,
	registerMigrators func(storeFactory migrations.StoreFactory) oobmigration.RegisterMigratorsFunc,
) *cli.Command {
	fromFlag := &cli.StringFlag{
		Name:     "from",
		Usage:    "The source (current) instance version. Must be of the form `v{Major}.{Minor}`.",
		Required: true,
	}
	toFlag := &cli.StringFlag{
		Name:     "to",
		Usage:    "The target instance version. Must be of the form `v{Major}.{Minor}`.",
		Required: true,
	}
	skipVersionCheckFlag := &cli.BoolFlag{
		Name:     "skip-version-check",
		Usage:    "Skip validation of the instance's current version.",
		Required: false,
	}
	dryRunFlag := &cli.BoolFlag{
		Name:     "dry-run",
		Usage:    "Print the upgrade plan but do not execute it.",
		Required: false,
	}

	action := makeAction(outFactory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		from, ok := oobmigration.NewVersionFromString(fromFlag.Get(cmd))
		if !ok {
			return errors.New("bad format for -from")
		}
		to, ok := oobmigration.NewVersionFromString(toFlag.Get(cmd))
		if !ok {
			return errors.New("bad format for -to")
		}

		// Construct inclusive upgrade version range `[from, to]`. This also checks
		// for known major version upgrades (e.g., 3.0.0 -> 4.0.0) and ensures that
		// the given values are in the correct order (e.g., from < to).
		versionRange, err := oobmigration.UpgradeRange(from, to)
		if err != nil {
			return err
		}

		// Find the relevant schema and data migrations to perform (and in what order)
		// for the given version range. Perform the upgrade on the configured databases.
		plan, err := planUpgrade(versionRange)
		if err != nil {
			return err
		}
		if len(plan.steps) == 0 {
			return errors.New("upgrade plan contains no steps")
		}

		if dryRunFlag.Get(cmd) {
			for i, step := range plan.steps {
				out.WriteLine(output.Linef(
					output.EmojiFingerPointRight,
					output.StyleReset,
					"Migrating to v%s (step %d of %d)",
					step.instanceVersion,
					i+1,
					len(plan.steps),
				))

				out.WriteLine(output.Line(output.EmojiFingerPointRight, output.StyleReset, "Running schema migrations"))

				for schemaName, leafIDs := range step.schemaMigrationLeafIDsBySchemaName {
					out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleReset, "%s -> %v", schemaName, leafIDs))
				}

				if len(step.outOfBandMigrationIDs) > 0 {
					out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleReset, "Running out of band migrations %v", step.outOfBandMigrationIDs))
				}

			}
		} else {
			if err := runUpgrade(ctx, runnerFactory, plan, skipVersionCheckFlag.Get(cmd), registerMigrators, out); err != nil {
				return err
			}
		}

		return nil
	})

	return &cli.Command{
		Name:        "upgrade",
		Usage:       "Upgrade Sourcegraph instance databases to a target version",
		Description: "",
		Action:      action,
		Flags: []cli.Flag{
			fromFlag,
			toFlag,
			skipVersionCheckFlag,
			dryRunFlag,
		},
	}
}
