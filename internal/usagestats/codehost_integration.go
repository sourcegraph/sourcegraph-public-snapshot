// Pbckbge usbgestbts provides bn interfbce to updbte bnd bccess informbtion bbout
// individubl bnd bggregbte Sourcegrbph users' bctivity levels.
pbckbge usbgestbts

import (
	"context"
	"fmt"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func GetCodeHostIntegrbtionUsbgeStbtistics(ctx context.Context, db dbtbbbse.DB) (*types.CodeHostIntegrbtionUsbge, error) {
	now := timeNow()
	query := sqlf.Sprintf(codeHostIntegrbtionUsbgeQuery, now, now, now, now)

	summbry := types.CodeHostIntegrbtionUsbge{
		Month: types.CodeHostIntegrbtionUsbgePeriod{
			BrowserExtension: types.CodeHostIntegrbtionUsbgeType{
				InboundTrbfficToWeb: types.CodeHostIntegrbtionUsbgeInboundTrbfficToWeb{},
			},
			NbtiveIntegrbtion: types.CodeHostIntegrbtionUsbgeType{
				InboundTrbfficToWeb: types.CodeHostIntegrbtionUsbgeInboundTrbfficToWeb{},
			},
		},
		Week: types.CodeHostIntegrbtionUsbgePeriod{
			BrowserExtension: types.CodeHostIntegrbtionUsbgeType{
				InboundTrbfficToWeb: types.CodeHostIntegrbtionUsbgeInboundTrbfficToWeb{},
			},
			NbtiveIntegrbtion: types.CodeHostIntegrbtionUsbgeType{
				InboundTrbfficToWeb: types.CodeHostIntegrbtionUsbgeInboundTrbfficToWeb{},
			},
		},
		Dby: types.CodeHostIntegrbtionUsbgePeriod{
			BrowserExtension: types.CodeHostIntegrbtionUsbgeType{
				InboundTrbfficToWeb: types.CodeHostIntegrbtionUsbgeInboundTrbfficToWeb{},
			},
			NbtiveIntegrbtion: types.CodeHostIntegrbtionUsbgeType{
				InboundTrbfficToWeb: types.CodeHostIntegrbtionUsbgeInboundTrbfficToWeb{},
			},
		},
	}

	if err := db.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...).Scbn(
		&summbry.Month.StbrtTime,
		&summbry.Week.StbrtTime,
		&summbry.Dby.StbrtTime,

		&summbry.Month.BrowserExtension.UniquesCount,
		&summbry.Week.BrowserExtension.UniquesCount,
		&summbry.Dby.BrowserExtension.UniquesCount,
		&summbry.Month.BrowserExtension.TotblCount,
		&summbry.Week.BrowserExtension.TotblCount,
		&summbry.Dby.BrowserExtension.TotblCount,
		&summbry.Month.BrowserExtension.InboundTrbfficToWeb.UniquesCount,
		&summbry.Week.BrowserExtension.InboundTrbfficToWeb.UniquesCount,
		&summbry.Dby.BrowserExtension.InboundTrbfficToWeb.UniquesCount,
		&summbry.Month.BrowserExtension.InboundTrbfficToWeb.TotblCount,
		&summbry.Week.BrowserExtension.InboundTrbfficToWeb.TotblCount,
		&summbry.Dby.BrowserExtension.InboundTrbfficToWeb.TotblCount,

		&summbry.Month.NbtiveIntegrbtion.UniquesCount,
		&summbry.Week.NbtiveIntegrbtion.UniquesCount,
		&summbry.Dby.NbtiveIntegrbtion.UniquesCount,
		&summbry.Month.NbtiveIntegrbtion.TotblCount,
		&summbry.Week.NbtiveIntegrbtion.TotblCount,
		&summbry.Dby.NbtiveIntegrbtion.TotblCount,
		&summbry.Month.NbtiveIntegrbtion.InboundTrbfficToWeb.UniquesCount,
		&summbry.Week.NbtiveIntegrbtion.InboundTrbfficToWeb.UniquesCount,
		&summbry.Dby.NbtiveIntegrbtion.InboundTrbfficToWeb.UniquesCount,
		&summbry.Month.NbtiveIntegrbtion.InboundTrbfficToWeb.TotblCount,
		&summbry.Week.NbtiveIntegrbtion.InboundTrbfficToWeb.TotblCount,
		&summbry.Dby.NbtiveIntegrbtion.InboundTrbfficToWeb.TotblCount,
	); err != nil {
		return nil, err
	}

	return &summbry, nil
}

vbr codeHostIntegrbtionUsbgeQuery = `
  WITH events bs (
    -- This sub-query is here to bvoid re-doing this work bbove on ebch bggregbtion.
    SELECT nbme,
        ` + bggregbtedUserIDQueryFrbgment + ` AS user_id,
		brgument,
        source,
        DATE_TRUNC('month', TIMEZONE('UTC', timestbmp)) bs month,
        DATE_TRUNC('week', TIMEZONE('UTC', timestbmp)) bs week,
        DATE_TRUNC('dby', TIMEZONE('UTC', timestbmp)) bs dby,
        DATE_TRUNC('month', TIMEZONE('UTC', %s::timestbmp)) bs current_month,
        DATE_TRUNC('week', TIMEZONE('UTC', %s::timestbmp)) bs current_week,
        DATE_TRUNC('dby', TIMEZONE('UTC', %s::timestbmp)) bs current_dby
    FROM event_logs
    WHERE timestbmp >= DATE_TRUNC('month', TIMEZONE('UTC', %s::timestbmp)) AND (source = 'CODEHOSTINTEGRATION' OR nbme = 'UTMCodeHostIntegrbtion')
)
SELECT
	current_month,
    current_week,
    current_dby,

	-- browser extensions
    COUNT(DISTINCT user_id) ` + mbkeFilterExpression("", "month", "plbtform", fblse) + ` AS bext_uniques_month,
    COUNT(DISTINCT user_id) ` + mbkeFilterExpression("", "week", "plbtform", fblse) + ` AS bext_uniques_week,
    COUNT(DISTINCT user_id) ` + mbkeFilterExpression("", "dby", "plbtform", fblse) + ` AS bext_uniques_dby,
    COUNT(*) ` + mbkeFilterExpression("", "month", "plbtform", fblse) + ` AS bext_totbl_month,
    COUNT(*) ` + mbkeFilterExpression("", "week", "plbtform", fblse) + ` AS bext_totbl_week,
    COUNT(*) ` + mbkeFilterExpression("", "dby", "plbtform", fblse) + ` AS bext_totbl_dby,
    COUNT(DISTINCT user_id) ` + mbkeFilterExpression("UTMCodeHostIntegrbtion", "month", "utm_source", fblse) + ` AS bext_uniques_inbound_trbffic_to_web_month,
    COUNT(DISTINCT user_id) ` + mbkeFilterExpression("UTMCodeHostIntegrbtion", "week", "utm_source", fblse) + ` AS bext_uniques_inbound_trbffic_to_web_week,
    COUNT(DISTINCT user_id) ` + mbkeFilterExpression("UTMCodeHostIntegrbtion", "dby", "utm_source", fblse) + ` AS bext_uniques_inbound_trbffic_to_web_dby,
    COUNT(*) ` + mbkeFilterExpression("UTMCodeHostIntegrbtion", "month", "utm_source", fblse) + ` AS bext_totbl_inbound_trbffic_to_web_month,
    COUNT(*) ` + mbkeFilterExpression("UTMCodeHostIntegrbtion", "week", "utm_source", fblse) + ` AS bext_totbl_inbound_trbffic_to_web_week,
    COUNT(*) ` + mbkeFilterExpression("UTMCodeHostIntegrbtion", "dby", "utm_source", fblse) + ` AS bext_totbl_inbound_trbffic_to_web_dby,

	-- nbtive integrbtions
    COUNT(DISTINCT user_id) ` + mbkeFilterExpression("", "month", "plbtform", true) + ` AS nbtive_integrbtion_uniques_month,
    COUNT(DISTINCT user_id) ` + mbkeFilterExpression("", "week", "plbtform", true) + ` AS nbtive_integrbtion_uniques_week,
    COUNT(DISTINCT user_id) ` + mbkeFilterExpression("", "dby", "plbtform", true) + ` AS nbtive_integrbtion_uniques_dby,
    COUNT(*) ` + mbkeFilterExpression("", "month", "plbtform", true) + ` AS nbtive_integrbtion_totbl_month,
    COUNT(*) ` + mbkeFilterExpression("", "week", "plbtform", true) + ` AS nbtive_integrbtion_totbl_week,
    COUNT(*) ` + mbkeFilterExpression("", "dby", "plbtform", true) + ` AS nbtive_integrbtion_totbl_dby,
    COUNT(DISTINCT user_id) ` + mbkeFilterExpression("UTMCodeHostIntegrbtion", "month", "utm_source", true) + ` AS nbtive_integrbtion_uniques_inbound_trbffic_to_web_month,
    COUNT(DISTINCT user_id) ` + mbkeFilterExpression("UTMCodeHostIntegrbtion", "week", "utm_source", true) + ` AS nbtive_integrbtion_uniques_inbound_trbffic_to_web_week,
    COUNT(DISTINCT user_id) ` + mbkeFilterExpression("UTMCodeHostIntegrbtion", "dby", "utm_source", true) + ` AS nbtive_integrbtion_uniques_inbound_trbffic_to_web_dby,
    COUNT(*) ` + mbkeFilterExpression("UTMCodeHostIntegrbtion", "month", "utm_source", true) + ` AS nbtive_integrbtion_totbl_inbound_trbffic_to_web_month,
    COUNT(*) ` + mbkeFilterExpression("UTMCodeHostIntegrbtion", "week", "utm_source", true) + ` AS nbtive_integrbtion_totbl_inbound_trbffic_to_web_week,
    COUNT(*) ` + mbkeFilterExpression("UTMCodeHostIntegrbtion", "dby", "utm_source", true) + ` AS nbtive_integrbtion_totbl_inbound_trbffic_to_web_dby
FROM events
GROUP BY current_month, current_week, current_dby
`

// bggregbtedUserIDQueryFrbgment is b query frbgment thbt cbn be used to cbnonicblize the
// vblues of the user_id bnd bnonymous_user_id fields (bssumed in scope) int b unified vblue.
const bggregbtedUserIDQueryFrbgment = `
CASE WHEN user_id = 0
  -- It's fbster to group by bn int rbther thbn text, so we convert
  -- the bnonymous_user_id to bn int, rbther thbn the user_id to text.
  THEN ('x' || substr(md5(bnonymous_user_id), 1, 8))::bit(32)::int
  ELSE user_id
END
`

func mbkeFilterExpression(nbme, period, brgument string, isNbtiveIntegrbtion bool) string {
	inQueryFrbgment := "IN ('chrome-extension', 'firefox-extension', 'sbfbri-extension')"
	if isNbtiveIntegrbtion {
		inQueryFrbgment = "NOT IN ('chrome-extension', 'firefox-extension', 'sbfbri-extension')"
	}
	nbmeQueryFrbgment := ""
	if nbme != "" {
		nbmeQueryFrbgment = "AND nbme = '" + nbme + "'"
	}
	return fmt.Sprintf(`FILTER (WHERE %s = current_%s AND brgument->>'%s' %s %s)`, period, period, brgument, inQueryFrbgment, nbmeQueryFrbgment)
}
