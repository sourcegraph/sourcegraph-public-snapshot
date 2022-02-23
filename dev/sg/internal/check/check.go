package check

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CheckFunc func(context.Context) error

func InPath(cmd string) func(context.Context) error {
	return func(ctx context.Context) error {
		hashCmd := fmt.Sprintf("hash %s 2>/dev/null", cmd)
		_, err := usershell.CombinedExec(ctx, hashCmd)
		if err != nil {
			return errors.Newf("executable %q not found in $PATH", cmd)
		}
		return nil
	}
}

func CommandExitCode(cmd string, exitCode int) func(context.Context) error {
	return func(ctx context.Context) error {
		cmd := usershell.Cmd(ctx, cmd)
		err := cmd.Run()
		var execErr *exec.ExitError
		if err != nil {
			if errors.As(err, &execErr) && execErr.ExitCode() != exitCode {
				return errors.Newf("command %q has wrong exit code, wanted %d but got %d", cmd, exitCode, execErr.ExitCode())
			}
			return err
		}
		return nil
	}
}

func Version(cmdName, haveVersion, versionConstraint string) error {
	c, err := semver.NewConstraint(versionConstraint)
	if err != nil {
		return err
	}

	version, err := semver.NewVersion(haveVersion)
	if err != nil {
		return errors.Newf("cannot decode version in %q: %w", haveVersion, err)
	}

	if !c.Check(version) {
		return errors.Newf("version %q from %q does not match constraint %q", haveVersion, cmdName, versionConstraint)
	}
	return nil
}

func CommandOutputContains(cmd, contains string) func(context.Context) error {
	return func(ctx context.Context) error {
		out, _ := usershell.CombinedExec(ctx, cmd)
		if !strings.Contains(string(out), contains) {
			return errors.Newf("command output of %q doesn't contain %q", cmd, contains)
		}
		return nil
	}
}

func FileContains(fileName, content string) func(context.Context) error {
	return func(context.Context) error {
		file, err := os.Open(fileName)
		if err != nil {
			return errors.Wrapf(err, "failed to check that %q contains %q", fileName, content)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, content) {
				return nil
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		return errors.Newf("file %q did not contain %q", fileName, content)
	}
}

// This ties the check to having the library installed with apt-get on Ubuntu,
// which against the principle of checking dependencies independently of their
// installation method. Given they're just there for comby and sqlite, the chances
// that someone needs to install them in a different way is fairly low, making this
// check acceptable for the time being.
func HasUbuntuLibrary(name string) func(context.Context) error {
	return func(ctx context.Context) error {
		_, err := usershell.CombinedExec(ctx, fmt.Sprintf("dpkg -s %s", name))
		return errors.Newf("dpkg: %w", err)
	}
}
