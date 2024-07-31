package processor

import (
	"context"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codegraph"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestCorrelateSCIP(t *testing.T) {
	testIndexPath := "./testdata/index1.scip.gz"
	fileInfo, err := os.Stat(testIndexPath)
	require.NoError(t, err)

	var smallLimit int64 = 1024
	require.Greater(t, fileInfo.Size(), smallLimit)
	implCorrelateSCIP(t, testIndexPath, smallLimit)

	var largeLimit int64 = 100 * 1024 * 1024
	require.Less(t, fileInfo.Size(), largeLimit)
	implCorrelateSCIP(t, testIndexPath, largeLimit)
}

func implCorrelateSCIP(t *testing.T, testIndexPath string, indexSizeLimit int64) {
	ctx := context.Background()

	oldValue := uncompressedSizeLimitBytes
	uncompressedSizeLimitBytes = indexSizeLimit
	t.Cleanup(func() {
		uncompressedSizeLimitBytes = oldValue
	})

	testReader := func() gzipReadSeeker {
		gzipped, err := os.Open(testIndexPath)
		if err != nil {
			t.Fatalf("unexpected error reading test file: %s", err)
		}
		indexReader, err := newGzipReadSeeker(gzipped)
		require.NoError(t, err, "failed to create reader for test file")

		return indexReader
	}

	// Correlate and consume channels from returned object
	scipDataStream, err := prepareSCIPDataStream(ctx, testReader(), "", func(ctx context.Context, dirnames []string) (map[string][]string, error) {
		return scipDirectoryChildren, nil
	})
	if err != nil {
		t.Fatalf("unexpected error processing SCIP: %s", err)
	}
	var documents []codegraph.ProcessedSCIPDocument
	packageData := codegraph.ProcessedPackageData{}
	err = scipDataStream.DocumentIterator.VisitAllDocuments(ctx, log.NoOp(), &packageData, func(d codegraph.ProcessedSCIPDocument) error {
		documents = append(documents, d)
		return nil
	})
	require.NoError(t, err)
	packageData.Normalize()
	packages := packageData.Packages
	packageReferences := packageData.PackageReferences
	if err != nil {
		t.Fatalf("unexpected error reading processed SCIP: %s", err)
	}

	// Check metadata values
	expectedMetadata := codegraph.ProcessedMetadata{
		TextDocumentEncoding: "UTF8",
		ToolName:             "scip-typescript",
		ToolVersion:          "0.3.3",
		ToolArguments:        nil,
		ProtocolVersion:      0,
	}
	if diff := cmp.Diff(expectedMetadata, scipDataStream.Metadata); diff != "" {
		t.Fatalf("unexpected metadata (-want +got):\n%s", diff)
	}

	// Check document values
	if len(documents) != 11 {
		t.Fatalf("unexpected number of documents. want=%d have=%d", 11, len(documents))
	} else {
		documentMap := map[string]codegraph.ProcessedSCIPDocument{}
		for _, document := range documents {
			documentMap[document.Path] = document
		}

		var paths []string
		for path := range documentMap {
			paths = append(paths, path)
		}
		sort.Strings(paths)

		expectedPaths := []string{
			"template/src/extension.ts",
			"template/src/indicators.ts",
			"template/src/language.ts",
			"template/src/logging.ts",
			"template/src/util/api.ts",
			"template/src/util/graphql.ts",
			"template/src/util/ix.test.ts",
			"template/src/util/ix.ts",
			"template/src/util/promise.ts",
			"template/src/util/uri.test.ts",
			"template/src/util/uri.ts",
		}
		if diff := cmp.Diff(expectedPaths, paths); diff != "" {
			t.Errorf("unexpected paths (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(testedInvertedRangeIndex, shared.ExtractSymbolIndexes(documentMap["template/src/util/graphql.ts"].Document)); diff != "" {
			t.Errorf("unexpected inverted symbols (-want +got):\n%s", diff)
		}
	}

	// Check package and references values
	expectedPackages := []precise.Package{
		{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "template",
			Version: "0.0.0-DEVELOPMENT",
		},
	}
	if diff := cmp.Diff(expectedPackages, packages); diff != "" {
		t.Errorf("unexpected packages (-want +got):\n%s", diff)
	}
	expectedReferences := []precise.PackageReference{
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "@types/lodash",
			Version: "4.14.178",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "@types/mocha",
			Version: "9.0.0",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "@types/node",
			Version: "14.17.15",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "js-base64",
			Version: "3.7.1",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "rxjs",
			Version: "6.6.7",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "sourcegraph",
			Version: "25.5.0",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "tagged-template-noop",
			Version: "2.1.01",
		}},
		{Package: precise.Package{
			Scheme:  "scip-typescript",
			Manager: "npm",
			Name:    "typescript",
			Version: "4.9.3",
		}},
	}
	if diff := cmp.Diff(expectedReferences, packageReferences); diff != "" {
		t.Errorf("unexpected references (-want +got):\n%s", diff)
	}
}

var testedInvertedRangeIndex = []shared.InvertedRangeIndex{
	{
		SymbolName:      "scip-typescript npm js-base64 3.7.1 `base64.d.ts`/",
		ReferenceRanges: []int32{0, 27, 0, 38},
	},
	{
		SymbolName:      "scip-typescript npm js-base64 3.7.1 `base64.d.ts`/decode.",
		ReferenceRanges: []int32{0, 9, 0, 19, 42, 22, 42, 32},
	},
	{
		SymbolName:      "scip-typescript npm sourcegraph 25.5.0 src/`sourcegraph.d.ts`/`'sourcegraph'`/",
		ReferenceRanges: []int32{1, 12, 1, 23, 1, 29, 1, 42, 25, 27, 25, 38},
	},
	{
		SymbolName:      "scip-typescript npm sourcegraph 25.5.0 src/`sourcegraph.d.ts`/`'sourcegraph'`/commands/",
		ReferenceRanges: []int32{25, 39, 25, 47},
	},
	{
		SymbolName:      "scip-typescript npm sourcegraph 25.5.0 src/`sourcegraph.d.ts`/`'sourcegraph'`/commands/executeCommand().",
		ReferenceRanges: []int32{25, 48, 25, 62},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/",
		DefinitionRanges: []int32{0, 0, 0, 0},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/GraphQLResponse#",
		DefinitionRanges: []int32{3, 5, 3, 20},
		ReferenceRanges:  []int32{25, 63, 25, 78},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/GraphQLResponse#[T]",
		DefinitionRanges: []int32{3, 21, 3, 22},
		ReferenceRanges:  []int32{3, 49, 3, 50},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/GraphQLResponseError#",
		DefinitionRanges: []int32{10, 10, 10, 30},
		ReferenceRanges:  []int32{3, 54, 3, 74},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/GraphQLResponseError#data.",
		DefinitionRanges: []int32{11, 4, 11, 8},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/GraphQLResponseError#errors.",
		DefinitionRanges: []int32{12, 4, 12, 10},
		ReferenceRanges:  []int32{27, 17, 27, 23, 28, 23, 28, 29, 28, 54, 28, 60, 28, 91, 28, 97},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/GraphQLResponseSuccess#",
		DefinitionRanges: []int32{5, 10, 5, 32},
		ReferenceRanges:  []int32{3, 26, 3, 48},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/GraphQLResponseSuccess#[T]",
		DefinitionRanges: []int32{5, 33, 5, 34},
		ReferenceRanges:  []int32{6, 10, 6, 11},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/GraphQLResponseSuccess#data.",
		DefinitionRanges: []int32{6, 4, 6, 8},
		ReferenceRanges:  []int32{31, 20, 31, 24},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/GraphQLResponseSuccess#errors.",
		DefinitionRanges: []int32{7, 4, 7, 10},
		ReferenceRanges:  []int32{27, 17, 27, 23},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/QueryGraphQLFn#",
		DefinitionRanges: []int32{16, 12, 16, 26},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/QueryGraphQLFn#[T]",
		DefinitionRanges: []int32{16, 27, 16, 28},
		ReferenceRanges:  []int32{16, 95, 16, 96},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/aggregateErrors().",
		DefinitionRanges: []int32{34, 9, 34, 24},
		ReferenceRanges:  []int32{28, 66, 28, 81},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/aggregateErrors().(errors)",
		DefinitionRanges: []int32{34, 25, 34, 31},
		ReferenceRanges:  []int32{35, 35, 35, 41, 37, 8, 37, 14},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/errors0:",
		DefinitionRanges: []int32{37, 8, 37, 14},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/graphqlIdToRepoId().",
		DefinitionRanges: []int32{41, 16, 41, 33},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/graphqlIdToRepoId().(id)",
		DefinitionRanges: []int32{41, 34, 41, 36},
		ReferenceRanges:  []int32{42, 33, 42, 35},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/name0:",
		DefinitionRanges: []int32{36, 8, 36, 12},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/queryGraphQL().",
		DefinitionRanges: []int32{24, 22, 24, 34},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/queryGraphQL().(query)",
		DefinitionRanges: []int32{24, 38, 24, 43},
		ReferenceRanges:  []int32{25, 99, 25, 104},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/queryGraphQL().(vars)",
		DefinitionRanges: []int32{24, 53, 24, 57},
		ReferenceRanges:  []int32{25, 106, 25, 110},
	},
	{
		SymbolName:       "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`graphql.ts`/queryGraphQL().[T]",
		DefinitionRanges: []int32{24, 35, 24, 36},
		ReferenceRanges:  []int32{24, 102, 24, 103, 25, 79, 25, 80},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es2015.core.d.ts`/ObjectConstructor#assign().",
		ReferenceRanges: []int32{35, 18, 35, 24},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es2015.iterable.d.ts`/Promise#",
		ReferenceRanges: []int32{16, 87, 16, 94, 24, 94, 24, 101},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es2015.promise.d.ts`/Promise.",
		ReferenceRanges: []int32{16, 87, 16, 94, 24, 94, 24, 101},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es2015.symbol.wellknown.d.ts`/Promise#",
		ReferenceRanges: []int32{16, 87, 16, 94, 24, 94, 24, 101},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es2015.symbol.wellknown.d.ts`/String#split().",
		ReferenceRanges: []int32{43, 30, 43, 35},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es2018.promise.d.ts`/Promise#",
		ReferenceRanges: []int32{16, 87, 16, 94, 24, 94, 24, 101},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es2022.error.d.ts`/Error#",
		ReferenceRanges: []int32{12, 12, 12, 17, 34, 33, 34, 38, 34, 43, 34, 48, 35, 29, 35, 34},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es5.d.ts`/Array#join().",
		ReferenceRanges: []int32{35, 70, 35, 74},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es5.d.ts`/Array#length.",
		ReferenceRanges: []int32{28, 30, 28, 36},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es5.d.ts`/Array#map().",
		ReferenceRanges: []int32{35, 42, 35, 45},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es5.d.ts`/Error#",
		ReferenceRanges: []int32{12, 12, 12, 17, 34, 33, 34, 38, 34, 43, 34, 48, 35, 29, 35, 34},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es5.d.ts`/Error#message.",
		ReferenceRanges: []int32{35, 61, 35, 68},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es5.d.ts`/Error.",
		ReferenceRanges: []int32{12, 12, 12, 17, 34, 33, 34, 38, 34, 43, 34, 48, 35, 29, 35, 34},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es5.d.ts`/Object#",
		ReferenceRanges: []int32{35, 11, 35, 17},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es5.d.ts`/Object.",
		ReferenceRanges: []int32{35, 11, 35, 17},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es5.d.ts`/Promise#",
		ReferenceRanges: []int32{16, 87, 16, 94, 24, 94, 24, 101},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es5.d.ts`/String#split().",
		ReferenceRanges: []int32{43, 30, 43, 35},
	},
	{
		SymbolName:      "scip-typescript npm typescript 4.9.3 lib/`lib.es5.d.ts`/parseInt().",
		ReferenceRanges: []int32{43, 11, 43, 19},
	},
}
