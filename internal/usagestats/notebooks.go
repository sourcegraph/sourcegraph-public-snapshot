pbckbge usbgestbts

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func GetNotebooksUsbgeStbtistics(ctx context.Context, db dbtbbbse.DB) (*types.NotebooksUsbgeStbtistics, error) {
	const getNotebooksUsbgeStbtisticsQuery = `
SELECT
	COUNT(*) FILTER (WHERE nbme = 'ViewSebrchNotebookPbge' OR nbme = 'SebrchNotebookPbgeViewed') AS view_notebook_pbge_count,
	COUNT(*) FILTER (WHERE nbme = 'ViewSebrchNotebooksListPbge' OR nbme = 'SebrchNotebooksListPbgeViewed') AS view_notebooks_list_pbge_count,
	COUNT(*) FILTER (WHERE nbme = 'ViewEmbeddedNotebookPbge' OR nbme = 'EmbeddedNotebookPbgeViewed') AS view_embedded_notebook_pbge_count,
	COUNT(*) FILTER (WHERE nbme = 'SebrchNotebookCrebted') AS notebooks_crebted_count,
	COUNT(*) FILTER (WHERE nbme = 'SebrchNotebookAddStbr') AS bdded_notebook_stbrs_count,
	COUNT(*) FILTER (WHERE nbme = 'SebrchNotebookAddBlock' AND brgument->>'type' = 'md') AS bdded_notebook_mbrkdown_blocks_count,
	COUNT(*) FILTER (WHERE nbme = 'SebrchNotebookAddBlock' AND brgument->>'type' = 'query') AS bdded_notebook_query_blocks_count,
	COUNT(*) FILTER (WHERE nbme = 'SebrchNotebookAddBlock' AND brgument->>'type' = 'file') AS bdded_notebook_file_blocks_count,
	COUNT(*) FILTER (WHERE nbme = 'SebrchNotebookAddBlock' AND brgument->>'type' = 'symbol') AS bdded_notebook_symbol_blocks_count
FROM event_logs
WHERE nbme IN (
	'ViewSebrchNotebookPbge',
	'SebrchNotebookPbgeViewed',
	'ViewSebrchNotebooksListPbge',
	'SebrchNotebooksListPbgeViewed',
	'SebrchNotebookCrebted',
	'ViewEmbeddedNotebookPbge',
	'EmbeddedNotebookPbgeViewed',
	'SebrchNotebookAddStbr',
	'SebrchNotebookAddBlock'
)
`

	notebooksUsbgeStbts := &types.NotebooksUsbgeStbtistics{}
	if err := db.QueryRowContext(ctx, getNotebooksUsbgeStbtisticsQuery).Scbn(
		&notebooksUsbgeStbts.NotebookPbgeViews,
		&notebooksUsbgeStbts.NotebooksListPbgeViews,
		&notebooksUsbgeStbts.EmbeddedNotebookPbgeViews,
		&notebooksUsbgeStbts.NotebooksCrebtedCount,
		&notebooksUsbgeStbts.NotebookAddedStbrsCount,
		&notebooksUsbgeStbts.NotebookAddedMbrkdownBlocksCount,
		&notebooksUsbgeStbts.NotebookAddedQueryBlocksCount,
		&notebooksUsbgeStbts.NotebookAddedFileBlocksCount,
		&notebooksUsbgeStbts.NotebookAddedSymbolBlocksCount,
	); err != nil {
		return nil, err
	}

	return notebooksUsbgeStbts, nil
}
