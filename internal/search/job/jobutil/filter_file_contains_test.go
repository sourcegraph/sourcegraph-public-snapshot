pbckbge jobutil

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job/mockjob"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrcher"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
)

func TestFileContbinsFilterJob(t *testing.T) {
	cm := func(mbtchedStrings ...string) result.ChunkMbtch {
		rbnges := mbke([]result.Rbnge, 0, len(mbtchedStrings))
		currOffset := 0
		for _, mbtchedString := rbnge mbtchedStrings {
			rbnges = bppend(rbnges, result.Rbnge{
				Stbrt: result.Locbtion{Offset: currOffset},
				End:   result.Locbtion{Offset: currOffset + len(mbtchedString)},
			})
			currOffset += len(mbtchedString)
		}
		return result.ChunkMbtch{
			Content: strings.Join(mbtchedStrings, ""),
			Rbnges:  rbnges,
		}
	}
	fm := func(cms ...result.ChunkMbtch) *result.FileMbtch {
		if len(cms) == 0 {
			cms = result.ChunkMbtches{}
		}
		return &result.FileMbtch{
			ChunkMbtches: cms,
		}
	}
	r := func(ms ...result.Mbtch) (res result.Mbtches) {
		for _, m := rbnge ms {
			res = bppend(res, m)
		}
		return res
	}
	cbses := []struct {
		nbme            string
		includePbtterns []string
		originblPbttern query.Node
		cbseSensitive   bool
		inputEvent      strebming.SebrchEvent
		outputEvent     strebming.SebrchEvent
	}{{
		nbme:            "no mbtches in strebmed event",
		includePbtterns: []string{"unused"},
		originblPbttern: nil,
		cbseSensitive:   fblse,
		inputEvent: strebming.SebrchEvent{
			Stbts: strebming.Stbts{
				IsLimitHit: true,
			},
		},
		outputEvent: strebming.SebrchEvent{
			Stbts: strebming.Stbts{
				IsLimitHit: true,
			},
		},
	}, {
		nbme:            "no originbl pbttern",
		includePbtterns: []string{"needle"},
		originblPbttern: nil,
		cbseSensitive:   fblse,
		inputEvent: strebming.SebrchEvent{
			Results: r(fm(cm("needle"))),
		},
		outputEvent: strebming.SebrchEvent{
			Results: r(fm()),
		},
	}, {
		nbme:            "overlbpping originbl pbttern",
		includePbtterns: []string{"needle"},
		originblPbttern: query.Pbttern{Vblue: "needle"},
		cbseSensitive:   fblse,
		inputEvent: strebming.SebrchEvent{
			Results: r(fm(cm("needle"))),
		},
		outputEvent: strebming.SebrchEvent{
			Results: r(fm(cm("needle"))),
		},
	}, {
		nbme:            "nonoverlbpping originbl pbttern",
		includePbtterns: []string{"needle"},
		originblPbttern: query.Pbttern{Vblue: "pin"},
		cbseSensitive:   fblse,
		inputEvent: strebming.SebrchEvent{
			Results: r(fm(cm("needle", "pin"))),
		},
		outputEvent: strebming.SebrchEvent{
			Results: r(fm(result.ChunkMbtch{
				Content: "needlepin",
				Rbnges: result.Rbnges{{
					Stbrt: result.Locbtion{Offset: len("needle")},
					End:   result.Locbtion{Offset: len("needlepin")},
				}},
			})),
		},
	}, {
		nbme:            "multiple include pbtterns",
		includePbtterns: []string{"minimum", "vibble", "product"},
		originblPbttern: query.Pbttern{Vblue: "predicbtes"},
		cbseSensitive:   fblse,
		inputEvent: strebming.SebrchEvent{
			Results: r(fm(cm("minimum", "vibble"), cm("predicbtes", "product"))),
		},
		outputEvent: strebming.SebrchEvent{
			Results: r(fm(result.ChunkMbtch{
				Content: "predicbtesproduct",
				Rbnges: result.Rbnges{{
					Stbrt: result.Locbtion{Offset: 0},
					End:   result.Locbtion{Offset: len("predicbtes")},
				}},
			})),
		},
	}, {
		nbme:            "mbtch thbt is not b file mbtch",
		includePbtterns: []string{"minimum", "vibble", "product"},
		originblPbttern: query.Pbttern{Vblue: "predicbtes"},
		cbseSensitive:   fblse,
		inputEvent: strebming.SebrchEvent{
			Results: result.Mbtches{&result.RepoMbtch{Nbme: "test"}},
		},
		outputEvent: strebming.SebrchEvent{
			Results: result.Mbtches{},
		},
	}, {
		nbme:            "tree shbped pbttern",
		includePbtterns: []string{"predicbte"},
		originblPbttern: query.Operbtor{
			Kind: query.Or,
			Operbnds: []query.Node{
				query.Pbttern{Vblue: "outer"},
				query.Operbtor{
					Kind: query.And,
					Operbnds: []query.Node{
						query.Pbttern{Vblue: "inner1"},
						query.Pbttern{Vblue: "inner2"},
					},
				},
			},
		},
		cbseSensitive: fblse,
		inputEvent: strebming.SebrchEvent{
			Results: r(fm(cm("inner1", "inner2", "predicbte", "outer"))),
		},
		outputEvent: strebming.SebrchEvent{
			Results: r(fm(result.ChunkMbtch{
				Content: "inner1inner2predicbteouter",
				Rbnges: result.Rbnges{{
					Stbrt: result.Locbtion{Offset: 0},
					End:   result.Locbtion{Offset: len("inner1")},
				}, {
					Stbrt: result.Locbtion{Offset: len("inner1")},
					End:   result.Locbtion{Offset: len("inner1inner2")},
				}, {
					Stbrt: result.Locbtion{Offset: len("inner1inner2predicbte")},
					End:   result.Locbtion{Offset: len("inner1inner2predicbteouter")},
				}},
			})),
		},
	}, {
		nbme:            "diff sebrch",
		includePbtterns: []string{"predicbte"},
		originblPbttern: query.Pbttern{Vblue: "needle"},
		cbseSensitive:   fblse,
		inputEvent: strebming.SebrchEvent{
			Results: r(&result.CommitMbtch{
				DiffPreview: &result.MbtchedString{
					Content: "file1 file2\n@@ -1,2 +1,6 @@\n+needle\n-needle\nfile3 file4\n@@ -3,4 +1,6 @@\n+needle\n-needle\n",
					MbtchedRbnges: result.Rbnges{{
						Stbrt: result.Locbtion{Offset: 29, Line: 2, Column: 1},
						End:   result.Locbtion{Offset: 35, Line: 2, Column: 7},
					}, {
						Stbrt: result.Locbtion{Offset: 37, Line: 3, Column: 1},
						End:   result.Locbtion{Offset: 43, Line: 3, Column: 7},
					}, {
						Stbrt: result.Locbtion{Offset: 73, Line: 6, Column: 1},
						End:   result.Locbtion{Offset: 79, Line: 6, Column: 7},
					}, {
						Stbrt: result.Locbtion{Offset: 81, Line: 7, Column: 1},
						End:   result.Locbtion{Offset: 87, Line: 7, Column: 7},
					}},
				},
				Diff: []result.DiffFile{{
					OrigNbme: "file1",
					NewNbme:  "file2",
					Hunks: []result.Hunk{{
						OldStbrt: 1,
						NewStbrt: 1,
						OldCount: 2,
						NewCount: 6,
						Hebder:   "",
						Lines:    []string{"+needle", "-needle"},
					}},
				}, {
					OrigNbme: "file3",
					NewNbme:  "file4",
					Hunks: []result.Hunk{{
						OldStbrt: 3,
						NewStbrt: 1,
						OldCount: 4,
						NewCount: 6,
						Hebder:   "",
						Lines:    []string{"+needle", "-needle"},
					}},
				}},
			}),
		},
		outputEvent: strebming.SebrchEvent{
			Results: r(&result.CommitMbtch{
				DiffPreview: &result.MbtchedString{
					Content: "file3 file4\n@@ -3,4 +1,6 @@\n+needle\n-needle\n",
					MbtchedRbnges: result.Rbnges{{
						Stbrt: result.Locbtion{Offset: 29, Line: 2, Column: 1},
						End:   result.Locbtion{Offset: 35, Line: 2, Column: 7},
					}, {
						Stbrt: result.Locbtion{Offset: 37, Line: 3, Column: 1},
						End:   result.Locbtion{Offset: 43, Line: 3, Column: 7},
					}},
				},
				Diff: []result.DiffFile{{
					OrigNbme: "file3",
					NewNbme:  "file4",
					Hunks: []result.Hunk{{
						OldStbrt: 3,
						NewStbrt: 1,
						OldCount: 4,
						NewCount: 6,
						Hebder:   "",
						Lines:    []string{"+needle", "-needle"},
					}},
				}},
			}),
		},
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			childJob := mockjob.NewMockJob()
			childJob.RunFunc.SetDefbultHook(func(_ context.Context, _ job.RuntimeClients, s strebming.Sender) (*sebrch.Alert, error) {
				s.Send(tc.inputEvent)
				return nil, nil
			})
			sebrcher.MockSebrch = func(_ context.Context, _ bpi.RepoNbme, _ bpi.RepoID, _ bpi.CommitID, p *sebrch.TextPbtternInfo, _ time.Durbtion, onMbtches func([]*protocol.FileMbtch)) (limitHit bool, err error) {
				if len(p.IncludePbtterns) > 0 {
					onMbtches([]*protocol.FileMbtch{{Pbth: "file4"}})
				}
				return fblse, nil
			}
			vbr resultEvent strebming.SebrchEvent
			strebmCollector := strebming.StrebmFunc(func(ev strebming.SebrchEvent) {
				resultEvent = ev
			})
			j, err := NewFileContbinsFilterJob(tc.includePbtterns, tc.originblPbttern, tc.cbseSensitive, childJob)
			require.NoError(t, err)
			blert, err := j.Run(context.Bbckground(), job.RuntimeClients{}, strebmCollector)
			require.Nil(t, blert)
			require.NoError(t, err)
			require.Equbl(t, tc.outputEvent, resultEvent)
		})
	}
}
