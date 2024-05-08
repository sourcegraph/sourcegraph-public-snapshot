package dependencies

import (
	"context"
	"os"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/sgconf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// cmdFix executes the given command as an action in a new user shell.
func cmdFix(cmd string) check.FixAction[CheckArgs] {
	return func(ctx context.Context, cio check.IO, args CheckArgs) error {
		c := usershell.Command(ctx, cmd)
		if cio.Input != nil {
			c = c.Input(cio.Input)
		}
		return c.Run().StreamLines(cio.Verbose)
	}
}

func cmdFixes(cmds ...string) check.FixAction[CheckArgs] {
	return func(ctx context.Context, cio check.IO, args CheckArgs) error {
		for _, cmd := range cmds {
			if err := cmdFix(cmd)(ctx, cio, args); err != nil {
				return err
			}
		}
		return nil
	}
}

func enableOnlyInSourcegraphRepo() check.EnableFunc[CheckArgs] {
	return func(ctx context.Context, args CheckArgs) error {
		_, err := root.RepositoryRoot()
		return err
	}
}

func disableInCI() check.EnableFunc[CheckArgs] {
	return func(ctx context.Context, args CheckArgs) error {
		// Docker is quite funky in CI
		if os.Getenv("CI") == "true" {
			return errors.New("disabled in CI")
		}
		return nil
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func checkSourcegraphDatabase(ctx context.Context, out *std.Output, args CheckArgs) error {
	getConfig := func() (*sgconf.Config, error) {
		var config *sgconf.Config
		var err error
		if args.DisableOverwrite {
			config, err = sgconf.GetWithoutOverwrites(args.ConfigFile)
		} else {
			config, err = sgconf.Get(args.ConfigFile, args.ConfigOverwriteFile)
		}
		if err != nil {
			return nil, err
		}
		if config == nil {
			return nil, errors.New("failed to read sg.config.yaml. This step of `sg setup` needs to be run in the `sourcegraph` repository")
		}

		return config, nil
	}

	return check.SourcegraphDatabase(getConfig)(ctx)
}

func checkSrcCliVersion(versionConstraint string) check.CheckFunc {
	return check.CompareSemanticVersion("src", "src version -client-only", versionConstraint)
}

func forceASDFPluginAdd(ctx context.Context, plugin string, source string) error {
	err := usershell.Run(ctx, "asdf plugin-add", plugin, source).Wait()
	if err != nil && strings.Contains(err.Error(), "already added") {
		return nil
	}
	return errors.Wrap(err, "asdf plugin-add")
}

// brewInstall returns a FixAction that installs a brew formula.
// If the brew output contains an autofix for adding the formula to the path
// (in the case of keg-only formula), it will be automatically applied.
func createBrewInstallFix(formula string, cask bool) check.FixAction[CheckArgs] {
	return func(ctx context.Context, cio check.IO, args CheckArgs) error {
		cmd := "brew install "
		if cask {
			cmd += "--cask "
		}
		cmd += formula
		c := usershell.Command(ctx, cmd)
		if cio.Input != nil {
			c = c.Input(cio.Input)
		}

		pathAddCommandIsNext := false
		return c.Run().StreamLines(func(line string) {
			if pathAddCommandIsNext {
				matches := exportPathRegexp.FindStringSubmatch(line)
				if len(matches) != 2 {
					cio.Output.WriteWarningf("unexpected output from brew install: %q\n"+
						"was not able to automatically update $PATH. Please add this to "+
						"your path manually.", line)
				} else {
					_ = usershell.Run(
						ctx,
						"echo -e 'export PATH="+matches[1],
						">>",
						usershell.ShellConfigPath(ctx),
					).Wait()
				}
				pathAddCommandIsNext = false
			}
			if strings.Contains(line, "If you need to have "+formula+" first in your PATH, run:") {
				pathAddCommandIsNext = true
			}
			cio.Verbose(line)
		})
	}
}

var exportPathRegexp = regexp.MustCompile(`export PATH=(.*) >>`)

func caskInstall(formula string) check.FixAction[CheckArgs] {
	return createBrewInstallFix(formula, true)
}

func brewInstall(formula string) check.FixAction[CheckArgs] {
	return createBrewInstallFix(formula, false)
}
