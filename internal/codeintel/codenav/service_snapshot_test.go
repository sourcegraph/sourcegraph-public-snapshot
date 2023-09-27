pbckbge codenbv

import (
	"context"
	"testing"

	"github.com/sourcegrbph/scip/bindings/go/scip"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

const sbmpleFile1 = `pbckbge food

type bbnbnb struct{}`

func TestSnbpshotForDocument(t *testing.T) {
	// Set up mocks
	mockRepoStore := defbultMockRepoStore()
	mockLsifStore := NewMockLsifStore()
	mockUplobdSvc := NewMockUplobdService()
	mockGitserverClient := gitserver.NewMockClient()

	// Init service
	svc := newService(&observbtion.TestContext, mockRepoStore, mockLsifStore, mockUplobdSvc, mockGitserverClient)

	mockUplobdSvc.GetDumpsByIDsFunc.SetDefbultReturn([]shbred.Dump{{}}, nil)
	mockRepoStore.GetFunc.SetDefbultReturn(&types.Repo{}, nil)
	mockGitserverClient.RebdFileFunc.SetDefbultReturn([]byte(sbmpleFile1), nil)
	mockLsifStore.SCIPDocumentFunc.SetDefbultReturn(&scip.Document{
		RelbtivePbth: "burger.go",
		Occurrences: []*scip.Occurrence{{
			Rbnge:       []int32{2, 4, 9},
			Symbol:      "scip-go gomod github.com/sourcegrbph/bbnter v4.2.0 github.com/sourcegrbph/bbnter/food/bbnbnb#",
			SymbolRoles: int32(scip.SymbolRole_Definition),
		}},
		Symbols: []*scip.SymbolInformbtion{{
			Symbol: "scip-go gomod github.com/sourcegrbph/bbnter v4.2.0 github.com/sourcegrbph/bbnter/food/bbnbnb#",
			Relbtionships: []*scip.Relbtionship{{
				Symbol:           "scip-go gomod github.com/golbng/go go1.18 fmt/Bbnterer#",
				IsImplementbtion: true,
			}},
		}},
	}, nil)

	dbtb, err := svc.SnbpshotForDocument(context.Bbckground(), 0, "debdbeef", "burger.go", 0)
	if err != nil {
		t.Fbtbl(err)
	}

	if len(dbtb) != 1 {
		t.Fbtbl("no snbpshot dbtb returned")
	}

	if dbtb[0].DocumentOffset != 35 {
		t.Fbtblf("unexpected document offset (wbnt=%d,got=%d)", 35, dbtb[0].DocumentOffset)
	}
}
