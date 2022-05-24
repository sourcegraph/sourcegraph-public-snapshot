package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Postgres() *monitoring.Container {
	const (
		// In docker-compose, codeintel-db container is called pgsql. In Kubernetes,
		// codeintel-db container is called codeintel-db Because of this, we track
		// all database cAdvisor metrics in a single panel using this container
		// name regex to ensure we have observability on all platforms.
		containerName = "(pgsql|codeintel-db|codeinsights)"
	)
	return &monitoring.Container{
		Name:                     "postgres",
		Title:                    "Postgres",
		Description:              "Postgres metrics, exported from postgres_exporter (not available on server).",
		NoSourcegraphDebugServer: true, // This is third-party service
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{{
					monitoring.Observable{
						Name:              "connections",
						Description:       "active connections",
						Owner:             monitoring.ObservableOwnerDevOps,
						DataMustExist:     false, // not deployed on docker-compose
						Query:             `sum by (job) (pg_stat_activity_count{datname!~"template.*|postgres|cloudsqladmin"}) OR sum by (job) (pg_stat_activity_count{job="codeinsights-db", datname!~"template.*|cloudsqladmin"})`,
						Panel:             monitoring.Panel().LegendFormat("{{datname}}"),
						Warning:           monitoring.Alert().LessOrEqual(5).For(5 * time.Minute),
						PossibleSolutions: "none",
					},
					monitoring.Observable{
						Name:          "usage_connections_percentage",
						Description:   "connection in use",
						Owner:         monitoring.ObservableOwnerDevOps,
						DataMustExist: false,
						Query:         `sum(pg_stat_activity_count) by (job) / (sum(pg_settings_max_connections) by (job) - sum(pg_settings_superuser_reserved_connections) by (job)) * 100`,
						Panel:         monitoring.Panel().LegendFormat("{{job}}").Unit(monitoring.Percentage).Max(100).Min(0),
						Warning:       monitoring.Alert().GreaterOrEqual(80).For(5 * time.Minute),
						Critical:      monitoring.Alert().GreaterOrEqual(100).For(5 * time.Minute),
						PossibleSolutions: `
							- Consider increasing [max_connections](https://www.postgresql.org/docs/current/runtime-config-connection.html#GUC-MAX-CONNECTIONS) of the database instance, [learn more](https://docs.sourcegraph.com/admin/config/postgres-conf)
						`,
					},
					monitoring.Observable{
						Name:              "transaction_durations",
						Description:       "maximum transaction durations",
						Owner:             monitoring.ObservableOwnerDevOps,
						DataMustExist:     false, // not deployed on docker-compose
						Query:             `sum by (job) (pg_stat_activity_max_tx_duration{datname!~"template.*|postgres|cloudsqladmin"}) OR sum by (job) (pg_stat_activity_max_tx_duration{job="codeinsights-db", datname!~"template.*|cloudsqladmin"})`,
						Panel:             monitoring.Panel().LegendFormat("{{datname}}").Unit(monitoring.Seconds),
						Warning:           monitoring.Alert().GreaterOrEqual(0.3).For(5 * time.Minute),
						Critical:          monitoring.Alert().GreaterOrEqual(0.5).For(10 * time.Minute),
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
							Owner:             monitoring.ObservableOwnerDevOps,
							DataMustExist:     false, // not deployed on docker-compose
							Query:             "pg_up",
							Panel:             monitoring.Panel().LegendFormat("{{app}}"),
							Critical:          monitoring.Alert().LessOrEqual(0).For(5 * time.Minute),
							PossibleSolutions: "none",
							Interpretation:    "A non-zero value indicates the database is online.",
						},
						monitoring.Observable{
							Name:          "invalid_indexes",
							Description:   "invalid indexes (unusable by the query planner)",
							Owner:         monitoring.ObservableOwnerDevOps,
							DataMustExist: false, // not deployed on docker-compose
							Query:         "max by (relname)(pg_invalid_index_count)",
							Panel:         monitoring.Panel().LegendFormat("{{relname}}"),
							Critical:      monitoring.Alert().GreaterOrEqual(1).AggregateBy(monitoring.AggregatorSum),
							PossibleSolutions: `
								- Drop and re-create the invalid trigger - please contact Sourcegraph to supply the trigger definition.
							`,
							Interpretation: "A non-zero value indicates the that Postgres failed to build an index. Expect degraded performance until the index is manually rebuilt.",
						},
					},
					{
						monitoring.Observable{
							Name:          "pg_exporter_err",
							Description:   "errors scraping postgres exporter",
							Owner:         monitoring.ObservableOwnerDevOps,
							DataMustExist: false, // not deployed on docker-compose
							Query:         "pg_exporter_last_scrape_error",
							Panel:         monitoring.Panel().LegendFormat("{{app}}"),
							Warning:       monitoring.Alert().GreaterOrEqual(1).For(5 * time.Minute),

							PossibleSolutions: `
								- Ensure the Postgres exporter can access the Postgres database. Also, check the Postgres exporter logs for errors.
							`,
							Interpretation: "This value indicates issues retrieving metrics from postgres_exporter.",
						},
						monitoring.Observable{
							Name:           "migration_in_progress",
							Description:    "active schema migration",
							Owner:          monitoring.ObservableOwnerDevOps,
							DataMustExist:  false, // not deployed on docker-compose
							Query:          "pg_sg_migration_status",
							Panel:          monitoring.Panel().LegendFormat("{{app}}"),
							Critical:       monitoring.Alert().GreaterOrEqual(1).For(5 * time.Minute),
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
						//	Owner:           monitoring.ObservableOwnerDevOps,
						//	Query:           `avg(rate(pg_stat_database_blks_hit{datname!~"template.*|postgres|cloudsqladmin"}[5m]) / (rate(pg_stat_database_blks_hit{datname!~"template.*|postgres|cloudsqladmin"}[5m]) + rate(pg_stat_database_blks_read{datname!~"template.*|postgres|cloudsqladmin"}[5m]))) by (datname) * 100`,
						//	DataMayNotExist: true,
						//	Warning:         monitoring.Alert().LessOrEqual(0.98).For(5 * time.Minute),
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
							Owner:          monitoring.ObservableOwnerDevOps,
							Query:          `max by (relname)(pg_table_bloat_size)`,
							Panel:          monitoring.Panel().LegendFormat("{{relname}}").Unit(monitoring.Bytes),
							NoAlert:        true,
							Interpretation: "Total size of this table",
						},
						monitoring.Observable{
							Name:           "pg_table_bloat_ratio",
							Description:    "table bloat ratio",
							Owner:          monitoring.ObservableOwnerDevOps,
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
							Owner:          monitoring.ObservableOwnerDevOps,
							Query:          `max by (relname)(pg_index_bloat_size)`,
							Panel:          monitoring.Panel().LegendFormat("{{relname}}").Unit(monitoring.Bytes),
							NoAlert:        true,
							Interpretation: "Total size of this index",
						},
						monitoring.Observable{
							Name:           "pg_index_bloat_ratio",
							Description:    "index bloat ratio",
							Owner:          monitoring.ObservableOwnerDevOps,
							Query:          `max by (relname)(pg_index_bloat_ratio) * 100`,
							Panel:          monitoring.Panel().LegendFormat("{{relname}}").Unit(monitoring.Percentage),
							NoAlert:        true,
							Interpretation: "Estimated bloat ratio of this index (high bloat = high overhead)",
						},
					},
				},
			},

			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerDevOps, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerDevOps, nil),
		},
	}
}
