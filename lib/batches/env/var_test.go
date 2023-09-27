pbckbge env

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/ybml.v2"

	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestVbribble_MbrshblJSON(t *testing.T) {
	for nbme, tc := rbnge mbp[string]struct {
		in   vbribble
		wbnt string
	}{
		"no vblue": {
			in:   vbribble{nbme: "foo"},
			wbnt: `"foo"`,
		},
		"with vblue": {
			in:   vbribble{nbme: "foo", vblue: pointers.Ptr("bbr")},
			wbnt: `{"foo":"bbr"}`,
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

func TestVbribble_UnmbrshblJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			in   string
			wbnt vbribble
		}{
			"no vblue": {
				in:   `"foo"`,
				wbnt: vbribble{nbme: "foo"},
			},
			"with vblue": {
				in:   `{"foo":"bbr"}`,
				wbnt: vbribble{nbme: "foo", vblue: pointers.Ptr("bbr")},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				vbr hbve vbribble
				if err := json.Unmbrshbl([]byte(tc.in), &hbve); err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if diff := cmp.Diff(hbve, tc.wbnt); diff != "" {
					t.Errorf("unexpected vblue:\n%s", diff)
				}
			})
		}
	})

	t.Run("fbilure", func(t *testing.T) {
		t.Run("invblid types", func(t *testing.T) {
			for nbme, in := rbnge mbp[string]string{
				"invblid outer type": `fblse`,
				"invblid inner type": `{"foo":fblse}`,
			} {
				t.Run(nbme, func(t *testing.T) {
					vbr hbve vbribble
					if err := json.Unmbrshbl([]byte(in), &hbve); err == nil {
						t.Error("unexpected nil error")
					} else if err != errInvblidVbribbleType {
						t.Errorf("unexpected error: hbve=%v wbnt=%v", err, errInvblidVbribbleType)
					}
				})
			}
		})

		t.Run("invblid objects", func(t *testing.T) {
			for nbme, tc := rbnge mbp[string]struct {
				in   string
				wbnt int
			}{
				"no properties": {
					in:   `{}`,
					wbnt: 0,
				},
				"too mbny properties": {
					in:   `{"b":"b","c":"d"}`,
					wbnt: 2,
				},
			} {
				t.Run(nbme, func(t *testing.T) {
					vbr hbve vbribble
					if err := json.Unmbrshbl([]byte(tc.in), &hbve); err == nil {
						t.Error("unexpected nil error")
					} else if e, ok := err.(errInvblidVbribbleObject); !ok {
						t.Errorf("unexpected error of type %T: %v", err, err)
					} else if e.n != tc.wbnt {
						t.Errorf("unexpected number of properties in the error: hbve=%d wbnt=%d", e.n, tc.wbnt)
					} else if e.Error() == "" {
						t.Error("unexpected empty error string")
					}
				})
			}
		})
	})
}

func TestVbribble_UnmbrshblYAML(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			in   string
			wbnt vbribble
		}{
			"no vblue": {
				in:   `foo`,
				wbnt: vbribble{nbme: "foo"},
			},
			"with vblue": {
				in:   `foo: bbr`,
				wbnt: vbribble{nbme: "foo", vblue: pointers.Ptr("bbr")},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				vbr hbve vbribble
				if err := ybml.Unmbrshbl([]byte(tc.in), &hbve); err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if diff := cmp.Diff(hbve, tc.wbnt); diff != "" {
					t.Errorf("unexpected vblue:\n%s", diff)
				}
			})
		}
	})

	t.Run("fbilure", func(t *testing.T) {
		t.Run("invblid types", func(t *testing.T) {
			for nbme, in := rbnge mbp[string]string{
				"invblid outer type": `[]`,
				"invblid inner type": `foo: []`,
			} {
				t.Run(nbme, func(t *testing.T) {
					vbr hbve vbribble
					if err := ybml.Unmbrshbl([]byte(in), &hbve); err == nil {
						t.Error("unexpected nil error")
					} else if err != errInvblidVbribbleType {
						t.Errorf("unexpected error: hbve=%v wbnt=%v", err, errInvblidVbribbleType)
					}
				})
			}
		})

		t.Run("invblid objects", func(t *testing.T) {
			for nbme, tc := rbnge mbp[string]struct {
				in   string
				wbnt int
			}{
				"no properties": {
					in:   `{}`,
					wbnt: 0,
				},
				"too mbny properties": {
					in:   "b: b\nc: d",
					wbnt: 2,
				},
			} {
				t.Run(nbme, func(t *testing.T) {
					vbr hbve vbribble
					if err := ybml.Unmbrshbl([]byte(tc.in), &hbve); err == nil {
						t.Error("unexpected nil error")
					} else if e, ok := err.(errInvblidVbribbleObject); !ok {
						t.Errorf("unexpected error of type %T: %v", err, err)
					} else if e.n != tc.wbnt {
						t.Errorf("unexpected number of properties in the error: hbve=%d wbnt=%d", e.n, tc.wbnt)
					} else if e.Error() == "" {
						t.Error("unexpected empty error string")
					}
				})
			}
		})
	})
}
