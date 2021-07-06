package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Database connections monitoring overview.
const TitleDatabaseConnectionsMonitoring = "Database connections"

func DatabaseConnectionsMonitoring(app string) []monitoring.Row {
	return []monitoring.Row{
		{
			{
				Name:           "max_open_conns",
				Description:    "maximum open",
				Query:          fmt.Sprintf(`sum by (app_name, db_name) (src_pgsql_conns_max_open{app_name=%q})`, app),
				Panel:          monitoring.Panel().LegendFormat("dbname={{db_name}}"),
				NoAlert:        true,
				Owner:          monitoring.ObservableOwnerCoreApplication,
				Interpretation: "none",
			},
			{
				Name:           "open_conns",
				Description:    "established",
				Query:          fmt.Sprintf(`sum by (app_name, db_name) (src_pgsql_conns_open{app_name=%q})`, app),
				Panel:          monitoring.Panel().LegendFormat("dbname={{db_name}}"),
				NoAlert:        true,
				Owner:          monitoring.ObservableOwnerCoreApplication,
				Interpretation: "none",
			},
		},
		{
			{
				Name:           "in_use",
				Description:    "used",
				Query:          fmt.Sprintf(`sum by (app_name, db_name) (src_pgsql_conns_in_use{app_name=%q})`, app),
				Panel:          monitoring.Panel().LegendFormat("dbname={{db_name}}"),
				NoAlert:        true,
				Owner:          monitoring.ObservableOwnerCoreApplication,
				Interpretation: "none",
			},
			{
				Name:           "idle",
				Description:    "idle",
				Query:          fmt.Sprintf(`sum by (app_name, db_name) (src_pgsql_conns_idle{app_name=%q})`, app),
				Panel:          monitoring.Panel().LegendFormat("dbname={{db_name}}"),
				NoAlert:        true,
				Owner:          monitoring.ObservableOwnerCoreApplication,
				Interpretation: "none",
			},
		},
		{
			{
				Name:           "waited_for",
				Description:    "waited for",
				Query:          fmt.Sprintf(`sum by (app_name, db_name) (increase(src_pgsql_conns_waited_for{app_name=%q}[1m]))`, app),
				Panel:          monitoring.Panel().LegendFormat("dbname={{db_name}}"),
				NoAlert:        true,
				Owner:          monitoring.ObservableOwnerCoreApplication,
				Interpretation: "none",
			},
			{
				Name:           "blocked_seconds",
				Description:    "blocked seconds (99th percentile)",
				Query:          fmt.Sprintf(`histogram_quantile(0.99, sum by (app_name, db_name, le) (rate(src_pgsql_conns_blocked_seconds_bucket{app_name=%q}[1m])))`, app),
				Panel:          monitoring.Panel().LegendFormat("dbname={{db_name}}").Unit(monitoring.Seconds),
				NoAlert:        true,
				Owner:          monitoring.ObservableOwnerCoreApplication,
				Interpretation: "none",
			},
		},
		{
			{
				Name:           "closed_max_idle",
				Description:    "closed by SetMaxIdleConns",
				Query:          fmt.Sprintf(`sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_idle{app_name=%q}[1m]))`, app),
				Panel:          monitoring.Panel().LegendFormat("dbname={{db_name}}"),
				NoAlert:        true,
				Owner:          monitoring.ObservableOwnerCoreApplication,
				Interpretation: "none",
			},
			{
				Name:           "closed_max_lifetime",
				Description:    "closed by SetConnMaxLifetime",
				Query:          fmt.Sprintf(`sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_lifetime{app_name=%q}[1m]))`, app),
				Panel:          monitoring.Panel().LegendFormat("dbname={{db_name}}"),
				NoAlert:        true,
				Owner:          monitoring.ObservableOwnerCoreApplication,
				Interpretation: "none",
			},
			{
				Name:           "closed_max_idle_time",
				Description:    "closed by SetConnMaxIdleTime",
				Query:          fmt.Sprintf(`sum by (app_name, db_name) (increase(src_pgsql_conns_closed_max_idle_time{app_name=%q}[1m]))`, app),
				Panel:          monitoring.Panel().LegendFormat("dbname={{db_name}}"),
				NoAlert:        true,
				Owner:          monitoring.ObservableOwnerCoreApplication,
				Interpretation: "none",
			},
		},
	}
}
