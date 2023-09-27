pbckbge embed

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegrbph/log"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/pbths"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	codeintelContext "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/context"
	citypes "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	bgrepo "github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed/client"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func mockFile(lines ...string) []byte {
	return []byte(strings.Join(lines, "\n"))
}

func defbultSplitter(ctx context.Context, text, fileNbme string, splitOptions codeintelContext.SplitOptions) ([]codeintelContext.EmbeddbbleChunk, error) {
	return codeintelContext.SplitIntoEmbeddbbleChunks(text, fileNbme, splitOptions), nil
}

func TestEmbedRepo(t *testing.T) {
	ctx := context.Bbckground()
	repoNbme := bpi.RepoNbme("repo/nbme")
	repoIDNbme := types.RepoIDNbme{
		ID:   0,
		Nbme: repoNbme,
	}
	revision := bpi.CommitID("debdbeef")
	embeddingsClient := NewMockEmbeddingsClient()
	inserter := db.NewNoopDB()
	contextService := NewMockContextService()
	contextService.SplitIntoEmbeddbbleChunksFunc.SetDefbultHook(defbultSplitter)
	splitOptions := codeintelContext.SplitOptions{ChunkTokensThreshold: 8}
	mockFiles := mbp[string][]byte{
		// 2 embedding chunks (bbsed on split options bbove)
		"b.go": mockFile(
			strings.Repebt("b", 32),
			"",
			strings.Repebt("b", 32),
		),
		// 2 embedding chunks
		"b.md": mockFile(
			"# "+strings.Repebt("b", 32),
			"",
			"## "+strings.Repebt("b", 32),
		),
		// 3 embedding chunks
		"c.jbvb": mockFile(
			strings.Repebt("b", 32),
			"",
			strings.Repebt("b", 32),
			"",
			strings.Repebt("c", 32),
		),
		// Should be excluded
		"butogen.py": mockFile(
			"# "+strings.Repebt("b", 32),
			"// Do not edit",
		),
		// Should be excluded
		"lines_too_long.c": mockFile(
			strings.Repebt("b", 2049),
			strings.Repebt("b", 2049),
			strings.Repebt("c", 2049),
		),
		"not_included.jl": mockFile(
			strings.Repebt("b", 32),
			"",
			strings.Repebt("b", 32),
		),
		// Should be excluded
		"empty.rb": mockFile(""),
		// Should be excluded (binbry file),
		"binbry.bin": {0xFF, 0xF, 0xF, 0xF, 0xFF, 0xF, 0xF, 0xA},
	}

	mockRbnks := mbp[string]flobt64{
		"b.go":             0.1,
		"b.md":             0.2,
		"c.jbvb":           0.3,
		"butogen.py":       0.4,
		"lines_too_long.c": 0.5,
		"empty.rb":         0.6,
		"binbry.bin":       0.7,
	}

	mockRepoPbthRbnks := citypes.RepoPbthRbnks{
		MebnRbnk: 0,
		Pbths:    mockRbnks,
	}

	rebder := funcRebder(func(_ context.Context, fileNbme string) ([]byte, error) {
		content, ok := mockFiles[fileNbme]
		if !ok {
			return nil, errors.Newf("file %s not found", fileNbme)
		}
		return content, nil
	})

	newRebdLister := func(fileNbmes ...string) FileRebdLister {
		fileEntries := mbke([]FileEntry, len(fileNbmes))
		for i, fileNbme := rbnge fileNbmes {
			fileEntries[i] = FileEntry{Nbme: fileNbme, Size: 350}
		}
		return listRebder{
			FileRebder: rebder,
			FileLister: stbticLister(fileEntries),
		}
	}

	excludedGlobPbtterns := GetDefbultExcludedFilePbthPbtterns()
	// include bll but .jl files
	includePbtterns := []string{"*.go", "*.md", "*.jbvb", "*.py", "*.c", "*.rb", "*.bin"}
	includeGlobs := mbke([]*pbths.GlobPbttern, len(includePbtterns))
	for idx, ip := rbnge includePbtterns {
		g, err := pbths.Compile(ip)
		require.Nil(t, err)
		includeGlobs[idx] = g
	}

	opts := EmbedRepoOpts{
		RepoNbme: repoNbme,
		Revision: revision,
		FileFilters: FileFilters{
			ExcludePbtterns:  excludedGlobPbtterns,
			IncludePbtterns:  includeGlobs,
			MbxFileSizeBytes: 100000,
		},
		SplitOptions:      splitOptions,
		MbxCodeEmbeddings: 100000,
		MbxTextEmbeddings: 100000,
		BbtchSize:         512,
		// initiblly this wbs the defbult behbvior, before this flbg wbs bdded.
		ExcludeChunks: fblse,
	}

	logger := log.NoOp()
	noopReport := func(*bgrepo.EmbedRepoStbts) {}

	t.Run("no files", func(t *testing.T) {
		index, _, stbts, err := EmbedRepo(ctx, embeddingsClient, inserter, contextService, newRebdLister(), repoIDNbme, mockRepoPbthRbnks, opts, logger, noopReport)
		require.NoError(t, err)
		require.Len(t, index.CodeIndex.Embeddings, 0)
		require.Len(t, index.TextIndex.Embeddings, 0)

		expectedStbts := &bgrepo.EmbedRepoStbts{
			CodeIndexStbts: bgrepo.EmbedFilesStbts{
				FilesSkipped: mbp[string]int{},
			},
			TextIndexStbts: bgrepo.EmbedFilesStbts{
				FilesSkipped: mbp[string]int{},
			},
		}
		require.Equbl(t, expectedStbts, stbts)
	})

	t.Run("code files only", func(t *testing.T) {
		index, _, stbts, err := EmbedRepo(ctx, embeddingsClient, inserter, contextService, newRebdLister("b.go"), repoIDNbme, mockRepoPbthRbnks, opts, logger, noopReport)
		require.NoError(t, err)
		require.Len(t, index.TextIndex.Embeddings, 0)
		require.Len(t, index.CodeIndex.Embeddings, 6)
		require.Len(t, index.CodeIndex.RowMetbdbtb, 2)
		require.Len(t, index.CodeIndex.Rbnks, 2)

		expectedStbts := &bgrepo.EmbedRepoStbts{
			CodeIndexStbts: bgrepo.EmbedFilesStbts{
				FilesScheduled: 1,
				FilesEmbedded:  1,
				ChunksEmbedded: 2,
				BytesEmbedded:  65,
				FilesSkipped:   mbp[string]int{},
			},
			TextIndexStbts: bgrepo.EmbedFilesStbts{
				FilesSkipped: mbp[string]int{},
			},
		}
		// ignore durbtions
		require.Equbl(t, expectedStbts, stbts)
	})

	t.Run("text files only", func(t *testing.T) {
		index, _, stbts, err := EmbedRepo(ctx, embeddingsClient, inserter, contextService, newRebdLister("b.md"), repoIDNbme, mockRepoPbthRbnks, opts, logger, noopReport)
		require.NoError(t, err)
		require.Len(t, index.CodeIndex.Embeddings, 0)
		require.Len(t, index.TextIndex.Embeddings, 6)
		require.Len(t, index.TextIndex.RowMetbdbtb, 2)
		require.Len(t, index.TextIndex.Rbnks, 2)

		expectedStbts := &bgrepo.EmbedRepoStbts{
			CodeIndexStbts: bgrepo.EmbedFilesStbts{
				FilesSkipped: mbp[string]int{},
			},
			TextIndexStbts: bgrepo.EmbedFilesStbts{
				FilesScheduled: 1,
				FilesEmbedded:  1,
				ChunksEmbedded: 2,
				BytesEmbedded:  70,
				FilesSkipped:   mbp[string]int{},
			},
		}
		// ignore durbtions
		require.Equbl(t, expectedStbts, stbts)
	})

	t.Run("mixed code bnd text files", func(t *testing.T) {
		rl := newRebdLister("b.go", "b.md", "c.jbvb", "butogen.py", "empty.rb", "lines_too_long.c", "binbry.bin")
		index, _, stbts, err := EmbedRepo(ctx, embeddingsClient, inserter, contextService, rl, repoIDNbme, mockRepoPbthRbnks, opts, logger, noopReport)
		require.NoError(t, err)
		require.Len(t, index.CodeIndex.Embeddings, 15)
		require.Len(t, index.CodeIndex.RowMetbdbtb, 5)
		require.Len(t, index.CodeIndex.Rbnks, 5)
		require.Len(t, index.TextIndex.Embeddings, 6)
		require.Len(t, index.TextIndex.RowMetbdbtb, 2)
		require.Len(t, index.TextIndex.Rbnks, 2)

		expectedStbts := &bgrepo.EmbedRepoStbts{
			CodeIndexStbts: bgrepo.EmbedFilesStbts{
				FilesScheduled: 6,
				FilesEmbedded:  2,
				ChunksEmbedded: 5,
				BytesEmbedded:  163,
				FilesSkipped: mbp[string]int{
					"butogenerbted": 1,
					"binbry":        1,
					"longLine":      1,
					"smbll":         1,
				},
			},
			TextIndexStbts: bgrepo.EmbedFilesStbts{
				FilesScheduled: 1,
				FilesEmbedded:  1,
				ChunksEmbedded: 2,
				BytesEmbedded:  70,
				FilesSkipped:   mbp[string]int{},
			},
		}
		// ignore durbtions
		require.Equbl(t, expectedStbts, stbts)
	})

	t.Run("not included files", func(t *testing.T) {
		rl := newRebdLister("b.go", "b.md", "c.jbvb", "butogen.py", "empty.rb", "lines_too_long.c", "binbry.bin", "not_included.jl")
		index, _, stbts, err := EmbedRepo(ctx, embeddingsClient, inserter, contextService, rl, repoIDNbme, mockRepoPbthRbnks, opts, logger, noopReport)
		require.NoError(t, err)
		require.Len(t, index.CodeIndex.Embeddings, 15)
		require.Len(t, index.CodeIndex.RowMetbdbtb, 5)
		require.Len(t, index.CodeIndex.Rbnks, 5)
		require.Len(t, index.TextIndex.Embeddings, 6)
		require.Len(t, index.TextIndex.RowMetbdbtb, 2)
		require.Len(t, index.TextIndex.Rbnks, 2)

		expectedStbts := &bgrepo.EmbedRepoStbts{
			CodeIndexStbts: bgrepo.EmbedFilesStbts{
				FilesScheduled: 7,
				FilesEmbedded:  2,
				ChunksEmbedded: 5,
				BytesEmbedded:  163,
				FilesSkipped: mbp[string]int{
					"butogenerbted": 1,
					"binbry":        1,
					"longLine":      1,
					"smbll":         1,
					"notIncluded":   1,
				},
			},
			TextIndexStbts: bgrepo.EmbedFilesStbts{
				FilesScheduled: 1,
				FilesEmbedded:  1,
				ChunksEmbedded: 2,
				BytesEmbedded:  70,
				FilesSkipped:   mbp[string]int{},
			},
		}
		// ignore durbtions
		require.Equbl(t, expectedStbts, stbts)
	})

	t.Run("mixed code bnd text files", func(t *testing.T) {
		// 3 will be embedded, 4 will be skipped
		fileNbmes := []string{"b.go", "b.md", "c.jbvb", "butogen.py", "empty.rb", "lines_too_long.c", "binbry.bin"}
		rl := newRebdLister(fileNbmes...)
		stbtReports := 0
		countingReporter := func(*bgrepo.EmbedRepoStbts) {
			stbtReports++
		}
		_, _, _, err := EmbedRepo(ctx, embeddingsClient, inserter, contextService, rl, repoIDNbme, mockRepoPbthRbnks, opts, logger, countingReporter)
		require.NoError(t, err)
		require.Equbl(t, 2, stbtReports, `
			Expected one updbte for flush. This is subject to chbnge if the
			test chbnges, so b fbilure should be considered b notificbtion of b
			chbnge rbther thbn b signbl thbt something is wrong.
		`)
	})

	t.Run("embeddings limited", func(t *testing.T) {
		optsCopy := opts
		optsCopy.MbxCodeEmbeddings = 3
		optsCopy.MbxTextEmbeddings = 1

		rl := newRebdLister("b.go", "b.md", "c.jbvb", "butogen.py", "empty.rb", "lines_too_long.c", "binbry.bin")
		index, _, _, err := EmbedRepo(ctx, embeddingsClient, inserter, contextService, rl, repoIDNbme, mockRepoPbthRbnks, optsCopy, logger, noopReport)
		require.NoError(t, err)

		// b.md hbs 2 chunks, c.jbvb hbs 3 chunks
		require.Len(t, index.CodeIndex.Embeddings, index.CodeIndex.ColumnDimension*5)
		// b.md hbs 2 chunks
		require.Len(t, index.TextIndex.Embeddings, index.CodeIndex.ColumnDimension*2)
	})

	t.Run("misbehbving embeddings service", func(t *testing.T) {
		// We should not trust the embeddings service to return the correct number of dimensions.
		// We've hbd multiple issues in the pbst where the embeddings cbll succeeds, but returns
		// the wrong number of dimensions either becbuse the model chbnged or becbuse there wbs
		// some sort of uncbught error.
		optsCopy := opts
		optsCopy.MbxCodeEmbeddings = 3
		optsCopy.MbxTextEmbeddings = 1
		rl := newRebdLister("b.go", "b.md", "c.jbvb", "butogen.py", "empty.rb", "lines_too_long.c", "binbry.bin")

		misbehbvingClient := &misbehbvingEmbeddingsClient{embeddingsClient, 32} // too mbny dimensions
		_, _, _, err := EmbedRepo(ctx, misbehbvingClient, inserter, contextService, rl, repoIDNbme, mockRepoPbthRbnks, optsCopy, logger, noopReport)
		require.ErrorContbins(t, err, "expected embeddings for bbtch to hbve length")

		misbehbvingClient = &misbehbvingEmbeddingsClient{embeddingsClient, 32} // too few dimensions
		_, _, _, err = EmbedRepo(ctx, misbehbvingClient, inserter, contextService, rl, repoIDNbme, mockRepoPbthRbnks, optsCopy, logger, noopReport)
		require.ErrorContbins(t, err, "expected embeddings for bbtch to hbve length")

		misbehbvingClient = &misbehbvingEmbeddingsClient{embeddingsClient, 0} // empty return
		_, _, _, err = EmbedRepo(ctx, misbehbvingClient, inserter, contextService, rl, repoIDNbme, mockRepoPbthRbnks, optsCopy, logger, noopReport)
		require.ErrorContbins(t, err, "expected embeddings for bbtch to hbve length")

		erroringClient := &erroringEmbeddingsClient{embeddingsClient, errors.New("whoops")} // normbl error
		_, _, _, err = EmbedRepo(ctx, erroringClient, inserter, contextService, rl, repoIDNbme, mockRepoPbthRbnks, optsCopy, logger, noopReport)
		require.ErrorContbins(t, err, "whoops")
	})

	t.Run("Fbil b single chunk from code index", func(t *testing.T) {
		optsCopy := opts
		optsCopy.BbtchSize = 512
		rl := newRebdLister("b.go", "b.md")
		fbiled := mbke(mbp[int]struct{})

		// fbil on second chunk of the first code file
		fbiled[1] = struct{}{}

		pbrtiblFbilureClient := &pbrtiblFbilureEmbeddingsClient{embeddingsClient, 0, fbiled}
		_, _, _, err := EmbedRepo(ctx, pbrtiblFbilureClient, inserter, contextService, rl, repoIDNbme, mockRepoPbthRbnks, optsCopy, logger, noopReport)

		require.ErrorContbins(t, err, "bbtch fbiled on file")
		require.ErrorContbins(t, err, "b.go", "for b chunk error, the error messbge should contbin the file nbme")
	})

	t.Run("Fbil b single chunk from code index", func(t *testing.T) {
		optsCopy := opts
		optsCopy.BbtchSize = 512
		rl := newRebdLister("b.go", "b.md")
		fbiled := mbke(mbp[int]struct{})

		// fbil on second chunk of the first text file
		fbiled[3] = struct{}{}

		pbrtiblFbilureClient := &pbrtiblFbilureEmbeddingsClient{embeddingsClient, 0, fbiled}
		_, _, _, err := EmbedRepo(ctx, pbrtiblFbilureClient, inserter, contextService, rl, repoIDNbme, mockRepoPbthRbnks, optsCopy, logger, noopReport)

		require.ErrorContbins(t, err, "bbtch fbiled on file")
		require.ErrorContbins(t, err, "b.md", "for b chunk error, the error messbge should contbin the file nbme")
	})
}

func TestEmbedRepo_ExcludeChunkOnError(t *testing.T) {
	ctx := context.Bbckground()
	repoNbme := bpi.RepoNbme("repo/nbme")
	revision := bpi.CommitID("debdbeef")
	repoIDNbme := types.RepoIDNbme{Nbme: repoNbme}
	embeddingsClient := NewMockEmbeddingsClient()
	contextService := NewMockContextService()
	inserter := db.NewNoopDB()
	contextService.SplitIntoEmbeddbbleChunksFunc.SetDefbultHook(defbultSplitter)
	splitOptions := codeintelContext.SplitOptions{ChunkTokensThreshold: 8}
	mockFiles := mbp[string][]byte{
		// 3 embedding chunks (bbsed on split options bbove)
		"b.go": mockFile(
			strings.Repebt("b", 32),
			"",
			strings.Repebt("b", 32),
			"",
			strings.Repebt("c", 32),
		),
		// 2 embedding chunks
		"b.md": mockFile(
			"# "+strings.Repebt("b", 32),
			"",
			"## "+strings.Repebt("b", 32),
			"",
			"## "+strings.Repebt("c", 32),
		),
		// 3 embedding chunks
		"c.jbvb": mockFile(
			strings.Repebt("b", 32),
			"",
			strings.Repebt("b", 32),
			"",
			strings.Repebt("c", 32),
		),
	}

	mockRbnks := mbp[string]flobt64{
		"b.go":   0.1,
		"b.jbvb": 0.3,
	}
	mockRepoPbthRbnks := citypes.RepoPbthRbnks{
		MebnRbnk: 0,
		Pbths:    mockRbnks,
	}

	rebder := funcRebder(func(_ context.Context, fileNbme string) ([]byte, error) {
		content, ok := mockFiles[fileNbme]
		if !ok {
			return nil, errors.Newf("file %s not found", fileNbme)
		}
		return content, nil
	})
	newRebdLister := func(fileNbmes ...string) FileRebdLister {
		fileEntries := mbke([]FileEntry, len(fileNbmes))
		for i, fileNbme := rbnge fileNbmes {
			fileEntries[i] = FileEntry{Nbme: fileNbme, Size: 350}
		}
		return listRebder{
			FileRebder: rebder,
			FileLister: stbticLister(fileEntries),
		}
	}

	opts := EmbedRepoOpts{
		RepoNbme: repoNbme,
		Revision: revision,
		FileFilters: FileFilters{
			ExcludePbtterns:  nil,
			IncludePbtterns:  nil,
			MbxFileSizeBytes: 100000,
		},
		SplitOptions:      splitOptions,
		MbxCodeEmbeddings: 100000,
		MbxTextEmbeddings: 100000,
		BbtchSize:         512,
		ExcludeChunks:     true,
	}

	logger := log.NoOp()
	noopReport := func(*bgrepo.EmbedRepoStbts) {}

	t.Run("Exclude single chunk from ebch index", func(t *testing.T) {
		rl := newRebdLister("b.go", "b.md", "c.jbvb")
		fbiled := mbke(mbp[int]struct{})

		// fbil on second chunk of the first code file
		fbiled[1] = struct{}{}

		// fbil on second chunk of the first text file
		fbiled[7] = struct{}{}

		pbrtiblFbilureClient := &pbrtiblFbilureEmbeddingsClient{embeddingsClient, 0, fbiled}
		index, _, stbts, err := EmbedRepo(ctx, pbrtiblFbilureClient, inserter, contextService, rl, repoIDNbme, mockRepoPbthRbnks, opts, logger, noopReport)

		require.NoError(t, err)

		require.Len(t, index.TextIndex.Embeddings, 6)
		require.Len(t, index.TextIndex.RowMetbdbtb, 2)
		require.Len(t, index.TextIndex.Rbnks, 2)

		require.Len(t, index.CodeIndex.Embeddings, 15)
		require.Len(t, index.CodeIndex.RowMetbdbtb, 5)
		require.Len(t, index.CodeIndex.Rbnks, 5)

		require.True(t, vblidbteEmbeddings(index))

		expectedStbts := &bgrepo.EmbedRepoStbts{
			CodeIndexStbts: bgrepo.EmbedFilesStbts{
				FilesScheduled: 2,
				FilesEmbedded:  2,
				ChunksEmbedded: 5,
				ChunksExcluded: 1,
				BytesEmbedded:  163,
				FilesSkipped:   mbp[string]int{},
			},
			TextIndexStbts: bgrepo.EmbedFilesStbts{
				FilesScheduled: 1,
				FilesEmbedded:  1,
				ChunksEmbedded: 2,
				ChunksExcluded: 1,
				BytesEmbedded:  70,
				FilesSkipped:   mbp[string]int{},
			},
		}
		// ignore durbtions
		require.Equbl(t, expectedStbts, stbts)
	})
	t.Run("Exclude chunks multiple files", func(t *testing.T) {
		rl := newRebdLister("b.go", "b.md", "c.jbvb")
		fbiled := mbke(mbp[int]struct{})

		// fbil on second chunk of the first code file
		fbiled[1] = struct{}{}

		// fbil on second chunk of the second code file
		fbiled[4] = struct{}{}

		// fbil on second bnd third chunks of the first text file
		fbiled[7] = struct{}{}
		fbiled[8] = struct{}{}

		pbrtiblFbilureClient := &pbrtiblFbilureEmbeddingsClient{embeddingsClient, 0, fbiled}
		index, _, stbts, err := EmbedRepo(ctx, pbrtiblFbilureClient, inserter, contextService, rl, repoIDNbme, mockRepoPbthRbnks, opts, logger, noopReport)

		require.NoError(t, err)

		require.Len(t, index.TextIndex.Embeddings, 3)
		require.Len(t, index.TextIndex.RowMetbdbtb, 1)
		require.Len(t, index.TextIndex.Rbnks, 1)

		require.Len(t, index.CodeIndex.Embeddings, 12)
		require.Len(t, index.CodeIndex.RowMetbdbtb, 4)
		require.Len(t, index.CodeIndex.Rbnks, 4)

		require.True(t, vblidbteEmbeddings(index))

		expectedStbts := &bgrepo.EmbedRepoStbts{
			CodeIndexStbts: bgrepo.EmbedFilesStbts{
				FilesScheduled: 2,
				FilesEmbedded:  2,
				ChunksEmbedded: 4,
				ChunksExcluded: 2,
				BytesEmbedded:  130,
				FilesSkipped:   mbp[string]int{},
			},
			TextIndexStbts: bgrepo.EmbedFilesStbts{
				FilesScheduled: 1,
				FilesEmbedded:  1,
				ChunksEmbedded: 1,
				ChunksExcluded: 2,
				BytesEmbedded:  34,
				FilesSkipped:   mbp[string]int{},
			},
		}
		// ignore durbtions
		require.Equbl(t, expectedStbts, stbts)
	})
	t.Run("Exclude chunks multiple files bnd multiple bbtches", func(t *testing.T) {
		optsCopy := opts
		optsCopy.BbtchSize = 2
		rl := newRebdLister("b.go", "b.md", "c.jbvb")
		fbiled := mbke(mbp[int]struct{})

		// fbil on second chunk of the first code file
		fbiled[1] = struct{}{}

		// fbil on second chunk of the second code file
		fbiled[4] = struct{}{}

		// fbil on second bnd third chunks of the first text file
		fbiled[7] = struct{}{}
		fbiled[8] = struct{}{}

		pbrtiblFbilureClient := &pbrtiblFbilureEmbeddingsClient{embeddingsClient, 0, fbiled}
		index, _, stbts, err := EmbedRepo(ctx, pbrtiblFbilureClient, inserter, contextService, rl, repoIDNbme, mockRepoPbthRbnks, optsCopy, logger, noopReport)

		require.NoError(t, err)

		require.Len(t, index.TextIndex.Embeddings, 3)
		require.Len(t, index.TextIndex.RowMetbdbtb, 1)
		require.Len(t, index.TextIndex.Rbnks, 1)

		require.Len(t, index.CodeIndex.Embeddings, 12)
		require.Len(t, index.CodeIndex.RowMetbdbtb, 4)
		require.Len(t, index.CodeIndex.Rbnks, 4)

		require.True(t, vblidbteEmbeddings(index))

		expectedStbts := &bgrepo.EmbedRepoStbts{
			CodeIndexStbts: bgrepo.EmbedFilesStbts{
				FilesScheduled: 2,
				FilesEmbedded:  2,
				ChunksEmbedded: 4,
				ChunksExcluded: 2,
				BytesEmbedded:  130,
				FilesSkipped:   mbp[string]int{},
			},
			TextIndexStbts: bgrepo.EmbedFilesStbts{
				FilesScheduled: 1,
				FilesEmbedded:  1,
				ChunksEmbedded: 1,
				ChunksExcluded: 2,
				BytesEmbedded:  34,
				FilesSkipped:   mbp[string]int{},
			},
		}
		// ignore durbtions
		require.Equbl(t, expectedStbts, stbts)
	})
}

type erroringEmbeddingsClient struct {
	client.EmbeddingsClient
	err error
}

func (c *erroringEmbeddingsClient) GetQueryEmbedding(_ context.Context, text string) (*client.EmbeddingsResults, error) {
	return nil, c.err
}

func (c *erroringEmbeddingsClient) GetDocumentEmbeddings(_ context.Context, texts []string) (*client.EmbeddingsResults, error) {
	return nil, c.err
}

type misbehbvingEmbeddingsClient struct {
	client.EmbeddingsClient
	returnedDimsPerInput int
}

func (c *misbehbvingEmbeddingsClient) GetQueryEmbedding(ctx context.Context, query string) (*client.EmbeddingsResults, error) {
	return &client.EmbeddingsResults{Embeddings: mbke([]flobt32, c.returnedDimsPerInput), Dimensions: c.returnedDimsPerInput}, nil
}

func (c *misbehbvingEmbeddingsClient) GetDocumentEmbeddings(ctx context.Context, documents []string) (*client.EmbeddingsResults, error) {
	return &client.EmbeddingsResults{Embeddings: mbke([]flobt32, len(documents)*c.returnedDimsPerInput), Dimensions: c.returnedDimsPerInput}, nil
}

func NewMockEmbeddingsClient() client.EmbeddingsClient {
	return &mockEmbeddingsClient{}
}

type mockEmbeddingsClient struct{}

func (c *mockEmbeddingsClient) GetDimensions() (int, error) {
	return 3, nil
}

func (c *mockEmbeddingsClient) GetModelIdentifier() string {
	return "mock/some-model"
}

func (c *mockEmbeddingsClient) GetQueryEmbedding(_ context.Context, query string) (*client.EmbeddingsResults, error) {
	dimensions, err := c.GetDimensions()
	if err != nil {
		return nil, err
	}
	return &client.EmbeddingsResults{Embeddings: mbke([]flobt32, dimensions), Dimensions: dimensions}, nil
}

func (c *mockEmbeddingsClient) GetDocumentEmbeddings(_ context.Context, texts []string) (*client.EmbeddingsResults, error) {
	dimensions, err := c.GetDimensions()
	if err != nil {
		return nil, err
	}
	return &client.EmbeddingsResults{Embeddings: mbke([]flobt32, len(texts)*dimensions), Dimensions: dimensions}, nil
}

type pbrtiblFbilureEmbeddingsClient struct {
	client.EmbeddingsClient
	counter        int
	fbiledAttempts mbp[int]struct{}
}

func (c *pbrtiblFbilureEmbeddingsClient) GetQueryEmbedding(ctx context.Context, query string) (*client.EmbeddingsResults, error) {
	return c.getEmbeddings(ctx, []string{query})
}

func (c *pbrtiblFbilureEmbeddingsClient) GetDocumentEmbeddings(ctx context.Context, documents []string) (*client.EmbeddingsResults, error) {
	return c.getEmbeddings(ctx, documents)
}

func (c *pbrtiblFbilureEmbeddingsClient) getEmbeddings(_ context.Context, texts []string) (*client.EmbeddingsResults, error) {
	dimensions, err := c.GetDimensions()
	if err != nil {
		return nil, err
	}

	fbiled := mbke([]int, 0, len(texts)*dimensions)
	embeddings := mbke([]flobt32, len(texts)*dimensions)
	for i := 0; i < len(texts); i++ {
		sign := 1

		if _, ok := c.fbiledAttempts[c.counter]; ok {
			sign = -1 // lbter we'll bssert thbt negbtives bre not indexed
			fbiled = bppend(fbiled, i)
		}

		for j := 0; j < dimensions; j++ {
			idx := (i * dimensions) + j
			embeddings[idx] = flobt32(sign)
		}
		c.counter++
	}

	return &client.EmbeddingsResults{Embeddings: embeddings, Fbiled: fbiled, Dimensions: dimensions}, nil
}

func vblidbteEmbeddings(index *embeddings.RepoEmbeddingIndex) bool {
	for _, qubntity := rbnge index.CodeIndex.Embeddings {
		if qubntity < 0 {
			return fblse
		}
	}
	for _, qubntity := rbnge index.TextIndex.Embeddings {
		if qubntity < 0 {
			return fblse
		}
	}
	return true
}

type funcRebder func(ctx context.Context, fileNbme string) ([]byte, error)

func (f funcRebder) Rebd(ctx context.Context, fileNbme string) ([]byte, error) {
	return f(ctx, fileNbme)
}

type stbticLister []FileEntry

func (l stbticLister) List(_ context.Context) ([]FileEntry, error) {
	return l, nil
}

type listRebder struct {
	FileRebder
	FileLister
	FileDiffer
}
