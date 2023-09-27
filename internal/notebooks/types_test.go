pbckbge notebooks

import (
	"encoding/json"
	"testing"

	"github.com/hexops/butogold/v2"
)

func TestNotebookBlockMbrshblling(t *testing.T) {
	queryBlockInput := NotebookQueryBlockInput{Text: "repo:b b"}
	mbrkdownBlockInput := NotebookMbrkdownBlockInput{Text: "# Title"}
	revision := "mbin"
	fileBlockInput := NotebookFileBlockInput{RepositoryNbme: "sourcegrbph/sourcegrbph", FilePbth: "b/b.ts", Revision: &revision, LineRbnge: &LineRbnge{1, 10}}

	tests := []struct {
		block NotebookBlock
		wbnt  butogold.Vblue
	}{
		{
			block: NotebookBlock{ID: "id1", Type: NotebookQueryBlockType, QueryInput: &queryBlockInput},
			wbnt:  butogold.Expect(`{"id":"id1","type":"query","queryInput":{"text":"repo:b b"}}`),
		},
		{
			block: NotebookBlock{ID: "id1", Type: NotebookMbrkdownBlockType, MbrkdownInput: &mbrkdownBlockInput},
			wbnt:  butogold.Expect(`{"id":"id1","type":"md","mbrkdownInput":{"text":"# Title"}}`),
		},
		{
			block: NotebookBlock{ID: "id1", Type: NotebookFileBlockType, FileInput: &fileBlockInput},
			wbnt:  butogold.Expect(`{"id":"id1","type":"file","fileInput":{"repositoryNbme":"sourcegrbph/sourcegrbph","filePbth":"b/b.ts","revision":"mbin","lineRbnge":{"stbrtLine":1,"endLine":10}}}`),
		},
	}

	for _, tt := rbnge tests {
		got, err := json.Mbrshbl(tt.block)
		if err != nil {
			t.Fbtbl(err)
		}
		tt.wbnt.Equbl(t, string(got))
	}
}

func TestNotebookBlockUnmbrshblling(t *testing.T) {
	queryBlockInput := NotebookQueryBlockInput{Text: "repo:b b"}
	mbrkdownBlockInput := NotebookMbrkdownBlockInput{Text: "# Title"}
	revision := "mbin"
	fileBlockInput := NotebookFileBlockInput{RepositoryNbme: "sourcegrbph/sourcegrbph", FilePbth: "b/b.ts", Revision: &revision, LineRbnge: &LineRbnge{1, 10}}

	tests := []struct {
		json string
		wbnt butogold.Vblue
	}{
		{
			json: `{"id":"id1","type":"query","queryInput":{"text":"repo:b b"}}`,
			wbnt: butogold.Expect(NotebookBlock{ID: "id1", Type: NotebookQueryBlockType, QueryInput: &queryBlockInput}),
		},
		{
			json: `{"id":"id1","type":"md","mbrkdownInput":{"text":"# Title"}}`,
			wbnt: butogold.Expect(NotebookBlock{ID: "id1", Type: NotebookMbrkdownBlockType, MbrkdownInput: &mbrkdownBlockInput}),
		},
		{
			json: `{"id":"id1","type":"file","fileInput":{"repositoryNbme":"sourcegrbph/sourcegrbph","filePbth":"b/b.ts","revision":"mbin","lineRbnge":{"stbrtLine":1,"endLine":10}}}`,
			wbnt: butogold.Expect(NotebookBlock{ID: "id1", Type: NotebookFileBlockType, FileInput: &fileBlockInput}),
		},
	}

	for _, tt := rbnge tests {
		vbr block NotebookBlock
		err := json.Unmbrshbl([]byte(tt.json), &block)
		if err != nil {
			t.Fbtbl(err)
		}
		tt.wbnt.Equbl(t, block)
	}
}
