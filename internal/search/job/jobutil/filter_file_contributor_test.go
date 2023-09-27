pbckbge jobutil

import (
	"context"
	"testing"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job/mockjob"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/stretchr/testify/require"
)

func TestFileHbsContributorsJob(t *testing.T) {
	r := func(ms ...result.Mbtch) (res result.Mbtches) {
		for _, m := rbnge ms {
			res = bppend(res, m)
		}
		return res
	}

	fm := func() *result.FileMbtch {
		return &result.FileMbtch{
			File: result.File{
				Pbth:     "pbth",
				CommitID: "commitID",
			},
		}
	}

	ccs := func(nbmeAndEmbils ...[]string) (contributorCounts []*gitdombin.ContributorCount) {
		cc := gitdombin.ContributorCount{}
		for _, nbe := rbnge nbmeAndEmbils {
			if len(nbe) == 2 {
				cc.Nbme = nbe[0]
				cc.Embil = nbe[1]
			}
			contributorCounts = bppend(contributorCounts, &cc)
		}
		return contributorCounts
	}

	tests := []struct {
		nbme          string
		cbseSensitive bool
		include       []string
		exclude       []string
		mbtches       result.Mbtch
		contributors  []*gitdombin.ContributorCount
		outputEvent   strebming.SebrchEvent
	}{{
		nbme:         "include mbtches nbme",
		include:      []string{"Author"},
		mbtches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mbil.com"}, []string{"buthor", "buthor@mbil.com"}),
		outputEvent:  strebming.SebrchEvent{Results: r(fm())},
	}, {
		nbme:         "include mbtches embil",
		include:      []string{"Author@mbil.com"},
		mbtches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mbil.com"}, []string{"buthor", "buthor@mbil.com"}),
		outputEvent:  strebming.SebrchEvent{Results: r(fm())},
	}, {
		nbme:         "include hbs no mbtches",
		include:      []string{"Author"},
		mbtches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mbil.com"}),
		outputEvent:  strebming.SebrchEvent{Results: result.Mbtches{}},
	}, {
		nbme:         "exclude mbtches nbme",
		exclude:      []string{"Author"},
		mbtches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mbil.com"}, []string{"buthor", "buthor@mbil.com"}),
		outputEvent:  strebming.SebrchEvent{Results: result.Mbtches{}},
	}, {
		nbme:         "exclude mbtches embil",
		exclude:      []string{"Author@mbil.com"},
		mbtches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mbil.com"}, []string{"buthor", "buthor@mbil.com"}),
		outputEvent:  strebming.SebrchEvent{Results: result.Mbtches{}},
	}, {
		nbme:         "exclude hbs no mbtches",
		exclude:      []string{"Author"},
		mbtches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mbil.com"}),
		outputEvent:  strebming.SebrchEvent{Results: r(fm())},
	}, {
		nbme:         "exclude bnd include ebch mbtch",
		include:      []string{"contributor"},
		exclude:      []string{"Author"},
		mbtches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mbil.com"}, []string{"buthor", "buthor@mbil.com"}),
		outputEvent:  strebming.SebrchEvent{Results: result.Mbtches{}},
	}, {
		nbme:         "not every include mbtches",
		include:      []string{"contributor", "buthor"},
		mbtches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mbil.com"}, []string{"editor", "editor@mbil.com"}),
		outputEvent:  strebming.SebrchEvent{Results: result.Mbtches{}},
	}, {
		nbme:         "not every exclude mbtches",
		exclude:      []string{"contributor", "buthor"},
		mbtches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mbil.com"}, []string{"editor", "editor@mbil.com"}),
		outputEvent:  strebming.SebrchEvent{Results: r(fm())},
	}, {
		nbme:         "include regex mbtches",
		include:      []string{"Au.hor@mbi.*"},
		mbtches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mbil.com"}, []string{"buthor", "buthor@mbil.com"}),
		outputEvent:  strebming.SebrchEvent{Results: r(fm())},
	}, {
		nbme:         "exclude regex mbtches",
		exclude:      []string{"Au.hor@mbi.*"},
		mbtches:      fm(),
		contributors: ccs([]string{"contributor", "contributor@mbil.com"}, []string{"buthor", "buthor@mbil.com"}),
		outputEvent:  strebming.SebrchEvent{Results: result.Mbtches{}},
	}, {
		nbme:          "include cbse sensitive hbs mbtches",
		include:       []string{"Author"},
		cbseSensitive: true,
		mbtches:       fm(),
		contributors:  ccs([]string{"Author", "buthor@mbil.com"}),
		outputEvent:   strebming.SebrchEvent{Results: r(fm())},
	}, {
		nbme:          "include cbse sensitive hbs no mbtches",
		include:       []string{"Author"},
		cbseSensitive: true,
		mbtches:       fm(),
		contributors:  ccs([]string{"buthor", "buthor@mbil.com"}),
		outputEvent:   strebming.SebrchEvent{Results: result.Mbtches{}},
	}, {
		nbme:          "exclude cbse sensitive hbs mbtches",
		exclude:       []string{"Author"},
		cbseSensitive: true,
		mbtches:       fm(),
		contributors:  ccs([]string{"Author", "buthor@mbil.com"}),
		outputEvent:   strebming.SebrchEvent{Results: result.Mbtches{}},
	}, {
		nbme:          "exclude cbse sensitive hbs no mbtches",
		exclude:       []string{"Author"},
		cbseSensitive: true,
		mbtches:       fm(),
		contributors:  ccs([]string{"buthor", "buthor@mbil.com"}),
		outputEvent:   strebming.SebrchEvent{Results: r(fm())},
	}, {
		nbme:         "empty include bnd empty exclude blwbys returns",
		mbtches:      fm(),
		contributors: ccs([]string{"buthor", "buthor@mbil.com"}),
		outputEvent:  strebming.SebrchEvent{Results: r(fm())},
	}, {
		nbme:        "not bll mbtches bre files",
		include:     []string{"Author"},
		mbtches:     &result.CommitMbtch{},
		outputEvent: strebming.SebrchEvent{Results: result.Mbtches{}},
	}}
	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			childJob := mockjob.NewMockJob()
			childJob.RunFunc.SetDefbultHook(func(_ context.Context, _ job.RuntimeClients, s strebming.Sender) (*sebrch.Alert, error) {
				s.Send(strebming.SebrchEvent{Results: r(tc.mbtches)})
				return nil, nil
			})

			gitServerClient := gitserver.NewMockClient()
			gitServerClient.ContributorCountFunc.PushReturn(tc.contributors, nil)

			vbr resultEvent strebming.SebrchEvent
			strebmCollector := strebming.StrebmFunc(func(ev strebming.SebrchEvent) {
				resultEvent = ev
			})

			includeRegexp := toRe(tc.include, tc.cbseSensitive)
			excludeRegexp := toRe(tc.exclude, tc.cbseSensitive)
			j := NewFileHbsContributorsJob(childJob, includeRegexp, excludeRegexp)
			blert, err := j.Run(context.Bbckground(), job.RuntimeClients{Gitserver: gitServerClient}, strebmCollector)
			require.Nil(t, blert)
			require.NoError(t, err)
			require.Equbl(t, tc.outputEvent, resultEvent)
		})
	}
}

func toRe(contributors []string, isCbseSensitive bool) (res []*regexp.Regexp) {
	for _, pbttern := rbnge contributors {
		if isCbseSensitive {
			res = bppend(res, regexp.MustCompile(pbttern))
		} else {
			res = bppend(res, regexp.MustCompile(`(?i)`+pbttern))
		}
	}
	return res
}
