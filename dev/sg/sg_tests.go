package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/cliutil/completions"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func init() {
	postInitHooks = append(postInitHooks, func(cmd *cli.Context) {
		// Create 'sg test' help text after flag (and config) initialization
		testCommand.Description = constructTestCmdLongHelp()
	})
}

var testCommand = &cli.Command{
	Name:      "test",
	ArgsUsage: "<testsuite>",
	Usage:     "Run the given test suite",
	UsageText: `
# Run different test suites:
sg test backend
sg test backend-integration
sg test client
sg test web-e2e

# List available test suites:
sg test -help

# Arguments are passed along to the command
sg test backend-integration -run TestSearch
`,
	Category: CategoryDev,
	BashComplete: completions.CompleteOptions(func() (options []string) {
		config, _ := getConfig()
		if config == nil {
			return
		}
		for name := range config.Tests {
			options = append(options, name)
		}
		return
	}),
	Action: testExec,
	Subcommands: []*cli.Command{
		quarantineCommand,
	},
}

var quarantineCommand = &cli.Command{
	Name:      "quarantine",
	ArgsUsage: "<bazel-target>",
	Usage:     "quarantine the given bazel test target which removes it from being executed with the main tests",
	UsageText: `
# Quarantine the test target //dev/build-tracker
sg test quarantine //dev/build-tracker

# Remove the test target from quarantine which allows it to be executed with the other main tests
sg test quarantine --remove //dev/build-tracker
`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "remove",
		},
	},
	Action: quarantineTarget,
}

const QuarantineFile = "test_quarantine.json"

func testExec(ctx *cli.Context) error {
	config, err := getConfig()
	if err != nil {
		return err
	}

	args := ctx.Args().Slice()
	if len(args) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "No test suite specified"))
		return flag.ErrHelp
	}

	cmd, ok := config.Tests[args[0]]
	if !ok {
		std.Out.WriteLine(output.Styledf(output.StyleWarning, "ERROR: test suite %q not found :(", args[0]))
		return flag.ErrHelp
	}

	return run.Test(ctx.Context, cmd, args[1:], config.Env)
}

func constructTestCmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, "Testsuites are defined in sg configuration.")

	// Attempt to parse config to list available testsuites, but don't fail on
	// error, because we should never error when the user wants --help output.
	config, err := getConfig()
	if err != nil {
		out.Write([]byte("\n"))
		// Do not treat error message as a format string
		std.NewOutput(&out, false).WriteWarningf("%s", err.Error())
		return out.String()
	}

	fmt.Fprintf(&out, "\n\n")
	fmt.Fprintf(&out, "Available testsuites in `%s`:\n", configFile)
	fmt.Fprintf(&out, "\n")

	var names []string
	for name := range config.Tests {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Fprint(&out, "* "+strings.Join(names, "\n* "))

	return out.String()
}

type quarantine struct {
	Tests []*quarantineTest `json:"tests"`
}

type quarantineTest struct {
	Timestamp time.Time `json:"timestamp"`
	Target    string    `json:"target"`
}

func newQuarantineTest(target string) *quarantineTest {
	now := time.Now().UTC()
	return &quarantineTest{
		Timestamp: now,
		Target:    target,
	}
}

func loadQuarantinedTests(path string) (*quarantine, error) {
	fd, err := os.Open(path)
	if os.IsNotExist(err) {
		return &quarantine{
			Tests: make([]*quarantineTest, 0),
		}, nil
	} else if err != nil {
		return nil, err
	}
	defer fd.Close()

	var q quarantine
	err = json.NewDecoder(fd).Decode(&q)
	if err != nil {
		return nil, err
	}
	return &q, err
}

func writeQuarantinedTests(path string, quarantine *quarantine) error {
	fd, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fd.Close()

	return json.NewEncoder(fd).Encode(quarantine)
}

func validTarget(ctx context.Context, target string) bool {
	cmd := exec.CommandContext(ctx, "bazel", "query", target)

	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func validateTargets(ctx context.Context, pending output.Pending, targets []string) ([]string, []string) {
	valid := make([]string, 0)
	invalid := make([]string, 0)
	for _, target := range targets {
		pending.Updatef("checking target %s", target)
		if strings.HasSuffix(target, "...") {
			pending.WriteLine(output.Emoji("⚠️", fmt.Sprintf("%s is not fully qualified", target)))
			invalid = append(invalid, fmt.Sprintf("%s is not fully qualified - please provide a fully qualified target", target))
		} else if validTarget(ctx, target) {
			pending.WriteLine(output.Emoji("☑️", fmt.Sprintf("%s is valid", target)))
			valid = append(valid, target)
		} else {
			pending.WriteLine(output.Emoji("⚠️", fmt.Sprintf("%s is invalid", target)))
			invalid = append(invalid, target)
		}
	}

	return valid, invalid
}

func addTargets(qMap map[string]*quarantineTest, targets ...string) error {
	for _, target := range targets {
		if _, ok := qMap[target]; ok {
			std.Out.WriteWarningf("%s is already quarantined", target)
		} else {
			qMap[target] = newQuarantineTest(target)
			std.Out.WriteSuccessf("%s added to quarantine", target)
		}
	}
	return nil
}

func removeTargets(qMap map[string]*quarantineTest, targets ...string) error {
	for _, target := range targets {
		if _, ok := qMap[target]; ok {
			delete(qMap, target)
			std.Out.WriteSuccessf("%s removed from quarantine", target)
		} else {
			std.Out.WriteWarningf("%s not found in quarantine", target)
		}
	}
	return nil
}

func toQuarantineMap(tests *quarantine) map[string]*quarantineTest {
	result := make(map[string]*quarantineTest, 0)
	for _, test := range tests.Tests {
		result[test.Target] = test
	}
	return result
}

func quarantineTarget(ctx *cli.Context) error {
	dirRoot, err := root.RepositoryRoot()
	if err != nil {
		return errors.Wrap(err, "failed to load quarantine file")
	}
	quarantinePath := filepath.Join(dirRoot, QuarantineFile)
	tests, err := loadQuarantinedTests(quarantinePath)
	if err != nil {
		return err
	}

	qMap := toQuarantineMap(tests)
	targets := ctx.Args().Slice()
	pending := std.Out.Pending(output.Emojif("🔍", "Validating targets by running 'bazel query <target>'"))
	targets, invalid := validateTargets(ctx.Context, pending, targets)
	if len(invalid) > 0 {
		pending.Complete(output.Emoji("🚧", "Cannot add targets to quarantine since some of them are invalid"))
		return errors.Newf("The following targets are invalid according to bazel: %s", strings.Join(invalid, " "))
	}
	pending.Complete(output.Emoji("✅", "All targets are valid"))

	if ctx.Bool("remove") {
		removeTargets(qMap, targets...)
	} else if err := addTargets(qMap, targets...); err != nil {
		return err
	}

	tests.Tests = make([]*quarantineTest, 0, len(qMap))
	for _, v := range qMap {
		tests.Tests = append(tests.Tests, v)
	}

	err = writeQuarantinedTests(quarantinePath, tests)
	if err == nil {
		std.Out.WriteSuggestionf("Quarantine updated with new targets. Please commit the updated quarantine!")
	}
	return err
}
