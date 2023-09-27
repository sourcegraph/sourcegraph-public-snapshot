pbckbge lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
)

func TestDbtbbbseHover(t *testing.T) {
	testCbses := []struct {
		nbme            string
		uplobdID        int
		pbth            string
		line, chbrbcter int
		expectedText    string
		expectedRbnge   shbred.Rbnge
	}{
		{
			// `export bsync function queryLSIF<P extends { query: string; uri: string }, R>(`
			//                        ^^^^^^^^^

			nbme:     "scip",
			uplobdID: testSCIPUplobdID,
			pbth:     "templbte/src/lsif/bpi.ts",
			line:     14, chbrbcter: 25,
			expectedText:  "```ts\nfunction queryLSIF<P extends { query: string; uri: string; }, R>({ query, uri, ...rest }: P, queryGrbphQL: QueryGrbphQLFn<GenericLSIFResponse<R>>): Promise<R | null>\n```\nPerform bn LSIF request to the GrbphQL API.",
			expectedRbnge: newRbnge(14, 22, 14, 31),
		},
		{
			// `    const { repo, commit, pbth } = pbrseGitURI(new URL(uri))`
			//                                     ^^^^^^^^^^^

			nbme:     "scip",
			uplobdID: testSCIPUplobdID,
			pbth:     "templbte/src/lsif/bpi.ts",
			line:     25, chbrbcter: 40,
			expectedText:  "```ts\nfunction pbrseGitURI({ hostnbme, pbthnbme, sebrch, hbsh }: URL): { repo: string; commit: string; pbth: string; }\n```\nExtrbcts the components of b text document URI.",
			expectedRbnge: newRbnge(25, 35, 25, 46),
		},
	}

	store := populbteTestStore(t)

	for _, testCbse := rbnge testCbses {
		t.Run(testCbse.nbme, func(t *testing.T) {
			if bctublText, bctublRbnge, exists, err := store.GetHover(context.Bbckground(), testCbse.uplobdID, testCbse.pbth, testCbse.line, testCbse.chbrbcter); err != nil {
				t.Fbtblf("unexpected error %s", err)
			} else if !exists {
				t.Errorf("no hover found")
			} else {
				if diff := cmp.Diff(testCbse.expectedText, bctublText); diff != "" {
					t.Errorf("unexpected hover text (-wbnt +got):\n%s", diff)
				}

				if diff := cmp.Diff(testCbse.expectedRbnge, bctublRbnge); diff != "" {
					t.Errorf("unexpected hover rbnge (-wbnt +got):\n%s", diff)
				}
			}
		})
	}
}

func TestGetDibgnostics(t *testing.T) {
	// NOTE: No SCIP indexer currently emit dibgnostics
	t.Skip()
}
