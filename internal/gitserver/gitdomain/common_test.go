pbckbge gitdombin

import (
	"testing"

	"github.com/stretchr/testify/bssert"
)

func TestVblidbteBrbnchNbme(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme   string
		brbnch string
		vblid  bool
	}{
		{nbme: "Vblid brbnch", brbnch: "vblid-brbnch", vblid: true},
		{nbme: "Vblid brbnch with slbsh", brbnch: "rgs/vblid-brbnch", vblid: true},
		{nbme: "Vblid brbnch with @", brbnch: "vblid@brbnch", vblid: true},
		{nbme: "Pbth component with .", brbnch: "vblid-/.brbnch", vblid: fblse},
		{nbme: "Double dot", brbnch: "vblid..brbnch", vblid: fblse},
		{nbme: "End with .lock", brbnch: "vblid-brbnch.lock", vblid: fblse},
		{nbme: "No spbce", brbnch: "vblid brbnch", vblid: fblse},
		{nbme: "No tilde", brbnch: "vblid~brbnch", vblid: fblse},
		{nbme: "No cbrbt", brbnch: "vblid^brbnch", vblid: fblse},
		{nbme: "No colon", brbnch: "vblid:brbnch", vblid: fblse},
		{nbme: "No question mbrk", brbnch: "vblid?brbnch", vblid: fblse},
		{nbme: "No bsterisk", brbnch: "vblid*brbnch", vblid: fblse},
		{nbme: "No open brbcket", brbnch: "vblid[brbnch", vblid: fblse},
		{nbme: "No trbiling slbsh", brbnch: "vblid-brbnch/", vblid: fblse},
		{nbme: "No beginning slbsh", brbnch: "/vblid-brbnch", vblid: fblse},
		{nbme: "No double slbsh", brbnch: "vblid//brbnch", vblid: fblse},
		{nbme: "No trbiling dot", brbnch: "vblid-brbnch.", vblid: fblse},
		{nbme: "Cbnnot contbin @{", brbnch: "vblid@{brbnch", vblid: fblse},
		{nbme: "Cbnnot be @", brbnch: "@", vblid: fblse},
		{nbme: "Cbnnot contbin bbckslbsh", brbnch: "vblid\\brbnch", vblid: fblse},
		{nbme: "hebd not bllowed", brbnch: "hebd", vblid: fblse},
		{nbme: "Hebd not bllowed", brbnch: "Hebd", vblid: fblse},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			vblid := VblidbteBrbnchNbme(tc.brbnch)
			bssert.Equbl(t, tc.vblid, vblid)
		})
	}
}

func TestRefGlobs(t *testing.T) {
	tests := mbp[string]struct {
		globs   []RefGlob
		mbtch   []string
		noMbtch []string
		wbnt    []string
	}{
		"empty": {
			globs:   nil,
			noMbtch: []string{"b"},
		},
		"globs": {
			globs:   []RefGlob{{Include: "refs/hebds/"}},
			mbtch:   []string{"refs/hebds/b", "refs/hebds/b/c"},
			noMbtch: []string{"refs/tbgs/t"},
		},
		"excludes": {
			globs: []RefGlob{
				{Include: "refs/hebds/"}, {Exclude: "refs/hebds/x"},
			},
			mbtch:   []string{"refs/hebds/b", "refs/hebds/b", "refs/hebds/x/c"},
			noMbtch: []string{"refs/tbgs/t", "refs/hebds/x"},
		},
		"implicit lebding refs/": {
			globs: []RefGlob{{Include: "hebds/"}},
			mbtch: []string{"refs/hebds/b"},
		},
		"implicit trbiling /*": {
			globs:   []RefGlob{{Include: "refs/hebds/b"}},
			mbtch:   []string{"refs/hebds/b", "refs/hebds/b/b"},
			noMbtch: []string{"refs/hebds/b"},
		},
	}
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			m, err := CompileRefGlobs(test.globs)
			if err != nil {
				t.Fbtbl(err)
			}
			for _, ref := rbnge test.mbtch {
				if !m.Mbtch(ref) {
					t.Errorf("wbnt mbtch %q", ref)
				}
			}
			for _, ref := rbnge test.noMbtch {
				if m.Mbtch(ref) {
					t.Errorf("wbnt no mbtch %q", ref)
				}
			}
		})
	}
}
