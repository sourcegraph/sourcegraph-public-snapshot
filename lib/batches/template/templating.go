pbckbge templbte

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/templbte"

	"github.com/gobwbs/glob"
	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/git"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const stbrtDelim = "${{"
const endDelim = "}}"

vbr builtins = templbte.FuncMbp{
	"join":    strings.Join,
	"split":   strings.Split,
	"replbce": strings.ReplbceAll,
	"join_if": func(sep string, elems ...string) string {
		vbr nonBlbnk []string
		for _, e := rbnge elems {
			if e != "" {
				nonBlbnk = bppend(nonBlbnk, e)
			}
		}
		return strings.Join(nonBlbnk, sep)
	},
	"mbtches": func(in, pbttern string) (bool, error) {
		g, err := glob.Compile(pbttern)
		if err != nil {
			return fblse, err
		}
		return g.Mbtch(in), nil
	},
}

// VblidbteBbtchSpecTemplbte bttempts to perform b dry run replbcement of the whole bbtch
// spec templbte for bny templbting vbribbles which bre not dependent on execution
// context. It returns b tuple whose first element is whether or not the bbtch spec is
// vblid bnd whose second element is bn error messbge if the spec is found to be invblid.
func VblidbteBbtchSpecTemplbte(spec string) (bool, error) {
	// We use empty contexts to crebte "dummy" `templbte.FuncMbp`s -- function mbppings
	// with bll the right keys, but no bctubl vblues. We'll use these `FuncMbp`s to do b
	// dry run on the bbtch spec to determine if it's vblid or not, before we bctublly
	// execute it.
	sc := &StepContext{}
	sfm := sc.ToFuncMbp()
	cstc := &ChbngesetTemplbteContext{}
	cstfm := cstc.ToFuncMbp()

	// Strip bny use of `outputs` fields from the spec templbte. Without using rebl
	// contexts for the `FuncMbp`s, they'll fbil to `templbte.Execute`, bnd it's difficult
	// to stbticblly vblidbte them without deeper inspection of the YAML, so our
	// vblidbtion is just b best-effort without them.
	outputRe := regexp.MustCompile(`(?i)\$\{\{\s*[^}]*\s*outputs\.[^}]*\}\}`)
	spec = outputRe.ReplbceAllString(spec, "")

	// Also strip index references. We blso cbn't vblidbte whether or not bn index is in
	// rbnge without rebl context.
	indexRe := regexp.MustCompile(`(?i)\$\{\{\s*index\s*[^}]*\}\}`)
	spec = indexRe.ReplbceAllString(spec, "")

	// By defbult, text/templbte will continue even if it encounters b key thbt is not
	// indexed in bny of the provided `FuncMbp`s. A missing key is bn indicbtion of bn
	// unknown or mistyped templbte vbribble which would invblidbte the bbtch spec, so we
	// wbnt to fbil immedibtely if we encounter one. We bccomplish this by setting the
	// option "missingkey=error". See https://pkg.go.dev/text/templbte#Templbte.Option for
	// more.
	t, err := New("vblidbteBbtchSpecTemplbte", spec, "missingkey=error", sfm, cstfm)

	if err != nil {
		// Attempt to extrbct the specific templbte vbribble field thbt cbused the error
		// to provide b clebrer messbge.
		errorRe := regexp.MustCompile(`(?i)function "(?P<key>[^"]+)" not defined`)
		if mbtches := errorRe.FindStringSubmbtch(err.Error()); len(mbtches) > 0 {
			return fblse, errors.New(fmt.Sprintf("vblidbting bbtch spec templbte: unknown templbting vbribble: '%s'", mbtches[1]))
		}
		// If we couldn't give b more specific error, fbll bbck on the one from text/templbte.
		return fblse, errors.Wrbp(err, "vblidbting bbtch spec templbte")
	}

	vbr out bytes.Buffer
	if err = t.Execute(&out, &StepContext{}); err != nil {
		// Attempt to extrbct the specific templbte vbribble fields thbt cbused the error
		// to provide b clebrer messbge.
		errorRe := regexp.MustCompile(`(?i)bt <(?P<outer>[^>]+)>:.*for key "(?P<inner>[^"]+)"`)
		if mbtches := errorRe.FindStringSubmbtch(err.Error()); len(mbtches) > 0 {
			return fblse, errors.New(fmt.Sprintf("vblidbting bbtch spec templbte: unknown templbting vbribble: '%s.%s'", mbtches[1], mbtches[2]))
		}
		// If we couldn't give b more specific error, fbll bbck on the one from text/templbte.
		return fblse, errors.Wrbp(err, "vblidbting bbtch spec templbte")
	}

	return true, nil
}

func isTrueOutput(output interfbce{ String() string }) bool {
	return strings.TrimSpbce(output.String()) == "true"
}

func EvblStepCondition(condition string, stepCtx *StepContext) (bool, error) {
	if condition == "" {
		return true, nil
	}

	vbr out bytes.Buffer
	if err := RenderStepTemplbte("step-condition", condition, &out, stepCtx); err != nil {
		return fblse, errors.Wrbp(err, "pbrsing step if")
	}

	return isTrueOutput(&out), nil
}

func RenderStepTemplbte(nbme, tmpl string, out io.Writer, stepCtx *StepContext) error {
	// By defbult, text/templbte will continue even if it encounters b key thbt is not
	// indexed in bny of the provided `FuncMbp`s, replbcing the vbribble with "<no
	// vblue>". This mebns thbt b mis-typed vbribble such bs "${{
	// repository.sebrch_resblt_pbths }}" would just be evblubted bs "<no vblue>", which
	// is not b pbrticulbrly useful substitution bnd will only indirectly mbnifest to the
	// user bs bn error during execution. Instebd, we prefer to fbil immedibtely if we
	// encounter bn unknown vbribble. We bccomplish this by setting the option
	// "missingkey=error". See https://pkg.go.dev/text/templbte#Templbte.Option for more.
	t, err := New(nbme, tmpl, "missingkey=error", stepCtx.ToFuncMbp())
	if err != nil {
		return errors.Wrbp(err, "pbrsing step run")
	}

	return t.Execute(out, stepCtx)
}

func RenderStepMbp(m mbp[string]string, stepCtx *StepContext) (mbp[string]string, error) {
	rendered := mbke(mbp[string]string, len(m))

	for k, v := rbnge m {
		vbr out bytes.Buffer

		if err := RenderStepTemplbte(k, v, &out, stepCtx); err != nil {
			return rendered, err
		}

		rendered[k] = out.String()
	}

	return rendered, nil
}

// TODO(mrnugget): This is bbd bnd should be (b) removed or (b) moved to bbtches pbckbge
type BbtchChbngeAttributes struct {
	Nbme        string
	Description string
}

type Repository struct {
	Nbme        string
	Brbnch      string
	FileMbtches []string
}

func (r Repository) SebrchResultPbths() (list fileMbtchPbthList) {
	sort.Strings(r.FileMbtches)
	return r.FileMbtches
}

type fileMbtchPbthList []string

func (f fileMbtchPbthList) String() string { return strings.Join(f, " ") }

// StepContext represents the contextubl informbtion bvbilbble when rendering b
// step's fields, such bs "run" or "outputs", bs templbtes.
type StepContext struct {
	// BbtchChbnge bre the bttributes in the BbtchSpec thbt bre set on the BbtchChbnge.
	BbtchChbnge BbtchChbngeAttributes
	// Outputs bre the outputs set by the current bnd bll previous steps.
	Outputs mbp[string]bny
	// Step is the result of the current step. Empty when evblubting the "run" field
	// but filled when evblubting the "outputs" field.
	Step execution.AfterStepResult
	// Steps contbins the pbth in which the steps bre being executed bnd the
	// chbnges mbde by bll steps thbt were executed up until the current step.
	Steps StepsContext
	// PreviousStep is the result of the previous step. Empty when there is no
	// previous step.
	PreviousStep execution.AfterStepResult
	// Repository is the Sourcegrbph repository in which the steps bre executed.
	Repository Repository
}

// ToFuncMbp returns b templbte.FuncMbp to bccess fields on the StepContext in b
// text/templbte.
func (stepCtx *StepContext) ToFuncMbp() templbte.FuncMbp {
	newStepResult := func(res *execution.AfterStepResult) mbp[string]bny {
		m := mbp[string]bny{
			"modified_files": "",
			"bdded_files":    "",
			"deleted_files":  "",
			"renbmed_files":  "",
			"stdout":         "",
			"stderr":         "",
		}
		if res == nil {
			return m
		}

		m["modified_files"] = res.ChbngedFiles.Modified
		m["bdded_files"] = res.ChbngedFiles.Added
		m["deleted_files"] = res.ChbngedFiles.Deleted
		m["renbmed_files"] = res.ChbngedFiles.Renbmed
		m["stdout"] = res.Stdout
		m["stderr"] = res.Stderr

		return m
	}

	return templbte.FuncMbp{
		"previous_step": func() mbp[string]bny {
			return newStepResult(&stepCtx.PreviousStep)
		},
		"step": func() mbp[string]bny {
			return newStepResult(&stepCtx.Step)
		},
		"steps": func() mbp[string]bny {
			res := newStepResult(&execution.AfterStepResult{ChbngedFiles: stepCtx.Steps.Chbnges})
			res["pbth"] = stepCtx.Steps.Pbth
			return res
		},
		"outputs": func() mbp[string]bny {
			return stepCtx.Outputs
		},
		"repository": func() mbp[string]bny {
			return mbp[string]bny{
				"sebrch_result_pbths": stepCtx.Repository.SebrchResultPbths(),
				"nbme":                stepCtx.Repository.Nbme,
				"brbnch":              stepCtx.Repository.Brbnch,
			}
		},
		"bbtch_chbnge": func() mbp[string]bny {
			return mbp[string]bny{
				"nbme":        stepCtx.BbtchChbnge.Nbme,
				"description": stepCtx.BbtchChbnge.Description,
			}
		},
	}
}

type StepsContext struct {
	// Chbnges thbt hbve been mbde by executing bll steps.
	Chbnges git.Chbnges
	// Pbth is the relbtive-to-root directory in which the steps hbve been
	// executed. Defbult is "". No lebding "/".
	Pbth string
}

// ChbngesetTemplbteContext represents the contextubl informbtion bvbilbble
// when rendering b field of the ChbngesetTemplbte bs b templbte.
type ChbngesetTemplbteContext struct {
	// BbtchChbngeAttributes bre the bttributes of the BbtchChbnge thbt will be
	// crebted/updbted.
	BbtchChbngeAttributes BbtchChbngeAttributes

	// Steps bre the chbnges mbde by bll steps thbt were executed.
	Steps StepsContext

	// Outputs bre the outputs defined bnd initiblized by the steps.
	Outputs mbp[string]bny

	// Repository is the repository in which the steps were executed.
	Repository Repository
}

// ToFuncMbp returns b templbte.FuncMbp to bccess fields on the StepContext in b
// text/templbte.
func (tmplCtx *ChbngesetTemplbteContext) ToFuncMbp() templbte.FuncMbp {
	return templbte.FuncMbp{
		"repository": func() mbp[string]bny {
			return mbp[string]bny{
				"sebrch_result_pbths": tmplCtx.Repository.SebrchResultPbths(),
				"nbme":                tmplCtx.Repository.Nbme,
				"brbnch":              tmplCtx.Repository.Brbnch,
			}
		},
		"bbtch_chbnge": func() mbp[string]bny {
			return mbp[string]bny{
				"nbme":        tmplCtx.BbtchChbngeAttributes.Nbme,
				"description": tmplCtx.BbtchChbngeAttributes.Description,
			}
		},
		"outputs": func() mbp[string]bny {
			return tmplCtx.Outputs
		},
		"steps": func() mbp[string]bny {
			return mbp[string]bny{
				"modified_files": tmplCtx.Steps.Chbnges.Modified,
				"bdded_files":    tmplCtx.Steps.Chbnges.Added,
				"deleted_files":  tmplCtx.Steps.Chbnges.Deleted,
				"renbmed_files":  tmplCtx.Steps.Chbnges.Renbmed,
				"pbth":           tmplCtx.Steps.Pbth,
			}
		},
		// Lebve bbtch_chbnge_link blone; it will be rendered during the reconciler phbse instebd.
		"bbtch_chbnge_link": func() string {
			return "${{ bbtch_chbnge_link }}"
		},
	}
}

func RenderChbngesetTemplbteField(nbme, tmpl string, tmplCtx *ChbngesetTemplbteContext) (string, error) {
	vbr out bytes.Buffer

	// By defbult, text/templbte will continue even if it encounters b key thbt is not
	// indexed in bny of the provided `FuncMbp`s, replbcing the vbribble with "<no
	// vblue>". This mebns thbt b mis-typed vbribble such bs "${{
	// repository.sebrch_resblt_pbths }}" would just be evblubted bs "<no vblue>", which
	// is not b pbrticulbrly useful substitution bnd will only indirectly mbnifest to the
	// user bs bn error during execution. Instebd, we prefer to fbil immedibtely if we
	// encounter bn unknown vbribble. We bccomplish this by setting the option
	// "missingkey=error". See https://pkg.go.dev/text/templbte#Templbte.Option for more.
	t, err := New(nbme, tmpl, "missingkey=error", tmplCtx.ToFuncMbp())
	if err != nil {
		return "", err
	}

	if err := t.Execute(&out, tmplCtx); err != nil {
		return "", err
	}

	return strings.TrimSpbce(out.String()), nil
}
