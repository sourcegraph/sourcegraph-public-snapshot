package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

var (
	versionFlagSet  = flag.NewFlagSet("sg version", flag.ExitOnError)
	versionRecord   = versionFlagSet.Bool("record", false, "Record the currently installed sg version")
	versionRecorded = versionFlagSet.Bool("recorded", false, "Print the currently recorded sg version")
)

var (
	versionCommand = &ffcli.Command{
		Name:        "version",
		ShortUsage:  "sg version",
		ShortHelp:   "Prints the sg version",
		FlagSet:     versionFlagSet,
		Exec:        versionExec,
		Subcommands: []*ffcli.Command{},
	}
)

func versionExec(ctx context.Context, args []string) error {
	if !*versionRecorded && !*versionRecord {
		stdout.Out.Write(BuildCommit)
		return nil
	}

	versionPath, err := getRecordedVersionPath()
	if err != nil {
		return fmt.Errorf("getRecordedVersionPath: %w", err)
	}
	if *versionRecorded {
		recordedVersion, err := os.ReadFile(versionPath)
		if err != nil {
			return fmt.Errorf("os.ReadFile: %w", err)
		}
		stdout.Out.Write(strings.TrimSpace(string(recordedVersion)))
	}
	if *versionRecord {
		if BuildCommit == "dev" {
			return fmt.Errorf("Refusing to record dev version without embedded commit")
		}

		versionFile, err := os.Create(versionPath)
		if err != nil {
			return fmt.Errorf("os.Create: %w", err)
		}
		realBuildCommit := strings.TrimPrefix(BuildCommit, "dev-")
		if _, err := versionFile.WriteString(realBuildCommit); err != nil {
			return fmt.Errorf("versionFile.WriteString: %w", err)
		}
		writeSuccessLinef("Currently installed version %q has been recorded to %s", realBuildCommit, versionPath)
	}
	return nil
}

func getRecordedVersionPath() (string, error) {
	homepath, err := root.GetSGHomePath()
	if err != nil {
		return "", fmt.Errorf("GetSGHomePath: %w", err)
	}
	return filepath.Join(homepath, "PREVIOUS_SG_VERSION"), nil
}
