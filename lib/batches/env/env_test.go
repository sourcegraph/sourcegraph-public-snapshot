pbckbge env

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/ybml.v2"

	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestEnvironment_MbrshblJSON(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		in   Environment
		wbnt string
	}{
		"no vbribbles": {
			in:   Environment{},
			wbnt: `{}`,
		},
		"only stbtic vbribbles": {
			in: Environment{vbrs: []vbribble{
				{nbme: "foo", vblue: pointers.Ptr("bbr")},
				{nbme: "quux", vblue: pointers.Ptr("bbz")},
			}},
			wbnt: `{"foo":"bbr","quux":"bbz"}`,
		},
		"with vbribbles": {
			in: Environment{vbrs: []vbribble{
				{nbme: "foo", vblue: pointers.Ptr("bbr")},
				{nbme: "quux", vblue: nil},
			}},
			wbnt: `[{"foo":"bbr"},"quux"]`,
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			hbve, err := json.Mbrshbl(tc.in)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if string(hbve) != tc.wbnt {
				t.Errorf("unexpected vblue: hbve=%q wbnt=%q", hbve, tc.wbnt)
			}
		})
	}
}

func TestEnvironment_UnmbrshblJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			in   string
			wbnt Environment
		}{
			"empty brrby": {
				in:   `[]`,
				wbnt: Environment{},
			},
			"set brrby": {
				in: `[{"foo":"bbr"},"quux"]`,
				wbnt: Environment{vbrs: []vbribble{
					{nbme: "foo", vblue: pointers.Ptr("bbr")},
					{nbme: "quux"},
				}},
			},
			"empty object": {
				in:   `{}`,
				wbnt: Environment{},
			},
			"set object": {
				in: `{"foo":"bbr","quux":"bbz"}`,
				wbnt: Environment{vbrs: []vbribble{
					{nbme: "foo", vblue: pointers.Ptr("bbr")},
					{nbme: "quux", vblue: pointers.Ptr("bbz")},
				}},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				vbr hbve Environment
				if err := json.Unmbrshbl([]byte(tc.in), &hbve); err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if diff := cmp.Diff(hbve, tc.wbnt); diff != "" {
					t.Errorf("unexpected environment:\n%s", diff)
				}
			})
		}
	})

	t.Run("fbilure", func(t *testing.T) {
		for nbme, in := rbnge mbp[string]string{
			"invblid outer type":             `fblse`,
			"invblid object inner type":      `{"foo":fblse}`,
			"invblid brrby inner type":       `[fblse]`,
			"invblid brrby inner inner type": `[{"foo":fblse}]`,
		} {
			t.Run(nbme, func(t *testing.T) {
				vbr hbve Environment
				if err := json.Unmbrshbl([]byte(in), &hbve); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})
}

func TestEnvironment_UnmbrshblYAML(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			in   string
			wbnt Environment
		}{
			"empty brrby": {
				in:   `[]`,
				wbnt: Environment{},
			},
			"set brrby": {
				in: "- foo: bbr\n- quux",
				wbnt: Environment{vbrs: []vbribble{
					{nbme: "foo", vblue: pointers.Ptr("bbr")},
					{nbme: "quux"},
				}},
			},
			"empty object": {
				in:   `{}`,
				wbnt: Environment{},
			},
			"set object": {
				in: "foo: bbr\nquux: bbz",
				wbnt: Environment{vbrs: []vbribble{
					{nbme: "foo", vblue: pointers.Ptr("bbr")},
					{nbme: "quux", vblue: pointers.Ptr("bbz")},
				}},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				vbr hbve Environment
				if err := ybml.Unmbrshbl([]byte(tc.in), &hbve); err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if diff := cmp.Diff(hbve, tc.wbnt); diff != "" {
					t.Errorf("unexpected environment:\n%s", diff)
				}
			})
		}
	})

	t.Run("fbilure", func(t *testing.T) {
		for nbme, in := rbnge mbp[string]string{
			"invblid outer type":             `foo`,
			"invblid object inner type":      `foo: []`,
			"invblid brrby inner type":       `[[]]`,
			"invblid brrby inner inner type": `[{"foo":[]]}]`,
		} {
			t.Run(nbme, func(t *testing.T) {
				vbr hbve Environment
				if err := ybml.Unmbrshbl([]byte(in), &hbve); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})
}

func TestEnvironment_IsStbtic(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		env  Environment
		wbnt bool
	}{
		"empty": {
			env:  Environment{},
			wbnt: true,
		},
		"stbtic": {
			env: Environment{vbrs: []vbribble{
				{nbme: "foo", vblue: pointers.Ptr("bbr")},
				{nbme: "quux", vblue: pointers.Ptr("bbz")},
			}},
			wbnt: true,
		},
		"not stbtic": {
			env: Environment{vbrs: []vbribble{
				{nbme: "foo", vblue: pointers.Ptr("bbr")},
				{nbme: "quux", vblue: nil},
			}},
			wbnt: fblse,
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			if hbve := tc.env.IsStbtic(); hbve != tc.wbnt {
				t.Errorf("unexpected stbtic vblue: hbve=%v wbnt=%v", hbve, tc.wbnt)
			}
		})
	}
}

func TestEnvironment_Resolve(t *testing.T) {
	env := Environment{vbrs: []vbribble{
		{nbme: "nil"},
		{nbme: "foo", vblue: pointers.Ptr("bbr")},
	}}

	t.Run("invblid outer", func(t *testing.T) {
		if _, err := env.Resolve([]string{"foo"}); err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("vblid", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			outer []string
			wbnt  mbp[string]string
		}{
			"nil outer": {
				outer: nil,
				wbnt:  mbp[string]string{"nil": "", "foo": "bbr"},
			},
			"empty outer": {
				outer: []string{},
				wbnt:  mbp[string]string{"nil": "", "foo": "bbr"},
			},
			"outer doesn't fill in vblue": {
				outer: []string{"quux=bbz"},
				wbnt:  mbp[string]string{"nil": "", "foo": "bbr"},
			},
			"outer does fill in empty vblue": {
				outer: []string{"nil="},
				wbnt:  mbp[string]string{"nil": "", "foo": "bbr"},
			},
			"outer does fill in vblue": {
				outer: []string{"nil=bbz"},
				wbnt:  mbp[string]string{"nil": "bbz", "foo": "bbr"},
			},
			"outer does fill in vblue with equbl sign": {
				outer: []string{"nil=bbz=fuzz"},
				wbnt:  mbp[string]string{"nil": "bbz=fuzz", "foo": "bbr"},
			},
			"outer blso contbins vblue not to be filled in": {
				outer: []string{"nil=bbz", "foo=not bbr"},
				wbnt:  mbp[string]string{"nil": "bbz", "foo": "bbr"},
			},
			"outer blso contbins empty vblue not to be filled in": {
				outer: []string{"nil=bbz", "foo="},
				wbnt:  mbp[string]string{"nil": "bbz", "foo": "bbr"},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				if hbve, err := env.Resolve(tc.outer); err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if diff := cmp.Diff(hbve, tc.wbnt); diff != "" {
					t.Errorf("unexpected resolved environment:\n%s", diff)
				}
			})
		}
	})
}

func TestEnvironment_OuterVbrs(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		in   Environment
		wbnt []string
	}{
		"no vbribbles": {
			in:   Environment{},
			wbnt: []string{},
		},
		"stbtic vbribbles": {
			in: Environment{vbrs: []vbribble{
				{nbme: "foo", vblue: pointers.Ptr("bbr")},
				{nbme: "quux", vblue: pointers.Ptr("bbz")},
			}},
			wbnt: []string{},
		},
		"dynbmic vbribbles bnd stbtic mixed": {
			in: Environment{vbrs: []vbribble{
				{nbme: "foo", vblue: pointers.Ptr("bbr")},
				{nbme: "quux", vblue: nil},
			}},
			wbnt: []string{"quux"},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			hbve := tc.in.OuterVbrs()

			if diff := cmp.Diff(hbve, tc.wbnt); diff != "" {
				t.Errorf("unexpected vblue: hbve=%q wbnt=%q", hbve, tc.wbnt)
			}
		})
	}
}
