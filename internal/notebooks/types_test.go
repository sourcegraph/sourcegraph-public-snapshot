package notebooks

import (
	"encoding/json"
	"testing"

	"github.com/hexops/autogold/v2"
)

func TestNotebookBlockMarshalling(t *testing.T) {
	queryBlockInput := NotebookQueryBlockInput{Text: "repo:a b"}
	markdownBlockInput := NotebookMarkdownBlockInput{Text: "# Title"}
	revision := "main"
	fileBlockInput := NotebookFileBlockInput{RepositoryName: "sourcegraph/sourcegraph", FilePath: "a/b.ts", Revision: &revision, LineRange: &LineRange{1, 10}}

	tests := []struct {
		block NotebookBlock
		want  autogold.Value
	}{
		{
			block: NotebookBlock{ID: "id1", Type: NotebookQueryBlockType, QueryInput: &queryBlockInput},
			want:  autogold.Expect(`{"id":"id1","type":"query","queryInput":{"text":"repo:a b"}}`),
		},
		{
			block: NotebookBlock{ID: "id1", Type: NotebookMarkdownBlockType, MarkdownInput: &markdownBlockInput},
			want:  autogold.Expect(`{"id":"id1","type":"md","markdownInput":{"text":"# Title"}}`),
		},
		{
			block: NotebookBlock{ID: "id1", Type: NotebookFileBlockType, FileInput: &fileBlockInput},
			want:  autogold.Expect(`{"id":"id1","type":"file","fileInput":{"repositoryName":"sourcegraph/sourcegraph","filePath":"a/b.ts","revision":"main","lineRange":{"startLine":1,"endLine":10}}}`),
		},
	}

	for _, tt := range tests {
		got, err := json.Marshal(tt.block)
		if err != nil {
			t.Fatal(err)
		}
		tt.want.Equal(t, string(got))
	}
}

func TestNotebookBlockUnmarshalling(t *testing.T) {
	queryBlockInput := NotebookQueryBlockInput{Text: "repo:a b"}
	markdownBlockInput := NotebookMarkdownBlockInput{Text: "# Title"}
	revision := "main"
	fileBlockInput := NotebookFileBlockInput{RepositoryName: "sourcegraph/sourcegraph", FilePath: "a/b.ts", Revision: &revision, LineRange: &LineRange{1, 10}}

	tests := []struct {
		json string
		want autogold.Value
	}{
		{
			json: `{"id":"id1","type":"query","queryInput":{"text":"repo:a b"}}`,
			want: autogold.Expect(NotebookBlock{ID: "id1", Type: NotebookQueryBlockType, QueryInput: &queryBlockInput}),
		},
		{
			json: `{"id":"id1","type":"md","markdownInput":{"text":"# Title"}}`,
			want: autogold.Expect(NotebookBlock{ID: "id1", Type: NotebookMarkdownBlockType, MarkdownInput: &markdownBlockInput}),
		},
		{
			json: `{"id":"id1","type":"file","fileInput":{"repositoryName":"sourcegraph/sourcegraph","filePath":"a/b.ts","revision":"main","lineRange":{"startLine":1,"endLine":10}}}`,
			want: autogold.Expect(NotebookBlock{ID: "id1", Type: NotebookFileBlockType, FileInput: &fileBlockInput}),
		},
	}

	for _, tt := range tests {
		var block NotebookBlock
		err := json.Unmarshal([]byte(tt.json), &block)
		if err != nil {
			t.Fatal(err)
		}
		tt.want.Equal(t, block)
	}
}
