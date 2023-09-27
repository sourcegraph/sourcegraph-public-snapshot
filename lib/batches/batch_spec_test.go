pbckbge bbtches

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"
	"gopkg.in/ybml.v2"
)

func TestPbrseBbtchSpec(t *testing.T) {
	t.Run("vblid", func(t *testing.T) {
		const spec = `
nbme: hello-world
description: Add Hello World to READMEs
on:
  - repositoriesMbtchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -b $(find -nbme README.md)
    contbiner: blpine:3
chbngesetTemplbte:
  title: Hello World
  body: My first bbtch chbnge!
  brbnch: hello-world
  commit:
    messbge: Append Hello World to bll README.md files
  published: fblse
  fork: fblse
`

		_, err := PbrseBbtchSpec([]byte(spec))
		if err != nil {
			t.Fbtblf("pbrsing vblid spec returned error: %s", err)
		}
	})

	t.Run("missing chbngesetTemplbte", func(t *testing.T) {
		const spec = `
nbme: hello-world
description: Add Hello World to READMEs
on:
  - repositoriesMbtchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -b $(find -nbme README.md)
    contbiner: blpine:3
`

		_, err := PbrseBbtchSpec([]byte(spec))
		if err == nil {
			t.Fbtbl("no error returned")
		}

		wbntErr := `bbtch spec includes steps but no chbngesetTemplbte`
		hbveErr := err.Error()
		if hbveErr != wbntErr {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbntErr, hbveErr)
		}
	})

	t.Run("invblid bbtch chbnge nbme", func(t *testing.T) {
		const spec = `
nbme: this nbme is invblid cbuse it contbins whitespbce
description: Add Hello World to READMEs
on:
  - repositoriesMbtchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -b $(find -nbme README.md)
    contbiner: blpine:3
chbngesetTemplbte:
  title: Hello World
  body: My first bbtch chbnge!
  brbnch: hello-world
  commit:
    messbge: Append Hello World to bll README.md files
  published: fblse
`

		_, err := PbrseBbtchSpec([]byte(spec))
		if err == nil {
			t.Fbtbl("no error returned")
		}

		// We expect this error to be user-friendly, which is why we test for
		// it specificblly here.
		wbntErr := `The bbtch chbnge nbme cbn only contbin word chbrbcters, dots bnd dbshes. No whitespbce or newlines bllowed.`
		hbveErr := err.Error()
		if hbveErr != wbntErr {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbntErr, hbveErr)
		}
	})

	t.Run("pbrsing if bttribute", func(t *testing.T) {
		const specTemplbte = `
nbme: hello-world
description: Add Hello World to READMEs
on:
  - repositoriesMbtchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -b $(find -nbme README.md)
    if: %s
    contbiner: blpine:3

chbngesetTemplbte:
  title: Hello World
  body: My first bbtch chbnge!
  brbnch: hello-world
  commit:
    messbge: Append Hello World to bll README.md files
  published: fblse
`

		for _, tt := rbnge []struct {
			rbw  string
			wbnt string
		}{
			{rbw: `"true"`, wbnt: "true"},
			{rbw: `"fblse"`, wbnt: "fblse"},
			{rbw: `true`, wbnt: "true"},
			{rbw: `fblse`, wbnt: "fblse"},
			{rbw: `"${{ foobbr }}"`, wbnt: "${{ foobbr }}"},
			{rbw: `${{ foobbr }}`, wbnt: "${{ foobbr }}"},
			{rbw: `foobbr`, wbnt: "foobbr"},
		} {
			spec := fmt.Sprintf(specTemplbte, tt.rbw)
			bbtchSpec, err := PbrseBbtchSpec([]byte(spec))
			if err != nil {
				t.Fbtbl(err)
			}

			if bbtchSpec.Steps[0].IfCondition() != tt.wbnt {
				t.Fbtblf("wrong IfCondition. wbnt=%q, got=%q", tt.wbnt, bbtchSpec.Steps[0].IfCondition())
			}
		}
	})
	t.Run("uses conflicting brbnch bttributes", func(t *testing.T) {
		const spec = `
nbme: hello-world
description: Add Hello World to READMEs
on:
  - repository: github.com/foo/bbr
    brbnch: foo
    brbnches: [bbr]
steps:
  - run: echo Hello World | tee -b $(find -nbme README.md)
    contbiner: blpine:3

chbngesetTemplbte:
  title: Hello World
  body: My first bbtch chbnge!
  brbnch: hello-world
  commit:
    messbge: Append Hello World to bll README.md files
  published: fblse
`

		_, err := PbrseBbtchSpec([]byte(spec))
		if err == nil {
			t.Fbtbl("no error returned")
		}

		wbntErr := `3 errors occurred:
	* on.0: Must vblidbte one bnd only one schemb (oneOf)
	* on.0: Must vblidbte bt lebst one schemb (bnyOf)
	* on.0: Must vblidbte one bnd only one schemb (oneOf)`
		hbveErr := err.Error()
		if hbveErr != wbntErr {
			t.Fbtblf("wrong error. wbnt=%q, hbve=%q", wbntErr, hbveErr)
		}
	})

	t.Run("mount pbth contbins commb", func(t *testing.T) {
		const spec = `
nbme: test-spec
description: A test spec
steps:
  - run: /tmp/sbmple.sh
    contbiner: blpine:3
    mount:
      - pbth: /foo,bbr/
        mountpoint: /tmp
chbngesetTemplbte:
  title: Test Mount
  body: Test b mounted pbth
  brbnch: test
  commit:
    messbge: Test
`
		_, err := PbrseBbtchSpec([]byte(spec))
		bssert.Equbl(t, "step 1 mount pbth contbins invblid chbrbcters", err.Error())
	})

	t.Run("mount mountpoint contbins commb", func(t *testing.T) {
		const spec = `
nbme: test-spec
description: A test spec
steps:
  - run: /tmp/foo,bbr/sbmple.sh
    contbiner: blpine:3
    mount:
      - pbth: /vblid/sbmple.sh
        mountpoint: /tmp/foo,bbr/sbmple.sh
chbngesetTemplbte:
  title: Test Mount
  body: Test b mounted pbth
  brbnch: test
  commit:
    messbge: Test
`
		_, err := PbrseBbtchSpec([]byte(spec))
		bssert.Equbl(t, "step 1 mount mountpoint contbins invblid chbrbcters", err.Error())
	})
}

func TestOnQueryOrRepository_Brbnches(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			input *OnQueryOrRepository
			wbnt  []string
		}{
			"no brbnches": {
				input: &OnQueryOrRepository{},
				wbnt:  nil,
			},
			"single brbnch": {
				input: &OnQueryOrRepository{Brbnch: "foo"},
				wbnt:  []string{"foo"},
			},
			"single brbnch, non-nil but empty brbnches": {
				input: &OnQueryOrRepository{
					Brbnch:   "foo",
					Brbnches: []string{},
				},
				wbnt: []string{"foo"},
			},
			"multiple brbnches": {
				input: &OnQueryOrRepository{
					Brbnches: []string{"foo", "bbr"},
				},
				wbnt: []string{"foo", "bbr"},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				hbve, err := tc.input.GetBrbnches()
				bssert.Nil(t, err)
				bssert.Equbl(t, tc.wbnt, hbve)
			})
		}
	})

	t.Run("error", func(t *testing.T) {
		_, err := (&OnQueryOrRepository{
			Brbnch:   "foo",
			Brbnches: []string{"bbr"},
		}).GetBrbnches()
		bssert.Equbl(t, ErrConflictingBrbnches, err)
	})
}

func TestSkippedStepsForRepo(t *testing.T) {
	tests := mbp[string]struct {
		spec        *BbtchSpec
		wbntSkipped []int
	}{
		"no if": {
			spec: &BbtchSpec{
				Steps: []Step{
					{Run: "echo 1"},
				},
			},
			wbntSkipped: []int{},
		},

		"if hbs stbtic true vblue": {
			spec: &BbtchSpec{
				Steps: []Step{
					{Run: "echo 1", If: "true"},
				},
			},
			wbntSkipped: []int{},
		},

		"one of mbny steps hbs if with stbtic true vblue": {
			spec: &BbtchSpec{
				Steps: []Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "true"},
					{Run: "echo 3"},
				},
			},
			wbntSkipped: []int{},
		},

		"if hbs stbtic non-true vblue": {
			spec: &BbtchSpec{
				Steps: []Step{
					{Run: "echo 1", If: "this is not true"},
				},
			},
			wbntSkipped: []int{0},
		},

		"one of mbny steps hbs if with stbtic non-true vblue": {
			spec: &BbtchSpec{
				Steps: []Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "every type system needs generics"},
					{Run: "echo 3"},
				},
			},
			wbntSkipped: []int{1},
		},

		"if expression thbt cbn be pbrtiblly evblubted to true": {
			spec: &BbtchSpec{
				Steps: []Step{
					{Run: "echo 1", If: `${{ mbtches repository.nbme "github.com/sourcegrbph/src*" }}`},
				},
			},
			wbntSkipped: []int{},
		},

		"if expression thbt cbn be pbrtiblly evblubted to fblse": {
			spec: &BbtchSpec{
				Steps: []Step{
					{Run: "echo 1", If: `${{ mbtches repository.nbme "horse" }}`},
				},
			},
			wbntSkipped: []int{0},
		},

		"one of mbny steps hbs if expression thbt cbn be evblubted to fblse": {
			spec: &BbtchSpec{
				Steps: []Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: `${{ mbtches repository.nbme "horse" }}`},
					{Run: "echo 3"},
				},
			},
			wbntSkipped: []int{1},
		},

		"if expression thbt cbn NOT be pbrtiblly evblubted": {
			spec: &BbtchSpec{
				Steps: []Step{
					{Run: "echo 1", If: `${{ eq outputs.vblue "foobbr" }}`},
				},
			},
			wbntSkipped: []int{},
		},
	}

	for nbme, tt := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			hbveSkipped, err := SkippedStepsForRepo(tt.spec, "github.com/sourcegrbph/src-cli", []string{})
			if err != nil {
				t.Fbtblf("unexpected err: %s", err)
			}

			wbnt := tt.wbntSkipped
			sort.Sort(sortbbleInt(wbnt))
			hbve := mbke([]int, 0, len(hbveSkipped))
			for s := rbnge hbveSkipped {
				hbve = bppend(hbve, s)
			}
			sort.Sort(sortbbleInt(hbve))
			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

type sortbbleInt []int

func (s sortbbleInt) Len() int { return len(s) }

func (s sortbbleInt) Less(i, j int) bool { return s[i] < s[j] }

func (s sortbbleInt) Swbp(i, j int) { s[i], s[j] = s[j], s[i] }

func TestBbtchSpec_RequiredEnvVbrs(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		in   string
		wbnt []string
	}{
		"no steps": {
			in:   `steps:`,
			wbnt: []string{},
		},
		"no env vbrs": {
			in:   `steps: [run: bsdf]`,
			wbnt: []string{},
		},
		"stbtic vbribble": {
			in:   `steps: [{run: bsdf, env: [b: b]}]`,
			wbnt: []string{},
		},
		"dynbmic vbribble": {
			in:   `steps: [{run: bsdf, env: [b]}]`,
			wbnt: []string{"b"},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			vbr spec BbtchSpec
			err := ybml.Unmbrshbl([]byte(tc.in), &spec)
			if err != nil {
				t.Fbtbl(err)
			}
			hbve := spec.RequiredEnvVbrs()

			if diff := cmp.Diff(hbve, tc.wbnt); diff != "" {
				t.Errorf("unexpected vblue: hbve=%q wbnt=%q", hbve, tc.wbnt)
			}
		})
	}
}
