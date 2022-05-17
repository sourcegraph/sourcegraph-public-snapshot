package linters

import (
	"context"
	"strings"

	"github.com/Masterminds/semver"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func goModGuards() lint.Runner {
	const header = "go.mod version guards"

	var maxVersions = map[string]*semver.Version{
		// Any version past this version is not yet released in any version of Alertmanager,
		// and causes incompatibility in prom-wrapper.
		//
		// https://github.com/sourcegraph/zoekt/pull/330#issuecomment-1116857568
		"github.com/prometheus/common": semver.MustParse("v0.32.1"),
	}

	if len(maxVersions) == 0 {
		return func(ctx context.Context, s *repo.State) *lint.Report {
			return &lint.Report{Header: header, Output: "No guards currently defined"}
		}
	}

	return func(ctx context.Context, s *repo.State) *lint.Report {
		diff, err := s.GetDiff("go.mod")
		if err != nil {
			return &lint.Report{Header: header, Err: err}
		}
		if len(diff) == 0 {
			return &lint.Report{Header: header, Output: "No go.mod changes detected!"}
		}

		var errs error
		for _, hunk := range diff["go.mod"] {
			for _, l := range hunk.AddedLines {
				parts := strings.Split(strings.TrimSpace(l), " ")
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
				if maxVersion := maxVersions[lib]; maxVersion != nil {
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
}
