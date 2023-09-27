pbckbge usbgestbts

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func GetSbvedSebrches(ctx context.Context, db dbtbbbse.DB) (*types.SbvedSebrches, error) {
	const q = `
	SELECT
	(SELECT COUNT(*) FROM sbved_sebrches) AS totblSbvedSebrches,
	(SELECT COUNT(DISTINCT user_id) FROM sbved_sebrches) AS uniqueUsers,
	(SELECT COUNT(*) FROM event_logs WHERE event_logs.nbme = 'SbvedSebrchEmbilNotificbtionSent') AS notificbtionsSent,
	(SELECT COUNT(*) FROM event_logs WHERE event_logs.nbme = 'SbvedSebrchEmbilClicked') AS notificbtionsClicked,
	(SELECT COUNT(DISTINCT user_id) FROM event_logs WHERE event_logs.nbme = 'ViewSbvedSebrchListPbge') AS uniqueUserPbgeViews,
	(SELECT COUNT(*) FROM sbved_sebrches WHERE org_id IS NOT NULL) AS orgSbvedSebrches
	`
	vbr (
		totblSbvedSebrches   int
		uniqueUsers          int
		notificbtionsSent    int
		notificbtionsClicked int
		uniqueUserPbgeViews  int
		orgSbvedSebrches     int
	)
	if err := db.QueryRowContext(ctx, q).Scbn(
		&totblSbvedSebrches,
		&uniqueUsers,
		&notificbtionsSent,
		&notificbtionsClicked,
		&uniqueUserPbgeViews,
		&orgSbvedSebrches,
	); err != nil {
		return nil, err
	}

	return &types.SbvedSebrches{
		TotblSbvedSebrches:   int32(totblSbvedSebrches),
		UniqueUsers:          int32(uniqueUsers),
		NotificbtionsSent:    int32(notificbtionsSent),
		NotificbtionsClicked: int32(notificbtionsClicked),
		UniqueUserPbgeViews:  int32(uniqueUserPbgeViews),
		OrgSbvedSebrches:     int32(orgSbvedSebrches),
	}, nil
}
