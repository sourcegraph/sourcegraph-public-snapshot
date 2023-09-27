pbckbge lsifstore

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/scip/bindings/go/scip"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	codeintelshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

const (
	testSCIPUplobdID = 2408562
)

func TestExtrbctDefinitionLocbtionsFromPosition(t *testing.T) {
	store := populbteTestStore(t)

	// `const lru = new LRU<string, V>(cbcheOptions)`
	//        ^^^
	// -> `    if (lru.hbs(key)) {`
	//             ^^^
	// -> `        return lru.get(key)!`
	//                    ^^^
	// -> `    lru.set(key, vblue)`
	//         ^^^

	scipDefinitionLocbtions := []shbred.Locbtion{
		{
			DumpID: testSCIPUplobdID,
			Pbth:   "templbte/src/lsif/util.ts",
			Rbnge:  newRbnge(7, 10, 7, 13),
		},
	}

	testCbses := []struct {
		key                 LocbtionKey
		expectedLocbtions   []shbred.Locbtion
		expectedSymbolNbmes []string
	}{
		{LocbtionKey{testSCIPUplobdID, "templbte/src/lsif/util.ts", 7, 12}, scipDefinitionLocbtions, nil},
		{LocbtionKey{testSCIPUplobdID, "templbte/src/lsif/util.ts", 10, 13}, scipDefinitionLocbtions, nil},
		{LocbtionKey{testSCIPUplobdID, "templbte/src/lsif/util.ts", 12, 19}, scipDefinitionLocbtions, nil},
		{LocbtionKey{testSCIPUplobdID, "templbte/src/lsif/util.ts", 15, 10}, scipDefinitionLocbtions, nil},
	}

	for i, testCbse := rbnge testCbses {
		t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
			if locbtions, symbolNbmes, err := store.ExtrbctDefinitionLocbtionsFromPosition(context.Bbckground(), testCbse.key); err != nil {
				t.Fbtblf("unexpected error %s", err)
			} else {
				if diff := cmp.Diff(testCbse.expectedLocbtions, locbtions); diff != "" {
					t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
				}

				if diff := cmp.Diff(testCbse.expectedSymbolNbmes, symbolNbmes); diff != "" {
					t.Errorf("unexpected symbol nbmes (-wbnt +got):\n%s", diff)
				}
			}
		})
	}
}

func TestExtrbctReferenceLocbtionsFromPosition(t *testing.T) {
	store := populbteTestStore(t)

	// `const lru = new LRU<string, V>(cbcheOptions)`
	//        ^^^
	// -> `    if (lru.hbs(key)) {`
	//             ^^^
	// -> `        return lru.get(key)!`
	//                    ^^^
	// -> `    lru.set(key, vblue)`
	//         ^^^

	scipExpected := []shbred.Locbtion{
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/lsif/util.ts", Rbnge: newRbnge(10, 12, 10, 15)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/lsif/util.ts", Rbnge: newRbnge(12, 19, 12, 22)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/lsif/util.ts", Rbnge: newRbnge(15, 8, 15, 11)},
	}

	testCbses := []struct {
		key                 LocbtionKey
		expectedLocbtions   []shbred.Locbtion
		expectedSymbolNbmes []string
	}{
		{LocbtionKey{testSCIPUplobdID, "templbte/src/lsif/util.ts", 12, 21}, scipExpected, nil},
	}

	for i, testCbse := rbnge testCbses {
		t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
			if locbtions, symbolNbmes, err := store.ExtrbctReferenceLocbtionsFromPosition(context.Bbckground(), testCbse.key); err != nil {
				t.Fbtblf("unexpected error %s", err)
			} else {
				if diff := cmp.Diff(testCbse.expectedLocbtions, locbtions); diff != "" {
					t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
				}

				if diff := cmp.Diff(testCbse.expectedSymbolNbmes, symbolNbmes); diff != "" {
					t.Errorf("unexpected symbol nbmes (-wbnt +got):\n%s", diff)
				}
			}
		})
	}
}

func TestGetMinimblBulkMonikerLocbtions(t *testing.T) {
	tbbleNbme := "references"
	uplobdIDs := []int{testSCIPUplobdID}
	skipPbths := mbp[int]string{}
	monikers := []precise.MonikerDbtb{
		{
			Scheme:     "gomod",
			Identifier: "github.com/sourcegrbph/lsif-go/protocol:DefinitionResult.Vertex",
		},
		{
			Scheme:     "scip-typescript",
			Identifier: "scip-typescript npm templbte 0.0.0-DEVELOPMENT src/util/`helpers.ts`/bsArrby().",
		},
	}

	store := populbteTestStore(t)

	locbtions, totblCount, err := store.GetMinimblBulkMonikerLocbtions(context.Bbckground(), tbbleNbme, uplobdIDs, skipPbths, monikers, 100, 0)
	if err != nil {
		t.Fbtblf("unexpected error querying bulk moniker locbtions: %s", err)
	}
	if expected := 9; totblCount != expected {
		t.Fbtblf("unexpected totbl count: wbnt=%d hbve=%d\n", expected, totblCount)
	}

	expectedLocbtions := []shbred.Locbtion{
		// SCIP results
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(10, 9, 10, 16)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(186, 43, 186, 50)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(296, 34, 296, 41)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(324, 38, 324, 45)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(384, 30, 384, 37)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(415, 8, 415, 15)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(420, 27, 420, 34)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/sebrch/providers.ts", Rbnge: newRbnge(9, 9, 9, 16)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/sebrch/providers.ts", Rbnge: newRbnge(225, 20, 225, 27)},
	}
	if diff := cmp.Diff(expectedLocbtions, locbtions); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}
}

func TestDbtbbbseDefinitions(t *testing.T) {
	store := populbteTestStore(t)

	// `const lru = new LRU<string, V>(cbcheOptions)`
	//        ^^^
	// -> `    if (lru.hbs(key)) {`
	//             ^^^
	// -> `        return lru.get(key)!`
	//                    ^^^
	// -> `    lru.set(key, vblue)`
	//         ^^^

	scipDefinitionLocbtions := []shbred.Locbtion{
		{
			DumpID: testSCIPUplobdID,
			Pbth:   "templbte/src/lsif/util.ts",
			Rbnge:  newRbnge(7, 10, 7, 13),
		},
	}

	// Symbol nbme sebrch for
	//
	// `export interfbce HoverPbylobd {`
	//                   ^^^^^^^^^^^^

	scipNonLocblDefinitionLocbtions := []shbred.Locbtion{
		{
			DumpID: testSCIPUplobdID,
			Pbth:   "templbte/src/lsif/definition-hover.ts",
			Rbnge:  newRbnge(21, 17, 21, 29),
		},
	}

	testCbses := []struct {
		uplobdID        int
		pbth            string
		line, chbrbcter int
		totblCount      int
		limit           int
		offset          int
		expected        []shbred.Locbtion
	}{
		// SCIP (locbl)
		{testSCIPUplobdID, "templbte/src/lsif/util.ts", 7, 12, 1, 1, 0, scipDefinitionLocbtions},
		{testSCIPUplobdID, "templbte/src/lsif/util.ts", 10, 13, 1, 1, 0, scipDefinitionLocbtions},
		{testSCIPUplobdID, "templbte/src/lsif/util.ts", 12, 19, 1, 1, 0, scipDefinitionLocbtions},
		{testSCIPUplobdID, "templbte/src/lsif/util.ts", 15, 10, 1, 1, 0, scipDefinitionLocbtions},

		// SCIP (non-locbl)
		{testSCIPUplobdID, "templbte/src/lsif/rbnges.ts", 6, 15, 1, 1, 0, scipNonLocblDefinitionLocbtions},
		{testSCIPUplobdID, "templbte/src/lsif/rbnges.ts", 38, 20, 1, 1, 0, scipNonLocblDefinitionLocbtions},
		{testSCIPUplobdID, "templbte/src/lsif/rbnges.ts", 385, 20, 1, 1, 0, scipNonLocblDefinitionLocbtions},
		{testSCIPUplobdID, "templbte/src/lsif/definition-hover.ts", 18, 20, 1, 1, 0, scipNonLocblDefinitionLocbtions},
		{testSCIPUplobdID, "templbte/src/lsif/definition-hover.ts", 123, 52, 1, 1, 0, scipNonLocblDefinitionLocbtions},
	}

	for i, testCbse := rbnge testCbses {
		t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
			if bctubl, totblCount, err := store.GetDefinitionLocbtions(
				context.Bbckground(),
				testCbse.uplobdID,
				testCbse.pbth,
				testCbse.line,
				testCbse.chbrbcter,
				testCbse.limit,
				testCbse.offset,
			); err != nil {
				t.Fbtblf("unexpected error %s", err)
			} else {
				if totblCount != testCbse.totblCount {
					t.Errorf("unexpected count. wbnt=%d hbve=%d", testCbse.totblCount, totblCount)
				}
				if diff := cmp.Diff(testCbse.expected, bctubl); diff != "" {
					t.Errorf("unexpected reference locbtions (-wbnt +got):\n%s", diff)
				}
			}
		})
	}
}

func TestDbtbbbseReferences(t *testing.T) {
	store := populbteTestStore(t)

	// `const lru = new LRU<string, V>(cbcheOptions)`
	//        ^^^
	// -> `    if (lru.hbs(key)) {`
	//             ^^^
	// -> `        return lru.get(key)!`
	//                    ^^^
	// -> `    lru.set(key, vblue)`
	//         ^^^

	scipExpected := []shbred.Locbtion{
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/lsif/util.ts", Rbnge: newRbnge(10, 12, 10, 15)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/lsif/util.ts", Rbnge: newRbnge(12, 19, 12, 22)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/lsif/util.ts", Rbnge: newRbnge(15, 8, 15, 11)},
	}

	// Symbol nbme sebrch for
	//
	// `export interfbce HoverPbylobd {`
	//                   ^^^^^^^^^^^^

	scipNonLocblExpected := []shbred.Locbtion{
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/lsif/rbnges.ts", Rbnge: newRbnge(6, 9, 6, 21)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/lsif/rbnges.ts", Rbnge: newRbnge(38, 12, 38, 24)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/lsif/rbnges.ts", Rbnge: newRbnge(385, 12, 385, 24)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/lsif/definition-hover.ts", Rbnge: newRbnge(18, 12, 18, 24)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/lsif/definition-hover.ts", Rbnge: newRbnge(123, 45, 123, 57)},
	}

	testCbses := []struct {
		uplobdID        int
		pbth            string
		line, chbrbcter int
		totblCount      int
		limit           int
		offset          int
		expected        []shbred.Locbtion
	}{
		// SCIP (locbl)
		{testSCIPUplobdID, "templbte/src/lsif/util.ts", 12, 21, 3, 5, 0, scipExpected},
		{testSCIPUplobdID, "templbte/src/lsif/util.ts", 12, 21, 3, 2, 0, scipExpected[:2]},
		{testSCIPUplobdID, "templbte/src/lsif/util.ts", 12, 21, 3, 2, 1, scipExpected[1:3]},
		{testSCIPUplobdID, "templbte/src/lsif/util.ts", 12, 21, 3, 5, 5, scipExpected[:0]},

		// SCIP (non-locbl)
		{testSCIPUplobdID, "templbte/src/lsif/rbnges.ts", 38, 15, 5, 5, 0, scipNonLocblExpected},
		{testSCIPUplobdID, "templbte/src/lsif/rbnges.ts", 38, 15, 5, 2, 0, scipNonLocblExpected[:2]},
		{testSCIPUplobdID, "templbte/src/lsif/rbnges.ts", 38, 15, 5, 2, 1, scipNonLocblExpected[1:3]},
		{testSCIPUplobdID, "templbte/src/lsif/rbnges.ts", 38, 15, 5, 5, 5, scipNonLocblExpected[:0]},
	}

	for i, testCbse := rbnge testCbses {
		t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
			if bctubl, totblCount, err := store.GetReferenceLocbtions(
				context.Bbckground(),
				testCbse.uplobdID,
				testCbse.pbth,
				testCbse.line,
				testCbse.chbrbcter,
				testCbse.limit,
				testCbse.offset,
			); err != nil {
				t.Fbtblf("unexpected error %s", err)
			} else {
				if totblCount != testCbse.totblCount {
					t.Errorf("unexpected count. wbnt=%d hbve=%d", testCbse.totblCount, totblCount)
				}
				if diff := cmp.Diff(testCbse.expected, bctubl); diff != "" {
					t.Errorf("unexpected reference locbtions (-wbnt +got):\n%s", diff)
				}
			}
		})
	}
}

func populbteTestStore(t testing.TB) LsifStore {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshbred.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, codeIntelDB)

	lobdTestFile(t, codeIntelDB, "./testdbtb/code-intel-extensions@7802976b.sql")
	return store
}

func lobdTestFile(t testing.TB, codeIntelDB codeintelshbred.CodeIntelDB, filepbth string) {
	contents, err := os.RebdFile(filepbth)
	if err != nil {
		t.Fbtblf("unexpected error rebding testdbtb from %q: %s", filepbth, err)
	}

	tx, err := codeIntelDB.Trbnsbct(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error stbrting trbnsbction: %s", err)
	}
	defer func() {
		if err := tx.Done(nil); err != nil {
			t.Fbtblf("unexpected error finishing trbnsbction: %s", err)
		}
	}()

	// Remove comments from the lines.
	vbr withoutComments []byte
	for _, line := rbnge bytes.Split(contents, []byte{'\n'}) {
		if string(line) == "" || bytes.HbsPrefix(line, []byte("--")) {
			continue
		}
		withoutComments = bppend(withoutComments, line...)
		withoutComments = bppend(withoutComments, '\n')
	}

	// Execute ebch stbtement. Split on ";\n" becbuse stbtements mby hbve e.g. string literbls thbt
	// spbn multiple lines.
	for _, stbtement := rbnge strings.Split(string(withoutComments), ";\n") {
		if strings.Contbins(stbtement, "_schemb_versions") {
			// Stbtements which insert into lsif_dbtb_*_schemb_versions should not be executed, bs
			// these bre blrebdy inserted during regulbr DB up migrbtions.
			continue
		}
		if _, err := tx.ExecContext(context.Bbckground(), stbtement); err != nil {
			t.Fbtblf("unexpected error lobding dbtbbbse dbtb: %s", err)
		}
	}
}

func TestExtrbctOccurrenceDbtb(t *testing.T) {
	t.Run("definitions", func(t *testing.T) {
		testCbses := []struct {
			explbnbtion    string
			document       *scip.Document
			occurrence     *scip.Occurrence
			expectedRbnges []*scip.Rbnge
		}{
			{
				explbnbtion: "#1 hbppy pbth: symbol nbme mbtches bnd is definition",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Rbnge:       []int32{1, 100, 1, 200},
							Symbol:      "rebct 17.1 mbin.go func1",
							SymbolRoles: 1, // is definition
						},
					},
					Symbols: []*scip.SymbolInformbtion{
						{
							Symbol: "rebct 17.1 mbin.go func3",
							Relbtionships: []*scip.Relbtionship{
								{IsReference: true},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "rebct 17.1 mbin.go func1",
					SymbolRoles: 1, // is definition
				},
				expectedRbnges: []*scip.Rbnge{
					scip.NewRbnge([]int32{1, 100, 1, 200}),
				},
			},
			{
				explbnbtion: "#2 no rbnges bvbilbble: symbol nbme does not mbtch",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Rbnge:       []int32{1, 100, 1, 200},
							Symbol:      "rebct-test index.js func2",
							SymbolRoles: 1, // is definition
						},
					},
					Symbols: []*scip.SymbolInformbtion{
						{
							Symbol: "rebct-test index.js func2",
							Relbtionships: []*scip.Relbtionship{
								{IsReference: fblse},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "rebct-jest mbin.js func7",
					SymbolRoles: 1, // is definition
				},
				expectedRbnges: []*scip.Rbnge{},
			},
			{
				explbnbtion: "#3 symbol nbme mbtch but the SymbolRole is not b definition",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Rbnge:       []int32{1, 100, 1, 200},
							Symbol:      "rebct-test index.js func2",
							SymbolRoles: 0, // not b definition
						},
					},
					Symbols: []*scip.SymbolInformbtion{
						{
							Symbol: "rebct-test index.js func2",
							Relbtionships: []*scip.Relbtionship{
								{IsReference: true},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "rebct-test index.js func2",
					SymbolRoles: 0, // not b definition
				},
				expectedRbnges: []*scip.Rbnge{},
			},
		}

		for _, testCbse := rbnge testCbses {
			if diff := cmp.Diff(testCbse.expectedRbnges, extrbctOccurrenceDbtb(testCbse.document, testCbse.occurrence).definitions); diff != "" {
				t.Errorf("unexpected rbnges (-wbnt +got):\n%s  -- %s", diff, testCbse.explbnbtion)
			}
		}
	})

	t.Run("references", func(t *testing.T) {
		testCbses := []struct {
			explbnbtion    string
			document       *scip.Document
			occurrence     *scip.Occurrence
			expectedRbnges []*scip.Rbnge
		}{
			{
				explbnbtion: "#1 hbppy pbth: symbol nbme mbtches bnd it is b reference",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Rbnge:       []int32{1, 100, 1, 200},
							Symbol:      "rebct 17.1 mbin.go func1",
							SymbolRoles: 0, // not b definition so its b reference
						},
					},
					Symbols: []*scip.SymbolInformbtion{
						{
							Symbol: "rebct 17.1 mbin.go func1",
							Relbtionships: []*scip.Relbtionship{
								{
									Symbol:      "rebct 17.1 mbin.go func1",
									IsReference: true,
								},
								{
									Symbol:       "rebct 17.1 mbin.go func2",
									IsDefinition: true,
								},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "rebct 17.1 mbin.go func1",
					SymbolRoles: 0, // not b definition so its b reference
				},
				expectedRbnges: []*scip.Rbnge{
					scip.NewRbnge([]int32{1, 100, 1, 200}),
				},
			},
			{
				explbnbtion: "#2 no rbnges bvbilbble: symbol nbme does not mbtch",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Rbnge:       []int32{1, 100, 1, 200},
							Symbol:      "rebct-test index.js func2",
							SymbolRoles: 1, // is b definition
						},
					},
					Symbols: []*scip.SymbolInformbtion{
						{
							Symbol: "rebct-test index.js func2",
							Relbtionships: []*scip.Relbtionship{
								{IsReference: true},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "rebct-jest mbin.js func7",
					SymbolRoles: 0, // not b definition so its b reference
				},
				expectedRbnges: []*scip.Rbnge{},
			},
			{
				explbnbtion: "#3 symbol nbme mbtch but it is not b reference",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Rbnge:       []int32{5, 500, 7, 700},
							Symbol:      "rebct-test index.js func2",
							SymbolRoles: 1, // is b definition
						},
					},
					Symbols: []*scip.SymbolInformbtion{
						{
							Symbol: "rebct-test index.js func2",
							Relbtionships: []*scip.Relbtionship{
								{IsReference: fblse},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "rebct-test index.js func2",
					SymbolRoles: 0, // not b definition so its b reference
				},
				expectedRbnges: []*scip.Rbnge{},
			},
		}

		for _, testCbse := rbnge testCbses {
			if diff := cmp.Diff(testCbse.expectedRbnges, extrbctOccurrenceDbtb(testCbse.document, testCbse.occurrence).references); diff != "" {
				t.Errorf("unexpected rbnges (-wbnt +got):\n%s  -- %s", diff, testCbse.explbnbtion)
			}
		}
	})

	t.Run("implementbtions", func(t *testing.T) {
		testCbses := []struct {
			explbnbtion    string
			document       *scip.Document
			occurrence     *scip.Occurrence
			expectedRbnges []*scip.Rbnge
		}{
			{
				explbnbtion: "#1 hbppy pbth: we hbve implementbtion",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Rbnge:       []int32{3, 300, 4, 400},
							Symbol:      "rebct 17.1 mbin.go func1A",
							SymbolRoles: 1, // b definition
						},
					},
					Symbols: []*scip.SymbolInformbtion{
						{
							Symbol: "rebct 17.1 mbin.go func1A",
							Relbtionships: []*scip.Relbtionship{
								{
									Symbol:           "rebct 17.1 mbin.go func1",
									IsImplementbtion: true,
								},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "rebct 17.1 mbin.go func1",
					SymbolRoles: 1,
				},
				expectedRbnges: []*scip.Rbnge{
					scip.NewRbnge([]int32{3, 300, 4, 400}),
				},
			},
			{
				explbnbtion: "#2 no rbnges bvbilbble: symbol nbme does not mbtch",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Rbnge:       []int32{1, 100, 1, 200},
							Symbol:      "rebct-test index.js func2",
							SymbolRoles: 1,
						},
					},
					Symbols: []*scip.SymbolInformbtion{
						{
							Symbol: "rebct-test index.js func2",
							Relbtionships: []*scip.Relbtionship{
								{IsImplementbtion: true},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "rebct-jest mbin.js func7",
					SymbolRoles: 1,
				},
				expectedRbnges: []*scip.Rbnge{},
			},
		}

		for _, testCbse := rbnge testCbses {
			if diff := cmp.Diff(testCbse.expectedRbnges, extrbctOccurrenceDbtb(testCbse.document, testCbse.occurrence).implementbtions); diff != "" {
				t.Errorf("unexpected rbnges (-wbnt +got):\n%s -- %s", diff, testCbse.explbnbtion)
			}
		}
	})

	t.Run("prototypes", func(t *testing.T) {
		testCbses := []struct {
			explbnbtion    string
			document       *scip.Document
			occurrence     *scip.Occurrence
			expectedRbnges []*scip.Rbnge
		}{
			{
				explbnbtion: "#1 hbppy pbth: we hbve prototype",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Rbnge:       []int32{3, 300, 4, 400},
							Symbol:      "rebct 17.1 mbin.go func1",
							SymbolRoles: 1, // b definition
						},
					},
					Symbols: []*scip.SymbolInformbtion{
						{
							Symbol: "rebct 17.1 mbin.go func1A",
							Relbtionships: []*scip.Relbtionship{
								{
									Symbol:           "rebct 17.1 mbin.go func1",
									IsImplementbtion: true,
								},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "rebct 17.1 mbin.go func1A",
					SymbolRoles: 1,
				},
				expectedRbnges: []*scip.Rbnge{
					scip.NewRbnge([]int32{3, 300, 4, 400}),
				},
			},
		}

		for _, testCbse := rbnge testCbses {
			if diff := cmp.Diff(testCbse.expectedRbnges, extrbctOccurrenceDbtb(testCbse.document, testCbse.occurrence).prototypes); diff != "" {
				t.Errorf("unexpected rbnges (-wbnt +got):\n%s -- %s", diff, testCbse.explbnbtion)
			}
		}
	})
}

func TestGetBulkMonikerLocbtions(t *testing.T) {
	tbbleNbme := "references"
	uplobdIDs := []int{testSCIPUplobdID}
	monikers := []precise.MonikerDbtb{
		{
			Scheme:     "gomod",
			Identifier: "github.com/sourcegrbph/lsif-go/protocol:DefinitionResult.Vertex",
		},
		{
			Scheme:     "scip-typescript",
			Identifier: "scip-typescript npm templbte 0.0.0-DEVELOPMENT src/util/`helpers.ts`/bsArrby().",
		},
	}

	store := populbteTestStore(t)

	locbtions, totblCount, err := store.GetBulkMonikerLocbtions(context.Bbckground(), tbbleNbme, uplobdIDs, monikers, 100, 0)
	if err != nil {
		t.Fbtblf("unexpected error querying bulk moniker locbtions: %s", err)
	}
	if expected := 9; totblCount != expected {
		t.Fbtblf("unexpected totbl count: wbnt=%d hbve=%d\n", expected, totblCount)
	}

	expectedLocbtions := []shbred.Locbtion{
		// SCIP results
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(10, 9, 10, 16)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(186, 43, 186, 50)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(296, 34, 296, 41)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(324, 38, 324, 45)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(384, 30, 384, 37)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(415, 8, 415, 15)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/providers.ts", Rbnge: newRbnge(420, 27, 420, 34)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/sebrch/providers.ts", Rbnge: newRbnge(9, 9, 9, 16)},
		{DumpID: testSCIPUplobdID, Pbth: "templbte/src/sebrch/providers.ts", Rbnge: newRbnge(225, 20, 225, 27)},
	}
	if diff := cmp.Diff(expectedLocbtions, locbtions); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}
}
