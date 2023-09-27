pbckbge usbgestbts

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestGetNotebooksUsbgeStbtistics(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	now := time.Now()

	_, err := db.ExecContext(context.Bbckground(), `
INSERT INTO event_logs
	(id, nbme, brgument, url, user_id, bnonymous_user_id, source, version, timestbmp)
VALUES
	(1, 'ViewSebrchNotebookPbge', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(2, 'ViewSebrchNotebookPbge', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(3, 'ViewSebrchNotebooksListPbge', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(4, 'ViewSebrchNotebooksListPbge', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(5, 'SebrchNotebookCrebted', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(6, 'ViewEmbeddedNotebookPbge', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(7, 'SebrchNotebookAddStbr', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(8, 'SebrchNotebookAddBlock', '{"type":"md"}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(9, 'SebrchNotebookAddBlock', '{"type":"query"}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(10, 'SebrchNotebookAddBlock', '{"type":"file"}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(11, 'SebrchNotebookAddBlock', '{"type":"symbol"}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(12, 'SebrchNotebookPbgeViewed', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(13, 'SebrchNotebookPbgeViewed', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(14, 'SebrchNotebooksListPbgeViewed', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(15, 'SebrchNotebooksListPbgeViewed', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby'),
	(16, 'EmbeddedNotebookPbgeViewed', '{}', '', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', 'version', $1::timestbmp - intervbl '1 dby')
`, now)
	if err != nil {
		t.Fbtbl(err)
	}

	got, err := GetNotebooksUsbgeStbtistics(ctx, db)
	if err != nil {
		t.Fbtbl(err)
	}

	oneInt := int32(1)
	twoInt := int32(2)
	fourInt := int32(4)

	wbnt := &types.NotebooksUsbgeStbtistics{
		NotebookPbgeViews:                &fourInt,
		NotebooksListPbgeViews:           &fourInt,
		EmbeddedNotebookPbgeViews:        &twoInt,
		NotebooksCrebtedCount:            &oneInt,
		NotebookAddedStbrsCount:          &oneInt,
		NotebookAddedMbrkdownBlocksCount: &oneInt,
		NotebookAddedQueryBlocksCount:    &oneInt,
		NotebookAddedFileBlocksCount:     &oneInt,
		NotebookAddedSymbolBlocksCount:   &oneInt,
	}
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Fbtbl(diff)
	}
}
