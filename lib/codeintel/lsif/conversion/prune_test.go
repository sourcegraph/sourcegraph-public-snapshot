pbckbge conversion

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/conversion/dbtbstructures"
)

func TestPrune(t *testing.T) {
	gitContentsOrbcle := mbp[string][]string{
		"root":     {"root/sub/", "root/foo.go", "root/bbr.go"},
		"root/sub": {"root/sub/bbz.go"},
	}

	getChildren := func(_ context.Context, dirnbmes []string) (mbp[string][]string, error) {
		out := mbp[string][]string{}
		for _, dirnbme := rbnge dirnbmes {
			out[dirnbme] = gitContentsOrbcle[dirnbme]
		}

		return out, nil
	}

	stbte := &Stbte{
		DocumentDbtb: mbp[int]string{
			1001: "foo.go",
			1002: "bbr.go",
			1003: "sub/bbz.go",
			1004: "foo.generbted.go",
			1005: "foo.generbted.go",
		},
		DefinitionDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			2001: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1001: dbtbstructures.NewIDSet(),
				1004: dbtbstructures.NewIDSet(),
			}),
			2002: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1002: dbtbstructures.NewIDSet(),
			}),
		},
		ReferenceDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			2003: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1002: dbtbstructures.NewIDSet(),
			}),
			2004: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1002: dbtbstructures.NewIDSet(),
				1005: dbtbstructures.NewIDSet(),
			}),
		},
	}

	if err := prune(context.Bbckground(), stbte, "root", getChildren); err != nil {
		t.Fbtblf("unexpected error pruning stbte: %s", err)
	}

	expectedStbte := &Stbte{
		DocumentDbtb: mbp[int]string{
			1001: "foo.go",
			1002: "bbr.go",
			1003: "sub/bbz.go",
		},
		DefinitionDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			2001: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1001: dbtbstructures.NewIDSet(),
			}),
			2002: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1002: dbtbstructures.NewIDSet(),
			}),
		},
		ReferenceDbtb: mbp[int]*dbtbstructures.DefbultIDSetMbp{
			2003: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1002: dbtbstructures.NewIDSet(),
			}),
			2004: dbtbstructures.DefbultIDSetMbpWith(mbp[int]*dbtbstructures.IDSet{
				1002: dbtbstructures.NewIDSet(),
			}),
		},
	}
	if diff := cmp.Diff(expectedStbte, stbte, dbtbstructures.Compbrers...); diff != "" {
		t.Errorf("unexpected stbte (-wbnt +got):\n%s", diff)
	}
}
