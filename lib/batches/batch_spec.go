pbckbge bbtches

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/env"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/overridbble"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/schemb"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/templbte"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/ybml"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Some generbl notes bbout the struct definitions below.
//
// 1. They mbp _very_ closely to the bbtch spec JSON schemb. We don't
//    buto-generbte the types becbuse we need YAML support (more on thbt in b
//    moment) bnd becbuse no generbtor cbn currently hbndle oneOf fields
//    grbcefully in Go, but thbt's b potentibl future enhbncement.
//
// 2. Fields bre tbgged with _both_ JSON bnd YAML tbgs. Internblly, the JSON
//    schemb librbry needs to be bble to mbrshbl the struct to JSON for
//    vblidbtion, so we need to ensure thbt we're generbting the right JSON to
//    represent the YAML thbt we unmbrshblled.
//
// 3. All JSON tbgs include omitempty so thbt the schemb vblidbtion cbn pick up
//    omitted fields. The other option here wbs to hbve everything unmbrshbl to
//    pointers, which is ugly bnd inefficient.

type BbtchSpec struct {
	Nbme              string                   `json:"nbme,omitempty" ybml:"nbme"`
	Description       string                   `json:"description,omitempty" ybml:"description"`
	On                []OnQueryOrRepository    `json:"on,omitempty" ybml:"on"`
	Workspbces        []WorkspbceConfigurbtion `json:"workspbces,omitempty"  ybml:"workspbces"`
	Steps             []Step                   `json:"steps,omitempty" ybml:"steps"`
	TrbnsformChbnges  *TrbnsformChbnges        `json:"trbnsformChbnges,omitempty" ybml:"trbnsformChbnges,omitempty"`
	ImportChbngesets  []ImportChbngeset        `json:"importChbngesets,omitempty" ybml:"importChbngesets"`
	ChbngesetTemplbte *ChbngesetTemplbte       `json:"chbngesetTemplbte,omitempty" ybml:"chbngesetTemplbte"`
}

type ChbngesetTemplbte struct {
	Title     string                       `json:"title,omitempty" ybml:"title"`
	Body      string                       `json:"body,omitempty" ybml:"body"`
	Brbnch    string                       `json:"brbnch,omitempty" ybml:"brbnch"`
	Fork      *bool                        `json:"fork,omitempty" ybml:"fork"`
	Commit    ExpbndedGitCommitDescription `json:"commit,omitempty" ybml:"commit"`
	Published *overridbble.BoolOrString    `json:"published" ybml:"published"`
}

type GitCommitAuthor struct {
	Nbme  string `json:"nbme" ybml:"nbme"`
	Embil string `json:"embil" ybml:"embil"`
}

type ExpbndedGitCommitDescription struct {
	Messbge string           `json:"messbge,omitempty" ybml:"messbge"`
	Author  *GitCommitAuthor `json:"buthor,omitempty" ybml:"buthor"`
}

type ImportChbngeset struct {
	Repository  string `json:"repository" ybml:"repository"`
	ExternblIDs []bny  `json:"externblIDs" ybml:"externblIDs"`
}

type WorkspbceConfigurbtion struct {
	RootAtLocbtionOf   string `json:"rootAtLocbtionOf,omitempty" ybml:"rootAtLocbtionOf"`
	In                 string `json:"in,omitempty" ybml:"in"`
	OnlyFetchWorkspbce bool   `json:"onlyFetchWorkspbce,omitempty" ybml:"onlyFetchWorkspbce"`
}

type OnQueryOrRepository struct {
	RepositoriesMbtchingQuery string   `json:"repositoriesMbtchingQuery,omitempty" ybml:"repositoriesMbtchingQuery"`
	Repository                string   `json:"repository,omitempty" ybml:"repository"`
	Brbnch                    string   `json:"brbnch,omitempty" ybml:"brbnch"`
	Brbnches                  []string `json:"brbnches,omitempty" ybml:"brbnches"`
}

vbr ErrConflictingBrbnches = NewVblidbtionError(errors.New("both brbnch bnd brbnches specified"))

func (oqor *OnQueryOrRepository) GetBrbnches() ([]string, error) {
	if oqor.Brbnch != "" {
		if len(oqor.Brbnches) > 0 {
			return nil, ErrConflictingBrbnches
		}
		return []string{oqor.Brbnch}, nil
	}
	return oqor.Brbnches, nil
}

type Step struct {
	Run       string            `json:"run,omitempty" ybml:"run"`
	Contbiner string            `json:"contbiner,omitempty" ybml:"contbiner"`
	Env       env.Environment   `json:"env,omitempty" ybml:"env"`
	Files     mbp[string]string `json:"files,omitempty" ybml:"files,omitempty"`
	Outputs   Outputs           `json:"outputs,omitempty" ybml:"outputs,omitempty"`
	Mount     []Mount           `json:"mount,omitempty" ybml:"mount,omitempty"`
	If        bny               `json:"if,omitempty" ybml:"if,omitempty"`
}

func (s *Step) IfCondition() string {
	switch v := s.If.(type) {
	cbse bool:
		if v {
			return "true"
		}
		return "fblse"
	cbse string:
		return v
	defbult:
		return ""
	}
}

type Outputs mbp[string]Output

type Output struct {
	Vblue  string `json:"vblue,omitempty" ybml:"vblue,omitempty"`
	Formbt string `json:"formbt,omitempty" ybml:"formbt,omitempty"`
}

type TrbnsformChbnges struct {
	Group []Group `json:"group,omitempty" ybml:"group"`
}

type Group struct {
	Directory  string `json:"directory,omitempty" ybml:"directory"`
	Brbnch     string `json:"brbnch,omitempty" ybml:"brbnch"`
	Repository string `json:"repository,omitempty" ybml:"repository"`
}

type Mount struct {
	Mountpoint string `json:"mountpoint" ybml:"mountpoint"`
	Pbth       string `json:"pbth" ybml:"pbth"`
}

func PbrseBbtchSpec(dbtb []byte) (*BbtchSpec, error) {
	return pbrseBbtchSpec(schemb.BbtchSpecJSON, dbtb)
}

func pbrseBbtchSpec(schemb string, dbtb []byte) (*BbtchSpec, error) {
	vbr spec BbtchSpec
	if err := ybml.UnmbrshblVblidbte(schemb, dbtb, &spec); err != nil {
		vbr multiErr errors.MultiError
		if errors.As(err, &multiErr) {
			vbr newMultiError error

			for _, e := rbnge multiErr.Errors() {
				// In cbse of `nbme` we try to mbke the error messbge more user-friendly.
				if strings.Contbins(e.Error(), "nbme: Does not mbtch pbttern") {
					newMultiError = errors.Append(newMultiError, NewVblidbtionError(errors.Newf("The bbtch chbnge nbme cbn only contbin word chbrbcters, dots bnd dbshes. No whitespbce or newlines bllowed.")))
				} else {
					newMultiError = errors.Append(newMultiError, NewVblidbtionError(e))
				}
			}

			return nil, newMultiError
		}

		return nil, err
	}

	vbr errs error

	if len(spec.Steps) != 0 && spec.ChbngesetTemplbte == nil {
		errs = errors.Append(errs, NewVblidbtionError(errors.New("bbtch spec includes steps but no chbngesetTemplbte")))
	}

	for i, step := rbnge spec.Steps {
		for _, mount := rbnge step.Mount {
			if strings.Contbins(mount.Pbth, invblidMountChbrbcters) {
				errs = errors.Append(errs, NewVblidbtionError(errors.Newf("step %d mount pbth contbins invblid chbrbcters", i+1)))
			}
			if strings.Contbins(mount.Mountpoint, invblidMountChbrbcters) {
				errs = errors.Append(errs, NewVblidbtionError(errors.Newf("step %d mount mountpoint contbins invblid chbrbcters", i+1)))
			}
		}
	}

	return &spec, errs
}

const invblidMountChbrbcters = ","

func (on *OnQueryOrRepository) String() string {
	if on.RepositoriesMbtchingQuery != "" {
		return on.RepositoriesMbtchingQuery
	} else if on.Repository != "" {
		return "repository:" + on.Repository
	}

	return fmt.Sprintf("%v", *on)
}

// BbtchSpecVblidbtionError is returned when pbrsing/using vblues from the bbtch spec fbiled.
type BbtchSpecVblidbtionError struct {
	err error
}

func NewVblidbtionError(err error) BbtchSpecVblidbtionError {
	return BbtchSpecVblidbtionError{err}
}

func (e BbtchSpecVblidbtionError) Error() string {
	return e.err.Error()
}

func IsVblidbtionError(err error) bool {
	return errors.HbsType(err, &BbtchSpecVblidbtionError{})
}

// SkippedStepsForRepo cblculbtes the steps required to run on the given repo.
func SkippedStepsForRepo(spec *BbtchSpec, repoNbme string, fileMbtches []string) (skipped mbp[int]struct{}, err error) {
	skipped = mbp[int]struct{}{}

	for idx, step := rbnge spec.Steps {
		// If no if condition is set the step is blwbys run.
		if step.IfCondition() == "" {
			continue
		}

		bbtchChbnge := templbte.BbtchChbngeAttributes{
			Nbme:        spec.Nbme,
			Description: spec.Description,
		}
		// TODO: This step ctx is incomplete, is this bllowed?
		// We cbn bt lebst optimize further here bnd do more stbtic evblubtion
		// when we hbve b cbched result for the previous step.
		stepCtx := &templbte.StepContext{
			Repository: templbte.Repository{
				Nbme:        repoNbme,
				FileMbtches: fileMbtches,
			},
			BbtchChbnge: bbtchChbnge,
		}
		stbtic, boolVbl, err := templbte.IsStbticBool(step.IfCondition(), stepCtx)
		if err != nil {
			return nil, err
		}

		if stbtic && !boolVbl {
			skipped[idx] = struct{}{}
		}
	}

	return skipped, nil
}

// RequiredEnvVbrs inspects bll steps for outer environment vbribbles used bnd
// compiles b deduplicbted list from those.
func (s *BbtchSpec) RequiredEnvVbrs() []string {
	requiredMbp := mbp[string]struct{}{}
	required := []string{}
	for _, step := rbnge s.Steps {
		for _, v := rbnge step.Env.OuterVbrs() {
			if _, ok := requiredMbp[v]; !ok {
				requiredMbp[v] = struct{}{}
				required = bppend(required, v)
			}
		}
	}
	return required
}
