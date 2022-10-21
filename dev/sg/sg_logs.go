package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/nxadm/tail"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	logsCommand = &cli.Command{
		Name:   "logs",
		Usage:  "",
		Action: logsExec,
	}
)

func logsExec(ctx *cli.Context) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "could not check pidfiles")
	}

	pattern := filepath.Join(homeDir, ".sourcegraph", "sg.pid.*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return errors.Wrap(err, "could not list pidfiles")
	}

	var tailFile *tail.Tail

	for _, match := range matches {
		f, err := os.Open(match)
		if err != nil {
			return errors.Wrapf(err, "could not check pidfile %q", match)
		}
		defer f.Close()

		var content run.PidFile
		if err := json.NewDecoder(f).Decode(&content); err != nil {
			return errors.Wrapf(err, "could not check pidfile %q", match)
		}

		tailFile, err = tail.TailFile(content.LogFile, tail.Config{})
		if err != nil {
			return err
		}
		break
	}

	out := std.Out.Output

	for lines := range tailFile.Lines {
		if lines.Err != nil {
			return lines.Err
		}
		out.Write(lines.Text)
	}
	tailFile.Wait() // this just waits ... it doesn't wait for MORE data

	return nil
}
