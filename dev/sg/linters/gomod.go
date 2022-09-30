package linters

import (
	"context"
	"strings"

	"github.com/Masterminds/semver"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func goModGuards() *linter {
	const header = "go.mod version guards"

	var maxVersions = map[string]*semver.Version{
		// Any version past this version is not yet released in any version of Alertmanager,
		// and causes incompatibility in prom-wrapper.
		//
		// https://github.com/sourcegraph/zoekt/pull/330#issuecomment-1116857568
		"github.com/prometheus/common": semver.MustParse("v0.32.1"),
	}

	return &linter{
		Name: header,
		Enabled: func(ctx context.Context, args *repo.State) error {
			if len(maxVersions) == 0 {
				return errors.New("no version restrictions declared")
			}
			return nil
		},
		Check: func(ctx context.Context, out *std.Output, state *repo.State) error {
			diff, err := state.GetDiff("go.mod")
			if err != nil {
				return err
			}
			if len(diff) == 0 {
				out.Verbose("No go.mod changes detected!")
				return nil
			}

			failedLibs := map[string]error{}
			for _, hunk := range diff["go.mod"] {
				for _, l := range hunk.AddedLines {
					parts := strings.Split(strings.TrimSpace(l), " ")
					switch len(parts) {
					// Dependencies: 'lib v1.2.3'
					case 2:
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
								failedLibs[lib] = errors.Wrapf(err, "invalid version", version)
								continue
							}
							if v.GreaterThan(maxVersion) {
								failedLibs[lib] = errors.Newf("must not exceed version %s", maxVersion)
							}
						}

					// Overrides: 'lib => lib v1.2.3'
					case 4:
						var (
							replaced = parts[0]
							lib      = parts[2]
							version  = parts[3]
						)
						if replaced != lib {
							continue
						}
						if !strings.HasPrefix(version, "v") {
							continue
						}

						if maxVersion := maxVersions[lib]; maxVersion != nil {
							v, err := semver.NewVersion(version)
							if err != nil {
								failedLibs[lib] = errors.Wrapf(err, "invalid version", version)
								continue
							}
							if !v.GreaterThan(maxVersion) {
								// reset error if override enforces a safe verison
								failedLibs[lib] = nil
							}
						}
					}
				}
			}

			var errs error
			for lib, err := range failedLibs {
				errs = errors.Append(errs, errors.Wrap(err, lib))
			}
			return errs
		},
	}
}
