package indexing

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	enterprisedbstore "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/shared"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func TestDependencySyncSchedulerJVM(t *testing.T) {
	newOperations(&observation.TestContext)
	mockWorkerStore := NewMockWorkerStore()
	mockDBStore := NewMockDBStore()
	mockExtsvcStore := NewMockExternalServiceStore()
	mockDBStore.WithFunc.SetDefaultReturn(mockDBStore)
	mockScanner := NewMockPackageReferenceScanner()
	mockDBStore.ReferencesForUploadFunc.SetDefaultReturn(mockScanner, nil)
	mockDBStore.GetUploadByIDFunc.SetDefaultReturn(enterprisedbstore.Upload{ID: 42, RepositoryID: 50, Indexer: "scip-java"}, true, nil)
	mockScanner.NextFunc.PushReturn(shared.PackageReference{Package: shared.Package{DumpID: 42, Scheme: dependencies.JVMPackagesScheme, Name: "name1", Version: "v2.2.0"}}, true, nil)

	handler := dependencySyncSchedulerHandler{
		dbStore:     mockDBStore,
		workerStore: mockWorkerStore,
		extsvcStore: mockExtsvcStore,
	}

	logger := logtest.Scoped(t)
	job := enterprisedbstore.DependencySyncingJob{
		UploadID: 42,
	}
	if err := handler.Handle(context.Background(), logger, job); err != nil {
		t.Fatalf("unexpected error performing update: %s", err)
	}

	if len(mockDBStore.InsertDependencyIndexingJobFunc.History()) != 1 {
		t.Errorf("unexpected number of calls to InsertDependencyIndexingJob. want=%d have=%d", 1, len(mockDBStore.InsertDependencyIndexingJobFunc.History()))
	} else {
		var kinds []string
		for _, call := range mockDBStore.InsertDependencyIndexingJobFunc.History() {
			kinds = append(kinds, call.Arg2)
		}

		expectedKinds := []string{extsvc.KindJVMPackages}
		if diff := cmp.Diff(expectedKinds, kinds); diff != "" {
			t.Errorf("unexpected kinds (-want +got):\n%s", diff)
		}
	}

	if len(mockExtsvcStore.ListFunc.History()) != 1 {
		t.Errorf("unexpected number of calls to extsvc.List. want=%d have=%d", 1, len(mockExtsvcStore.ListFunc.History()))
	}

	if len(mockDBStore.InsertCloneableDependencyRepoFunc.History()) != 1 {
		t.Errorf("unexpected number of calls to InsertCloneableDependencyRepo. want=%d have=%d", 1, len(mockDBStore.InsertCloneableDependencyRepoFunc.History()))
	}
}

func TestDependencySyncSchedulerGomod(t *testing.T) {
	newOperations(&observation.TestContext)
	mockWorkerStore := NewMockWorkerStore()
	mockDBStore := NewMockDBStore()
	mockExtsvcStore := NewMockExternalServiceStore()
	mockDBStore.WithFunc.SetDefaultReturn(mockDBStore)
	mockScanner := NewMockPackageReferenceScanner()
	mockDBStore.ReferencesForUploadFunc.SetDefaultReturn(mockScanner, nil)
	mockDBStore.GetUploadByIDFunc.SetDefaultReturn(enterprisedbstore.Upload{ID: 42, RepositoryID: 50, Indexer: "lsif-go"}, true, nil)
	mockScanner.NextFunc.PushReturn(shared.PackageReference{Package: shared.Package{DumpID: 42, Scheme: "gomod", Name: "name1", Version: "v2.2.0"}}, true, nil)

	handler := dependencySyncSchedulerHandler{
		dbStore:     mockDBStore,
		workerStore: mockWorkerStore,
		extsvcStore: mockExtsvcStore,
	}

	logger := logtest.Scoped(t)
	job := enterprisedbstore.DependencySyncingJob{
		UploadID: 42,
	}
	if err := handler.Handle(context.Background(), logger, job); err != nil {
		t.Fatalf("unexpected error performing update: %s", err)
	}

	if len(mockDBStore.InsertDependencyIndexingJobFunc.History()) != 1 {
		t.Errorf("unexpected number of calls to InsertDependencyIndexingJob. want=%d have=%d", 1, len(mockDBStore.InsertDependencyIndexingJobFunc.History()))
	} else {
		var kinds []string
		for _, call := range mockDBStore.InsertDependencyIndexingJobFunc.History() {
			kinds = append(kinds, call.Arg2)
		}

		expectedKinds := []string{""}

		if diff := cmp.Diff(expectedKinds, kinds); diff != "" {
			t.Errorf("unexpected kinds (-want +got):\n%s", diff)
		}
	}

	if len(mockExtsvcStore.ListFunc.History()) != 0 {
		t.Errorf("unexpected number of calls to extsvc.List. want=%d have=%d", 0, len(mockExtsvcStore.ListFunc.History()))
	}

	if len(mockDBStore.InsertCloneableDependencyRepoFunc.History()) != 0 {
		t.Errorf("unexpected number of calls to InsertCloneableDependencyRepo. want=%d have=%d", 0, len(mockDBStore.InsertCloneableDependencyRepoFunc.History()))
	}
}

func TestNewPackage(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   shared.Package
		out  precise.Package
	}{
		{
			name: "jvm name normalization",
			in: shared.Package{
				Scheme:  dependencies.JVMPackagesScheme,
				Name:    "maven/junit/junit",
				Version: "4.2",
			},
			out: precise.Package{
				Scheme:  dependencies.JVMPackagesScheme,
				Name:    "junit:junit",
				Version: "4.2",
			},
		},
		{
			name: "jvm name normalization no-op",
			in: shared.Package{
				Scheme:  dependencies.JVMPackagesScheme,
				Name:    "junit:junit",
				Version: "4.2",
			},
			out: precise.Package{
				Scheme:  dependencies.JVMPackagesScheme,
				Name:    "junit:junit",
				Version: "4.2",
			},
		},
		{
			name: "npm no-op",
			in: shared.Package{
				Scheme:  dependencies.NpmPackagesScheme,
				Name:    "@graphql-mesh/graphql",
				Version: "0.24.0",
			},
			out: precise.Package{
				Scheme:  dependencies.NpmPackagesScheme,
				Name:    "@graphql-mesh/graphql",
				Version: "0.24.0",
			},
		},
		{
			name: "go no-op",
			in: shared.Package{
				Scheme:  dependencies.GoModulesScheme,
				Name:    "github.com/tsenart/vegeta/v12",
				Version: "12.7.0",
			},
			out: precise.Package{
				Scheme:  dependencies.GoModulesScheme,
				Name:    "github.com/tsenart/vegeta/v12",
				Version: "12.7.0",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			have := newPackage(tc.in)
			want := tc.out

			if diff := cmp.Diff(want, have); diff != "" {
				t.Fatalf("mismatch (-want, +have): %s", diff)
			}
		})
	}
}
