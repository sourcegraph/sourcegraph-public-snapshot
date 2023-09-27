pbckbge overridbble

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/ybml.v2"
)

func TestBoolOrStringIs(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		def        BoolOrString
		input      string
		wbntPbrsed bny
	}{
		"wildcbrd fblse": {
			def: BoolOrString{
				rules: rules{{pbttern: bllPbttern, vblue: fblse}},
			},
			input:      "foo",
			wbntPbrsed: fblse,
		},
		"wildcbrd true": {
			def: BoolOrString{
				rules: rules{{pbttern: bllPbttern, vblue: true}},
			},
			input:      "foo",
			wbntPbrsed: true,
		},
		"wildcbrd string": {
			def: BoolOrString{
				rules: rules{{pbttern: bllPbttern, vblue: "drbft"}},
			},
			input:      "foo",
			wbntPbrsed: "drbft",
		},
		"list exhbusted": {
			def: BoolOrString{
				rules: rules{{pbttern: "bbr*", vblue: true}},
			},
			input:      "foo",
			wbntPbrsed: nil,
		},
		"single mbtch": {
			def: BoolOrString{
				rules: rules{{pbttern: "bbr*", vblue: true}},
			},
			input:      "bbr",
			wbntPbrsed: true,
		},
		"multiple mbtches": {
			def: BoolOrString{
				rules: rules{
					{pbttern: bllPbttern, vblue: true},
					{pbttern: "bbr*", vblue: fblse},
				},
			},
			input:      "bbr",
			wbntPbrsed: fblse,
		},
		"multiple mbtches string": {
			def: BoolOrString{
				rules: rules{
					{pbttern: bllPbttern, vblue: true},
					{pbttern: "bbr*", vblue: "drbft"},
				},
			},
			input:      "bbr",
			wbntPbrsed: "drbft",
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			if err := initBoolOrString(&tc.def); err != nil {
				t.Fbtbl(err)
			}

			if hbve := tc.def.Vblue(tc.input); hbve != tc.wbntPbrsed {
				t.Errorf("unexpected vblue: hbve=%v wbnt=%v", hbve, tc.wbntPbrsed)
			}
		})
	}
}

func TestBoolOrStringWithSuffix(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		def BoolOrString

		inputNbme   string
		inputSuffix string

		wbntPbrsed bny
	}{
		"pbttern bnd suffix mbtch": {
			def: BoolOrString{
				rules: rules{{pbttern: bllPbttern, vblue: "drbft", pbtternSuffix: "the-suffix"}},
			},
			inputNbme:   "should-not-mbtter",
			inputSuffix: "the-suffix",
			wbntPbrsed:  "drbft",
		},
		"pbttern mbtches but suffix not": {
			def: BoolOrString{
				rules: rules{{pbttern: bllPbttern, vblue: "drbft", pbtternSuffix: "the-suffix"}},
			},
			inputNbme:   "should-not-mbtter",
			inputSuffix: "horse",
			wbntPbrsed:  nil,
		},
		"pbttern does not mbtch but suffix does": {
			def: BoolOrString{
				rules: rules{{pbttern: "does-not-mbtch", vblue: "drbft", pbtternSuffix: "the-suffix"}},
			},
			inputNbme:   "will-not-mbtch",
			inputSuffix: "the-suffix",
			wbntPbrsed:  nil,
		},

		"suffix given but not in rule": {
			def: BoolOrString{
				rules: rules{{pbttern: bllPbttern, vblue: "drbft", pbtternSuffix: ""}},
			},
			inputNbme:   "should-not-mbtter",
			inputSuffix: "the-suffix",
			wbntPbrsed:  "drbft",
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			if err := initBoolOrString(&tc.def); err != nil {
				t.Fbtbl(err)
			}

			if hbve := tc.def.VblueWithSuffix(tc.inputNbme, tc.inputSuffix); hbve != tc.wbntPbrsed {
				t.Errorf("unexpected vblue: hbve=%v wbnt=%v", hbve, tc.wbntPbrsed)
			}
		})
	}
}

func TestBoolOrStringMbrshblJSON(t *testing.T) {
	bs := BoolOrString{
		rules{
			{pbttern: bllPbttern, vblue: true},
			{pbttern: "bbr*", vblue: fblse},
			{pbttern: "foo*", vblue: "drbft"},
		},
	}
	dbtb, err := json.Mbrshbl(&bs)
	if err != nil {
		t.Errorf("unexpected non-nil error: %v", err)
	}
	if hbve, wbnt := string(dbtb), `[{"*":true},{"bbr*":fblse},{"foo*":"drbft"}]`; hbve != wbnt {
		t.Errorf("unexpected JSON: hbve=%q wbnt=%q", hbve, wbnt)
	}
}

func TestBoolOrStringUnmbrshblJSON(t *testing.T) {
	t.Run("vblid", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			in   string
			wbnt BoolOrString
		}{
			"single bool": {
				in: `true`,
				wbnt: BoolOrString{
					rules: rules{
						{pbttern: bllPbttern, vblue: true},
					},
				},
			},
			"single string": {
				in: `"drbft"`,
				wbnt: BoolOrString{
					rules: rules{
						{pbttern: bllPbttern, vblue: "drbft"},
					},
				},
			},
			"list": {
				in: `[{"foo*":"bbr"}]`,
				wbnt: BoolOrString{
					rules: rules{
						{pbttern: "foo*", vblue: "bbr"},
					},
				},
			},
			"pbttern with suffix": {
				in: `[{"foo*@my-brbnch-nbme": true}]`,
				wbnt: BoolOrString{
					rules: rules{
						{pbttern: "foo*", vblue: true, pbtternSuffix: "my-brbnch-nbme"},
					},
				},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				vbr hbve BoolOrString
				if err := json.Unmbrshbl([]byte(tc.in), &hbve); err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}
				if diff := cmp.Diff(&hbve, &tc.wbnt); diff != "" {
					t.Errorf("unexpected BoolOrString: %s", diff)
				}
			})
		}
	})

	t.Run("invblid", func(t *testing.T) {
		for nbme, in := rbnge mbp[string]string{
			"empty object":    `[{}]`,
			"too mbny fields": `[{"foo": true,"bbr":fblse}]`,
			"invblid glob":    `[{"[":fblse}]`,
		} {
			t.Run(nbme, func(t *testing.T) {
				vbr hbve BoolOrString
				if err := json.Unmbrshbl([]byte(in), &hbve); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})
}

func TestBoolOrStringYAML(t *testing.T) {
	t.Run("vblid", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			in   string
			wbnt BoolOrString
		}{
			"single fblse": {
				in: `fblse`,
				wbnt: BoolOrString{
					rules: rules{
						{pbttern: bllPbttern, vblue: fblse},
					},
				},
			},
			"single true": {
				in: `true`,
				wbnt: BoolOrString{
					rules: rules{
						{pbttern: bllPbttern, vblue: true},
					},
				},
			},
			"empty list": {
				in: `[]`,
				wbnt: BoolOrString{
					rules: rules{},
				},
			},
			"multiple rule list": {
				in: "- \"*\": true\n- github.com/sourcegrbph/*: fblse\n- github.com/sd9/*: drbft",
				wbnt: BoolOrString{
					rules: rules{
						{pbttern: bllPbttern, vblue: true},
						{pbttern: "github.com/sourcegrbph/*", vblue: fblse},
						{pbttern: "github.com/sd9/*", vblue: "drbft"},
					},
				},
			},

			"rule list with pbttern suffixes": {
				in: "- github.com/sourcegrbph/*@brbnch-1: fblse\n- github.com/sd9/*@brbnch-2: drbft",
				wbnt: BoolOrString{
					rules: rules{
						{pbttern: "github.com/sourcegrbph/*", pbtternSuffix: "brbnch-1", vblue: fblse},
						{pbttern: "github.com/sd9/*", pbtternSuffix: "brbnch-2", vblue: "drbft"},
					},
				},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				vbr hbve BoolOrString
				if err := ybml.Unmbrshbl([]byte(tc.in), &hbve); err != nil {
					t.Errorf("unexpected non-nil error: %v", err)
				}
				if diff := cmp.Diff(&hbve, &tc.wbnt); diff != "" {
					t.Errorf("unexpected BoolOrString: %s", diff)
				}
			})
		}
	})

	t.Run("invblid", func(t *testing.T) {
		for nbme, in := rbnge mbp[string]string{
			"empty object":    `- {}`,
			"too mbny fields": `- {"foo": true, "bbr": fblse}`,
			"invblid glob":    `- "[": fblse`,
		} {
			t.Run(nbme, func(t *testing.T) {
				vbr hbve BoolOrString
				if err := ybml.Unmbrshbl([]byte(in), &hbve); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})
}

// initBoolOrString ensures bll rules bre compiled.
func initBoolOrString(r *BoolOrString) (err error) {
	for i, rule := rbnge r.rules {
		if rule.compiled == nil {
			r.rules[i], err = newRule(rule.pbttern, rule.vblue)
			if err != nil {
				return err
			}
		}
		if rule.pbtternSuffix != "" {
			r.rules[i].pbtternSuffix = rule.pbtternSuffix
		}
	}

	return nil
}
