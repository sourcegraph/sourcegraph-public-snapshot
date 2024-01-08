package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func main() {
	liblog := log.Init(log.Resource{
		Name:    "migrationdump",
		Version: "0.0.1",
	})
	defer liblog.Sync()
	logger := log.Scoped("cli")

	app := cli.App{
		Name:  "migrationdump",
		Usage: "Utility to dump migrations folder across versions",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "version.min",
				Usage: "First version to dump migrations for",
				Value: "3.20.0",
			},
			&cli.StringFlag{
				Name:     "version.max",
				Usage:    "First version to dump migrations for",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "output-dir",
				Usage:    "Directory to store tarballs into",
				Required: true,
			},
		},
		Action: func(ctx *cli.Context) error {
			return Main(ctx, logger)
		},
	}

	if err := app.RunContext(context.Background(), os.Args); err != nil {
		logger.Fatal("fatal error", log.Error(err))
		os.Exit(1)
	}

}

func Main(cctx *cli.Context, logger log.Logger) error {
	min, ok := oobmigration.NewVersionFromString(cctx.String("version.min"))
	if !ok {
		return fmt.Errorf("invalid minimum version: %q", cctx.String("version.min"))
	}

	max, ok := oobmigration.NewVersionFromString(cctx.String("version.max"))
	if !ok {
		return fmt.Errorf("invalid maximum version: %q", cctx.String("version.max"))
	}

	versions, err := oobmigration.UpgradeRange(min, max)
	if err != nil {
		return err
	}
	versionTags := make([]string, 0, len(versions))
	for _, version := range versions {
		versionTags = append(versionTags, version.GitTag())
	}

	for _, tag := range versionTags {
		if err := archiveMigrationsForVersion(cctx.String("output-dir"), tag, logger); err != nil {
			return err
		}
	}

	return nil
}

func archiveMigrationsForVersion(tmpdir string, rev string, logger log.Logger) error {
	cmd := exec.Command("git", "archive", "--format=tar.gz", rev, "migrations")
	out, err := cmd.CombinedOutput()

	if err != nil {
		if branch, ok := tagRevToBranch(rev); ok && err.Error() == "not a valid object name" {
			cmd := exec.Command("git", "archive", "--format=tar.gz", "origin/"+branch, "migrations")
			out, err = cmd.CombinedOutput()
		}
		if err != nil {
			return errors.Wrapf(err, "failed to run git archive: %s", out)
		}
	}

	f, err := os.Create(filepath.Join(tmpdir, fmt.Sprintf("migrations-%s.tar.gz", rev)))
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, bytes.NewReader(out)); err != nil {
		return err
	}
	return nil
}

func tagRevToBranch(rev string) (string, bool) {
	version, ok := oobmigration.NewVersionFromString(rev)
	if !ok {
		return "", false
	}

	return fmt.Sprintf("%d.%d", version.Major, version.Minor), true
}
