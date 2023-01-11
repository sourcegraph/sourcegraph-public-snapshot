// Package usagestats provides an interface to update and access information about
// individual and aggregate Sourcegraph users' activity levels.
package usagestats

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func GetCodeHostIntegrationUsageStatistics(ctx context.Context, db database.DB) (*types.CodeHostIntegrationUsage, error) {
	now := timeNow()
	query := sqlf.Sprintf(codeHostIntegrationUsageQuery, now, now, now, now)

	summary := types.CodeHostIntegrationUsage{
		Month: types.CodeHostIntegrationUsagePeriod{
			BrowserExtension: types.CodeHostIntegrationUsageType{
				InboundTrafficToWeb: types.CodeHostIntegrationUsageInboundTrafficToWeb{},
			},
			NativeIntegration: types.CodeHostIntegrationUsageType{
				InboundTrafficToWeb: types.CodeHostIntegrationUsageInboundTrafficToWeb{},
			},
		},
		Week: types.CodeHostIntegrationUsagePeriod{
			BrowserExtension: types.CodeHostIntegrationUsageType{
				InboundTrafficToWeb: types.CodeHostIntegrationUsageInboundTrafficToWeb{},
			},
			NativeIntegration: types.CodeHostIntegrationUsageType{
				InboundTrafficToWeb: types.CodeHostIntegrationUsageInboundTrafficToWeb{},
			},
		},
		Day: types.CodeHostIntegrationUsagePeriod{
			BrowserExtension: types.CodeHostIntegrationUsageType{
				InboundTrafficToWeb: types.CodeHostIntegrationUsageInboundTrafficToWeb{},
			},
			NativeIntegration: types.CodeHostIntegrationUsageType{
				InboundTrafficToWeb: types.CodeHostIntegrationUsageInboundTrafficToWeb{},
			},
		},
	}

	if err := db.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...).Scan(
		&summary.Month.StartTime,
		&summary.Week.StartTime,
		&summary.Day.StartTime,

		&summary.Month.BrowserExtension.UniquesCount,
		&summary.Week.BrowserExtension.UniquesCount,
		&summary.Day.BrowserExtension.UniquesCount,
		&summary.Month.BrowserExtension.TotalCount,
		&summary.Week.BrowserExtension.TotalCount,
		&summary.Day.BrowserExtension.TotalCount,
		&summary.Month.BrowserExtension.InboundTrafficToWeb.UniquesCount,
		&summary.Week.BrowserExtension.InboundTrafficToWeb.UniquesCount,
		&summary.Day.BrowserExtension.InboundTrafficToWeb.UniquesCount,
		&summary.Month.BrowserExtension.InboundTrafficToWeb.TotalCount,
		&summary.Week.BrowserExtension.InboundTrafficToWeb.TotalCount,
		&summary.Day.BrowserExtension.InboundTrafficToWeb.TotalCount,

		&summary.Month.NativeIntegration.UniquesCount,
		&summary.Week.NativeIntegration.UniquesCount,
		&summary.Day.NativeIntegration.UniquesCount,
		&summary.Month.NativeIntegration.TotalCount,
		&summary.Week.NativeIntegration.TotalCount,
		&summary.Day.NativeIntegration.TotalCount,
		&summary.Month.NativeIntegration.InboundTrafficToWeb.UniquesCount,
		&summary.Week.NativeIntegration.InboundTrafficToWeb.UniquesCount,
		&summary.Day.NativeIntegration.InboundTrafficToWeb.UniquesCount,
		&summary.Month.NativeIntegration.InboundTrafficToWeb.TotalCount,
		&summary.Week.NativeIntegration.InboundTrafficToWeb.TotalCount,
		&summary.Day.NativeIntegration.InboundTrafficToWeb.TotalCount,
	); err != nil {
		return nil, err
	}

	return &summary, nil
}

var codeHostIntegrationUsageQuery = `
  WITH events as (
    -- This sub-query is here to avoid re-doing this work above on each aggregation.
    SELECT name,
        ` + aggregatedUserIDQueryFragment + ` AS user_id,
		argument,
        source,
        DATE_TRUNC('month', TIMEZONE('UTC', timestamp)) as month,
        DATE_TRUNC('week', TIMEZONE('UTC', timestamp)) as week,
        DATE_TRUNC('day', TIMEZONE('UTC', timestamp)) as day,
        DATE_TRUNC('month', TIMEZONE('UTC', %s::timestamp)) as current_month,
        DATE_TRUNC('week', TIMEZONE('UTC', %s::timestamp)) as current_week,
        DATE_TRUNC('day', TIMEZONE('UTC', %s::timestamp)) as current_day
    FROM event_logs
    WHERE timestamp >= DATE_TRUNC('month', TIMEZONE('UTC', %s::timestamp)) AND (source = 'CODEHOSTINTEGRATION' OR name = 'UTMCodeHostIntegration')
)
SELECT
	current_month,
    current_week,
    current_day,

	-- browser extensions
    COUNT(DISTINCT user_id) ` + makeFilterExpression("", "month", "platform", false) + ` AS bext_uniques_month,
    COUNT(DISTINCT user_id) ` + makeFilterExpression("", "week", "platform", false) + ` AS bext_uniques_week,
    COUNT(DISTINCT user_id) ` + makeFilterExpression("", "day", "platform", false) + ` AS bext_uniques_day,
    COUNT(*) ` + makeFilterExpression("", "month", "platform", false) + ` AS bext_total_month,
    COUNT(*) ` + makeFilterExpression("", "week", "platform", false) + ` AS bext_total_week,
    COUNT(*) ` + makeFilterExpression("", "day", "platform", false) + ` AS bext_total_day,
    COUNT(DISTINCT user_id) ` + makeFilterExpression("UTMCodeHostIntegration", "month", "utm_source", false) + ` AS bext_uniques_inbound_traffic_to_web_month,
    COUNT(DISTINCT user_id) ` + makeFilterExpression("UTMCodeHostIntegration", "week", "utm_source", false) + ` AS bext_uniques_inbound_traffic_to_web_week,
    COUNT(DISTINCT user_id) ` + makeFilterExpression("UTMCodeHostIntegration", "day", "utm_source", false) + ` AS bext_uniques_inbound_traffic_to_web_day,
    COUNT(*) ` + makeFilterExpression("UTMCodeHostIntegration", "month", "utm_source", false) + ` AS bext_total_inbound_traffic_to_web_month,
    COUNT(*) ` + makeFilterExpression("UTMCodeHostIntegration", "week", "utm_source", false) + ` AS bext_total_inbound_traffic_to_web_week,
    COUNT(*) ` + makeFilterExpression("UTMCodeHostIntegration", "day", "utm_source", false) + ` AS bext_total_inbound_traffic_to_web_day,

	-- native integrations
    COUNT(DISTINCT user_id) ` + makeFilterExpression("", "month", "platform", true) + ` AS native_integration_uniques_month,
    COUNT(DISTINCT user_id) ` + makeFilterExpression("", "week", "platform", true) + ` AS native_integration_uniques_week,
    COUNT(DISTINCT user_id) ` + makeFilterExpression("", "day", "platform", true) + ` AS native_integration_uniques_day,
    COUNT(*) ` + makeFilterExpression("", "month", "platform", true) + ` AS native_integration_total_month,
    COUNT(*) ` + makeFilterExpression("", "week", "platform", true) + ` AS native_integration_total_week,
    COUNT(*) ` + makeFilterExpression("", "day", "platform", true) + ` AS native_integration_total_day,
    COUNT(DISTINCT user_id) ` + makeFilterExpression("UTMCodeHostIntegration", "month", "utm_source", true) + ` AS native_integration_uniques_inbound_traffic_to_web_month,
    COUNT(DISTINCT user_id) ` + makeFilterExpression("UTMCodeHostIntegration", "week", "utm_source", true) + ` AS native_integration_uniques_inbound_traffic_to_web_week,
    COUNT(DISTINCT user_id) ` + makeFilterExpression("UTMCodeHostIntegration", "day", "utm_source", true) + ` AS native_integration_uniques_inbound_traffic_to_web_day,
    COUNT(*) ` + makeFilterExpression("UTMCodeHostIntegration", "month", "utm_source", true) + ` AS native_integration_total_inbound_traffic_to_web_month,
    COUNT(*) ` + makeFilterExpression("UTMCodeHostIntegration", "week", "utm_source", true) + ` AS native_integration_total_inbound_traffic_to_web_week,
    COUNT(*) ` + makeFilterExpression("UTMCodeHostIntegration", "day", "utm_source", true) + ` AS native_integration_total_inbound_traffic_to_web_day
FROM events
GROUP BY current_month, current_week, current_day
`

// aggregatedUserIDQueryFragment is a query fragment that can be used to canonicalize the
// values of the user_id and anonymous_user_id fields (assumed in scope) int a unified value.
const aggregatedUserIDQueryFragment = `
CASE WHEN user_id = 0
  -- It's faster to group by an int rather than text, so we convert
  -- the anonymous_user_id to an int, rather than the user_id to text.
  THEN ('x' || substr(md5(anonymous_user_id), 1, 8))::bit(32)::int
  ELSE user_id
END
`

func makeFilterExpression(name, period, argument string, isNativeIntegration bool) string {
	inQueryFragment := "IN ('chrome-extension', 'firefox-extension', 'safari-extension')"
	if isNativeIntegration {
		inQueryFragment = "NOT IN ('chrome-extension', 'firefox-extension', 'safari-extension')"
	}
	nameQueryFragment := ""
	if name != "" {
		nameQueryFragment = "AND name = '" + name + "'"
	}
	return fmt.Sprintf(`FILTER (WHERE %s = current_%s AND argument->>'%s' %s %s)`, period, period, argument, inQueryFragment, nameQueryFragment)
}
