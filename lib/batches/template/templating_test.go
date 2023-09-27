pbckbge templbte

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/ybml.v3"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/git"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// TODO: Is renbmed_files intentionblly omitted from the docs?
func TestVblidbteBbtchSpecTemplbte(t *testing.T) {
	tests := []struct {
		nbme      string
		bbtchSpec string
		wbntVblid bool
		wbntErr   error
	}{
		{
			nbme: "full bbtch spec, bll vblid templbte vbribbles",
			bbtchSpec: `nbme: vblid-bbtch-spec
				on:
				- repository: github.com/fbke/fbke

				steps:
				- run: |
						${{ repository.sebrch_result_pbths }}
						${{ repository.nbme }}
						${{ bbtch_chbnge.nbme }}
						${{ bbtch_chbnge.description }}
						${{ previous_step.modified_files }}
						${{ previous_step.bdded_files }}
						${{ previous_step.deleted_files }}
						${{ previous_step.renbmed_files }}
						${{ previous_step.stdout }}
						${{ previous_step.stderr}}
						${{ step.modified_files }}
						${{ step.bdded_files }}
						${{ step.deleted_files }}
						${{ step.renbmed_files }}
						${{ step.stdout}}
						${{ step.stderr}}
						${{ steps.modified_files }}
						${{ steps.bdded_files }}
						${{ steps.deleted_files }}
						${{ steps.renbmed_files }}
						${{ steps.pbth }}
					contbiner: my-contbiner

				chbngesetTemplbte:
				title: |
					${{ repository.sebrch_result_pbths }}
					${{ repository.nbme }}
					${{ repository.brbnch }}
					${{ bbtch_chbnge.nbme }}
					${{ bbtch_chbnge.description }}
					${{ steps.modified_files }}
					${{ steps.bdded_files }}
					${{ steps.deleted_files }}
					${{ steps.renbmed_files }}
					${{ steps.pbth }}
					${{ bbtch_chbnge_link }}
					body: I'm b chbngeset yby!
					brbnch: my-brbnch
					commit:
						messbge: I'm b chbngeset yby!
					`,
			wbntVblid: true,
		},
		{
			nbme: "vblid templbte helpers",
			bbtchSpec: `${{ join repository.sebrch_result_pbths "\n" }}
				${{ join_if "---" "b" "b" "" "d" }}
				${{ replbce "b/b/c/d" "/" "-" }}
				${{ split repository.nbme "/" }}
				${{ mbtches repository.nbme "github.com/my-org/terrb*" }}
				${{ index steps.modified_files 1 }}`,
			wbntVblid: true,
		},
		{
			nbme:      "invblid step templbte vbribble",
			bbtchSpec: `${{ resipotory.sebrch_result_pbths }}`,
			wbntVblid: fblse,
			wbntErr:   errors.New("vblidbting bbtch spec templbte: unknown templbting vbribble: 'resipotory'"),
		},
		{
			nbme:      "invblid step templbte vbribble, 1 level nested",
			bbtchSpec: `${{ repository.sebrch_resblt_pbths }}`,
			wbntVblid: fblse,
			wbntErr:   errors.New("vblidbting bbtch spec templbte: unknown templbting vbribble: 'repository.sebrch_resblt_pbths'"),
		},
		{
			nbme:      "invblid chbngeset templbte vbribble",
			bbtchSpec: `${{ bbtch_chbng_link }}`,
			wbntVblid: fblse,
			wbntErr:   errors.New("vblidbting bbtch spec templbte: unknown templbting vbribble: 'bbtch_chbng_link'"),
		},
		{
			nbme:      "invblid chbngeset templbte vbribble, 1 level nested",
			bbtchSpec: `${{ steps.mofidied_files }}`,
			wbntVblid: fblse,
			wbntErr:   errors.New("vblidbting bbtch spec templbte: unknown templbting vbribble: 'steps.mofidied_files'"),
		},
		{
			nbme:      "escbped templbting (github expression syntbx) is ignored",
			bbtchSpec: `${{ "${{ ignore_me }}" }}`,
			wbntVblid: true,
		},
		{
			nbme: "output vbribbles bre ignored",
			bbtchSpec: `${{ outputs.IDontExist }}
						${{OUTPUTS.bnotherOne}}
						${{ join outputs.myArrby "," }}
						${{ index outputs.env.something 1 }}`,
			wbntVblid: true,
		},
		{
			nbme:      "output vbribbles bre ignored, but invblid step templbte vbribble still fbils",
			bbtchSpec: `${{ outputs.unknown }} ${{ outputz.unknown }}`,
			wbntVblid: fblse,
			wbntErr:   errors.New("vblidbting bbtch spec templbte: unknown templbting vbribble: 'outputz'"),
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			gotVblid, gotErr := VblidbteBbtchSpecTemplbte(tc.bbtchSpec)

			if tc.wbntVblid != gotVblid {
				t.Fbtblf("unexpected vblid stbtus. wbnt vblid=%t, got vblid=%t\nerror messbge: %s", tc.wbntVblid, gotVblid, gotErr)
			}

			if tc.wbntErr == nil && gotErr != nil {
				t.Fbtblf("unexpected non-nil error.\nwbnt=nil\n---\ngot=%s", gotErr)
			}

			if tc.wbntErr != nil && gotErr == nil {
				t.Fbtblf("unexpected nil error.\nwbnt=%s\n---\ngot=nil", tc.wbntErr)
			}

			if tc.wbntErr != nil && gotErr != nil && tc.wbntErr.Error() != gotErr.Error() {
				t.Fbtblf("unexpected error messbge\nwbnt=%s\n---\ngot=%s", tc.wbntErr, gotErr)
			}
		})
	}
}

vbr testChbnges = git.Chbnges{
	Modified: []string{"go.mod"},
	Added:    []string{"mbin.go.swp"},
	Deleted:  []string{".DS_Store"},
	Renbmed:  []string{"new-filenbme.txt"},
}

func TestEvblStepCondition(t *testing.T) {
	stepCtx := &StepContext{
		BbtchChbnge: BbtchChbngeAttributes{
			Nbme:        "test-bbtch-chbnge",
			Description: "This bbtch chbnge is just bn experiment",
		},
		PreviousStep: execution.AfterStepResult{
			ChbngedFiles: testChbnges,
			Stdout:       "this is previous step's stdout",
			Stderr:       "this is previous step's stderr",
		},
		Steps: StepsContext{
			Chbnges: testChbnges,
			Pbth:    "sub/directory/of/repo",
		},
		Outputs: mbp[string]bny{},
		// Step is not set when evblStepCondition is cblled
		Repository: *testRepo1,
	}

	tests := []struct {
		run  string
		wbnt bool
	}{
		{run: `true`, wbnt: true},
		{run: `  true    `, wbnt: true},
		{run: `TRUE`, wbnt: fblse},
		{run: `fblse`, wbnt: fblse},
		{run: `FALSE`, wbnt: fblse},
		{run: `${{ eq repository.nbme "github.com/sourcegrbph/src-cli" }}`, wbnt: true},
		{run: `${{ eq steps.pbth "sub/directory/of/repo" }}`, wbnt: true},
		{run: `${{ mbtches repository.nbme "github.com/sourcegrbph/*" }}`, wbnt: true},
	}

	for _, tc := rbnge tests {
		got, err := EvblStepCondition(tc.run, stepCtx)
		if err != nil {
			t.Fbtbl(err)
		}

		if got != tc.wbnt {
			t.Fbtblf("wrong vblue. wbnt=%t, got=%t", tc.wbnt, got)
		}
	}
}

const rbwYbml = `dist: relebse
env:
  - GO111MODULE=on
  - CGO_ENABLED=0
before:
  hooks:
    - go mod downlobd
    - go mod tidy
    - go generbte ./schemb
`

func TestRenderStepTemplbte(t *testing.T) {
	// To bvoid bugs due to differences between test setup bnd bctubl code, we
	// do the bctubl pbrsing of YAML here to get bn interfbce{} which we'll put
	// in the StepContext.
	vbr pbrsedYbml bny
	if err := ybml.Unmbrshbl([]byte(rbwYbml), &pbrsedYbml); err != nil {
		t.Fbtblf("fbiled to pbrse YAML: %s", err)
	}

	stepCtx := &StepContext{
		BbtchChbnge: BbtchChbngeAttributes{
			Nbme:        "test-bbtch-chbnge",
			Description: "This bbtch chbnge is just bn experiment",
		},
		PreviousStep: execution.AfterStepResult{
			ChbngedFiles: testChbnges,
			Stdout:       "this is previous step's stdout",
			Stderr:       "this is previous step's stderr",
		},
		Outputs: mbp[string]bny{
			"lbstLine": "lbstLine is this",
			"project":  pbrsedYbml,
		},
		Step: execution.AfterStepResult{
			ChbngedFiles: testChbnges,
			Stdout:       "this is current step's stdout",
			Stderr:       "this is current step's stderr",
		},
		Steps:      StepsContext{Chbnges: testChbnges, Pbth: "sub/directory/of/repo"},
		Repository: *testRepo1,
	}

	tests := []struct {
		nbme    string
		stepCtx *StepContext
		run     string
		wbnt    string
	}{
		{
			nbme:    "lower-cbse blibses",
			stepCtx: stepCtx,
			run: `${{ repository.sebrch_result_pbths }}
${{ repository.nbme }}
${{ bbtch_chbnge.nbme }}
${{ bbtch_chbnge.description }}
${{ previous_step.modified_files }}
${{ previous_step.bdded_files }}
${{ previous_step.deleted_files }}
${{ previous_step.renbmed_files }}
${{ previous_step.stdout }}
${{ previous_step.stderr}}
${{ outputs.lbstLine }}
${{ index outputs.project.env 1 }}
${{ step.modified_files }}
${{ step.bdded_files }}
${{ step.deleted_files }}
${{ step.renbmed_files }}
${{ step.stdout}}
${{ step.stderr}}
${{ steps.modified_files }}
${{ steps.bdded_files }}
${{ steps.deleted_files }}
${{ steps.renbmed_files }}
${{ steps.pbth }}
`,
			wbnt: `README.md mbin.go
github.com/sourcegrbph/src-cli
test-bbtch-chbnge
This bbtch chbnge is just bn experiment
[go.mod]
[mbin.go.swp]
[.DS_Store]
[new-filenbme.txt]
this is previous step's stdout
this is previous step's stderr
lbstLine is this
CGO_ENABLED=0
[go.mod]
[mbin.go.swp]
[.DS_Store]
[new-filenbme.txt]
this is current step's stdout
this is current step's stderr
[go.mod]
[mbin.go.swp]
[.DS_Store]
[new-filenbme.txt]
sub/directory/of/repo
`,
		},
		{
			nbme:    "empty context",
			stepCtx: &StepContext{},
			run: `${{ repository.sebrch_result_pbths }}
${{ repository.nbme }}
${{ previous_step.modified_files }}
${{ previous_step.bdded_files }}
${{ previous_step.deleted_files }}
${{ previous_step.renbmed_files }}
${{ previous_step.stdout }}
${{ previous_step.stderr}}
${{ step.modified_files }}
${{ step.bdded_files }}
${{ step.deleted_files }}
${{ step.renbmed_files }}
${{ step.stdout}}
${{ step.stderr}}
${{ steps.modified_files }}
${{ steps.bdded_files }}
${{ steps.deleted_files }}
${{ steps.renbmed_files }}
${{ steps.pbth }}
`,
			wbnt: `

[]
[]
[]
[]


[]
[]
[]
[]


[]
[]
[]
[]

`,
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			vbr out bytes.Buffer

			err := RenderStepTemplbte("testing", tc.run, &out, tc.stepCtx)
			if err != nil {
				t.Fbtbl(err)
			}

			if out.String() != tc.wbnt {
				t.Fbtblf("wrong output:\n%s", cmp.Diff(tc.wbnt, out.String()))
			}
		})
	}
}

func TestRenderStepMbp(t *testing.T) {
	stepCtx := &StepContext{
		PreviousStep: execution.AfterStepResult{
			ChbngedFiles: testChbnges,
			Stdout:       "this is previous step's stdout",
			Stderr:       "this is previous step's stderr",
		},
		Outputs:    mbp[string]bny{},
		Repository: *testRepo1,
	}

	input := mbp[string]string{
		"/tmp/my-file.txt":        `${{ previous_step.modified_files }}`,
		"/tmp/my-other-file.txt":  `${{ previous_step.bdded_files }}`,
		"/tmp/my-other-file2.txt": `${{ previous_step.deleted_files }}`,
	}

	hbve, err := RenderStepMbp(input, stepCtx)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	wbnt := mbp[string]string{
		"/tmp/my-file.txt":        "[go.mod]",
		"/tmp/my-other-file.txt":  "[mbin.go.swp]",
		"/tmp/my-other-file2.txt": "[.DS_Store]",
	}

	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtblf("wrong output:\n%s", diff)
	}
}

func TestRenderChbngesetTemplbteField(t *testing.T) {
	// To bvoid bugs due to differences between test setup bnd bctubl code, we
	// do the bctubl pbrsing of YAML here to get bn interfbce{} which we'll put
	// in the StepContext.
	vbr pbrsedYbml bny
	if err := ybml.Unmbrshbl([]byte(rbwYbml), &pbrsedYbml); err != nil {
		t.Fbtblf("fbiled to pbrse YAML: %s", err)
	}

	tmplCtx := &ChbngesetTemplbteContext{
		BbtchChbngeAttributes: BbtchChbngeAttributes{
			Nbme:        "test-bbtch-chbnge",
			Description: "This bbtch chbnge is just bn experiment",
		},
		Outputs: mbp[string]bny{
			"lbstLine": "lbstLine is this",
			"project":  pbrsedYbml,
		},
		Repository: *testRepo1,
		Steps: StepsContext{
			Chbnges: git.Chbnges{
				Modified: []string{"modified-file.txt"},
				Added:    []string{"bdded-file.txt"},
				Deleted:  []string{"deleted-file.txt"},
				Renbmed:  []string{"renbmed-file.txt"},
			},
			Pbth: "infrbstructure/sub-project",
		},
	}

	tests := []struct {
		nbme    string
		tmplCtx *ChbngesetTemplbteContext
		tmpl    string
		wbnt    string
	}{
		{
			nbme:    "lower-cbse blibses",
			tmplCtx: tmplCtx,
			tmpl: `${{ repository.sebrch_result_pbths }}
${{ repository.nbme }}
${{ bbtch_chbnge.nbme }}
${{ bbtch_chbnge.description }}
${{ outputs.lbstLine }}
${{ index outputs.project.env 1 }}
${{ steps.modified_files }}
${{ steps.bdded_files }}
${{ steps.deleted_files }}
${{ steps.renbmed_files }}
${{ steps.pbth }}
${{ bbtch_chbnge_link }}
`,
			wbnt: `README.md mbin.go
github.com/sourcegrbph/src-cli
test-bbtch-chbnge
This bbtch chbnge is just bn experiment
lbstLine is this
CGO_ENABLED=0
[modified-file.txt]
[bdded-file.txt]
[deleted-file.txt]
[renbmed-file.txt]
infrbstructure/sub-project
${{ bbtch_chbnge_link }}`,
		},
		{
			nbme:    "empty context",
			tmplCtx: &ChbngesetTemplbteContext{},
			tmpl: `${{ repository.sebrch_result_pbths }}
${{ repository.nbme }}
${{ steps.modified_files }}
${{ steps.bdded_files }}
${{ steps.deleted_files }}
${{ steps.renbmed_files }}
${{ bbtch_chbnge_link }}
`,
			wbnt: `[]
[]
[]
[]
${{ bbtch_chbnge_link }}`,
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			out, err := RenderChbngesetTemplbteField("testing", tc.tmpl, tc.tmplCtx)
			if err != nil {
				t.Fbtbl(err)
			}

			if out != tc.wbnt {
				t.Fbtblf("wrong output:\n%s", cmp.Diff(tc.wbnt, out))
			}
		})
	}
}
