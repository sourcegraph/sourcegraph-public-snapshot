package lint

import (
	"context"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
)

// Target denotes a set of linter tasks that can be run by `sg lint`
type Target struct {
	Name string
	Help string

	// Linters can be linters that only support checks, or linters that support
	// automatically fixing issues (FixableLinter).
	Linters []Linter
}

// Linter is a linter runner. It can make programmatic checks, call out to a bash script,
// or anything you want, and should return a report with helpful feedback for the user to
// act upon.
//
// To build a linter that can be fixed as well by 'sg lint -fix', you can implement the
// FixableLinter interface.
//
// Linter can be tested by providing a mock state with repo.NewMockState().
type Linter interface {
	Check(context.Context, *repo.State) *Report
}

// FixableLinter is a linter runner that can also fix lint issues automatically.
type FixableLinter interface {
	Linter

	// Fix, if implemented, should modify the workspace such that linter issues are all
	// fixed, and return an error if not all issues are fixed.
	//
	// Note that repo.State only denotes the initial state of the repository.
	Fix(context.Context, *repo.State) *Report
}

// Fixable returns a FixableLinter if the given Linter is fixable.
func Fixable(l Linter) (FixableLinter, bool) {
	fx, ok := l.(FixableLinter)
	return fx, ok
}

// Report describes the result of a linter runner.
type Report struct {
	// Header is the title for this report.
	Header string
	// Output will be expanded on failure. Optional if Err is provided.
	Output string
	// Err indicates a failure has been detected, and is mainly used to detect if an the
	// check has failed - its contents are only presented when Output is not provided.
	Err error
}

// Summary renders a summary of the report based on Output or Err.
func (r *Report) Summary() string {
	if r.Output == "" && r.Err != nil {
		return r.Err.Error()
	}
	return r.Output
}
