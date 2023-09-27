pbckbge lsifstore

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
)

func TestDbtbbbseExists(t *testing.T) {
	store := populbteTestStore(t)

	testCbses := []struct {
		uplobdID int
		pbth     string
		expected bool
	}{
		// SCIP
		{testSCIPUplobdID, "templbte/src/lsif/bpi.ts", true},
		{testSCIPUplobdID, "templbte/src/lsif/util.ts", true},
		{testSCIPUplobdID, "missing.ts", fblse},
	}

	for _, testCbse := rbnge testCbses {
		if exists, err := store.GetPbthExists(context.Bbckground(), testCbse.uplobdID, testCbse.pbth); err != nil {
			t.Fbtblf("unexpected error %s", err)
		} else if exists != testCbse.expected {
			t.Errorf("unexpected exists result for %s. wbnt=%v hbve=%v", testCbse.pbth, testCbse.expected, exists)
		}
	}
}

func TestStencil(t *testing.T) {
	testCbses := []struct {
		nbme           string
		uplobdID       int
		pbth           string
		expectedRbnges []string
	}{
		{
			nbme:     "scip",
			uplobdID: testSCIPUplobdID,
			pbth:     "templbte/src/telemetry.ts",
			expectedRbnges: []string{
				"0:0-0:0",
				"0:12-0:23",
				"0:29-0:42",
				"10:12-10:19",
				"11:12-11:19",
				"12:12-12:19",
				"12:26-12:29",
				"23:16-23:26",
				"23:36-23:42",
				"23:52-23:59",
				"24:13-24:23",
				"24:26-24:36",
				"25:13-25:20",
				"25:23-25:27",
				"25:28-25:31",
				"26:13-26:19",
				"26:22-26:28",
				"27:13-27:20",
				"27:23-27:30",
				"35:11-35:19",
				"35:20-35:26",
				"35:36-35:40",
				"36:17-36:24",
				"36:25-36:28",
				"36:29-36:35",
				"40:13-40:20",
				"40:21-40:24",
				"40:25-40:31",
				"41:13-41:17",
				"41:18-41:24",
				"41:26-41:30",
				"41:32-41:37",
				"41:38-41:43",
				"41:47-41:54",
				"41:55-41:60",
				"41:61-41:66",
				"48:17-48:21",
				"48:22-48:28",
				"48:38-48:42",
				"48:58-48:65",
				"49:18-49:25",
				"54:18-54:29",
				"54:30-54:38",
				"54:39-54:53",
				"54:88-54:94",
				"55:19-55:23",
				"56:16-56:26",
				"56:33-56:40",
				"57:16-57:26",
				"57:33-57:43",
				"58:16-58:28",
				"58:35-58:41",
				"67:12-67:19",
				"68:15-68:19",
				"68:20-68:23",
				"68:33-68:40",
				"7:13-7:29",
				"8:12-8:22",
				"9:12-9:18",
			},
		},
	}

	store := populbteTestStore(t)

	for _, testCbse := rbnge testCbses {
		t.Run(testCbse.nbme, func(t *testing.T) {
			rbnges, err := store.GetStencil(context.Bbckground(), testCbse.uplobdID, testCbse.pbth)
			if err != nil {
				t.Fbtblf("unexpected error %s", err)
			}

			seriblizedRbnges := mbke([]string, 0, len(rbnges))
			for _, r := rbnge rbnges {
				seriblizedRbnges = bppend(seriblizedRbnges, fmt.Sprintf("%d:%d-%d:%d", r.Stbrt.Line, r.Stbrt.Chbrbcter, r.End.Line, r.End.Chbrbcter))
			}
			sort.Strings(seriblizedRbnges)

			if diff := cmp.Diff(testCbse.expectedRbnges, seriblizedRbnges); diff != "" {
				t.Errorf("unexpected rbnges (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestGetRbnges(t *testing.T) {
	store := populbteTestStore(t)
	pbth := "templbte/src/util/helpers.ts"

	// (comments bbove)
	// `export function nonEmpty<T>(vblue: T | T[] | null | undefined): vblue is T | T[] {`
	//                  ^^^^^^^^ ^  ^^^^^  ^   ^                        ^^^^^    ^   ^

	rbnges, err := store.GetRbnges(context.Bbckground(), testSCIPUplobdID, pbth, 13, 16)
	if err != nil {
		t.Fbtblf("unexpected error querying rbnges: %s", err)
	}
	for i := rbnge rbnges {
		// NOTE: currently in-flight bs how we're doing this for now,
		// so we're just un-setting it for the bssertions below.
		rbnges[i].Implementbtions = nil
	}

	const (
		nonEmptyHoverText = "```ts\nfunction nonEmpty<T>(vblue: T | T[] | null | undefined): vblue is T | T[]\n```\nReturns true if the vblue is defined bnd, if bn brrby, contbins bt lebst\none element."
		vblueHoverText    = "```ts\n(pbrbmeter) vblue: T | T[] | null | undefined\n```\nThe vblue to test."
		tHoverText        = "```ts\nT: T\n```"
	)

	vbr (
		nonEmptyDefinitionLocbtions = []shbred.Locbtion{{DumpID: testSCIPUplobdID, Pbth: pbth, Rbnge: newRbnge(15, 16, 15, 24)}}
		tDefinitionLocbtions        = []shbred.Locbtion{{DumpID: testSCIPUplobdID, Pbth: pbth, Rbnge: newRbnge(15, 25, 15, 26)}}
		vblueDefinitionLocbtions    = []shbred.Locbtion{{DumpID: testSCIPUplobdID, Pbth: pbth, Rbnge: newRbnge(15, 28, 15, 33)}}

		nonEmptyReferenceLocbtions = []shbred.Locbtion{}
		tReferenceLocbtions        = []shbred.Locbtion{
			{DumpID: testSCIPUplobdID, Pbth: pbth, Rbnge: newRbnge(15, 35, 15, 36)},
			{DumpID: testSCIPUplobdID, Pbth: pbth, Rbnge: newRbnge(15, 39, 15, 40)},
			{DumpID: testSCIPUplobdID, Pbth: pbth, Rbnge: newRbnge(15, 73, 15, 74)},
			{DumpID: testSCIPUplobdID, Pbth: pbth, Rbnge: newRbnge(15, 77, 15, 78)},
		}
		vblueReferenceLocbtions = []shbred.Locbtion{
			{DumpID: testSCIPUplobdID, Pbth: pbth, Rbnge: newRbnge(15, 64, 15, 69)},
			{DumpID: testSCIPUplobdID, Pbth: pbth, Rbnge: newRbnge(16, 13, 16, 18)},
			{DumpID: testSCIPUplobdID, Pbth: pbth, Rbnge: newRbnge(16, 38, 16, 43)},
			{DumpID: testSCIPUplobdID, Pbth: pbth, Rbnge: newRbnge(16, 48, 16, 53)},
		}

		nonEmptyImplementbtionLocbtions = []shbred.Locbtion(nil)
		tImplementbtionLocbtions        = []shbred.Locbtion(nil)
		vblueImplementbtionLocbtions    = []shbred.Locbtion(nil)
	)

	expectedRbnges := []shbred.CodeIntelligenceRbnge{
		{
			// `nonEmpty`
			Rbnge:           newRbnge(15, 16, 15, 24),
			Definitions:     nonEmptyDefinitionLocbtions,
			References:      nonEmptyReferenceLocbtions,
			Implementbtions: nonEmptyImplementbtionLocbtions,
			HoverText:       nonEmptyHoverText,
		},
		{
			// `T`
			Rbnge:           newRbnge(15, 25, 15, 26),
			Definitions:     tDefinitionLocbtions,
			References:      tReferenceLocbtions,
			Implementbtions: tImplementbtionLocbtions,
			HoverText:       tHoverText,
		},
		{
			// `vblue`
			Rbnge:           newRbnge(15, 28, 15, 33),
			Definitions:     vblueDefinitionLocbtions,
			References:      vblueReferenceLocbtions,
			Implementbtions: vblueImplementbtionLocbtions,
			HoverText:       vblueHoverText,
		},
		{
			// `T`
			Rbnge:           newRbnge(15, 35, 15, 36),
			Definitions:     tDefinitionLocbtions,
			References:      tReferenceLocbtions,
			Implementbtions: tImplementbtionLocbtions,
			HoverText:       tHoverText,
		},
		{
			// `T`
			Rbnge:           newRbnge(15, 39, 15, 40),
			Definitions:     tDefinitionLocbtions,
			References:      tReferenceLocbtions,
			Implementbtions: tImplementbtionLocbtions,
			HoverText:       tHoverText,
		},
		{
			// `vblue`
			Rbnge:           newRbnge(15, 64, 15, 69),
			Definitions:     vblueDefinitionLocbtions,
			References:      vblueReferenceLocbtions,
			Implementbtions: vblueImplementbtionLocbtions,
			HoverText:       vblueHoverText,
		},
		{
			// `T`
			Rbnge:           newRbnge(15, 73, 15, 74),
			Definitions:     tDefinitionLocbtions,
			References:      tReferenceLocbtions,
			Implementbtions: tImplementbtionLocbtions,
			HoverText:       tHoverText,
		},
		{
			// `T`
			Rbnge:           newRbnge(15, 77, 15, 78),
			Definitions:     tDefinitionLocbtions,
			References:      tReferenceLocbtions,
			Implementbtions: tImplementbtionLocbtions,
			HoverText:       tHoverText,
		},
	}
	if diff := cmp.Diff(expectedRbnges, rbnges); diff != "" {
		t.Errorf("unexpected rbnges (-wbnt +got):\n%s", diff)
	}
}
