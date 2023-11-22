package notebooks

import (
	"testing"
)

func TestNotebookBlocksValidation(t *testing.T) {
	tests := []struct {
		blocks  NotebookBlocks
		wantErr string
	}{
		{blocks: NotebookBlocks{
			{ID: "id1", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:a b"}},
			{ID: "id1", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:a b"}},
		}, wantErr: "duplicate block id found: id1"},
		{blocks: NotebookBlocks{
			{ID: "id1", Type: NotebookBlockType("t"), QueryInput: &NotebookQueryBlockInput{"repo:a b"}},
		}, wantErr: "invalid block type: t"},
		{blocks: NotebookBlocks{{ID: "id1", Type: NotebookQueryBlockType}}, wantErr: "invalid query block with id: id1"},
		{blocks: NotebookBlocks{{ID: "id1", Type: NotebookMarkdownBlockType}}, wantErr: "invalid markdown block with id: id1"},
		{blocks: NotebookBlocks{{ID: "id1", Type: NotebookFileBlockType}}, wantErr: "invalid file block with id: id1"},
		{blocks: NotebookBlocks{
			{ID: "id1", SymbolInput: &NotebookSymbolBlockInput{LineContext: -10}, Type: NotebookSymbolBlockType},
		}, wantErr: "symbol block line context cannot be negative, block id: id1"},
	}

	for _, tt := range tests {
		err := validateNotebookBlocks(tt.blocks)
		if err == nil {
			t.Fatal("expected error, got nil")
		} else if err.Error() != tt.wantErr {
			t.Fatalf("wanted '%s' error, got '%s'", tt.wantErr, err.Error())
		}
	}
}
