package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetNotebooksUsageStatistics(ctx context.Context, db database.DB) (*types.NotebooksUsageStatistics, error) {
	const getNotebooksUsageStatisticsQuery = `
SELECT
	COUNT(*) FILTER (WHERE name = 'ViewSearchNotebookPage' OR name = 'SearchNotebookPageViewed') AS view_notebook_page_count,
	COUNT(*) FILTER (WHERE name = 'ViewSearchNotebooksListPage' OR name = 'SearchNotebooksListPageViewed') AS view_notebooks_list_page_count,
	COUNT(*) FILTER (WHERE name = 'ViewEmbeddedNotebookPage' OR name = 'EmbeddedNotebookPageViewed') AS view_embedded_notebook_page_count,
	COUNT(*) FILTER (WHERE name = 'SearchNotebookCreated') AS notebooks_created_count,
	COUNT(*) FILTER (WHERE name = 'SearchNotebookAddStar') AS added_notebook_stars_count,
	COUNT(*) FILTER (WHERE name = 'SearchNotebookAddBlock' AND argument->>'type' = 'md') AS added_notebook_markdown_blocks_count,
	COUNT(*) FILTER (WHERE name = 'SearchNotebookAddBlock' AND argument->>'type' = 'query') AS added_notebook_query_blocks_count,
	COUNT(*) FILTER (WHERE name = 'SearchNotebookAddBlock' AND argument->>'type' = 'file') AS added_notebook_file_blocks_count,
	COUNT(*) FILTER (WHERE name = 'SearchNotebookAddBlock' AND argument->>'type' = 'symbol') AS added_notebook_symbol_blocks_count
FROM event_logs
WHERE name IN (
	'ViewSearchNotebookPage',
	'SearchNotebookPageViewed',
	'ViewSearchNotebooksListPage',
	'SearchNotebooksListPageViewed',
	'SearchNotebookCreated',
	'ViewEmbeddedNotebookPage',
	'EmbeddedNotebookPageViewed',
	'SearchNotebookAddStar',
	'SearchNotebookAddBlock'
)
`

	notebooksUsageStats := &types.NotebooksUsageStatistics{}
	if err := db.QueryRowContext(ctx, getNotebooksUsageStatisticsQuery).Scan(
		&notebooksUsageStats.NotebookPageViews,
		&notebooksUsageStats.NotebooksListPageViews,
		&notebooksUsageStats.EmbeddedNotebookPageViews,
		&notebooksUsageStats.NotebooksCreatedCount,
		&notebooksUsageStats.NotebookAddedStarsCount,
		&notebooksUsageStats.NotebookAddedMarkdownBlocksCount,
		&notebooksUsageStats.NotebookAddedQueryBlocksCount,
		&notebooksUsageStats.NotebookAddedFileBlocksCount,
		&notebooksUsageStats.NotebookAddedSymbolBlocksCount,
	); err != nil {
		return nil, err
	}

	return notebooksUsageStats, nil
}
