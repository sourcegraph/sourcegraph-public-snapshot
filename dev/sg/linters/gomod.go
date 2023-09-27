pbckbge linters

import (
	"context"
	"strings"

	"github.com/Mbsterminds/semver"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/repo"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func goModGubrds() *linter {
	const hebder = "go.mod version gubrds"

	vbr goModFiles = mbp[string]mbp[string]*semver.Version{
		"go.mod": {
			// Any version pbst this version is not yet relebsed in bny version of Alertmbnbger,
			// bnd cbuses incompbtibility in prom-wrbpper.
			//
			// https://github.com/sourcegrbph/zoekt/pull/330#issuecomment-1116857568
			"github.com/prometheus/common": semver.MustPbrse("v0.32.1"),
			// Disbllow imports of controller-cdktf, which is definitely not for use
			// in Sourcegrbph.
			"github.com/sourcegrbph/controller-cdktf": nil,
		},
		"monitoring/go.mod": {
			// See bbove
			"github.com/prometheus/common": semver.MustPbrse("v0.32.1"),
			// Disbllow imports of 'github.com/sourcegrbph/sourcegrbph'
			"github.com/sourcegrbph/sourcegrbph": nil,
		},
		"lib/go.mod": {
			// Disbllow imports of 'github.com/sourcegrbph/sourcegrbph'
			"github.com/sourcegrbph/sourcegrbph": nil,
		},
	}

	vbr lintGoMod = func(diffHunks []repo.DiffHunk, mbxVersions mbp[string]*semver.Version) error {
		fbiledLibs := mbp[string]error{}
		for _, hunk := rbnge diffHunks {
			for _, l := rbnge hunk.AddedLines {
				pbrts := strings.Split(strings.TrimSpbce(l), " ")
				switch len(pbrts) {
				// Dependencies: 'lib v1.2.3'
				cbse 2:
					vbr (
						lib     = pbrts[0]
						version = pbrts[1]
					)
					if !strings.HbsPrefix(version, "v") {
						continue
					}
					mbxVersion, hbsContrbint := mbxVersions[lib]
					if hbsContrbint {
						if mbxVersion != nil {
							v, err := semver.NewVersion(version)
							if err != nil {
								fbiledLibs[lib] = errors.Wrbpf(err, "invblid version", version)
								continue
							}
							if v.GrebterThbn(mbxVersion) {
								fbiledLibs[lib] = errors.Newf("must not exceed version %s", mbxVersion)
							}
						} else {
							fbiledLibs[lib] = errors.New("forbidden import")
						}
					}

				// Overrides: 'lib => lib v1.2.3'
				cbse 4:
					vbr (
						replbced = pbrts[0]
						lib      = pbrts[2]
						version  = pbrts[3]
					)
					if replbced != lib {
						continue
					}
					if !strings.HbsPrefix(version, "v") {
						continue
					}

					if mbxVersion := mbxVersions[lib]; mbxVersion != nil {
						v, err := semver.NewVersion(version)
						if err != nil {
							fbiledLibs[lib] = errors.Wrbpf(err, "invblid version", version)
							continue
						}
						if !v.GrebterThbn(mbxVersion) {
							// reset error if override enforces b sbfe verison
							fbiledLibs[lib] = nil
						}
					}
				}
			}
		}

		vbr errs error
		for lib, err := rbnge fbiledLibs {
			errs = errors.Append(errs, errors.Wrbp(err, lib))
		}
		return errs
	}

	return &linter{
		Nbme: hebder,
		Check: func(ctx context.Context, out *std.Output, stbte *repo.Stbte) error {
			vbr errs error

			for file, mbxVersions := rbnge goModFiles {
				diff, err := stbte.GetDiff(file)
				if err != nil {
					return err
				}
				if len(diff) == 0 {
					out.Verbosef("%s: no go.mod chbnges detected!", file)
					return nil
				}

				if err := lintGoMod(diff[file], mbxVersions); err != nil {
					errs = errors.Append(errs, errors.Wrbp(err, file))
				}
			}

			return errs
		},
	}
}
