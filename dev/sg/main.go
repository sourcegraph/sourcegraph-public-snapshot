package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/cockroachdb/errors"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const (
	defaultConfigFile          = "sg.config.yaml"
	defaultConfigOverwriteFile = "sg.config.overwrite.yaml"
	defaultSecretsFile         = "sg.secrets.json"
)

var secretsStore *secrets.Store

var (
	BuildCommit = "dev"

	// globalConf is the global config. If a command needs to access it, it *must* call
	// `parseConf` before.
	globalConf *Config

	rootFlagSet         = flag.NewFlagSet("sg", flag.ExitOnError)
	verboseFlag         = rootFlagSet.Bool("v", false, "verbose mode")
	configFlag          = rootFlagSet.String("config", defaultConfigFile, "configuration file")
	overwriteConfigFlag = rootFlagSet.String("overwrite", defaultConfigOverwriteFile, "configuration overwrites file that is gitignored and can be used to, for example, add credentials")

	rootCommand = &ffcli.Command{
		ShortUsage: "sg [flags] <subcommand>",
		FlagSet:    rootFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			runCommand,
			runSetCommand,
			startCommand,
			testCommand,
			doctorCommand,
			liveCommand,
			migrationCommand,
			rfcCommand,
			funkyLogoCommand,
			teammateCommand,
			ciCommand,
			installCommand,
			versionCommand,
			secretCommand,
			setupCommand,
			checkCommand,
			resetCommand,
		},
	}
)

// setMaxOpenFiles will bump the maximum opened files count.
// It's harmless since the limit only persists for the lifetime of the process and it's quick too.
func setMaxOpenFiles() error {
	const maxOpenFiles = 10000

	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return errors.Wrap(err, "getrlimit failed")
	}

	if rLimit.Cur < maxOpenFiles {
		rLimit.Cur = maxOpenFiles

		// This may not succeed, see https://github.com/golang/go/issues/30401
		err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			return errors.Wrap(err, "setrlimit failed")
		}
	}

	return nil
}

func checkSgVersion() {
	_, err := root.RepositoryRoot()
	if err != nil {
		// Ignore the error, because we only want to check the version if we're
		// in sourcegraph/sourcegraph
		return
	}

	if BuildCommit == "dev" {
		// If `sg` was built with a dirty `./dev/sg` directory it's a dev build
		// and we don't need to display this message.
		return
	}

	rev := BuildCommit
	if strings.HasPrefix(BuildCommit, "dev-") {
		rev = BuildCommit[len("dev-"):]
	}

	out, err := run.GitCmd("rev-list", fmt.Sprintf("%s..HEAD", rev), "./dev/sg")
	if err != nil {
		fmt.Printf("error getting new commits since %s in ./dev/sg: %s\n", rev, err)
		fmt.Println("try reinstalling sg with `./dev/sg/install.sh`.")
		os.Exit(1)
	}

	out = strings.TrimSpace(out)
	if out != "" {
		stdout.Out.WriteLine(output.Linef("", output.StyleSearchMatch, "--------------------------------------------------------------------------"))
		stdout.Out.WriteLine(output.Linef("", output.StyleSearchMatch, "HEY! New version of sg available. Run `./dev/sg/install.sh` to install it."))
		stdout.Out.WriteLine(output.Linef("", output.StyleSearchMatch, "--------------------------------------------------------------------------"))
	}
}

func loadSecrets() error {
	homePath, err := root.GetSGHomePath()
	if err != nil {
		return err
	}
	fp := filepath.Join(homePath, defaultSecretsFile)
	secretsStore, err = secrets.LoadFile(fp)
	return err
}

func main() {
	if err := loadSecrets(); err != nil {
		fmt.Printf("failed to open secrets: %s\n", err)
	}
	ctx := secrets.WithContext(context.Background(), secretsStore)

	if err := rootCommand.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}

	checkSgVersion()

	// We always try to set this, since we
	// often want to watch files, start commands, etc...
	if err := setMaxOpenFiles(); err != nil {
		fmt.Printf("failed to set max open files: %s\n", err)
		os.Exit(1)
	}

	if err := rootCommand.Run(ctx); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

// parseConf parses the config file and the optional overwrite file.
// Iear the conf has already been parsed it's a noop.
func parseConf(confFile, overwriteFile string) (bool, output.FancyLine) {
	if globalConf != nil {
		return true, output.FancyLine{}
	}

	// Try to determine root of repository, so we can look for config there
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return false, output.Linef("", output.StyleWarning, "Failed to determine repository root location: %s", err)
	}

	// If the configFlag/overwriteConfigFlag flags have their default value, we
	// take the value as relative to the root of the repository.
	if confFile == defaultConfigFile {
		confFile = filepath.Join(repoRoot, confFile)
	}

	if overwriteFile == defaultConfigOverwriteFile {
		overwriteFile = filepath.Join(repoRoot, overwriteFile)
	}

	globalConf, err = ParseConfigFile(confFile)
	if err != nil {
		return false, output.Linef("", output.StyleWarning, "Failed to parse %s%s%s%s as configuration file:%s\n%s", output.StyleBold, confFile, output.StyleReset, output.StyleWarning, output.StyleReset, err)
	}

	if ok, _ := fileExists(overwriteFile); ok {
		overwriteConf, err := ParseConfigFile(overwriteFile)
		if err != nil {
			return false, output.Linef("", output.StyleWarning, "Failed to parse %s%s%s%s as overwrites configuration file:%s\n%s", output.StyleBold, overwriteFile, output.StyleReset, output.StyleWarning, output.StyleReset, err)
		}
		globalConf.Merge(overwriteConf)
	}

	return true, output.FancyLine{}
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
