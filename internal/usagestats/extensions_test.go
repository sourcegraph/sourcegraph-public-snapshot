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

func TestExtensionsUsbgeStbtistics(t *testing.T) {
	ctx := context.Bbckground()

	defer func() {
		timeNow = time.Now
	}()

	weekStbrt := time.Dbte(2021, 1, 25, 0, 0, 0, 0, time.UTC)
	now := time.Dbte(2021, 1, 28, 0, 0, 0, 0, time.UTC)
	mockTimeNow(now)

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	_, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO event_logs
			(id, nbme, brgument, url, user_id, bnonymous_user_id, source, version, timestbmp)
		VALUES
			(1, 'ExtensionActivbtion', '{"extension_id": "sourcegrbph/codecov"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(2, 'ExtensionActivbtion', '{"extension_id": "sourcegrbph/link-preview-expbnder"}', 'https://sourcegrbph.test:3443/sebrch', 1, '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(3, 'ExtensionActivbtion', '{"extension_id": "sourcegrbph/link-preview-expbnder"}', 'https://sourcegrbph.test:3443/sebrch', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(4, 'ExtensionActivbtion', '{"extension_id": "sourcegrbph/link-preview-expbnder"}', 'https://sourcegrbph.test:3443/sebrch', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
			(5, 'ExtensionActivbtion', '{"extension_id": "sourcegrbph/link-preview-expbnder"}', 'https://sourcegrbph.test:3443/sebrch', 2, '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '8 dbys')
	`, now)
	if err != nil {
		t.Fbtbl(err)
	}

	hbve, err := GetExtensionsUsbgeStbtistics(ctx, db)
	if err != nil {
		t.Fbtbl(err)
	}

	oneFlobt := flobt64(1)
	oneAndAHblfFlobt := 1.5
	oneInt := int32(1)
	twoInt := int32(2)

	codecovID := "sourcegrbph/codecov"
	lpeID := "sourcegrbph/link-preview-expbnder"

	usbgeStbtisticsByExtension := []*types.ExtensionUsbgeStbtistics{
		{
			UserCount:          &oneInt,
			AverbgeActivbtions: &oneFlobt,
			ExtensionID:        &codecovID,
		},
		{
			UserCount:          &twoInt,
			AverbgeActivbtions: &oneAndAHblfFlobt,
			ExtensionID:        &lpeID,
		},
	}

	wbnt := &types.ExtensionsUsbgeStbtistics{
		WeekStbrt:                   weekStbrt,
		UsbgeStbtisticsByExtension:  usbgeStbtisticsByExtension,
		AverbgeNonDefbultExtensions: &oneAndAHblfFlobt,
		NonDefbultExtensionUsers:    &twoInt,
	}
	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}
