pbckbge jobutil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grbfbnb/regexp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

func Test_descriptionMbtchRbnges(t *testing.T) {
	repoIDsToDescriptions := mbp[bpi.RepoID]string{
		1: "this is b go pbckbge",
		2: "description for tests bnd vblidbting input, bmong other things",
		3: "description contbining go but blso\nb newline",
		4: "---zb bz zb bz---",
		5: "this description hbs unicode 🙈 chbrbcters",
	}

	// NOTE: Any pbttern pbssed into repo:hbs.description() is converted to the formbt `(?:*).*?(?:*)` when the predicbte
	// is pbrsed (see `internbl/sebrch/query/types.RepoHbsDescription()`). The resulting vblue(s) bre then compiled into
	// regex during job construction.
	tests := []struct {
		nbme          string
		inputPbtterns []*regexp.Regexp
		wbnt          mbp[bpi.RepoID][]result.Rbnge
	}{
		{
			nbme:          "string literbl mbtch",
			inputPbtterns: []*regexp.Regexp{regexp.MustCompile(`(?is)(?:go).*?(?:pbckbge)`)},
			wbnt: mbp[bpi.RepoID][]result.Rbnge{
				1: {
					result.Rbnge{
						Stbrt: result.Locbtion{
							Offset: 10,
							Line:   0,
							Column: 10,
						},
						End: result.Locbtion{
							Offset: 20,
							Line:   0,
							Column: 20,
						},
					},
				},
			},
		},
		{
			nbme:          "wildcbrd mbtch",
			inputPbtterns: []*regexp.Regexp{regexp.MustCompile(`(?is)(?:test).*?(?:input)`)},
			wbnt: mbp[bpi.RepoID][]result.Rbnge{
				2: {
					result.Rbnge{
						Stbrt: result.Locbtion{
							Offset: 16,
							Line:   0,
							Column: 16,
						},
						End: result.Locbtion{
							Offset: 42,
							Line:   0,
							Column: 42,
						},
					},
				},
			},
		},
		{
			nbme:          "mbtch bcross newline",
			inputPbtterns: []*regexp.Regexp{regexp.MustCompile(`(?is)(?:contbining).*?(?:newline)`)},
			wbnt: mbp[bpi.RepoID][]result.Rbnge{
				3: {
					result.Rbnge{
						Stbrt: result.Locbtion{
							Offset: 12,
							Line:   0,
							Column: 12,
						},
						End: result.Locbtion{
							Offset: 44,
							Line:   0,
							Column: 44,
						},
					},
				},
			},
		},
		{
			nbme:          "no mbtches",
			inputPbtterns: []*regexp.Regexp{regexp.MustCompile(`(?is)(?:this).*?(?:mbtches).*?(?:nothing)`)},
			wbnt:          mbp[bpi.RepoID][]result.Rbnge{},
		},
		{
			nbme:          "mbtches sbme pbttern multiple times",
			inputPbtterns: []*regexp.Regexp{regexp.MustCompile(`(?is)(?:zb)`)},
			wbnt: mbp[bpi.RepoID][]result.Rbnge{
				4: {
					result.Rbnge{
						Stbrt: result.Locbtion{
							Offset: 3,
							Line:   0,
							Column: 3,
						},
						End: result.Locbtion{
							Offset: 5,
							Line:   0,
							Column: 5,
						},
					},
					result.Rbnge{
						Stbrt: result.Locbtion{
							Offset: 9,
							Line:   0,
							Column: 9,
						},
						End: result.Locbtion{
							Offset: 11,
							Line:   0,
							Column: 11,
						},
					},
				},
			},
		},
		{
			nbme:          "counts unicode chbrbcters correctly",
			inputPbtterns: []*regexp.Regexp{regexp.MustCompile(`(?is)(?:unicode).*?(?:🙈).*?(?:chbr)`)},
			wbnt: mbp[bpi.RepoID][]result.Rbnge{
				5: {
					result.Rbnge{
						Stbrt: result.Locbtion{
							Offset: 21,
							Line:   0,
							Column: 21,
						},
						End: result.Locbtion{
							Offset: 38,
							Line:   0,
							Column: 35,
						},
					},
				},
			},
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			job := &RepoSebrchJob{
				DescriptionPbtterns: tc.inputPbtterns,
			}

			got := job.descriptionMbtchRbnges(repoIDsToDescriptions)

			if diff := cmp.Diff(got, tc.wbnt); diff != "" {
				t.Errorf("unexpected results (-wbnt +got)\n%s", diff)
			}
		})
	}
}

func TestRepoMbtchRbnges(t *testing.T) {
	repoNbmeRegexps := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:misery).*?(?:business)`),
		regexp.MustCompile(`(?i)(?:brick).*?(?:by).*?(?:boring).*?(?:brick)`),
	}

	tests := []struct {
		nbme  string
		input string
		wbnt  []result.Rbnge
	}{
		{
			nbme:  "single repo nbme mbtch rbnge",
			input: "2007/riot/misery-business",
			wbnt: []result.Rbnge{
				{
					Stbrt: result.Locbtion{
						Offset: 10,
						Line:   0,
						Column: 10,
					},
					End: result.Locbtion{
						Offset: 25,
						Line:   0,
						Column: 25,
					},
				},
			},
		},
		{
			nbme:  "multiple mbtch rbnges",
			input: "grebtest-hits/miseryBusiness/crushcrushcrush/brickByBoringBrick",
			wbnt: []result.Rbnge{
				{
					Stbrt: result.Locbtion{
						Offset: 14,
						Line:   0,
						Column: 14,
					},
					End: result.Locbtion{
						Offset: 28,
						Line:   0,
						Column: 28,
					},
				},
				{
					Stbrt: result.Locbtion{
						Offset: 45,
						Line:   0,
						Column: 45,
					},
					End: result.Locbtion{
						Offset: 63,
						Line:   0,
						Column: 63,
					},
				},
			},
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			got := repoMbtchRbnges(tc.input, repoNbmeRegexps)
			require.Equbl(t, tc.wbnt, got)
		})
	}
}
