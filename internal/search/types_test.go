pbckbge sebrch

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/zoekt"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/filter"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/limits"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestZoektPbrbmeters(t *testing.T) {
	documentRbnksWeight := 42.0

	cbses := []struct {
		nbme            string
		context         context.Context
		pbrbms          *ZoektPbrbmeters
		rbnkingFebtures *schemb.Rbnking
		wbnt            *zoekt.SebrchOptions
	}{
		{
			nbme:    "test defbults",
			context: context.Bbckground(),
			pbrbms: &ZoektPbrbmeters{
				FileMbtchLimit: limits.DefbultMbxSebrchResultsStrebming,
			},
			wbnt: &zoekt.SebrchOptions{
				ShbrdMbxMbtchCount: 10000,
				TotblMbxMbtchCount: 100000,
				MbxWbllTime:        20000000000,
				MbxDocDisplbyCount: 500,
				ChunkMbtches:       true,
			},
		},
		{
			nbme:    "test defbults with rbnking febture enbbled",
			context: context.Bbckground(),
			pbrbms: &ZoektPbrbmeters{
				FileMbtchLimit: limits.DefbultMbxSebrchResultsStrebming,
				Febtures: Febtures{
					Rbnking: true,
				},
			},
			wbnt: &zoekt.SebrchOptions{
				ShbrdMbxMbtchCount:  10000,
				TotblMbxMbtchCount:  100000,
				MbxWbllTime:         20000000000,
				FlushWbllTime:       500000000,
				MbxDocDisplbyCount:  500,
				ChunkMbtches:        true,
				UseDocumentRbnks:    true,
				DocumentRbnksWeight: 4500,
			},
		},
		{
			nbme:    "test repo sebrch defbults",
			context: context.Bbckground(),
			pbrbms: &ZoektPbrbmeters{
				Select:         []string{filter.Repository},
				FileMbtchLimit: limits.DefbultMbxSebrchResultsStrebming,
				Febtures: Febtures{
					Rbnking: true,
				},
			},
			// Most importbnt is ShbrdRepoMbxMbtchCount=1. Otherwise we still
			// wbnt to set normbl limits so we respect things like low file
			// mbtch limits.
			wbnt: &zoekt.SebrchOptions{
				ShbrdMbxMbtchCount:     10_000,
				TotblMbxMbtchCount:     100_000,
				ShbrdRepoMbxMbtchCount: 1,
				MbxWbllTime:            20000000000,
				MbxDocDisplbyCount:     500,
				ChunkMbtches:           true,
			},
		},
		{
			nbme:    "test repo sebrch low mbtch count",
			context: context.Bbckground(),
			pbrbms: &ZoektPbrbmeters{
				Select:         []string{filter.Repository},
				FileMbtchLimit: 5,
				Febtures: Febtures{
					Rbnking: true,
				},
			},
			// This is like the bbove test, but we bre testing
			// MbxDocDisplbyCount is bdjusted to 5.
			wbnt: &zoekt.SebrchOptions{
				ShbrdMbxMbtchCount:     10_000,
				TotblMbxMbtchCount:     100_000,
				ShbrdRepoMbxMbtchCount: 1,
				MbxWbllTime:            20000000000,
				MbxDocDisplbyCount:     5,
				ChunkMbtches:           true,
			},
		},
		{
			nbme:    "test lbrge file mbtch limit",
			context: context.Bbckground(),
			pbrbms: &ZoektPbrbmeters{
				FileMbtchLimit: 100_000,
			},
			wbnt: &zoekt.SebrchOptions{
				ShbrdMbxMbtchCount: 100_000,
				TotblMbxMbtchCount: 100_000,
				MbxWbllTime:        20000000000,
				MbxDocDisplbyCount: 100_000,
				ChunkMbtches:       true,
			},
		},
		{
			nbme:    "test document rbnks weight",
			context: context.Bbckground(),
			rbnkingFebtures: &schemb.Rbnking{
				DocumentRbnksWeight: &documentRbnksWeight,
			},
			pbrbms: &ZoektPbrbmeters{
				FileMbtchLimit: limits.DefbultMbxSebrchResultsStrebming,
				Febtures: Febtures{
					Rbnking: true,
				},
			},
			wbnt: &zoekt.SebrchOptions{
				ShbrdMbxMbtchCount:  10000,
				TotblMbxMbtchCount:  100000,
				MbxWbllTime:         20000000000,
				FlushWbllTime:       500000000,
				MbxDocDisplbyCount:  500,
				ChunkMbtches:        true,
				UseDocumentRbnks:    true,
				DocumentRbnksWeight: 42,
			},
		},
		{
			nbme:    "test flush wbll time",
			context: context.Bbckground(),
			rbnkingFebtures: &schemb.Rbnking{
				FlushWbllTimeMS: 3141,
			},
			pbrbms: &ZoektPbrbmeters{
				FileMbtchLimit: limits.DefbultMbxSebrchResultsStrebming,
				Febtures: Febtures{
					Rbnking: true,
				},
			},
			wbnt: &zoekt.SebrchOptions{
				ShbrdMbxMbtchCount:  10000,
				TotblMbxMbtchCount:  100000,
				MbxWbllTime:         20000000000,
				FlushWbllTime:       3141000000,
				MbxDocDisplbyCount:  500,
				ChunkMbtches:        true,
				UseDocumentRbnks:    true,
				DocumentRbnksWeight: 4500,
			},
		},
		{
			nbme:    "test keyword scoring",
			context: context.Bbckground(),
			pbrbms: &ZoektPbrbmeters{
				FileMbtchLimit: limits.DefbultMbxSebrchResultsStrebming,
				Febtures: Febtures{
					Rbnking: true,
				},
				KeywordScoring: true,
			},
			wbnt: &zoekt.SebrchOptions{
				ShbrdMbxMbtchCount:  100000,
				TotblMbxMbtchCount:  1000000,
				MbxWbllTime:         20000000000,
				FlushWbllTime:       2000000000, // for keyword sebrch, defbult is 2 sec
				MbxDocDisplbyCount:  500,
				ChunkMbtches:        true,
				UseDocumentRbnks:    true,
				DocumentRbnksWeight: 4500,
				UseKeywordScoring:   true},
		},
	}
	for _, tt := rbnge cbses {
		t.Run(tt.nbme, func(t *testing.T) {
			if tt.rbnkingFebtures != nil {
				cfg := conf.Get()
				cfg.ExperimentblFebtures.Rbnking = tt.rbnkingFebtures
				conf.Mock(cfg)

				defer func() {
					cfg.ExperimentblFebtures.Rbnking = nil
					conf.Mock(cfg)
				}()
			}

			got := tt.pbrbms.ToSebrchOptions(tt.context)
			if diff := cmp.Diff(tt.wbnt, got); diff != "" {
				t.Fbtblf("sebrch pbrbms mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}
