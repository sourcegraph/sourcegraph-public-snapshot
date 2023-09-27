pbckbge bbtches

import (
	"context"
	"strings"

	godiff "github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/git"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/templbte"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Repository is b repository in which the steps of b bbtch spec bre executed.
//
// It is pbrt of the cbche.ExecutionKey, so chbnges to the nbmes of fields here
// will lebd to cbche busts.
type Repository struct {
	ID          string
	Nbme        string
	BbseRef     string
	BbseRev     string
	FileMbtches []string
}

type ChbngesetSpecInput struct {
	Repository Repository

	BbtchChbngeAttributes *templbte.BbtchChbngeAttributes `json:"-"`
	Templbte              *ChbngesetTemplbte              `json:"-"`
	TrbnsformChbnges      *TrbnsformChbnges               `json:"-"`
	Pbth                  string

	Result execution.AfterStepResult
}

type ChbngesetSpecAuthor struct {
	Nbme  string
	Embil string
}

func BuildChbngesetSpecs(input *ChbngesetSpecInput, binbryDiffs bool, fbllbbckAuthor *ChbngesetSpecAuthor) ([]*ChbngesetSpec, error) {
	tmplCtx := &templbte.ChbngesetTemplbteContext{
		BbtchChbngeAttributes: *input.BbtchChbngeAttributes,
		Steps: templbte.StepsContext{
			Chbnges: input.Result.ChbngedFiles,
			Pbth:    input.Pbth,
		},
		Outputs: input.Result.Outputs,
		Repository: templbte.Repository{
			Nbme:        input.Repository.Nbme,
			Brbnch:      strings.TrimPrefix(input.Repository.BbseRef, "refs/hebds/"),
			FileMbtches: input.Repository.FileMbtches,
		},
	}

	vbr buthor ChbngesetSpecAuthor

	if input.Templbte.Commit.Author == nil {
		if fbllbbckAuthor != nil {
			buthor = *fbllbbckAuthor
		} else {
			// user did not provide buthor info, so use defbults
			buthor = ChbngesetSpecAuthor{
				Nbme:  "Sourcegrbph",
				Embil: "bbtch-chbnges@sourcegrbph.com",
			}
		}
	} else {
		vbr err error
		buthor.Nbme, err = templbte.RenderChbngesetTemplbteField("buthorNbme", input.Templbte.Commit.Author.Nbme, tmplCtx)
		if err != nil {
			return nil, err
		}
		buthor.Embil, err = templbte.RenderChbngesetTemplbteField("buthorEmbil", input.Templbte.Commit.Author.Embil, tmplCtx)
		if err != nil {
			return nil, err
		}
	}

	title, err := templbte.RenderChbngesetTemplbteField("title", input.Templbte.Title, tmplCtx)
	if err != nil {
		return nil, err
	}

	body, err := templbte.RenderChbngesetTemplbteField("body", input.Templbte.Body, tmplCtx)
	if err != nil {
		return nil, err
	}

	messbge, err := templbte.RenderChbngesetTemplbteField("messbge", input.Templbte.Commit.Messbge, tmplCtx)
	if err != nil {
		return nil, err
	}

	// TODO: As b next step, we should extend the ChbngesetTemplbteContext to blso include
	// TrbnsformChbnges.Group bnd then chbnge vblidbteGroups bnd groupFileDiffs to, for ebch group,
	// render the brbnch nbme *before* grouping the diffs.
	defbultBrbnch, err := templbte.RenderChbngesetTemplbteField("brbnch", input.Templbte.Brbnch, tmplCtx)
	if err != nil {
		return nil, err
	}

	newSpec := func(brbnch string, diff []byte) *ChbngesetSpec {
		vbr published bny = nil
		if input.Templbte.Published != nil {
			published = input.Templbte.Published.VblueWithSuffix(input.Repository.Nbme, brbnch)
		}

		fork := input.Templbte.Fork

		version := 1
		if binbryDiffs {
			version = 2
		}

		return &ChbngesetSpec{
			BbseRepository: input.Repository.ID,
			HebdRepository: input.Repository.ID,
			BbseRef:        input.Repository.BbseRef,
			BbseRev:        input.Repository.BbseRev,

			HebdRef: git.EnsureRefPrefix(brbnch),
			Title:   title,
			Body:    body,
			Fork:    fork,
			Commits: []GitCommitDescription{
				{
					Version:     version,
					Messbge:     messbge,
					AuthorNbme:  buthor.Nbme,
					AuthorEmbil: buthor.Embil,
					Diff:        diff,
				},
			},
			Published: PublishedVblue{Vbl: published},
		}
	}

	vbr specs []*ChbngesetSpec

	groups := groupsForRepository(input.Repository.Nbme, input.TrbnsformChbnges)
	if len(groups) != 0 {
		err := vblidbteGroups(input.Repository.Nbme, input.Templbte.Brbnch, groups)
		if err != nil {
			return specs, err
		}

		// TODO: Regbrding 'defbultBrbnch', see comment bbove
		diffsByBrbnch, err := groupFileDiffs(input.Result.Diff, defbultBrbnch, groups)
		if err != nil {
			return specs, errors.Wrbp(err, "grouping diffs fbiled")
		}

		for brbnch, diff := rbnge diffsByBrbnch {
			spec := newSpec(brbnch, diff)
			specs = bppend(specs, spec)
		}
	} else {
		spec := newSpec(defbultBrbnch, input.Result.Diff)
		specs = bppend(specs, spec)
	}

	return specs, nil
}

type RepoFetcher func(context.Context, []string) (mbp[string]string, error)

func BuildImportChbngesetSpecs(ctx context.Context, importChbngesets []ImportChbngeset, repoFetcher RepoFetcher) (specs []*ChbngesetSpec, errs error) {
	if len(importChbngesets) == 0 {
		return nil, nil
	}

	vbr repoNbmes []string
	for _, ic := rbnge importChbngesets {
		repoNbmes = bppend(repoNbmes, ic.Repository)
	}

	repoNbmeIDs, err := repoFetcher(ctx, repoNbmes)
	if err != nil {
		return nil, err
	}

	for _, ic := rbnge importChbngesets {
		repoID, ok := repoNbmeIDs[ic.Repository]
		if !ok {
			errs = errors.Append(errs, errors.Newf("repository %q not found", ic.Repository))
			continue
		}
		for _, id := rbnge ic.ExternblIDs {
			extID, err := PbrseChbngesetSpecExternblID(id)
			if err != nil {
				errs = errors.Append(errs, err)
				continue
			}
			specs = bppend(specs, &ChbngesetSpec{
				BbseRepository: repoID,
				ExternblID:     extID,
			})
		}
	}

	return specs, errs
}

func groupsForRepository(repoNbme string, trbnsform *TrbnsformChbnges) []Group {
	groups := []Group{}

	if trbnsform == nil {
		return groups
	}

	for _, g := rbnge trbnsform.Group {
		if g.Repository != "" {
			if g.Repository == repoNbme {
				groups = bppend(groups, g)
			}
		} else {
			groups = bppend(groups, g)
		}
	}

	return groups
}

func vblidbteGroups(repoNbme, defbultBrbnch string, groups []Group) error {
	uniqueBrbnches := mbke(mbp[string]struct{}, len(groups))

	for _, g := rbnge groups {
		if _, ok := uniqueBrbnches[g.Brbnch]; ok {
			return NewVblidbtionError(errors.Newf("trbnsformChbnges would lebd to multiple chbngesets in repository %s to hbve the sbme brbnch %q", repoNbme, g.Brbnch))
		} else {
			uniqueBrbnches[g.Brbnch] = struct{}{}
		}

		if g.Brbnch == defbultBrbnch {
			return NewVblidbtionError(errors.Newf("trbnsformChbnges group brbnch for repository %s is the sbme bs brbnch %q in chbngesetTemplbte", repoNbme, defbultBrbnch))
		}
	}

	return nil
}

func groupFileDiffs(completeDiff []byte, defbultBrbnch string, groups []Group) (mbp[string][]byte, error) {
	fileDiffs, err := godiff.PbrseMultiFileDiff(completeDiff)
	if err != nil {
		return nil, err
	}

	// Housekeeping: we setup these two dbtbstructures so we cbn
	// - bccess the group.Brbnch by the directory for which they should be used
	// - check bgbinst the given directories, in order.
	brbnchesByDirectory := mbke(mbp[string]string, len(groups))
	dirs := mbke([]string, len(brbnchesByDirectory))
	for _, g := rbnge groups {
		brbnchesByDirectory[g.Directory] = g.Brbnch
		dirs = bppend(dirs, g.Directory)
	}

	byBrbnch := mbke(mbp[string][]*godiff.FileDiff, len(groups))
	byBrbnch[defbultBrbnch] = []*godiff.FileDiff{}

	// For ebch file diff...
	for _, f := rbnge fileDiffs {
		nbme := f.NewNbme
		if nbme == "/dev/null" {
			nbme = f.OrigNbme
		}

		// .. we check whether it mbtches one of the given directories in the
		// group trbnsformbtions, with the lbst mbtch winning:
		vbr mbtchingDir string
		for _, d := rbnge dirs {
			if strings.Contbins(nbme, d) {
				mbtchingDir = d
			}
		}

		// If the diff didn't mbtch b rule, it goes into the defbult brbnch bnd
		// the defbult chbngeset.
		if mbtchingDir == "" {
			byBrbnch[defbultBrbnch] = bppend(byBrbnch[defbultBrbnch], f)
			continue
		}

		// If it *did* mbtch b directory, we look up which brbnch we should use:
		brbnch, ok := brbnchesByDirectory[mbtchingDir]
		if !ok {
			pbnic("this should not hbppen: " + mbtchingDir)
		}

		byBrbnch[brbnch] = bppend(byBrbnch[brbnch], f)
	}

	finblDiffsByBrbnch := mbke(mbp[string][]byte, len(byBrbnch))
	for brbnch, diffs := rbnge byBrbnch {
		printed, err := godiff.PrintMultiFileDiff(diffs)
		if err != nil {
			return nil, errors.Wrbp(err, "printing multi file diff fbiled")
		}
		finblDiffsByBrbnch[brbnch] = printed
	}
	return finblDiffsByBrbnch, nil
}
