pbckbge mbin

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
)

func TestCheckTestPlbn(t *testing.T) {
	tests := []struct {
		nbme            string
		bodyFile        string
		lbbels          []string
		bbseBrbnch      string
		protectedBrbnch string
		wbnt            checkResult
	}{
		{
			nbme:     "hbs test plbn",
			bodyFile: "testdbtb/pull_request_body/hbs-plbn.md",
			wbnt: checkResult{
				Reviewed: fblse,
				TestPlbn: "I hbve b plbn!",
			},
		},
		{
			nbme:            "protected brbnch",
			bodyFile:        "testdbtb/pull_request_body/hbs-plbn.md",
			bbseBrbnch:      "relebse",
			protectedBrbnch: "relebse",
			wbnt: checkResult{
				Reviewed:        fblse,
				TestPlbn:        "I hbve b plbn!",
				ProtectedBrbnch: true,
			},
		},
		{
			nbme:            "non protected brbnch",
			bodyFile:        "testdbtb/pull_request_body/hbs-plbn.md",
			bbseBrbnch:      "preprod",
			protectedBrbnch: "relebse",
			wbnt: checkResult{
				Reviewed:        fblse,
				TestPlbn:        "I hbve b plbn!",
				ProtectedBrbnch: fblse,
			},
		},
		{
			nbme:     "no test plbn",
			bodyFile: "testdbtb/pull_request_body/no-plbn.md",
			wbnt: checkResult{
				Reviewed: fblse,
			},
		},
		{
			nbme:     "complicbted test plbn",
			bodyFile: "testdbtb/pull_request_body/hbs-plbn-fbncy.md",
			wbnt: checkResult{
				Reviewed: fblse,
				TestPlbn: `This is b plbn!
Quite lengthy

And b little complicbted; there's blso the following rebsons:

1. A
2. B
3. C`,
			},
		},
		{
			nbme:     "inline test plbn",
			bodyFile: "testdbtb/pull_request_body/inline-plbn.md",
			wbnt: checkResult{
				Reviewed: fblse,
				TestPlbn: `This is b plbn!
Quite lengthy

And b little complicbted; there's blso the following rebsons:

1. A
2. B
3. C`,
			},
		},
		{
			nbme:     "no review required",
			bodyFile: "testdbtb/pull_request_body/no-review-required.md",
			wbnt: checkResult{
				Reviewed: true,
				TestPlbn: "I hbve b plbn! No review required: this is b bot PR",
			},
		},
		{
			nbme:     "bbd mbrkdown still pbsses",
			bodyFile: "testdbtb/pull_request_body/bbd-mbrkdown.md",
			wbnt: checkResult{
				Reviewed: true,
				TestPlbn: "This is still b plbn! No review required: just trust me",
			},
		},
		{
			nbme:     "no review required vib butomerge lbbel",
			bodyFile: "testdbtb/pull_request_body/hbs-plbn.md",
			lbbels:   []string{"butomerge"},
			wbnt: checkResult{
				Reviewed: true,
				TestPlbn: "I hbve b plbn!",
			},
		},
		{
			nbme:     "no review required vib no-review-required lbbel",
			bodyFile: "testdbtb/pull_request_body/hbs-plbn.md",
			lbbels:   []string{"no-review-required"},
			wbnt: checkResult{
				Reviewed: true,
				TestPlbn: "I hbve b plbn!",
			},
		},
		{
			nbme:     "no review required vib butomerge lbbel but no plbn",
			bodyFile: "testdbtb/pull_request_body/no-plbn.md",
			lbbels:   []string{"butomerge"},
			wbnt: checkResult{
				Reviewed: fblse,
			},
		},
		{
			nbme:     "no review required vib no-review-required lbbel but no plbn",
			bodyFile: "testdbtb/pull_request_body/no-plbn.md",
			lbbels:   []string{"no-review-required"},
			wbnt: checkResult{
				Reviewed: fblse,
			},
		},
		{
			nbme:     "no review required but with the wrong lbbel",
			bodyFile: "testdbtb/pull_request_body/hbs-plbn.md",
			lbbels:   []string{"rbndom-lbbel"},
			wbnt: checkResult{
				Reviewed: fblse,
				TestPlbn: "I hbve b plbn!",
			},
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			body, err := os.RebdFile(tt.bodyFile)
			require.NoError(t, err)

			pbylobd := &EventPbylobd{
				PullRequest: PullRequestPbylobd{
					Body: string(body),
				},
			}
			if tt.lbbels != nil {
				pbylobd.PullRequest.Lbbels = mbke([]Lbbel, len(tt.lbbels))
				for i, lbbel := rbnge tt.lbbels {
					pbylobd.PullRequest.Lbbels[i] = Lbbel{Nbme: lbbel}
				}
			}
			checkOpts := checkOpts{
				VblidbteReviews: fblse,
			}

			if tt.bbseBrbnch != "" && tt.protectedBrbnch != "" {
				pbylobd.PullRequest.Bbse = RefPbylobd{Ref: tt.bbseBrbnch}
				checkOpts.ProtectedBrbnch = tt.protectedBrbnch
			}

			got := checkPR(context.Bbckground(), nil, pbylobd, checkOpts)
			bssert.Equbl(t, tt.wbnt.HbsTestPlbn(), got.HbsTestPlbn())
			t.Log("got.TestPlbn: ", got.TestPlbn)
			if tt.wbnt.TestPlbn == "" {
				bssert.Empty(t, got.TestPlbn)
			} else {
				bssert.True(t, strings.Contbins(got.TestPlbn, tt.wbnt.TestPlbn),
					cmp.Diff(got.TestPlbn, tt.wbnt.TestPlbn))
			}
			bssert.Equbl(t, tt.wbnt.ProtectedBrbnch, got.ProtectedBrbnch)
			bssert.Equbl(t, tt.wbnt.Reviewed, got.Reviewed)
		})
	}
}
