package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var setupCommand = &cli.Command{
	Name:     "setup",
	Usage:    "Set up your local dev environment!",
	Category: CategoryEnv,
	Action:   execAdapter(setupExec),
}

func setupExec(ctx context.Context, args []string) error {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "'sg setup' currently only supports macOS and Linux"))
		os.Exit(1)
	}

	currentOS := runtime.GOOS
	if overridesOS, ok := os.LookupEnv("SG_FORCE_OS"); ok {
		currentOS = overridesOS
	}

	ctx, err := usershell.Context(ctx)
	if err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			writeOrangeLinef("\n💡 You may need to restart your shell for the changes to work in this terminal.")
			writeOrangeLinef("   Close this terminal and open a new one or type the following command and press ENTER: %s", filepath.Base(usershell.ShellPath(ctx)))
			os.Exit(0)
		}
	}()

	var categories []dependencyCategory
	if currentOS == "darwin" {
		categories = macOSDependencies
	} else {
		categories = ubuntuOSDependencies
	}

	// Check whether we're in the sourcegraph/sourcegraph repository so we can
	// skip categories/dependencies that depend on the repository.
	_, err = root.RepositoryRoot()
	inRepo := err == nil

	failed := []int{}
	all := []int{}
	skipped := []int{}
	teammateFailed := []int{}
	for i := range categories {
		failed = append(failed, i)
		all = append(all, i)
	}

	for len(failed) != 0 {
		stdout.Out.ClearScreen()

		printSgSetupWelcomeScreen()
		writeOrangeLinef("                INFO: You can quit any time by typing ctrl-c\n")

		for i, category := range categories {
			idx := i + 1

			if category.requiresRepository && !inRepo {
				writeSkippedLinef("%d. %s %s[SKIPPED. Requires 'sg setup' to be run in 'sourcegraph' repository]%s", idx, category.name, output.StyleBold, output.StyleReset)
				skipped = append(skipped, idx)
				failed = removeEntry(failed, i)
				continue
			}

			pending := stdout.Out.Pending(output.Linef("", output.StylePending, "%d. %s - Determining status...", idx, category.name))
			category.Update(ctx)
			pending.Destroy()

			if combined := category.CombinedState(); combined {
				writeSuccessLinef("%d. %s", idx, category.name)
				failed = removeEntry(failed, i)
			} else {
				nonTeammateState := category.CombinedStateNonTeammates()
				if nonTeammateState {
					writeWarningLinef("%d. %s", idx, category.name)
					teammateFailed = append(skipped, idx)
				} else {
					writeFailureLinef("%d. %s", idx, category.name)
				}
			}
		}

		if len(failed) == 0 && len(teammateFailed) == 0 {
			if len(skipped) == 0 && len(teammateFailed) == 0 {
				stdout.Out.Write("")
				stdout.Out.WriteLine(output.Linef(output.EmojiOk, output.StyleBold, "Everything looks good! Happy hacking!"))
			}

			if len(skipped) != 0 {
				stdout.Out.Write("")
				writeWarningLinef("Some checks were skipped because 'sg setup' is not run in the 'sourcegraph' repository.")
				writeFingerPointingLinef("Restart 'sg setup' in the 'sourcegraph' repository to continue.")
			}

			return nil
		}

		stdout.Out.Write("")

		if len(teammateFailed) != 0 && len(failed) == len(teammateFailed) {
			writeWarningLinef("Some checks that are only relevant for Sourcegraph teammates failed.\nIf you're not a Sourcegraph teammate you're good to go. Hit Ctrl-C.\n\nIf you're a Sourcegraph teammate: which one do you want to fix?")
		} else {
			writeWarningLinef("Some checks failed. Which one do you want to fix?")
		}

		idx, err := getNumberOutOf(all)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		selectedCategory := categories[idx]

		stdout.Out.ClearScreen()

		err = presentFailedCategoryWithOptions(ctx, idx, &selectedCategory)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
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

func presentFailedCategoryWithOptions(ctx context.Context, categoryIdx int, category *dependencyCategory) error {
	printCategoryHeaderAndDependencies(categoryIdx+1, category)

	choices := map[int]string{1: "I want to fix these manually"}
	if category.autoFixing {
		choices[2] = "I'm feeling lucky. You try fixing all of it for me."
		choices[3] = "Go back"
	} else {
		choices[2] = "Go back"
	}

	choice, err := getChoice(choices)
	if err != nil {
		return err
	}

	switch choice {
	case 1:
		err = fixCategoryManually(ctx, categoryIdx, category)
	case 2:
		if category.autoFixing {
			stdout.Out.ClearScreen()
			err = fixCategoryAutomatically(ctx, category)
		}
	case 3:
		return nil
	}
	return err
}

func printCategoryHeaderAndDependencies(categoryIdx int, category *dependencyCategory) {
	stdout.Out.WriteLine(output.Linef(output.EmojiLightbulb, output.CombineStyles(output.StyleSearchQuery, output.StyleBold), "%d. %s", categoryIdx, category.name))
	stdout.Out.Write("")
	stdout.Out.Write("Checks:")

	for i, dep := range category.dependencies {
		idx := i + 1
		if dep.IsMet() {
			writeSuccessLinef("%d. %s", idx, dep.name)
		} else {
			var printer func(fmtStr string, args ...interface{})
			if dep.onlyTeammates {
				printer = writeWarningLinef
			} else {
				printer = writeFailureLinef
			}

			if dep.err != nil {
				printer("%d. %s: %s", idx, dep.name, dep.err)
			} else {
				printer("%d. %s: %s", idx, dep.name, "check failed")
			}
		}
	}
}

func fixCategoryAutomatically(ctx context.Context, category *dependencyCategory) error {
	// for go through sub dependencies that may be required to fix the dependencies themselves.
	for _, dep := range category.autoFixingDependencies {
		if dep.IsMet() {
			continue
		}
		if err := fixDependencyAutomatically(ctx, dep); err != nil {
			return err
		}
	}
	// now go through the real dependencies
	for _, dep := range category.dependencies {
		if dep.IsMet() {
			continue
		}

		if err := fixDependencyAutomatically(ctx, dep); err != nil {
			return err
		}
	}

	return nil
}

func fixDependencyAutomatically(ctx context.Context, dep *dependency) error {
	writeFingerPointingLinef("Trying my hardest to fix %q automatically...", dep.name)

	cmdStr := dep.InstructionsCommands(ctx)
	if cmdStr == "" {
		return nil
	}
	cmd := usershell.Cmd(ctx, cmdStr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		writeFailureLinef("Failed to run command: %s", err)
		return err
	}

	writeSuccessLinef("Done! %q should be fixed now!", dep.name)

	if dep.requiresSgSetupRestart {
		writeFingerPointingLinef("This command requires restarting of 'sg setup' to pick up the changes.")
		os.Exit(0)
	}

	return nil
}

func fixCategoryManually(ctx context.Context, categoryIdx int, category *dependencyCategory) error {
	for {
		toFix := []int{}

		for i, dep := range category.dependencies {
			if dep.IsMet() {
				continue
			}

			toFix = append(toFix, i)
		}

		if len(toFix) == 0 {
			break
		}

		var idx int

		if len(toFix) == 1 {
			idx = toFix[0]
		} else {
			writeFingerPointingLinef("Which one do you want to fix?")
			var err error
			idx, err = getNumberOutOf(toFix)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}

		dep := category.dependencies[idx]

		stdout.Out.WriteLine(output.Linef(output.EmojiFailure, output.CombineStyles(output.StyleWarning, output.StyleBold), "%s", dep.name))
		stdout.Out.Write("")

		if dep.err != nil {
			stdout.Out.WriteLine(output.Linef("", output.StyleBold, "Encountered the following error:\n\n%s%s\n", output.StyleReset, dep.err))
		}

		stdout.Out.WriteLine(output.Linef("", output.StyleBold, "How to fix:"))

		if dep.instructionsComment != "" {
			if dep.requiresSgSetupRestart {
				// Make sure to highlight the manual fix, if any.
				stdout.Out.Write("")
				stdout.Out.WriteLine(output.Linef(output.EmojiWarningSign, output.StyleYellow, "%s", dep.instructionsComment))
			} else {
				stdout.Out.Write("")
				stdout.Out.Write(dep.instructionsComment)
			}
		}

		// If we don't have anything do run, we simply print instructions to
		// the user
		if dep.InstructionsCommands(ctx) == "" {
			writeFingerPointingLinef("Hit return once you're done")
			waitForReturn()
		} else {
			// Otherwise we print the command(s) and ask the user whether we should run it or not
			stdout.Out.Write("")
			if category.requiresRepository {
				stdout.Out.Writef("Run the following command(s) %sin the 'sourcegraph' repository%s:", output.StyleBold, output.StyleReset)
			} else {
				stdout.Out.Write("Run the following command(s):")
			}
			stdout.Out.Write("")

			stdout.Out.WriteLine(output.Line("", output.CombineStyles(output.StyleBold, output.StyleYellow), strings.TrimSpace(dep.InstructionsCommands(ctx))))

			choice, err := getChoice(map[int]string{
				1: "I'll fix this manually (either by running the command or doing something else)",
				2: "You can run the command for me",
				3: "Go back",
			})
			if err != nil {
				return err
			}

			switch choice {
			case 1:
				writeFingerPointingLinef("Hit return once you're done")
				waitForReturn()
			case 2:
				if err := fixDependencyAutomatically(ctx, dep); err != nil {
					return err
				}
			case 3:
				return nil
			}
		}

		pending := stdout.Out.Pending(output.Linef("", output.StylePending, "Determining status..."))
		for _, dep := range category.dependencies {
			dep.Update(ctx)
		}
		pending.Destroy()

		printCategoryHeaderAndDependencies(categoryIdx, category)
	}

	return nil
}

func removeEntry(s []int, val int) (result []int) {
	for _, e := range s {
		if e != val {
			result = append(result, e)
		}
	}
	return result
}

type dependency struct {
	name string

	check check.CheckFunc

	onlyTeammates bool

	err error

	instructionsComment         string
	instructionsCommands        string
	instructionsCommandsBuilder commandBuilder
	requiresSgSetupRestart      bool
}

func (d *dependency) IsMet() bool { return d.err == nil }

func (d *dependency) InstructionsCommands(ctx context.Context) string {
	if d.instructionsCommandsBuilder != nil {
		return d.instructionsCommandsBuilder.Build(ctx)
	}
	return d.instructionsCommands
}

func (d *dependency) Update(ctx context.Context) {
	d.err = nil
	d.err = d.check(ctx)
}

type dependencyCategory struct {
	name         string
	dependencies []*dependency

	autoFixing             bool
	autoFixingDependencies []*dependency
	requiresRepository     bool
}

func (cat *dependencyCategory) CombinedState() bool {
	for _, dep := range cat.dependencies {
		if !dep.IsMet() {
			return false
		}
	}
	return true
}

func (cat *dependencyCategory) CombinedStateNonTeammates() bool {
	for _, dep := range cat.dependencies {
		if !dep.IsMet() && !dep.onlyTeammates {
			return false
		}
	}
	return true
}

func (cat *dependencyCategory) Update(ctx context.Context) {
	for _, dep := range cat.autoFixingDependencies {
		dep.Update(ctx)
	}
	for _, dep := range cat.dependencies {
		dep.Update(ctx)
	}
}

func getNumberOutOf(numbers []int) (int, error) {
	var strs []string
	var idx = make(map[int]struct{})
	for _, num := range numbers {
		strs = append(strs, fmt.Sprintf("%d", num+1))
		idx[num+1] = struct{}{}
	}

	for {
		fmt.Printf("[%s]: ", strings.Join(strs, ","))
		var num int
		_, err := fmt.Scan(&num)
		if err != nil {
			return 0, err
		}

		if _, ok := idx[num]; ok {
			return num - 1, nil
		}
		fmt.Printf("%d is an invalid choice :( Let's try again?\n", num)
	}
}

func waitForReturn() { fmt.Scanln() }

func getChoice(choices map[int]string) (int, error) {
	for {
		stdout.Out.Write("")
		writeFingerPointingLinef("What do you want to do?")

		for i := 0; i < len(choices); i++ {
			num := i + 1
			desc, ok := choices[num]
			if !ok {
				return 0, errors.Newf("internal error: %d not found in provided choices", i)
			}
			stdout.Out.Writef("%s[%d]%s: %s", output.StyleBold, num, output.StyleReset, desc)
		}

		fmt.Printf("Enter choice: ")

		var s int
		_, err := fmt.Scan(&s)
		if err != nil {
			return 0, err
		}

		if _, ok := choices[s]; ok {
			return s, nil
		}
		writeFailureLinef("Invalid choice")
	}
}

func checkCaddyTrusted(_ context.Context) error {
	certPath, err := caddySourcegraphCertificatePath()
	if err != nil {
		return errors.Wrap(err, "failed to determine path where proxy stores certificates")
	}

	ok, err := pathExists(certPath)
	if !ok || err != nil {
		return errors.New("sourcegraph.test certificate not found. highly likely it's not trusted by system")
	}

	rawCert, err := os.ReadFile(certPath)
	if err != nil {
		return errors.Wrap(err, "could not read certificate")
	}

	cert, err := pemDecodeSingleCert(rawCert)
	if err != nil {
		return errors.Wrap(err, "decoding cert failed")
	}

	if trusted(cert) {
		return nil
	}
	return errors.New("doesn't look like certificate is trusted")
}

// caddyAppDataDir returns the location of the sourcegraph.test certificate
// that Caddy created or would create.
//
// It's copy&pasted&modified from here: https://sourcegraph.com/github.com/caddyserver/caddy@9ee68c1bd57d72e8a969f1da492bd51bfa5ed9a0/-/blob/storage.go?L114
func caddySourcegraphCertificatePath() (string, error) {
	if basedir := os.Getenv("XDG_DATA_HOME"); basedir != "" {
		return filepath.Join(basedir, "caddy"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var appDataDir string
	switch runtime.GOOS {
	case "darwin":
		appDataDir = filepath.Join(home, "Library", "Application Support", "Caddy")
	case "linux":
		appDataDir = filepath.Join(home, ".local", "share", "caddy")
	default:
		return "", errors.Newf("unsupported OS: %s", runtime.GOOS)
	}

	return filepath.Join(appDataDir, "pki", "authorities", "local", "root.crt"), nil
}

func trusted(cert *x509.Certificate) bool {
	chains, err := cert.Verify(x509.VerifyOptions{})
	return len(chains) > 0 && err == nil
}

func pemDecodeSingleCert(pemDER []byte) (*x509.Certificate, error) {
	pemBlock, _ := pem.Decode(pemDER)
	if pemBlock == nil {
		return nil, errors.Newf("no PEM block found")
	}
	if pemBlock.Type != "CERTIFICATE" {
		return nil, errors.Newf("expected PEM block type to be CERTIFICATE, but got '%s'", pemBlock.Type)
	}
	return x509.ParseCertificate(pemBlock.Bytes)
}

type commandBuilder interface {
	Build(context.Context) string
}

type stringCommandBuilder func(context.Context) string

func (l stringCommandBuilder) Build(ctx context.Context) string {
	return l(ctx)
}

var dependencyCategoryAdditionalSgConfiguration = dependencyCategory{
	name:       "Additional sg configuration",
	autoFixing: true,
	dependencies: []*dependency{
		{
			name: "autocompletions",
			check: func(ctx context.Context) error {
				if !usershell.IsSupportedShell(ctx) {
					return nil // dont do setup
				}
				sgHome, err := root.GetSGHomePath()
				if err != nil {
					return err
				}
				shell := usershell.ShellType(ctx)
				autocompletePath := usershell.AutocompleteScriptPath(sgHome, shell)
				if _, err := os.Stat(autocompletePath); err != nil {
					return errors.Wrapf(err, "autocomplete script for shell %s not found", shell)
				}

				shellConfig := usershell.ShellConfigPath(ctx)
				conf, err := os.ReadFile(shellConfig)
				if err != nil {
					return err
				}
				if !strings.Contains(string(conf), autocompletePath) {
					return errors.Newf("autocomplete script %s not found in shell config %s",
						autocompletePath, shellConfig)
				}
				return nil
			},
			instructionsCommandsBuilder: stringCommandBuilder(func(ctx context.Context) string {
				sgHome, err := root.GetSGHomePath()
				if err != nil {
					return fmt.Sprintf("echo %s && exit 1", err.Error())
				}

				var commands []string

				shell := usershell.ShellType(ctx)
				if shell == "" {
					return "echo 'Failed to detect shell type' && exit 1"
				}
				autocompleteScript := usershell.AutocompleteScripts[shell]
				autocompletePath := usershell.AutocompleteScriptPath(sgHome, shell)
				commands = append(commands,
					fmt.Sprintf(`echo "Writing autocomplete script to %s"`, autocompletePath),
					fmt.Sprintf(`echo '%s' > %s`, autocompleteScript, autocompletePath))

				shellConfig := usershell.ShellConfigPath(ctx)
				if shellConfig == "" {
					return "echo 'Failed to detect shell config path' && exit 1"
				}
				conf, err := os.ReadFile(shellConfig)
				if err != nil {
					return fmt.Sprintf("echo %s && exit 1", err.Error())
				}
				if !strings.Contains(string(conf), autocompletePath) {
					commands = append(commands,
						fmt.Sprintf(`echo "Adding configuration to %s"`, shellConfig),
						fmt.Sprintf(`echo "PROG=sg source %s" >> %s`,
							autocompletePath, shellConfig))
				}

				return strings.Join(commands, "\n")
			}),
		},
	},
}
