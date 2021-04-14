package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Postgres() *monitoring.Container {
	sum := "sum"

	// In docker-compose, codeintel-db container is called pgsql
	// In Kubernetes, codeintel-db container is called codeintel-db
	// Because of this, we track all database cAdvisor metrics in a single panel using this
	// container name regex to ensure we have observability on all platforms.
	const databaseContainerNames = "(pgsql|codeintel-db)"

	return &monitoring.Container{
		Name:        "postgres",
		Title:       "Postgres",
		Description: "Postgres metrics, exported from postgres_exporter (only available on Kubernetes).",
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{{
					monitoring.Observable{
						Name:              "connections",
						Description:       "active connections",
						Owner:             monitoring.ObservableOwnerCoreApplication,
						Query:             `sum by (datname) (pg_stat_activity_count{datname!~"template.*|postgres|cloudsqladmin"})`,
						Panel:             monitoring.Panel().LegendFormat("{{datname}}"),
						Warning:           monitoring.Alert().LessOrEqual(5, nil).For(5 * time.Minute),
						PossibleSolutions: "none",
					},
					monitoring.Observable{
						Name:              "transaction_durations",
						Description:       "maximum transaction durations",
						Owner:             monitoring.ObservableOwnerCoreApplication,
						Query:             `sum by (datname) (pg_stat_activity_max_tx_duration{datname!~"template.*|postgres|cloudsqladmin"})`,
						Panel:             monitoring.Panel().LegendFormat("{{datname}}").Unit(monitoring.Milliseconds),
						Warning:           monitoring.Alert().GreaterOrEqual(300, nil).For(5 * time.Minute),
						Critical:          monitoring.Alert().GreaterOrEqual(500, nil).For(10 * time.Minute),
						PossibleSolutions: "none",
					},
				},
				},
			},
			{
				Title:  "Database and collector status",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						monitoring.Observable{
							Name:              "postgres_up",
							Description:       "database availability",
							Owner:             monitoring.ObservableOwnerCoreApplication,
							Query:             "pg_up",
							Panel:             monitoring.Panel().LegendFormat("{{app}}"),
							Critical:          monitoring.Alert().LessOrEqual(0, nil).For(5 * time.Minute),
							PossibleSolutions: "none",
							Interpretation:    "A non-zero value indicates the database is online.",
						},
						monitoring.Observable{
							Name:        "invalid_indexes",
							Description: "invalid indexes (unusable by the query planner)",
							Owner:       monitoring.ObservableOwnerCoreApplication,
							Query:       "max by (relname)(pg_invalid_index_count)",
							Panel:       monitoring.Panel().LegendFormat("{{relname}}"),
							Critical:    monitoring.Alert().GreaterOrEqual(1, &sum).For(0),
							PossibleSolutions: `
								- Drop and re-create the invalid trigger - please contact Sourcegraph to supply the trigger definition.
							`,
							Interpretation: "A non-zero value indicates the that Postgres failed to build an index. Expect degraded performance until the index is manually rebuilt.",
						},
					},
					{
						monitoring.Observable{
							Name:        "pg_exporter_err",
							Description: "errors scraping postgres exporter",
							Owner:       monitoring.ObservableOwnerCoreApplication,
							Query:       "pg_exporter_last_scrape_error",
							Panel:       monitoring.Panel().LegendFormat("{{app}}"),
							Warning:     monitoring.Alert().GreaterOrEqual(1, nil).For(5 * time.Minute),
							PossibleSolutions: `
								- Ensure the Postgres exporter can access the Postgres database. Also, check the Postgres exporter logs for errors.
							`,
							Interpretation: "This value indicates issues retrieving metrics from postgres_exporter.",
						},
						monitoring.Observable{
							Name:           "migration_in_progress",
							Description:    "active schema migration",
							Owner:          monitoring.ObservableOwnerCoreApplication,
							Query:          "pg_sg_migration_status",
							Panel:          monitoring.Panel().LegendFormat("{{app}}"),
							Critical:       monitoring.Alert().GreaterOrEqual(1, nil).For(5 * time.Minute),
							Interpretation: "A 0 value indicates that no migration is in progress.",
							PossibleSolutions: `
								The database migration has been in progress for 5 or more minutes - please contact Sourcegraph if this persists.
							`,
						},
						// TODO(@daxmc99): Blocked by https://github.com/sourcegraph/sourcegraph/issues/13300
						// need to enable `pg_stat_statements` in Postgres conf
						// monitoring.Observable{
						//	Name:            "cache_hit_ratio",
						//	Description:     "ratio of cache hits over 5m",
						//	Owner:           monitoring.ObservableOwnerCoreApplication,
						//	Query:           `avg(rate(pg_stat_database_blks_hit{datname!~"template.*|postgres|cloudsqladmin"}[5m]) / (rate(pg_stat_database_blks_hit{datname!~"template.*|postgres|cloudsqladmin"}[5m]) + rate(pg_stat_database_blks_read{datname!~"template.*|postgres|cloudsqladmin"}[5m]))) by (datname) * 100`,
						//	DataMayNotExist: true,
						//	Warning:         monitoring.Alert().LessOrEqual(0.98, nil).For(5 * time.Minute),
						//	PossibleSolutions: "Cache hit ratio should be at least 99%, please [open an issue](https://github.com/sourcegraph/sourcegraph/issues/new/choose) " +
						//		"to add additional indexes",
						//	PanelOptions: monitoring.PanelOptions().Unit(monitoring.Percentage)},
					},
				},
			},
			{
				Title:  "Object size and bloat",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						monitoring.Observable{
							Name:           "pg_table_size",
							Description:    "table size",
							Owner:          monitoring.ObservableOwnerCoreApplication,
							Query:          `max by (relname)(pg_table_bloat_size)`,
							Panel:          monitoring.Panel().LegendFormat("{{relname}}").Unit(monitoring.Bytes),
							NoAlert:        true,
							Interpretation: "Total size of this table",
						},
						monitoring.Observable{
							Name:           "pg_table_bloat_ratio",
							Description:    "table bloat ratio",
							Owner:          monitoring.ObservableOwnerCoreApplication,
							Query:          `max by (relname)(pg_table_bloat_ratio) * 100`,
							Panel:          monitoring.Panel().LegendFormat("{{relname}}").Unit(monitoring.Percentage),
							NoAlert:        true,
							Interpretation: "Estimated bloat ratio of this table (high bloat = high overhead)",
						},
					},
					{
						monitoring.Observable{
							Name:           "pg_index_size",
							Description:    "index size",
							Owner:          monitoring.ObservableOwnerCoreApplication,
							Query:          `max by (relname)(pg_index_bloat_size)`,
							Panel:          monitoring.Panel().LegendFormat("{{relname}}").Unit(monitoring.Bytes),
							NoAlert:        true,
							Interpretation: "Total size of this index",
						},
						monitoring.Observable{
							Name:           "pg_index_bloat_ratio",
							Description:    "index bloat ratio",
							Owner:          monitoring.ObservableOwnerCoreApplication,
							Query:          `max by (relname)(pg_index_bloat_ratio) * 100`,
							Panel:          monitoring.Panel().LegendFormat("{{relname}}").Unit(monitoring.Percentage),
							NoAlert:        true,
							Interpretation: "Estimated bloat ratio of this index (high bloat = high overhead)",
						},
					},
				},
			},
			{
				Title:  shared.TitleProvisioningIndicators,
				Hidden: true,
				// See docstring for databaseContainerNames
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm(databaseContainerNames, monitoring.ObservableOwnerCoreApplication).Observable(),
						shared.ProvisioningMemoryUsageLongTerm(databaseContainerNames, monitoring.ObservableOwnerCoreApplication).Observable(),
					},
					{
						shared.ProvisioningCPUUsageShortTerm(databaseContainerNames, monitoring.ObservableOwnerCoreApplication).Observable(),
						shared.ProvisioningMemoryUsageShortTerm(databaseContainerNames, monitoring.ObservableOwnerCoreApplication).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleKubernetesMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable(databaseContainerNames, monitoring.ObservableOwnerCoreApplication).Observable(),
					},
				},
			},
		},

		// This is third-party service
		NoSourcegraphDebugServer: true,
	}
}
