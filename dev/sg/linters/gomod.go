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

	var goModFiles = map[string]map[string]*semver.Version{
		"go.mod": {
			// Any version past this version is not yet released in any version of Alertmanager,
			// and causes incompatibility in prom-wrapper.
			//
			// https://github.com/sourcegraph/zoekt/pull/330#issuecomment-1116857568
			"github.com/prometheus/common": semver.MustParse("v0.32.1"),
			// Disallow imports of controller-cdktf, which is definitely not for use
			// in Sourcegraph.
			"github.com/sourcegraph/controller-cdktf": nil,
		},
		"monitoring/go.mod": {
			// See above
			"github.com/prometheus/common": semver.MustParse("v0.32.1"),
			// Disallow imports of 'github.com/sourcegraph/sourcegraph'
			"github.com/sourcegraph/sourcegraph": nil,
		},
		"lib/go.mod": {
			// Disallow imports of 'github.com/sourcegraph/sourcegraph'
			"github.com/sourcegraph/sourcegraph": nil,
		},
	}

	var lintGoMod = func(diffHunks []repo.DiffHunk, maxVersions map[string]*semver.Version) error {
		failedLibs := map[string]error{}
		for _, hunk := range diffHunks {
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
					maxVersion, hasContraint := maxVersions[lib]
					if hasContraint {
						if maxVersion != nil {
							v, err := semver.NewVersion(version)
							if err != nil {
								failedLibs[lib] = errors.Wrapf(err, "invalid version", version)
								continue
							}
							if v.GreaterThan(maxVersion) {
								failedLibs[lib] = errors.Newf("must not exceed version %s", maxVersion)
							}
						} else {
							failedLibs[lib] = errors.New("forbidden import")
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
	}

	return &linter{
		Name: header,
		Check: func(ctx context.Context, out *std.Output, state *repo.State) error {
			var errs error

			for file, maxVersions := range goModFiles {
				diff, err := state.GetDiff(file)
				if err != nil {
					return err
				}
				if len(diff) == 0 {
					out.Verbosef("%s: no go.mod changes detected!", file)
					return nil
				}

				if err := lintGoMod(diff[file], maxVersions); err != nil {
					errs = errors.Append(errs, errors.Wrap(err, file))
				}
			}

			return errs
		},
	}
}
