package linters

import (
	"context"
	"strings"

	"github.com/Masterminds/semver"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type goModVersionsLinter struct {
	maxVersions map[string]*semver.Version
}

func (l *goModVersionsLinter) Check(ctx context.Context, state *repo.State) *lint.Report {
	const header = "go.mod version guards"

	if len(l.maxVersions) == 0 {
		return &lint.Report{Header: header, Output: "No guards currently defined"}
	}

	diff, err := state.GetDiff("go.mod")
	if err != nil {
		return &lint.Report{Header: header, Err: err}
	}
	if len(diff) == 0 {
		return &lint.Report{Header: header, Output: "No go.mod changes detected!"}
	}

	var errs error
	for _, hunk := range diff["go.mod"] {
		for _, line := range hunk.AddedLines {
			parts := strings.Split(strings.TrimSpace(line), " ")
			if len(parts) != 2 {
				continue
			}
			var (
				lib     = parts[0]
				version = parts[1]
			)
			if !strings.HasPrefix(version, "v") {
				continue
			}
			if maxVersion := l.maxVersions[lib]; maxVersion != nil {
				v, err := semver.NewVersion(version)
				if err != nil {
					errs = errors.Append(errs, errors.Wrapf(err, "dependency %s has invalid version", lib))
					continue
				}
				if v.GreaterThan(maxVersion) {
					errs = errors.Append(errs, errors.Newf("dependency %s must not exceed version %s",
						lib, maxVersion))
				}
			}
		}
	}

	return &lint.Report{
		Header: header,
		Output: func() string {
			if errs != nil {
				return strings.TrimSpace(errs.Error())
			}
			return ""
		}(),
		Err: errs,
	}
}
