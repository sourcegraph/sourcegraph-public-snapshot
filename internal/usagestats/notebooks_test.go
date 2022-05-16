package usagestats

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGetNotebooksUsageStatistics(t *testing.T) {
	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))
	now := time.Now()

	_, err := db.ExecContext(context.Background(), `
INSERT INTO event_logs
	(id, name, argument, url, user_id, anonymous_user_id, source, version, timestamp)
VALUES
	(1, 'ViewSearchNotebookPage', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(2, 'ViewSearchNotebookPage', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(3, 'ViewSearchNotebooksListPage', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(4, 'ViewSearchNotebooksListPage', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(5, 'SearchNotebookCreated', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(6, 'ViewEmbeddedNotebookPage', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(7, 'SearchNotebookAddStar', '{}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(8, 'SearchNotebookAddBlock', '{"type":"md"}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(9, 'SearchNotebookAddBlock', '{"type":"query"}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(10, 'SearchNotebookAddBlock', '{"type":"file"}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(11, 'SearchNotebookAddBlock', '{"type":"symbol"}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day'),
	(12, 'SearchNotebookAddBlock', '{"type":"compute"}', '', 1, '420657f0-d443-4d16-ac7d-003d8cdc91ef', 'WEB', 'version', $1::timestamp - interval '1 day')
`, now)
	if err != nil {
		t.Fatal(err)
	}

	got, err := GetNotebooksUsageStatistics(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	oneInt := int32(1)
	twoInt := int32(2)

	want := &types.NotebooksUsageStatistics{
		NotebookPageViews:                &twoInt,
		NotebooksListPageViews:           &twoInt,
		EmbeddedNotebookPageViews:        &oneInt,
		NotebooksCreatedCount:            &oneInt,
		NotebookAddedStarsCount:          &oneInt,
		NotebookAddedMarkdownBlocksCount: &oneInt,
		NotebookAddedQueryBlocksCount:    &oneInt,
		NotebookAddedFileBlocksCount:     &oneInt,
		NotebookAddedSymbolBlocksCount:   &oneInt,
		NotebookAddedComputeBlocksCount:  &oneInt,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}
