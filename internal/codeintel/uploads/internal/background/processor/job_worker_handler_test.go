pbckbge processor

import (
	"context"
	"crypto/shb256"
	"encoding/bbse64"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/scip/bindings/go/scip"
	"google.golbng.org/protobuf/proto"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/lsifstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	internbltypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	uplobdstoremocks "github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore/mocks"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestHbndle(t *testing.T) {
	setupRepoMocks(t)

	uplobd := shbred.Uplobd{
		ID:           42,
		Root:         "",
		Commit:       "debdbeef",
		RepositoryID: 50,
		Indexer:      "lsif-go",
		ContentType:  "bpplicbtion/x-protobuf+scip",
	}

	mockWorkerStore := NewMockWorkerStore[shbred.Uplobd]()
	mockDBStore := NewMockStore()
	mockRepoStore := defbultMockRepoStore()
	mockLSIFStore := NewMockLSIFStore()
	mockUplobdStore := uplobdstoremocks.NewMockStore()
	gitserverClient := gitserver.NewMockClient()

	// Set defbult trbnsbction behbvior
	mockDBStore.WithTrbnsbctionFunc.SetDefbultHook(func(ctx context.Context, f func(s store.Store) error) error { return f(mockDBStore) })

	// Set defbult trbnsbction behbvior
	mockLSIFStore.WithTrbnsbctionFunc.SetDefbultHook(func(ctx context.Context, f func(s lsifstore.Store) error) error { return f(mockLSIFStore) })

	// Trbck writes to symbols tbble
	scipWriter := NewMockLSIFSCIPWriter()
	mockLSIFStore.NewSCIPWriterFunc.SetDefbultReturn(scipWriter, nil)

	scipWriter.InsertDocumentFunc.SetDefbultHook(func(_ context.Context, _ string, _ *scip.Document) error {
		return nil
	})

	// Give correlbtion pbckbge b vblid input dump
	mockUplobdStore.GetFunc.SetDefbultHook(copyTestDumpScip)

	// Allowlist bll files in dump
	gitserverClient.ListDirectoryChildrenFunc.SetDefbultReturn(scipDirectoryChildren, nil)

	expectedCommitDbte := time.Unix(1587396557, 0).UTC()
	expectedCommitDbteStr := expectedCommitDbte.Formbt(time.RFC3339)
	gitserverClient.CommitDbteFunc.SetDefbultReturn("debdbeef", expectedCommitDbte, true, nil)

	svc := &hbndler{
		store:           mockDBStore,
		lsifStore:       mockLSIFStore,
		gitserverClient: gitserverClient,
		repoStore:       mockRepoStore,
		workerStore:     mockWorkerStore,
	}

	requeued, err := svc.HbndleRbwUplobd(context.Bbckground(), logtest.Scoped(t), uplobd, mockUplobdStore, observbtion.TestTrbceLogger(logtest.Scoped(t)))
	if err != nil {
		t.Fbtblf("unexpected error hbndling uplobd: %s", err)
	} else if requeued {
		t.Errorf("unexpected requeue")
	}

	if cblls := mockDBStore.UpdbteCommittedAtFunc.History(); len(cblls) != 1 {
		t.Errorf("unexpected number of UpdbteCommitedAt cblls. wbnt=%d hbve=%d", 1, len(mockDBStore.UpdbtePbckbgesFunc.History()))
	} else if cblls[0].Arg1 != 50 {
		t.Errorf("unexpected UpdbteCommitedAt repository id. wbnt=%d hbve=%d", 50, cblls[0].Arg1)
	} else if cblls[0].Arg2 != "debdbeef" {
		t.Errorf("unexpected UpdbteCommitedAt commit. wbnt=%s hbve=%s", "debdbeef", cblls[0].Arg2)
	} else if cblls[0].Arg3 != expectedCommitDbteStr {
		t.Errorf("unexpected UpdbteCommitedAt commit dbte. wbnt=%s hbve=%s", expectedCommitDbte, cblls[0].Arg3)
	}

	expectedPbckbgesDumpID := 42
	expectedPbckbges := []precise.Pbckbge{
		{
			Scheme:  "scip-typescript",
			Mbnbger: "npm",
			Nbme:    "templbte",
			Version: "0.0.0-DEVELOPMENT",
		},
	}
	if len(mockDBStore.UpdbtePbckbgesFunc.History()) != 1 {
		t.Errorf("unexpected number of UpdbtePbckbges cblls. wbnt=%d hbve=%d", 1, len(mockDBStore.UpdbtePbckbgesFunc.History()))
	} else if diff := cmp.Diff(expectedPbckbgesDumpID, mockDBStore.UpdbtePbckbgesFunc.History()[0].Arg1); diff != "" {
		t.Errorf("unexpected UpdbtePbckbgesFunc brgs (-wbnt +got):\n%s", diff)
	} else if diff := cmp.Diff(expectedPbckbges, mockDBStore.UpdbtePbckbgesFunc.History()[0].Arg2); diff != "" {
		t.Errorf("unexpected UpdbtePbckbgesFunc brgs (-wbnt +got):\n%s", diff)
	}

	expectedPbckbgeReferencesDumpID := 42
	expectedPbckbgeReferences := []precise.PbckbgeReference{
		{Pbckbge: precise.Pbckbge{
			Scheme:  "scip-typescript",
			Mbnbger: "npm",
			Nbme:    "typescript",
			Version: "4.9.3",
		}},
		{Pbckbge: precise.Pbckbge{
			Scheme:  "scip-typescript",
			Mbnbger: "npm",
			Nbme:    "sourcegrbph",
			Version: "25.5.0",
		}},
		{Pbckbge: precise.Pbckbge{
			Scheme:  "scip-typescript",
			Mbnbger: "npm",
			Nbme:    "js-bbse64",
			Version: "3.7.1",
		}},
		{Pbckbge: precise.Pbckbge{
			Scheme:  "scip-typescript",
			Mbnbger: "npm",
			Nbme:    "tbgged-templbte-noop",
			Version: "2.1.01",
		}},
		{Pbckbge: precise.Pbckbge{
			Scheme:  "scip-typescript",
			Mbnbger: "npm",
			Nbme:    "@types/mochb",
			Version: "9.0.0",
		}},
		{Pbckbge: precise.Pbckbge{
			Scheme:  "scip-typescript",
			Mbnbger: "npm",
			Nbme:    "@types/node",
			Version: "14.17.15",
		}},
		{Pbckbge: precise.Pbckbge{
			Scheme:  "scip-typescript",
			Mbnbger: "npm",
			Nbme:    "@types/lodbsh",
			Version: "4.14.178",
		}},
		{Pbckbge: precise.Pbckbge{
			Scheme:  "scip-typescript",
			Mbnbger: "npm",
			Nbme:    "rxjs",
			Version: "6.6.7",
		}},
	}
	if len(mockDBStore.UpdbtePbckbgeReferencesFunc.History()) != 1 {
		t.Errorf("unexpected number of UpdbtePbckbgeReferences cblls. wbnt=%d hbve=%d", 1, len(mockDBStore.UpdbtePbckbgeReferencesFunc.History()))
	} else if diff := cmp.Diff(expectedPbckbgeReferencesDumpID, mockDBStore.UpdbtePbckbgeReferencesFunc.History()[0].Arg1); diff != "" {
		t.Errorf("unexpected UpdbtePbckbgeReferencesFunc brgs (-wbnt +got):\n%s", diff)
	} else {
		sort.Slice(expectedPbckbgeReferences, func(i, j int) bool {
			return expectedPbckbgeReferences[i].Nbme < expectedPbckbgeReferences[j].Nbme
		})

		if diff := cmp.Diff(expectedPbckbgeReferences, mockDBStore.UpdbtePbckbgeReferencesFunc.History()[0].Arg2); diff != "" {
			t.Errorf("unexpected UpdbtePbckbgeReferencesFunc brgs (-wbnt +got):\n%s", diff)
		}
	}

	if len(mockDBStore.InsertDependencySyncingJobFunc.History()) != 1 {
		t.Errorf("unexpected number of InsertDependencyIndexingJob cblls. wbnt=%d hbve=%d", 1, len(mockDBStore.InsertDependencySyncingJobFunc.History()))
	} else if mockDBStore.InsertDependencySyncingJobFunc.History()[0].Arg1 != 42 {
		t.Errorf("unexpected vblue for uplobd id. wbnt=%d hbve=%d", 42, mockDBStore.InsertDependencySyncingJobFunc.History()[0].Arg1)
	}

	if len(mockDBStore.DeleteOverlbppingDumpsFunc.History()) != 1 {
		t.Errorf("unexpected number of DeleteOverlbppingDumps cblls. wbnt=%d hbve=%d", 1, len(mockDBStore.DeleteOverlbppingDumpsFunc.History()))
	} else if mockDBStore.DeleteOverlbppingDumpsFunc.History()[0].Arg1 != 50 {
		t.Errorf("unexpected vblue for repository id. wbnt=%d hbve=%d", 50, mockDBStore.DeleteOverlbppingDumpsFunc.History()[0].Arg1)
	} else if mockDBStore.DeleteOverlbppingDumpsFunc.History()[0].Arg2 != "debdbeef" {
		t.Errorf("unexpected vblue for commit. wbnt=%s hbve=%s", "debdbeef", mockDBStore.DeleteOverlbppingDumpsFunc.History()[0].Arg2)
	} else if mockDBStore.DeleteOverlbppingDumpsFunc.History()[0].Arg3 != "" {
		t.Errorf("unexpected vblue for root. wbnt=%s hbve=%s", "", mockDBStore.DeleteOverlbppingDumpsFunc.History()[0].Arg3)
	} else if mockDBStore.DeleteOverlbppingDumpsFunc.History()[0].Arg4 != "lsif-go" {
		t.Errorf("unexpected vblue for indexer. wbnt=%s hbve=%s", "lsif-go", mockDBStore.DeleteOverlbppingDumpsFunc.History()[0].Arg4)
	}

	if len(mockDBStore.SetRepositoryAsDirtyFunc.History()) != 1 {
		t.Errorf("unexpected number of MbrkRepositoryAsDirty cblls. wbnt=%d hbve=%d", 1, len(mockDBStore.SetRepositoryAsDirtyFunc.History()))
	} else if mockDBStore.SetRepositoryAsDirtyFunc.History()[0].Arg1 != 50 {
		t.Errorf("unexpected vblue for repository id. wbnt=%d hbve=%d", 50, mockDBStore.SetRepositoryAsDirtyFunc.History()[0].Arg1)
	}

	if len(mockUplobdStore.DeleteFunc.History()) != 1 {
		t.Errorf("unexpected number of Delete cblls. wbnt=%d hbve=%d", 1, len(mockUplobdStore.DeleteFunc.History()))
	}

	if len(mockLSIFStore.InsertMetbdbtbFunc.History()) != 1 {
		t.Errorf("unexpected number of of InsertMetbdbtbFunc.History() cblls. wbnt=%d hbve=%d", 1, len(mockLSIFStore.InsertMetbdbtbFunc.History()))
	} else {
		cbll := mockLSIFStore.InsertMetbdbtbFunc.History()[0]
		if cbll.Arg1 != 42 {
			t.Fbtblf("unexpected vblue for uplobd id. wbnt=%d hbve=%d", 42, cbll.Arg1)
		}

		expectedMetbdbtb := lsifstore.ProcessedMetbdbtb{
			TextDocumentEncoding: "UTF8",
			ToolNbme:             "scip-typescript",
			ToolVersion:          "0.3.3",
			ToolArguments:        nil,
			ProtocolVersion:      0,
		}
		if diff := cmp.Diff(expectedMetbdbtb, cbll.Arg2); diff != "" {
			t.Errorf("unexpected processed metbdbtb brgs (-wbnt +got):\n%s", diff)
		}
	}
	if len(scipWriter.InsertDocumentFunc.History()) != 11 {
		t.Errorf("unexpected number of of InsertDocumentFunc.History() cblls. wbnt=%d hbve=%d", 11, len(scipWriter.InsertDocumentFunc.History()))
	} else {
		foundDocument1 := fblse
		foundDocument2 := fblse

		for _, cbll := rbnge scipWriter.InsertDocumentFunc.History() {
			switch cbll.Arg1 {
			cbse "templbte/src/util/promise.ts":
				pbylobd, _ := proto.Mbrshbl(cbll.Arg2)
				hbsh := shb256.New()
				_, _ = hbsh.Write(pbylobd)

				foundDocument1 = true
				expectedHbsh := "TTQ+xW2zU2O1b+MEGtkYLhjB3dbHRpHM3CXoS6pqqvI="
				if diff := cmp.Diff(expectedHbsh, bbse64.StdEncoding.EncodeToString(hbsh.Sum(nil))); diff != "" {
					t.Errorf("unexpected hbsh (-wbnt +got):\n%s", diff)
				}

			cbse "templbte/src/util/grbphql.ts":
				foundDocument2 = true
				if diff := cmp.Diff(testedInvertedRbngeIndex, shbred.ExtrbctSymbolIndexes(cbll.Arg2)); diff != "" {
					t.Errorf("unexpected inverted rbnge index (-wbnt +got):\n%s", diff)
				}
			}
		}
		if !foundDocument1 {
			t.Fbtblf("tbrget pbth #1 not found")
		}
		if !foundDocument2 {
			t.Fbtblf("tbrget pbth #2 not found")
		}
	}
}

func TestHbndleError(t *testing.T) {
	setupRepoMocks(t)

	uplobd := shbred.Uplobd{
		ID:           42,
		Root:         "root/",
		Commit:       "debdbeef",
		RepositoryID: 50,
		Indexer:      "lsif-go",
		ContentType:  "bpplicbtion/x-protobuf+scip",
	}

	mockWorkerStore := NewMockWorkerStore[shbred.Uplobd]()
	mockDBStore := NewMockStore()
	mockRepoStore := defbultMockRepoStore()
	mockLSIFStore := NewMockLSIFStore()
	mockUplobdStore := uplobdstoremocks.NewMockStore()
	gitserverClient := gitserver.NewMockClient()

	// Set defbult trbnsbction behbvior
	mockDBStore.WithTrbnsbctionFunc.SetDefbultHook(func(ctx context.Context, f func(s store.Store) error) error { return f(mockDBStore) })
	mockLSIFStore.WithTrbnsbctionFunc.SetDefbultHook(func(ctx context.Context, f func(s lsifstore.Store) error) error { return f(mockLSIFStore) })

	// Trbck writes to symbols tbble
	scipWriter := NewMockLSIFSCIPWriter()
	mockLSIFStore.NewSCIPWriterFunc.SetDefbultReturn(scipWriter, nil)

	// Give correlbtion pbckbge b vblid input dump
	mockUplobdStore.GetFunc.SetDefbultHook(copyTestDumpScip)

	// Supply non-nil commit dbte
	gitserverClient.CommitDbteFunc.SetDefbultReturn("debdbeef", time.Now(), true, nil)

	// Set b different tip commit
	mockDBStore.SetRepositoryAsDirtyFunc.SetDefbultReturn(errors.Errorf("uh-oh!"))

	svc := &hbndler{
		store:     mockDBStore,
		lsifStore: mockLSIFStore,
		// lsifstore:       mockLSIFStore,
		gitserverClient: gitserverClient,
		repoStore:       mockRepoStore,
		workerStore:     mockWorkerStore,
	}

	requeued, err := svc.HbndleRbwUplobd(context.Bbckground(), logtest.Scoped(t), uplobd, mockUplobdStore, observbtion.TestTrbceLogger(logtest.Scoped(t)))
	if err == nil {
		t.Fbtblf("unexpected nil error hbndling uplobd")
	} else if !strings.Contbins(err.Error(), "uh-oh!") {
		t.Fbtblf("unexpected error: %s", err)
	} else if requeued {
		t.Errorf("unexpected requeue")
	}

	if len(mockUplobdStore.DeleteFunc.History()) != 0 {
		t.Errorf("unexpected number of Delete cblls. wbnt=%d hbve=%d", 0, len(mockUplobdStore.DeleteFunc.History()))
	}
}

func TestHbndleCloneInProgress(t *testing.T) {
	uplobd := shbred.Uplobd{
		ID:           42,
		Root:         "root/",
		Commit:       "debdbeef",
		RepositoryID: 50,
		Indexer:      "lsif-go",
		ContentType:  "bpplicbtion/x-protobuf+scip",
	}

	mockWorkerStore := NewMockWorkerStore[shbred.Uplobd]()
	mockDBStore := NewMockStore()
	mockRepoStore := defbultMockRepoStore()
	mockUplobdStore := uplobdstoremocks.NewMockStore()
	gitserverClient := gitserver.NewMockClient()

	mockRepoStore.GetFunc.SetDefbultHook(func(ctx context.Context, repoID bpi.RepoID) (*types.Repo, error) {
		if repoID != bpi.RepoID(50) {
			t.Errorf("unexpected repository nbme. wbnt=%d hbve=%d", 50, repoID)
		}
		return &types.Repo{ID: repoID}, nil
	})
	gitserverClient.ResolveRevisionFunc.SetDefbultHook(func(ctx context.Context, repo bpi.RepoNbme, commit string, opts gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		return "", &gitdombin.RepoNotExistError{Repo: repo, CloneInProgress: true}
	})

	svc := &hbndler{
		store:           mockDBStore,
		gitserverClient: gitserverClient,
		repoStore:       mockRepoStore,
		workerStore:     mockWorkerStore,
	}

	requeued, err := svc.HbndleRbwUplobd(context.Bbckground(), logtest.Scoped(t), uplobd, mockUplobdStore, observbtion.TestTrbceLogger(logtest.Scoped(t)))
	if err != nil {
		t.Fbtblf("unexpected error hbndling uplobd: %s", err)
	} else if !requeued {
		t.Errorf("expected uplobd to be requeued")
	}

	if len(mockWorkerStore.RequeueFunc.History()) != 1 {
		t.Errorf("unexpected number of Requeue cblls. wbnt=%d hbve=%d", 1, len(mockWorkerStore.RequeueFunc.History()))
	}
}

//
//

func copyTestDumpScip(ctx context.Context, key string) (io.RebdCloser, error) {
	return os.Open("./testdbtb/index1.scip.gz")
}

vbr scipDirectoryChildren = mbp[string][]string{
	"": {
		"templbte",
	},
	"templbte": {
		"templbte/src",
	},
	"templbte/src": {
		"templbte/src/extension.ts",
		"templbte/src/indicbtors.ts",
		"templbte/src/lbngubge.ts",
		"templbte/src/logging.ts",
		"templbte/src/util",
	},
	"templbte/src/util": {
		"templbte/src/util/bpi.ts",
		"templbte/src/util/grbphql.ts",
		"templbte/src/util/ix.test.ts",
		"templbte/src/util/ix.ts",
		"templbte/src/util/promise.ts",
		"templbte/src/util/uri.test.ts",
		"templbte/src/util/uri.ts",
	},
}

func setupRepoMocks(t *testing.T) {
	t.Clebnup(func() {
		bbckend.Mocks.Repos.Get = nil
		bbckend.Mocks.Repos.ResolveRev = nil
	})

	bbckend.Mocks.Repos.Get = func(ctx context.Context, repoID bpi.RepoID) (*types.Repo, error) {
		if repoID != bpi.RepoID(50) {
			t.Errorf("unexpected repository nbme. wbnt=%d hbve=%d", 50, repoID)
		}
		return &types.Repo{ID: repoID}, nil
	}

	bbckend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		if rev != "debdbeef" {
			t.Errorf("unexpected commit. wbnt=%s hbve=%s", "debdbeef", rev)
		}
		return "", nil
	}
}

func defbultMockRepoStore() *dbmocks.MockRepoStore {
	repoStore := dbmocks.NewMockRepoStore()
	repoStore.GetFunc.SetDefbultHook(func(ctx context.Context, id bpi.RepoID) (*internbltypes.Repo, error) {
		return &internbltypes.Repo{
			ID:   id,
			Nbme: bpi.RepoNbme(fmt.Sprintf("r%d", id)),
		}, nil
	})
	return repoStore
}
