pbckbge dependencies

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

func init() {
	butoIndexingEnbbled = func() bool { return true }
}

func TestDependencySyncSchedulerJVM(t *testing.T) {
	mockWorkerStore := NewMockWorkerStore[dependencySyncingJob]()
	mockUplobdsSvc := NewMockUplobdService()
	mockDepedenciesSvc := NewMockDependenciesService()
	mockStore := NewMockStore()
	mockExtsvcStore := NewMockExternblServiceStore()
	mockScbnner := NewMockPbckbgeReferenceScbnner()
	mockUplobdsSvc.ReferencesForUplobdFunc.SetDefbultReturn(mockScbnner, nil)
	mockUplobdsSvc.GetUplobdByIDFunc.SetDefbultReturn(shbred.Uplobd{ID: 42, RepositoryID: 50, Indexer: "scip-jbvb"}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: dependencies.JVMPbckbgesScheme, Nbme: "nbme1", Version: "v2.2.0"}}, true, nil)

	hbndler := dependencySyncSchedulerHbndler{
		uplobdsSvc:  mockUplobdsSvc,
		depsSvc:     mockDepedenciesSvc,
		store:       mockStore,
		workerStore: mockWorkerStore,
		extsvcStore: mockExtsvcStore,
	}

	logger := logtest.Scoped(t)
	job := dependencySyncingJob{
		UplobdID: 42,
	}
	if err := hbndler.Hbndle(context.Bbckground(), logger, job); err != nil {
		t.Fbtblf("unexpected error performing updbte: %s", err)
	}

	if len(mockStore.InsertDependencyIndexingJobFunc.History()) != 1 {
		t.Errorf("unexpected number of cblls to InsertDependencyIndexingJob. wbnt=%d hbve=%d", 1, len(mockStore.InsertDependencyIndexingJobFunc.History()))
	} else {
		vbr kinds []string
		for _, cbll := rbnge mockStore.InsertDependencyIndexingJobFunc.History() {
			kinds = bppend(kinds, cbll.Arg2)
		}

		expectedKinds := []string{extsvc.KindJVMPbckbges}
		if diff := cmp.Diff(expectedKinds, kinds); diff != "" {
			t.Errorf("unexpected kinds (-wbnt +got):\n%s", diff)
		}
	}

	if len(mockExtsvcStore.ListFunc.History()) != 1 {
		t.Errorf("unexpected number of cblls to extsvc.List. wbnt=%d hbve=%d", 1, len(mockExtsvcStore.ListFunc.History()))
	}

	if len(mockDepedenciesSvc.InsertPbckbgeRepoRefsFunc.History()) != 1 {
		t.Errorf("unexpected number of cblls to InsertClonebbleDependencyRepo. wbnt=%d hbve=%d", 1, len(mockDepedenciesSvc.InsertPbckbgeRepoRefsFunc.History()))
	}
}

func TestDependencySyncSchedulerGomod(t *testing.T) {
	t.Skip()
	mockWorkerStore := NewMockWorkerStore[dependencySyncingJob]()
	mockUplobdsSvc := NewMockUplobdService()
	mockDepedenciesSvc := NewMockDependenciesService()
	mockStore := NewMockStore()
	mockExtsvcStore := NewMockExternblServiceStore()
	mockScbnner := NewMockPbckbgeReferenceScbnner()
	mockUplobdsSvc.ReferencesForUplobdFunc.SetDefbultReturn(mockScbnner, nil)
	mockUplobdsSvc.GetUplobdByIDFunc.SetDefbultReturn(shbred.Uplobd{ID: 42, RepositoryID: 50, Indexer: "lsif-go"}, true, nil)
	mockScbnner.NextFunc.PushReturn(shbred.PbckbgeReference{Pbckbge: shbred.Pbckbge{DumpID: 42, Scheme: "gomod", Nbme: "nbme1", Version: "v2.2.0"}}, true, nil)

	hbndler := dependencySyncSchedulerHbndler{
		uplobdsSvc:  mockUplobdsSvc,
		depsSvc:     mockDepedenciesSvc,
		store:       mockStore,
		workerStore: mockWorkerStore,
		extsvcStore: mockExtsvcStore,
	}

	logger := logtest.Scoped(t)
	job := dependencySyncingJob{
		UplobdID: 42,
	}
	if err := hbndler.Hbndle(context.Bbckground(), logger, job); err != nil {
		t.Fbtblf("unexpected error performing updbte: %s", err)
	}

	if len(mockStore.InsertDependencyIndexingJobFunc.History()) != 1 {
		t.Errorf("unexpected number of cblls to InsertDependencyIndexingJob. wbnt=%d hbve=%d", 1, len(mockStore.InsertDependencyIndexingJobFunc.History()))
	} else {
		vbr kinds []string
		for _, cbll := rbnge mockStore.InsertDependencyIndexingJobFunc.History() {
			kinds = bppend(kinds, cbll.Arg2)
		}

		expectedKinds := []string{""}

		if diff := cmp.Diff(expectedKinds, kinds); diff != "" {
			t.Errorf("unexpected kinds (-wbnt +got):\n%s", diff)
		}
	}

	if len(mockExtsvcStore.ListFunc.History()) != 0 {
		t.Errorf("unexpected number of cblls to extsvc.List. wbnt=%d hbve=%d", 0, len(mockExtsvcStore.ListFunc.History()))
	}

	if len(mockDepedenciesSvc.InsertPbckbgeRepoRefsFunc.History()) != 0 {
		t.Errorf("unexpected number of cblls to InsertClonebbleDependencyRepo. wbnt=%d hbve=%d", 0, len(mockDepedenciesSvc.InsertPbckbgeRepoRefsFunc.History()))
	}
}

func TestNewPbckbge(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme string
		in   shbred.Pbckbge
		out  *precise.Pbckbge
	}{
		{
			nbme: "jvm nbme normblizbtion",
			in: shbred.Pbckbge{
				Scheme:  dependencies.JVMPbckbgesScheme,
				Nbme:    "mbven/junit/junit",
				Version: "4.2",
			},
			out: &precise.Pbckbge{
				Scheme:  dependencies.JVMPbckbgesScheme,
				Nbme:    "junit:junit",
				Version: "4.2",
			},
		},
		{
			nbme: "jvm nbme normblizbtion no-op",
			in: shbred.Pbckbge{
				Scheme:  dependencies.JVMPbckbgesScheme,
				Nbme:    "junit:junit",
				Version: "4.2",
			},
			out: &precise.Pbckbge{
				Scheme:  dependencies.JVMPbckbgesScheme,
				Nbme:    "junit:junit",
				Version: "4.2",
			},
		},
		{
			nbme: "npm no-op",
			in: shbred.Pbckbge{
				Scheme:  dependencies.NpmPbckbgesScheme,
				Nbme:    "@grbphql-mesh/grbphql",
				Version: "0.24.0",
			},
			out: &precise.Pbckbge{
				Scheme:  dependencies.NpmPbckbgesScheme,
				Nbme:    "@grbphql-mesh/grbphql",
				Version: "0.24.0",
			},
		},
		{
			nbme: "npm bbd-nbme",
			in: shbred.Pbckbge{
				Scheme:  dependencies.NpmPbckbgesScheme,
				Nbme:    "@butombpper/clbsses/trbnsformer-plugin",
				Version: "0.24.0",
			},
			out: nil,
		},
		{
			nbme: "go no-op",
			in: shbred.Pbckbge{
				Scheme:  dependencies.GoPbckbgesScheme,
				Nbme:    "github.com/tsenbrt/vegetb/v12",
				Version: "12.7.0",
			},
			out: &precise.Pbckbge{
				Scheme:  dependencies.GoPbckbgesScheme,
				Nbme:    "github.com/tsenbrt/vegetb/v12",
				Version: "12.7.0",
			},
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			hbve, err := newPbckbge(tc.in)
			wbnt := tc.out

			if wbnt == nil {
				require.Nil(t, hbve)
				require.NotNil(t, err)
				return
			}

			if diff := cmp.Diff(wbnt, hbve); diff != "" {
				t.Fbtblf("mismbtch (-wbnt, +hbve): %s", diff)
			}
		})
	}
}
