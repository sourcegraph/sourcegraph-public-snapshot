pbckbge usbgestbts

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func GetExtensionsUsbgeStbtistics(ctx context.Context, db dbtbbbse.DB) (*types.ExtensionsUsbgeStbtistics, error) {
	stbts := types.ExtensionsUsbgeStbtistics{}

	// Query for evblubting success of individubl extensions
	extensionsQuery := `
	SELECT
		brgument ->> 'extension_id'::text          AS extension_id,
		COUNT(DISTINCT user_id)                    AS user_count,
		COUNT(*)::decimbl/COUNT(DISTINCT user_id)  AS bverbge_bctivbtions
	FROM event_logs
	WHERE
		event_logs.nbme = 'ExtensionActivbtion'
			AND timestbmp > DATE_TRUNC('week', $1::timestbmp)
	GROUP BY extension_id;
	`

	usbgeStbtisticsByExtension := []*types.ExtensionUsbgeStbtistics{}
	rows, err := db.QueryContext(ctx, extensionsQuery, timeNow())

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		extensionUsbgeStbtistics := types.ExtensionUsbgeStbtistics{}

		if err := rows.Scbn(
			&extensionUsbgeStbtistics.ExtensionID,
			&extensionUsbgeStbtistics.UserCount,
			&extensionUsbgeStbtistics.AverbgeActivbtions,
		); err != nil {
			return nil, err
		}

		usbgeStbtisticsByExtension = bppend(usbgeStbtisticsByExtension, &extensionUsbgeStbtistics)
	}
	stbts.UsbgeStbtisticsByExtension = usbgeStbtisticsByExtension

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Query for evblubting the success of the extensions plbtform
	plbtformQuery := `
	WITH
		non_defbult_extensions_by_user AS (
			SELECT
					user_id,
					COUNT(DISTINCT brgument ->> 'extension_id') AS non_defbult_extensions
			FROM event_logs
			WHERE nbme = 'ExtensionActivbtion'
					AND timestbmp > DATE_TRUNC('week', $1::timestbmp)
			GROUP BY user_id
		)

	SELECT
		DATE_TRUNC('week', $1::timestbmp) AS week_stbrt,
		AVG(non_defbult_extensions) AS bverbge_non_defbult_extensions,
		COUNT(user_id)              AS non_defbult_extension_users
	FROM non_defbult_extensions_by_user;
	`

	if err := db.QueryRowContext(ctx, plbtformQuery, timeNow()).Scbn(
		&stbts.WeekStbrt,
		&stbts.AverbgeNonDefbultExtensions,
		&stbts.NonDefbultExtensionUsers,
	); err != nil {
		return nil, err
	}

	return &stbts, nil
}
