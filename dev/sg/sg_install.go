package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	installFlagSet = flag.NewFlagSet("sg install", flag.ExitOnError)
	installCommand = &ffcli.Command{
		Name:       "install",
		ShortUsage: "sg install",
		ShortHelp:  "Installs sg to a user-defined location by copying sg itself",
		FlagSet:    installFlagSet,
		Exec:       installExec,
	}
)

func installExec(ctx context.Context, args []string) error {
	const location = "/usr/local/bin/sg"

	var logoOut bytes.Buffer
	printLogo(&logoOut)
	stdout.Out.Write(logoOut.String())

	stdout.Out.Write("")
	stdout.Out.WriteLine(output.Linef("", output.StyleLogo, "Welcome to the sg installation!"))

	stdout.Out.Write("")
	stdout.Out.Writef("We are going to install %ssg%s to %s%s%s. Okay?", output.StyleBold, output.StyleReset, output.StyleBold, location, output.StyleReset)

	locationOkay := getBool()
	if !locationOkay {
		return errors.New("user not happy with location :(")
	}

	currentLocation, err := os.Executable()
	if err != nil {
		panic(err)
	}

	pending := stdout.Out.Pending(output.Linef("", output.StylePending, "Copying from %s%s%s to %s%s%s...", output.StyleBold, currentLocation, output.StyleReset, output.StyleBold, location, output.StyleReset))

	original, err := os.Open(currentLocation)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleWarning, "Failed: %s", err))
		return err
	}
	defer original.Close()

	// Create new file
	newFile, err := os.Create(location)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleWarning, "Failed: %s", err))
		return err
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, original)
	if err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleWarning, "Failed: %s", err))
		return err
	}
	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Done!"))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	paths := []string{
		filepath.Join(homeDir, ".profile"),
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".zshenv"),
	}

	stdout.Out.Write("")
	stdout.Out.Writef("The path %s%s%s will be added to your %sPATH%s environment variable by", output.StyleBold, filepath.Dir(location), output.StyleReset, output.StyleBold, output.StyleReset)
	stdout.Out.Writef("modifying the profile files located at:")
	stdout.Out.Write("")
	for _, p := range paths {
		stdout.Out.Writef("  %s%s", output.StyleBold, p)
	}

	addToShellOkay := getBool()
	if !addToShellOkay {
		return errors.New("user not happy with adding stuff to shell:(")
	}

	return nil
}
func getBool() bool {
	var s string

	fmt.Printf("(y/N): ")
	_, err := fmt.Scan(&s)
	if err != nil {
		panic(err)
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "y" || s == "yes" {
		return true
	}
	return false
}
