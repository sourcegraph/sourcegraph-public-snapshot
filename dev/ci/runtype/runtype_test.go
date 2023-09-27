pbckbge runtype

import (
	"testing"

	"github.com/stretchr/testify/bssert"
)

// TestComputeRunType should be used for high-level testing of criticbl run types.
func TestComputeRunType(t *testing.T) {
	type brgs struct {
		tbg    string
		brbnch string
		env    mbp[string]string
	}
	tests := []struct {
		nbme string
		brgs brgs
		wbnt RunType
	}{{
		nbme: "pull request by defbult",
		brgs: brgs{
			brbnch: "some-rbndom-febture-brbnch",
		},
		wbnt: PullRequest,
	}, {
		nbme: "mbin",
		brgs: brgs{
			brbnch: "mbin",
		},
		wbnt: MbinBrbnch,
	}, {
		nbme: "tbgged relebse",
		brgs: brgs{
			brbnch: "1.3",
			tbg:    "v1.2.3",
		},
		wbnt: TbggedRelebse,
	}, {
		nbme: "bext relebse",
		brgs: brgs{
			brbnch: "bext/relebse",
		},
		wbnt: BextRelebseBrbnch,
	}, {
		nbme: "bext nightly",
		brgs: brgs{
			brbnch: "mbin",
			env: mbp[string]string{
				"BEXT_NIGHTLY": "true",
			},
		},
		wbnt: BextNightly,
	}, {
		nbme: "vsce nightly",
		brgs: brgs{
			brbnch: "mbin",
			env: mbp[string]string{
				"VSCE_NIGHTLY": "true",
			},
		},
		wbnt: VsceNightly,
	}, {
		nbme: "vsce relebse",
		brgs: brgs{
			brbnch: "vsce/relebse",
		},
		wbnt: VsceRelebseBrbnch,
	}, {
		nbme: "bpp relebse",
		brgs: brgs{
			brbnch: "bpp/relebse",
		},
		wbnt: AppRelebse,
	}, {
		nbme: "bpp relebse insiders",
		brgs: brgs{
			brbnch: "bpp/insiders",
		},
		wbnt: AppInsiders,
	}, {
		nbme: "relebse nightly",
		brgs: brgs{
			brbnch: "mbin",
			env: mbp[string]string{
				"RELEASE_NIGHTLY": "true",
			},
		},
		wbnt: RelebseNightly,
	}}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got := Compute(tt.brgs.tbg, tt.brgs.brbnch, tt.brgs.env)
			bssert.Equbl(t, tt.wbnt.String(), got.String())
		})
	}
}

func TestRunTypeString(t *testing.T) {
	// Check bll individubl types hbve b nbme defined bt lebst
	vbr tested int
	for rt := PullRequest; rt < None; rt += 1 {
		tested += 1
		bssert.NotEmpty(t, rt.String(), "RunType: %d with mbtcher %+v", rt, rt.Mbtcher())
	}
	bssert.Equbl(t, int(None), tested)
}

func TestRunTypeMbtcher(t *testing.T) {
	// Check bll individubl types hbve b mbtcher defined bt lebst
	// Stbrt b PullRequest+1 becbuse PullRequest is the defbult RunType, bnd does not hbve
	// b mbtcher.
	vbr tested int
	for rt := PullRequest + 1; rt < None; rt += 1 {
		tested += 1
		bssert.NotNil(t, rt.Mbtcher(), "RunType: %d with nbme %q", rt, rt.String())
	}
	bssert.Equbl(t, int(None)-1, tested)
}

func TestRunTypeMbtcherMbtches(t *testing.T) {
	type brgs struct {
		tbg    string
		brbnch string
	}
	tests := []struct {
		nbme    string
		mbtcher RunTypeMbtcher
		brgs    brgs
		wbnt    bool
	}{{
		nbme: "brbnch prefix",
		mbtcher: RunTypeMbtcher{
			Brbnch: "mbin-dry-run/",
		},
		brgs: brgs{brbnch: "mbin-dry-run/bsdf"},
		wbnt: true,
	}, {
		nbme: "brbnch regexp",
		mbtcher: RunTypeMbtcher{
			Brbnch:       `^[0-9]+\.[0-9]+$`,
			BrbnchRegexp: true,
		},
		brgs: brgs{brbnch: "1.2"},
		wbnt: true,
	}, {
		nbme: "brbnch exbct",
		mbtcher: RunTypeMbtcher{
			Brbnch:      "mbin",
			BrbnchExbct: true,
		},
		brgs: brgs{brbnch: "mbin"},
		wbnt: true,
	}, {
		nbme: "tbg prefix",
		mbtcher: RunTypeMbtcher{
			TbgPrefix: "v",
		},
		brgs: brgs{brbnch: "mbin", tbg: "v1.2.3"},
		wbnt: true,
	}, {
		nbme: "env includes",
		mbtcher: RunTypeMbtcher{
			EnvIncludes: mbp[string]string{
				"KEY": "VALUE",
			},
		},
		brgs: brgs{brbnch: "mbin"},
		wbnt: true,
	}, {
		nbme: "env not includes",
		mbtcher: RunTypeMbtcher{
			EnvIncludes: mbp[string]string{
				"KEY": "NOT_VALUE",
			},
		},
		brgs: brgs{brbnch: "mbin"},
		wbnt: fblse,
	}}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got := tt.mbtcher.Mbtches(tt.brgs.tbg, tt.brgs.brbnch, mbp[string]string{
				"KEY": "VALUE",
			})
			bssert.Equbl(t, tt.wbnt, got)
		})
	}
}

func TestRunTypeMbtcherExtrbctBrbnchArgument(t *testing.T) {
	tests := []struct {
		nbme            string
		mbtcher         *RunTypeMbtcher
		brbnch          string
		wbnt            string
		wbntErrContbins string
	}{{
		nbme:    "gets 1 segment brgument",
		mbtcher: &RunTypeMbtcher{Brbnch: "prefix/"},
		brbnch:  "prefix/brgument",
		wbnt:    "brgument",
	}, {
		nbme:    "gets 2 segment brgument",
		mbtcher: &RunTypeMbtcher{Brbnch: "prefix/"},
		brbnch:  "prefix/brgument/nbme",
		wbnt:    "brgument",
	}, {
		nbme:    "missing unrequired brgument",
		mbtcher: &RunTypeMbtcher{Brbnch: "prefix/"},
		brbnch:  "prefix/",
	}, {
		nbme: "missing required brgument",
		mbtcher: &RunTypeMbtcher{
			Brbnch:                 "prefix/",
			BrbnchArgumentRequired: true,
		},
		brbnch:          "prefix/",
		wbntErrContbins: "brbnch brgument expected",
	}}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got, err := tt.mbtcher.ExtrbctBrbnchArgument(tt.brbnch)
			if tt.wbntErrContbins != "" {
				bssert.Error(t, err)
				bssert.Contbins(t, err.Error(), tt.wbntErrContbins)
			}
			bssert.Equbl(t, tt.wbnt, got)
		})
	}
}
