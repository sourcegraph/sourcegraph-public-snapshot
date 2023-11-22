package apitest

import (
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/notebooks"
)

func BlockToAPIResponse(block notebooks.NotebookBlock) NotebookBlock {
	switch block.Type {
	case notebooks.NotebookMarkdownBlockType:
		return NotebookBlock{Typename: "MarkdownBlock", ID: block.ID, MarkdownInput: block.MarkdownInput.Text}
	case notebooks.NotebookQueryBlockType:
		return NotebookBlock{Typename: "QueryBlock", ID: block.ID, QueryInput: block.QueryInput.Text}
	case notebooks.NotebookFileBlockType:
		return NotebookBlock{Typename: "FileBlock", ID: block.ID, FileInput: FileInput{
			RepositoryName: block.FileInput.RepositoryName,
			FilePath:       block.FileInput.FilePath,
			Revision:       block.FileInput.Revision,
			LineRange:      &LineRange{StartLine: block.FileInput.LineRange.StartLine, EndLine: block.FileInput.LineRange.EndLine},
		}}
	case notebooks.NotebookSymbolBlockType:
		return NotebookBlock{Typename: "SymbolBlock", ID: block.ID, SymbolInput: SymbolInput{
			RepositoryName:      block.SymbolInput.RepositoryName,
			FilePath:            block.SymbolInput.FilePath,
			Revision:            block.SymbolInput.Revision,
			LineContext:         block.SymbolInput.LineContext,
			SymbolName:          block.SymbolInput.SymbolName,
			SymbolContainerName: block.SymbolInput.SymbolContainerName,
			SymbolKind:          block.SymbolInput.SymbolKind,
		}}
	}
	panic("unknown block type")
}

func NotebookToAPIResponse(notebook *notebooks.Notebook, id graphql.ID, creatorUsername string, updaterUsername string, viewerCanManage bool) Notebook {
	blocks := make([]NotebookBlock, 0, len(notebook.Blocks))
	for _, block := range notebook.Blocks {
		blocks = append(blocks, BlockToAPIResponse(block))
	}
	return Notebook{
		ID:              string(id),
		Title:           notebook.Title,
		Creator:         NotebookUser{Username: creatorUsername},
		Updater:         NotebookUser{Username: updaterUsername},
		CreatedAt:       notebook.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       notebook.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Public:          notebook.Public,
		ViewerCanManage: viewerCanManage,
		Blocks:          blocks,
	}
}

func BlockToAPIInput(block notebooks.NotebookBlock) graphqlbackend.CreateNotebookBlockInputArgs {
	switch block.Type {
	case notebooks.NotebookMarkdownBlockType:
		return graphqlbackend.CreateNotebookBlockInputArgs{ID: block.ID, Type: graphqlbackend.NotebookMarkdownBlockType, MarkdownInput: &block.MarkdownInput.Text}
	case notebooks.NotebookQueryBlockType:
		return graphqlbackend.CreateNotebookBlockInputArgs{ID: block.ID, Type: graphqlbackend.NotebookQueryBlockType, QueryInput: &block.QueryInput.Text}
	case notebooks.NotebookFileBlockType:
		return graphqlbackend.CreateNotebookBlockInputArgs{ID: block.ID, Type: graphqlbackend.NotebookFileBlockType, FileInput: &graphqlbackend.CreateFileBlockInput{
			RepositoryName: block.FileInput.RepositoryName,
			FilePath:       block.FileInput.FilePath,
			Revision:       block.FileInput.Revision,
			LineRange:      &graphqlbackend.CreateFileBlockLineRangeInput{StartLine: block.FileInput.LineRange.StartLine, EndLine: block.FileInput.LineRange.EndLine},
		}}
	case notebooks.NotebookSymbolBlockType:
		return graphqlbackend.CreateNotebookBlockInputArgs{ID: block.ID, Type: graphqlbackend.NotebookSymbolBlockType, SymbolInput: &graphqlbackend.CreateSymbolBlockInput{
			RepositoryName:      block.SymbolInput.RepositoryName,
			FilePath:            block.SymbolInput.FilePath,
			Revision:            block.SymbolInput.Revision,
			LineContext:         block.SymbolInput.LineContext,
			SymbolName:          block.SymbolInput.SymbolName,
			SymbolContainerName: block.SymbolInput.SymbolContainerName,
			SymbolKind:          block.SymbolInput.SymbolKind,
		}}
	}
	panic("unknown block type")
}

func marshalNamespaceID(notebook *notebooks.Notebook) graphql.ID {
	if notebook.NamespaceUserID != 0 {
		return graphqlbackend.MarshalUserID(notebook.NamespaceUserID)
	} else {
		return graphqlbackend.MarshalOrgID(notebook.NamespaceOrgID)
	}
}

func NotebookToAPIInput(notebook *notebooks.Notebook) graphqlbackend.NotebookInputArgs {
	blocks := make([]graphqlbackend.CreateNotebookBlockInputArgs, 0, len(notebook.Blocks))
	for _, block := range notebook.Blocks {
		blocks = append(blocks, BlockToAPIInput(block))
	}
	return graphqlbackend.NotebookInputArgs{
		Title:     notebook.Title,
		Public:    notebook.Public,
		Blocks:    blocks,
		Namespace: marshalNamespaceID(notebook),
	}
}
