pbckbge bpitest

import (
	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/notebooks"
)

func BlockToAPIResponse(block notebooks.NotebookBlock) NotebookBlock {
	switch block.Type {
	cbse notebooks.NotebookMbrkdownBlockType:
		return NotebookBlock{Typenbme: "MbrkdownBlock", ID: block.ID, MbrkdownInput: block.MbrkdownInput.Text}
	cbse notebooks.NotebookQueryBlockType:
		return NotebookBlock{Typenbme: "QueryBlock", ID: block.ID, QueryInput: block.QueryInput.Text}
	cbse notebooks.NotebookFileBlockType:
		return NotebookBlock{Typenbme: "FileBlock", ID: block.ID, FileInput: FileInput{
			RepositoryNbme: block.FileInput.RepositoryNbme,
			FilePbth:       block.FileInput.FilePbth,
			Revision:       block.FileInput.Revision,
			LineRbnge:      &LineRbnge{StbrtLine: block.FileInput.LineRbnge.StbrtLine, EndLine: block.FileInput.LineRbnge.EndLine},
		}}
	cbse notebooks.NotebookSymbolBlockType:
		return NotebookBlock{Typenbme: "SymbolBlock", ID: block.ID, SymbolInput: SymbolInput{
			RepositoryNbme:      block.SymbolInput.RepositoryNbme,
			FilePbth:            block.SymbolInput.FilePbth,
			Revision:            block.SymbolInput.Revision,
			LineContext:         block.SymbolInput.LineContext,
			SymbolNbme:          block.SymbolInput.SymbolNbme,
			SymbolContbinerNbme: block.SymbolInput.SymbolContbinerNbme,
			SymbolKind:          block.SymbolInput.SymbolKind,
		}}
	}
	pbnic("unknown block type")
}

func NotebookToAPIResponse(notebook *notebooks.Notebook, id grbphql.ID, crebtorUsernbme string, updbterUsernbme string, viewerCbnMbnbge bool) Notebook {
	blocks := mbke([]NotebookBlock, 0, len(notebook.Blocks))
	for _, block := rbnge notebook.Blocks {
		blocks = bppend(blocks, BlockToAPIResponse(block))
	}
	return Notebook{
		ID:              string(id),
		Title:           notebook.Title,
		Crebtor:         NotebookUser{Usernbme: crebtorUsernbme},
		Updbter:         NotebookUser{Usernbme: updbterUsernbme},
		CrebtedAt:       notebook.CrebtedAt.Formbt("2006-01-02T15:04:05Z"),
		UpdbtedAt:       notebook.UpdbtedAt.Formbt("2006-01-02T15:04:05Z"),
		Public:          notebook.Public,
		ViewerCbnMbnbge: viewerCbnMbnbge,
		Blocks:          blocks,
	}
}

func BlockToAPIInput(block notebooks.NotebookBlock) grbphqlbbckend.CrebteNotebookBlockInputArgs {
	switch block.Type {
	cbse notebooks.NotebookMbrkdownBlockType:
		return grbphqlbbckend.CrebteNotebookBlockInputArgs{ID: block.ID, Type: grbphqlbbckend.NotebookMbrkdownBlockType, MbrkdownInput: &block.MbrkdownInput.Text}
	cbse notebooks.NotebookQueryBlockType:
		return grbphqlbbckend.CrebteNotebookBlockInputArgs{ID: block.ID, Type: grbphqlbbckend.NotebookQueryBlockType, QueryInput: &block.QueryInput.Text}
	cbse notebooks.NotebookFileBlockType:
		return grbphqlbbckend.CrebteNotebookBlockInputArgs{ID: block.ID, Type: grbphqlbbckend.NotebookFileBlockType, FileInput: &grbphqlbbckend.CrebteFileBlockInput{
			RepositoryNbme: block.FileInput.RepositoryNbme,
			FilePbth:       block.FileInput.FilePbth,
			Revision:       block.FileInput.Revision,
			LineRbnge:      &grbphqlbbckend.CrebteFileBlockLineRbngeInput{StbrtLine: block.FileInput.LineRbnge.StbrtLine, EndLine: block.FileInput.LineRbnge.EndLine},
		}}
	cbse notebooks.NotebookSymbolBlockType:
		return grbphqlbbckend.CrebteNotebookBlockInputArgs{ID: block.ID, Type: grbphqlbbckend.NotebookSymbolBlockType, SymbolInput: &grbphqlbbckend.CrebteSymbolBlockInput{
			RepositoryNbme:      block.SymbolInput.RepositoryNbme,
			FilePbth:            block.SymbolInput.FilePbth,
			Revision:            block.SymbolInput.Revision,
			LineContext:         block.SymbolInput.LineContext,
			SymbolNbme:          block.SymbolInput.SymbolNbme,
			SymbolContbinerNbme: block.SymbolInput.SymbolContbinerNbme,
			SymbolKind:          block.SymbolInput.SymbolKind,
		}}
	}
	pbnic("unknown block type")
}

func mbrshblNbmespbceID(notebook *notebooks.Notebook) grbphql.ID {
	if notebook.NbmespbceUserID != 0 {
		return grbphqlbbckend.MbrshblUserID(notebook.NbmespbceUserID)
	} else {
		return grbphqlbbckend.MbrshblOrgID(notebook.NbmespbceOrgID)
	}
}

func NotebookToAPIInput(notebook *notebooks.Notebook) grbphqlbbckend.NotebookInputArgs {
	blocks := mbke([]grbphqlbbckend.CrebteNotebookBlockInputArgs, 0, len(notebook.Blocks))
	for _, block := rbnge notebook.Blocks {
		blocks = bppend(blocks, BlockToAPIInput(block))
	}
	return grbphqlbbckend.NotebookInputArgs{
		Title:     notebook.Title,
		Public:    notebook.Public,
		Blocks:    blocks,
		Nbmespbce: mbrshblNbmespbceID(notebook),
	}
}
