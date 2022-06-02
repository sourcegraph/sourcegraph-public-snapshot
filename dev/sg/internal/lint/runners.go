package lint

import (
	"context"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

type scriptRunner struct {
	header string
	script string
}

// ScriptCheck runs the given script from the root of sourcegraph/sourcegraph as a check.
func ScriptCheck(header string, script string) Linter {
	return &scriptRunner{header: header, script: script}
}

func (s *scriptRunner) Check(ctx context.Context, state *repo.State) *Report {
	out, err := root.Run(run.Bash(ctx, s.script)).String()
	return &Report{
		Header: s.header,
		Output: out,
		Err:    err,
	}
}

// CheckFunc can be used to build simple linter checks.
type CheckFunc func(ctx context.Context, state *repo.State) *Report

type checkFuncRunner struct{ check CheckFunc }

// FuncCheck is a Linter that executes the given runner as a check.
func FuncCheck(check CheckFunc) Linter {
	return &checkFuncRunner{check: check}
}

func (f *checkFuncRunner) Check(ctx context.Context, state *repo.State) *Report {
	return f.check(ctx, state)
}
