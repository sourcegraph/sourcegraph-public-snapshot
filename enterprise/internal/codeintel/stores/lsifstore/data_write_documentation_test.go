package lsifstore

import (
	"context"
	"testing"

	"github.com/hexops/autogold"
	"github.com/hexops/valast"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Note: You can `go test ./pkg -update` to update the expected `want` values in these tests.
// See https://github.com/hexops/autogold for more information.

func TestWriteDocumentationUpload(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()

	// Enable API docs search, so WriteDocumentationSearch is tested.
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				ApidocsSearchIndexing: "enabled",
			},
		},
	})
	defer conf.Mock(nil)

	// Get a documentation page we can use to test writes to the DB.
	tmpStore := populateTestStore(t)
	page, err := tmpStore.DocumentationPage(ctx, testBundleID, "/github.com/sourcegraph/lsif-go/internal/index")
	if err != nil {
		t.Fatal(err)
	}

	// Mock the actual upload.
	repo := &types.Repo{
		Name:    "github.com/sourcegraph/lsif-go",
		ID:      443, // arbitrary
		Private: false,
	}
	upload := dbstore.Upload{
		ID:             testBundleID,
		RepositoryName: string(repo.Name),
		RepositoryID:   int(repo.ID),
		Indexer:        "lsif-go",
	}
	isDefaultBranch := true

	db := dbtest.NewDB(t)
	store := NewStore(db, conf.DefaultClient(), &observation.TestContext)

	{
		tx, err := store.Transact(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// Write the page.
		repositoryNameID, languageNameID, err := tx.WriteDocumentationSearchPrework(ctx, upload, repo, isDefaultBranch)
		if err != nil {
			t.Fatal(err)
		}
		documentationPages := make(chan *precise.DocumentationPageData, 1)
		documentationPages <- page
		close(documentationPages)
		err = tx.WriteDocumentationPages(ctx, upload, repo, isDefaultBranch, documentationPages, repositoryNameID, languageNameID)
		if err != nil {
			t.Fatal(err)
		}
		if err = tx.Done(err); err != nil {
			t.Fatal(err)
		}
	}

	// Query the page and snapshot it.
	t.Run("query the page", func(t *testing.T) {
		gotPage, err := store.DocumentationPage(ctx, testBundleID, page.Tree.PathID)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Equal(t, gotPage)
	})
}

func TestWriteDocumentationPathInfo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()

	db := dbtest.NewDB(t)
	store := NewStore(db, conf.DefaultClient(), &observation.TestContext)

	pathInfo := &precise.DocumentationPathInfoData{
		PathID:  "/github.com/sourcegraph/lsif-go/internal",
		IsIndex: true,
		Children: []string{
			"/github.com/sourcegraph/lsif-go/internal/gomod",
			"/github.com/sourcegraph/lsif-go/internal/index",
		},
	}

	// Write documentation path info.
	documentationPathInfo := make(chan *precise.DocumentationPathInfoData, 1)
	documentationPathInfo <- pathInfo
	close(documentationPathInfo)
	err := store.WriteDocumentationPathInfo(ctx, testBundleID, documentationPathInfo)
	if err != nil {
		t.Fatal(err)
	}

	// Query the path info and snapshot it.
	t.Run("query path info", func(t *testing.T) {
		got, err := store.DocumentationPathInfo(ctx, testBundleID, pathInfo.PathID)
		if err != nil {
			t.Fatal(err)
		}
		autogold.Equal(t, got)
	})
}

func TestWriteDocumentationMappings(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()

	db := dbtest.NewDB(t)
	store := NewStore(db, conf.DefaultClient(), &observation.TestContext)

	filePath := "internal/index/indexer.go"
	mapping := precise.DocumentationMapping{
		ResultID: 1337, // arbitrary for sake of testing
		PathID:   "/github.com/sourcegraph/lsif-go/internal/index#NewIndexer",
		FilePath: &filePath,
	}

	// Write documentation mapping.
	documentationMappings := make(chan precise.DocumentationMapping, 1)
	documentationMappings <- mapping
	close(documentationMappings)
	err := store.WriteDocumentationMappings(ctx, testBundleID, documentationMappings)
	if err != nil {
		t.Fatal(err)
	}

	// Confirm lookup from PathID -> result ID works
	gotID, err := store.documentationPathIDToID(ctx, testBundleID, mapping.PathID)
	if err != nil {
		t.Fatal(err)
	}
	autogold.Want("ID", precise.ID("1337")).Equal(t, gotID)

	// Confirm lookup from PathID -> filepath works
	gotFilePath, err := store.documentationPathIDToFilePath(ctx, testBundleID, mapping.PathID)
	if err != nil {
		t.Fatal(err)
	}
	autogold.Want("FilePath", valast.Addr("internal/index/indexer.go").(*string)).Equal(t, gotFilePath)
}
