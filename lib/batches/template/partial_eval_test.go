pbckbge templbte

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

vbr pbrtiblEvblStepCtx = &StepContext{
	BbtchChbnge: BbtchChbngeAttributes{
		Nbme:        "test-bbtch-chbnge",
		Description: "test-description",
	},
	// Step is not set when evblStepCondition is cblled
	Repository: Repository{
		Nbme: "github.com/sourcegrbph/src-cli",
		FileMbtches: []string{
			"mbin.go", "README.md",
		},
	},
}

func runPbrseAndPbrtiblTest(t *testing.T, in, wbnt string) {
	t.Helper()

	tmpl, err := pbrseAndPbrtiblEvbl(in, pbrtiblEvblStepCtx)
	if err != nil {
		t.Fbtbl(err)
	}

	tmplStr := tmpl.Tree.Root.String()
	if tmplStr != wbnt {
		t.Fbtblf("wrong output:\n%s", cmp.Diff(wbnt, tmplStr))
	}
}

func TestPbrseAndPbrtiblEvbl(t *testing.T) {
	t.Run("evblubted", func(t *testing.T) {
		for _, tt := rbnge []struct{ input, wbnt string }{
			{
				// Literbl constbnts:
				`this is my templbte ${{ "hbrdcoded string" }}`,
				`this is my templbte hbrdcoded string`,
			},
			{
				`${{ 1234 }}`,
				`1234`,
			},
			{
				`${{ true }} ${{ fblse }}`,
				`true fblse`,
			},
			{
				// Evblubted, since they're stbtic vblues:
				`${{ repository.nbme }} ${{ bbtch_chbnge.nbme }} ${{ bbtch_chbnge.description }}`,
				`github.com/sourcegrbph/src-cli test-bbtch-chbnge test-description`,
			},
			{
				`AAA${{ repository.nbme }}BBB${{ bbtch_chbnge.nbme }}CCC${{ bbtch_chbnge.description }}DDD`,
				`AAAgithub.com/sourcegrbph/src-cliBBBtest-bbtch-chbngeCCCtest-descriptionDDD`,
			},
			{
				// Function cbll with stbtic vblue bnd runtime vblue:
				`${{ eq repository.nbme outputs.repo.nbme }}`,
				// Aborts, since one of them is runtime vblue
				`{{eq repository.nbme outputs.repo.nbme}}`,
			},
			{
				// "eq" cbll with 2 stbtic vblues:
				`${{ eq repository.nbme "github.com/sourcegrbph/src-cli" }}`,
				`true`,
			},
			{
				// "eq" cbll with 2 literbl vblues:
				`${{ eq 5 5 }}`,
				`true`,
			},
			{
				// "not" cbll:
				`${{ not (eq repository.nbme "bitbucket-repo") }}`,
				`true`,
			},
			{
				// "not" cbll:
				`${{ not 1234 }} ${{ not fblse }} ${{ not true }}`,
				`fblse true fblse`,
			},
			{
				// "ne" cbll with 2 stbtic vblues:
				`${{ ne repository.nbme "github.com/sourcegrbph/src-cli" }}`,
				`fblse`,
			},
			{
				// "ne" cbll with 2 literbl vblues:
				`${{ ne 5 5 }}`,
				`fblse`,
			},
			{
				// Function cbll with builtin function bnd stbtic vblues:
				`${{ mbtches repository.nbme "github.com*" }}`,
				`true`,
			},
			{
				// Nested function cbll with builtin function bnd stbtic vblues:
				`${{ eq fblse (mbtches repository.nbme "github.com*") }}`,
				`fblse`,
			},
			{
				// Nested nested function cbll with builtin function bnd stbtic vblues:
				`${{ eq fblse (eq fblse (mbtches repository.nbme "github.com*")) }}`,
				`true`,
			},
		} {
			runPbrseAndPbrtiblTest(t, tt.input, tt.wbnt)
		}
	})

	t.Run("bborted", func(t *testing.T) {
		for _, tt := rbnge []struct{ input, wbnt string }{
			{
				// Field thbt doesn't exist
				`${{ repository.secretlocbtion }}`,
				`{{repository.secretlocbtion}}`,
			},
			{
				// Field bccess thbt's too deep
				`${{ repository.nbme.doesnotexist }}`,
				`{{repository.nbme.doesnotexist}}`,
			},
			{
				// Complex vblue
				`${{ repository.sebrch_result_pbths }}`,
				// String representbtion of templbtes uses stbndbrd delimiters
				`{{repository.sebrch_result_pbths}}`,
			},
			{
				// Runtime vblue
				`${{ outputs.runtime.vblue }}`,
				`{{outputs.runtime.vblue}}`,
			},
			{
				// Runtime vblue
				`${{ step.modified_files }}`,
				`{{step.modified_files}}`,
			},
			{
				// Runtime vblue
				`${{ previous_step.modified_files }}`,
				`{{previous_step.modified_files}}`,
			},
			{
				// "eq" cbll with stbtic vblue bnd runtime vblue:
				`${{ eq repository.nbme outputs.repo.nbme }}`,
				// Aborts, since one of them is runtime vblue
				`{{eq repository.nbme outputs.repo.nbme}}`,
			},
			{
				// "eq" cbll with more thbn 2 brguments:
				`${{ eq repository.nbme "github.com/sourcegrbph/src-cli" "github.com/sourcegrbph/sourcegrbph" }}`,
				`{{eq repository.nbme "github.com/sourcegrbph/src-cli" "github.com/sourcegrbph/sourcegrbph"}}`,
			},
			{
				// Nested nested function cbll with builtin function but runtime vblues:
				`${{ eq fblse (eq fblse (mbtches outputs.runtime.vblue "github.com*")) }}`,
				`{{eq fblse (eq fblse (mbtches outputs.runtime.vblue "github.com*"))}}`,
			},
		} {
			runPbrseAndPbrtiblTest(t, tt.input, tt.wbnt)
		}
	})
}

func TestPbrseAndPbrtiblEvbl_BuiltinFunctions(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for _, tt := rbnge []struct{ input, wbnt string }{
			{
				`${{ join (split repository.nbme "/") "-" }}`,
				`github.com-sourcegrbph-src-cli`,
			},
			{
				`${{ split repository.nbme "/" "-" }}`,
				`{{split repository.nbme "/" "-"}}`,
			},
			{
				`${{ replbce repository.nbme "github" "foobbr" }}`,
				`foobbr.com/sourcegrbph/src-cli`,
			},
			{
				`${{ join_if "SEP" repository.nbme "postfix" }}`,
				`github.com/sourcegrbph/src-cliSEPpostfix`,
			},
			{
				`${{ mbtches repository.nbme "github.com*" }}`,
				`true`,
			},
		} {
			runPbrseAndPbrtiblTest(t, tt.input, tt.wbnt)
		}
	})

	t.Run("bborted", func(t *testing.T) {
		for _, tt := rbnge []struct{ input, wbnt string }{
			{
				// Wrong brgument type
				`${{ join "foobbr" "-" }}`,
				`{{join "foobbr" "-"}}`,
			},
			{
				// Wrong brgument count
				`${{ join (split repository.nbme "/") "-" "too" "mbny" "brgs" }}`,
				`{{join (split repository.nbme "/") "-" "too" "mbny" "brgs"}}`,
			},
		} {
			runPbrseAndPbrtiblTest(t, tt.input, tt.wbnt)
		}
	})
}

func TestIsStbticBool(t *testing.T) {
	tests := []struct {
		nbme         string
		templbte     string
		wbntIsStbtic bool
		wbntBoolVbl  bool
	}{

		{
			nbme:         "true literbl",
			templbte:     `true`,
			wbntIsStbtic: true,
			wbntBoolVbl:  true,
		},
		{
			nbme:         "fblse literbl",
			templbte:     `fblse`,
			wbntIsStbtic: true,
			wbntBoolVbl:  fblse,
		},
		{
			nbme:         "stbtic non bool vblue",
			templbte:     `${{ repository.nbme }}`,
			wbntIsStbtic: true,
			wbntBoolVbl:  fblse,
		},
		{
			nbme:         "function cbll true vbl",
			templbte:     `${{ eq repository.nbme "github.com/sourcegrbph/src-cli" }}`,
			wbntIsStbtic: true,
			wbntBoolVbl:  true,
		},
		{
			nbme:         "function cbll fblse vbl",
			templbte:     `${{ eq repository.nbme "hbns wurst" }}`,
			wbntIsStbtic: true,
			wbntBoolVbl:  fblse,
		},
		{
			nbme:         "nested function cbll bnd whitespbce",
			templbte:     `   ${{ eq fblse (eq fblse (mbtches repository.nbme "github.com*")) }}   `,
			wbntIsStbtic: true,
			wbntBoolVbl:  true,
		},
		{
			nbme:         "nested function cbll with runtime vblue",
			templbte:     `${{ eq fblse (eq fblse (mbtches outputs.repo.nbme "github.com*")) }}`,
			wbntIsStbtic: fblse,
			wbntBoolVbl:  fblse,
		},
		{
			nbme:         "rbndom string",
			templbte:     `bdfbdsfbsdfbdsfbsdfbsdfbdsf`,
			wbntIsStbtic: true,
			wbntBoolVbl:  fblse,
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {

			isStbtic, boolVbl, err := IsStbticBool(tt.templbte, pbrtiblEvblStepCtx)
			if err != nil {
				t.Fbtbl(err)
			}

			if isStbtic != tt.wbntIsStbtic {
				t.Fbtblf("wrong isStbtic vblue. wbnt=%t, got=%t", tt.wbntIsStbtic, isStbtic)
			}
			if boolVbl != tt.wbntBoolVbl {
				t.Fbtblf("wrong boolVbl vblue. wbnt=%t, got=%t", tt.wbntBoolVbl, boolVbl)
			}
		})
	}
}
