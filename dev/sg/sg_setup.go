package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/interrupt"
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
		std.Out.WriteLine(output.Styled(output.StyleWarning, "'sg setup' currently only supports macOS and Linux"))
		return NewEmptyExitErr(1)
	}

	currentOS := runtime.GOOS
	if overridesOS, ok := os.LookupEnv("SG_FORCE_OS"); ok {
		currentOS = overridesOS
	}

	ctx, err := usershell.Context(ctx)
	if err != nil {
		return err
	}

	// Before a user interrupts and exits, let them know that they may need to take
	// additiional actions.
	interrupt.Register(func() {
		std.Out.WriteAlertf("\n💡 You may need to restart your shell for the changes to work in this terminal.")
		std.Out.WriteAlertf("   Close this terminal and open a new one or type the following command and press ENTER: %s", filepath.Base(usershell.ShellPath(ctx)))
	})

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
		std.Out.ClearScreen()

		printSgSetupWelcomeScreen()
		std.Out.WriteAlertf("                INFO: You can quit any time by typing ctrl-c\n")

		for i, category := range categories {
			idx := i + 1

			if category.requiresRepository && !inRepo {
				std.Out.WriteSkippedf("%d. %s %s[SKIPPED. Requires 'sg setup' to be run in 'sourcegraph' repository]%s", idx, category.name, output.StyleBold, output.StyleReset)
				skipped = append(skipped, idx)
				failed = removeEntry(failed, i)
				continue
			}

			pending := std.Out.Pending(output.Styledf(output.StylePending, "%d. %s - Determining status...", idx, category.name))
			category.Update(ctx)
			pending.Destroy()

			if combined := category.CombinedState(); combined {
				std.Out.WriteSuccessf("%d. %s", idx, category.name)
				failed = removeEntry(failed, i)
			} else {
				nonTeammateState := category.CombinedStateNonTeammates()
				if nonTeammateState {
					std.Out.WriteWarningf("%d. %s", idx, category.name)
					teammateFailed = append(skipped, idx)
				} else {
					std.Out.WriteFailuref("%d. %s", idx, category.name)
				}
			}
		}

		if len(failed) == 0 && len(teammateFailed) == 0 {
			if len(skipped) == 0 && len(teammateFailed) == 0 {
				std.Out.Write("")
				std.Out.WriteLine(output.Linef(output.EmojiOk, output.StyleBold, "Everything looks good! Happy hacking!"))
			}

			if len(skipped) != 0 {
				std.Out.Write("")
				std.Out.WriteWarningf("Some checks were skipped because 'sg setup' is not run in the 'sourcegraph' repository.")
				std.Out.WriteSuggestionf("Restart 'sg setup' in the 'sourcegraph' repository to continue.")
			}

			return nil
		}

		std.Out.Write("")

		if len(teammateFailed) != 0 && len(failed) == len(teammateFailed) {
			std.Out.WriteWarningf("Some checks that are only relevant for Sourcegraph teammates failed.\nIf you're not a Sourcegraph teammate you're good to go. Hit Ctrl-C.\n\nIf you're a Sourcegraph teammate: which one do you want to fix?")
		} else {
			std.Out.WriteWarningf("Some checks failed. Which one do you want to fix?")
		}

		idx, err := getNumberOutOf(all)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		selectedCategory := categories[idx]

		std.Out.ClearScreen()

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
			std.Out.ClearScreen()
			err = fixCategoryAutomatically(ctx, category)
		}
	case 3:
		return nil
	}
	return err
}

func printCategoryHeaderAndDependencies(categoryIdx int, category *dependencyCategory) {
	std.Out.WriteLine(output.Linef(output.EmojiLightbulb, output.CombineStyles(output.StyleSearchQuery, output.StyleBold), "%d. %s", categoryIdx, category.name))
	std.Out.Write("")
	std.Out.Write("Checks:")

	for i, dep := range category.dependencies {
		idx := i + 1
		if dep.IsMet() {
			std.Out.WriteSuccessf("%d. %s", idx, dep.name)
		} else {
			var printer func(fmtStr string, args ...any)
			if dep.onlyTeammates {
				printer = std.Out.WriteWarningf
			} else {
				printer = std.Out.WriteFailuref
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
	std.Out.WriteNoticef("Trying my hardest to fix %q automatically...", dep.name)

	cmdStr := dep.InstructionsCommands(ctx)
	if cmdStr == "" {
		return nil
	}
	cmd := usershell.Cmd(ctx, cmdStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		std.Out.WriteFailuref("Failed to run command: %s", err)
		return err
	}

	std.Out.WriteSuccessf("Done! %q should be fixed now!", dep.name)

	if dep.requiresSgSetupRestart {
		std.Out.WriteNoticef("This command requires restarting of 'sg setup' to pick up the changes.")
		return nil
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
			std.Out.WriteNoticef("Which one do you want to fix?")
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

		std.Out.WriteLine(output.Linef(output.EmojiFailure, output.CombineStyles(output.StyleWarning, output.StyleBold), "%s", dep.name))
		std.Out.Write("")

		if dep.err != nil {
			std.Out.WriteLine(output.Styledf(output.StyleBold, "Encountered the following error:\n\n%s%s\n", output.StyleReset, dep.err))
		}

		std.Out.WriteLine(output.Styled(output.StyleBold, "How to fix:"))

		if dep.instructionsComment != "" {
			if dep.requiresSgSetupRestart {
				// Make sure to highlight the manual fix, if any.
				std.Out.Write("")
				std.Out.WriteLine(output.Linef(output.EmojiWarningSign, output.StyleYellow, "%s", dep.instructionsComment))
			} else {
				std.Out.Write("")
				std.Out.Write(dep.instructionsComment)
			}
		}

		// If we don't have anything do run, we simply print instructions to
		// the user
		if dep.InstructionsCommands(ctx) == "" {
			std.Out.WriteNoticef("Hit return once you're done")
			waitForReturn()
		} else {
			// Otherwise we print the command(s) and ask the user whether we should run it or not
			std.Out.Write("")
			if category.requiresRepository {
				std.Out.Writef("Run the following command(s) %sin the 'sourcegraph' repository%s:", output.StyleBold, output.StyleReset)
			} else {
				std.Out.Write("Run the following command(s):")
			}
			std.Out.Write("")

			std.Out.WriteLine(output.Line("", output.CombineStyles(output.StyleBold, output.StyleYellow), strings.TrimSpace(dep.InstructionsCommands(ctx))))

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
				std.Out.WriteNoticef("Hit return once you're done")
				waitForReturn()
			case 2:
				if err := fixDependencyAutomatically(ctx, dep); err != nil {
					return err
				}
			case 3:
				return nil
			}
		}

		pending := std.Out.Pending(output.Styled(output.StylePending, "Determining status..."))
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
	name               string
	dependencies       []*dependency
	requiresRepository bool

	// autoFixingDependencies are only accounted for it the user asks to fix the category.
	// Otherwise, they'll never be checked nor print an error, because the only thing that
	// matters to run Sourcegraph are the final dependencies defined in the dependencies
	// field itself.
	autoFixingDependencies []*dependency
	autoFixing             bool
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
		std.Out.Write("")
		std.Out.WriteNoticef("What do you want to do?")

		for i := 0; i < len(choices); i++ {
			num := i + 1
			desc, ok := choices[num]
			if !ok {
				return 0, errors.Newf("internal error: %d not found in provided choices", i)
			}
			std.Out.Writef("%s[%d]%s: %s", output.StyleBold, num, output.StyleReset, desc)
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
		std.Out.WriteFailuref("Invalid choice")
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

func printSgSetupWelcomeScreen() {
	genLine := func(style output.Style, content string) string {
		return fmt.Sprintf("%s%s%s", output.CombineStyles(output.StyleBold, style), content, output.StyleReset)
	}

	boxContent := func(content string) string { return genLine(output.StyleWhiteOnPurple, content) }
	shadow := func(content string) string { return genLine(output.StyleGreyBackground, content) }

	std.Out.Write(boxContent(`┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ sg ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓`))
	std.Out.Write(boxContent(`┃            _       __     __                             __                ┃`))
	std.Out.Write(boxContent(`┃           | |     / /__  / /________  ____ ___  ___     / /_____           ┃`) + shadow(`  `))
	std.Out.Write(boxContent(`┃           | | /| / / _ \/ / ___/ __ \/ __ '__ \/ _ \   / __/ __ \          ┃`) + shadow(`  `))
	std.Out.Write(boxContent(`┃           | |/ |/ /  __/ / /__/ /_/ / / / / / /  __/  / /_/ /_/ /          ┃`) + shadow(`  `))
	std.Out.Write(boxContent(`┃           |__/|__/\___/_/\___/\____/_/ /_/ /_/\___/   \__/\____/           ┃`) + shadow(`  `))
	std.Out.Write(boxContent(`┃                                           __              __               ┃`) + shadow(`  `))
	std.Out.Write(boxContent(`┃                  ___________   ________  / /___  ______  / /               ┃`) + shadow(`  `))
	std.Out.Write(boxContent(`┃                 / ___/ __  /  / ___/ _ \/ __/ / / / __ \/ /                ┃`) + shadow(`  `))
	std.Out.Write(boxContent(`┃                (__  ) /_/ /  (__  )  __/ /_/ /_/ / /_/ /_/                 ┃`) + shadow(`  `))
	std.Out.Write(boxContent(`┃               /____/\__, /  /____/\___/\__/\__,_/ .___(_)                  ┃`) + shadow(`  `))
	std.Out.Write(boxContent(`┃                    /____/                      /_/                         ┃`) + shadow(`  `))
	std.Out.Write(boxContent(`┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛`) + shadow(`  `))
	std.Out.Write(`  ` + shadow(`                                                                              `))
	std.Out.Write(`  ` + shadow(`                                                                              `))
}
