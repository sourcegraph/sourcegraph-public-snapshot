// Package linters defines all available linters.
package linters

import (
	"bytes"
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/run"
	"go.bobheadxi.dev/streamline/pipeline"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/check"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Target = check.Category[*repo.State]

type linter = check.Check[*repo.State]

// Targets lists all available linter targets. Each target consists of multiple linters.
//
// These should align with the names in 'dev/ci/internal/ci/changed'
var Targets = []Target{
	{
		Name:        "urls",
		Description: "Check for broken urls in the codebase",
		Checks: []*linter{
			runScript("Broken urls", "dev/check/broken-urls.bash"),
		},
	},
	{
		Name:        "go",
		Description: "Check go code for linting errors, forbidden imports, generated files, etc",
		Checks: []*linter{
			timeCheck(goGenerateLinter),
			onlyLocal(goDBConnImport),
			onlyLocal(noLocalHost),
			timeCheck(lintGoDirectives()),
			timeCheck(lintLoggingLibraries()),
			onlyLocal(lintTracingLibraries()),
			timeCheck(goModGuards()),
			onlyLocal(lintSGExit()),
		},
	},
	{
		Name:        "graphql",
		Description: "Checks the graphql code for linting errors [bazel]",
		Checks: []*linter{
			onlyLocal(bazelExec("graphql schema lint (bazel)", "test //cmd/frontend/graphqlbackend:graphql_schema_lint_test")),
		},
	},
	{
		Name:        "docs",
		Description: "Documentation checks",
		Checks: []*linter{
			onlyLocal(bazelExec("Docsite lint (bazel)", "test //doc:test")),
			timeCheck(docChangesLint()),
		},
	},
	{
		Name:        "dockerfiles",
		Description: "Check Dockerfiles for Sourcegraph best practices",
		Checks: []*linter{
			// TODO move to pre-commit
			timeCheck(hadolint()),
		},
	},
	{
		Name:        "client",
		Description: "Check client code for linting errors, forbidden imports, etc",
		Checks: []*linter{
			timeCheck(inlineTemplates),
			timeCheck(runScriptSerialized("pnpm dedupe", "dev/check/pnpm-deduplicate.sh")),
			// we only run this linter locally, since on CI it has it's own job
			onlyLocal(runScriptSerialized("pnpm lint:js:web", "dev/ci/pnpm-run.sh lint:js:web")),
			timeCheck(checkUnversionedDocsLinks()),
		},
	},
	{
		Name:        "shell",
		Description: "Check shell code for linting errors, formatting, etc",
		Checks: []*linter{
			timeCheck(shFmt),
			timeCheck(shellCheck),
			timeCheck(bashSyntax),
		},
	},
	{
		Name:        "protobuf",
		Description: "Check protobuf code for linting errors, formatting, etc",
		Checks: []*linter{
			timeCheck(bufFormat),
			timeCheck(bufGenerate),
			timeCheck(bufLint),
		},
	},
	{
		Name:        "bazel generated",
		Description: "Ensures documentation and source generated by Bazel is up to date",
		Checks: []*linter{
			onlyLocal(bazelExec("bazel generate files", "run //:write_all_generated")),
		},
	},
	Formatting,
}

var Formatting = Target{
	Name:        "format",
	Description: "Check client code and docs for formatting errors",
	Checks: []*linter{
		timeCheck(prettier),
	},
}

func onlyLocal(l *linter) *linter {
	if os.Getenv("CI") == "true" {
		l.Enabled = func(ctx context.Context, args *repo.State) error {
			return errors.New("check is disabled in CI")
		}
	}
	return l
}

// runScript creates check that runs the given script from the root of sourcegraph/sourcegraph.
func runScript(name string, script string) *linter {
	return &linter{
		Name: name,
		Check: func(ctx context.Context, out *std.Output, args *repo.State) error {
			return root.Run(run.Bash(ctx, script)).StreamLines(out.Write)
		},
	}
}

var runScriptSerializedMu sync.Mutex

// runScriptSerialized is exactly like runScript, but ensure that all the check functions
// are run serially by acquiring a lock.
//
// This is useful for pnpm for examples, as some tasks might end up writing to the same files
// concurrently, leading to race conditions and thus CI failures.
func runScriptSerialized(name string, script string) *linter {
	return &linter{
		Name: name,
		Check: func(ctx context.Context, out *std.Output, args *repo.State) error {
			event := honey.FromContext(ctx)

			t1 := time.Now()
			runScriptSerializedMu.Lock()
			t2 := time.Since(t1)
			event.AddField("pnpm_lock_duration", t2.Seconds())
			event.AddField("pnpm_lock_duration_ms", t2.Milliseconds())
			defer runScriptSerializedMu.Unlock()
			return root.Run(run.Bash(ctx, script)).StreamLines(out.Write)
		},
	}
}

// runCheck creates a check that runs the given check func.
func runCheck(name string, check check.CheckAction[*repo.State]) *linter {
	return &linter{
		Name:  name,
		Check: check,
	}
}

func bazelExec(name, args string) *linter {
	cmd := []string{"bazel"}
	cmd = append(cmd, strings.Split(args, " ")...)
	return &linter{
		Name: name,
		Check: func(ctx context.Context, out *std.Output, args *repo.State) error {
			return root.Run(run.Cmd(ctx, cmd...)).StreamLines(out.Write)
		},
	}
}

// pnpmInstallFilter is a pipeline that filters out all the warning junk that pnpm install
// emits that seem inconsequential, for example:
//
//	warning "@storybook/addon-foo > react-test-renderer@16.14.0" has incorrect peer dependency "react@^16.14.0".
//	warning " > @storybook/react@6.5.9" has unmet peer dependency "require-from-string@^2.0.2".
//	warning "@storybook/react > react-element-to-jsx-string@14.3.4" has incorrect peer dependency "react@^0.14.8 || ^15.0.1 || ^16.0.0 || ^17.0.1".
//	warning " > @testing-library/react-hooks@8.0.0" has incorrect peer dependency "react@^16.9.0 || ^17.0.0".
//	warning Workspaces can only be enabled in private projects.
//	warning Workspaces can only be enabled in private projects.
func pnpmInstallFilter() pipeline.Pipeline {
	return pipeline.Filter(func(line []byte) bool { return !bytes.Contains(line, []byte("warning")) })
}

// disabled can be used to mark a category or check as disabled.
func disabled(reason string) check.EnableFunc[*repo.State] {
	return func(context.Context, *repo.State) error {
		return errors.Newf("disabled: %s", reason)
	}
}
