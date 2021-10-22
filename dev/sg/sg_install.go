package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var location string
	switch runtime.GOOS {
	case "linux":
		location = filepath.Join(homeDir, ".local", "bin", "sg")
	case "darwin":
		location = "/usr/local/bin/sg"
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

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

	// Make sure directory for new file exists
	sgDir := filepath.Dir(location)
	if err := os.MkdirAll(sgDir, os.ModePerm); err != nil {
		pending.Complete(output.Linef(output.EmojiFailure, output.StyleWarning, "Failed: %s", err))
		return err
	}

	// Create new file
	newFile, err := os.OpenFile(location, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
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

	paths := []struct {
		path              string
		createIfNotExists bool
		install           bool
	}{
		{path: filepath.Join(homeDir, ".zshenv"), createIfNotExists: false},
		{path: filepath.Join(homeDir, ".bashrc"), createIfNotExists: false},
		{path: filepath.Join(homeDir, ".profile"), createIfNotExists: true},
	}

	for i, p := range paths {
		// If the file is not .profile and doesn't exist we don't want to append/create
		if _, err := os.Stat(p.path); !p.createIfNotExists && os.IsNotExist(err) {
			continue
		}
		paths[i].install = true
	}

	stdout.Out.Write("")
	stdout.Out.Writef("The path %s%s%s will be added to your %sPATH%s environment variable by", output.StyleBold, sgDir, output.StyleReset, output.StyleBold, output.StyleReset)
	stdout.Out.Writef("modifying the profile files located at:")
	stdout.Out.Write("")
	for _, p := range paths {
		if !p.install {
			continue
		}
		stdout.Out.Writef("  %s%s", output.StyleBold, p.path)
	}

	addToShellOkay := getBool()
	if !addToShellOkay {
		stdout.Out.Writef("Alright! Make sure to add %s to your $PATH, restart your shell and run 'sg logo'. See you!", sgDir)
		return nil
	}

	pending = stdout.Out.Pending(output.Linef("", output.StylePending, "Writing to files..."))

	exportLine := fmt.Sprintf("\nexport PATH=%s:$PATH\n", sgDir)
	lineWrittenTo := []string{}
	for _, p := range paths {
		if !p.install {
			continue
		}

		f, err := os.OpenFile(p.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return errors.Wrapf(err, "failed to open %s", p.path)
		}
		defer f.Close()

		if _, err := f.WriteString(exportLine); err != nil {
			return errors.Wrapf(err, "failed to write to %s", p.path)
		}

		lineWrittenTo = append(lineWrittenTo, p.path)
	}

	pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Done!"))

	stdout.Out.Writef("Modified the following files:")
	stdout.Out.Write("")
	for _, p := range lineWrittenTo {
		stdout.Out.Writef("  %s%s", output.StyleBold, p)
	}

	stdout.Out.Write("")
	stdout.Out.Writef("Restart your shell and run 'sg logo' to make sure it worked!")

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
