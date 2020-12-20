package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

const (
	dbCodeIntel   = "codeintel-db"
	dbSourcegraph = "pgsql"
)

func Postgres() *monitoring.Container {
	return &monitoring.Container{
		Name:        "postgres",
		Title:       "Postgres",
		Description: "Metrics from postgres_exporter.",
		Groups: []monitoring.Group{
			{
				Title: "Default postgres dashboard",
				Rows: []monitoring.Row{{
					monitoring.Observable{
						Name:              "connections",
						Description:       "connections",
						Owner:             monitoring.ObservableOwnerCloud,
						Query:             `sum by (datname) (pg_stat_activity_count{datname!~"template.*|postgres|cloudsqladmin"})`,
						DataMayNotExist:   true,
						DataMayNotBeNaN:   false,
						Warning:           monitoring.Alert().LessOrEqual(5).For(5 * time.Minute),
						PossibleSolutions: "none",
					},
					monitoring.Observable{
						Name:              "transactions",
						Description:       "transaction durations",
						Owner:             monitoring.ObservableOwnerCloud,
						Query:             `sum by (datname) (pg_stat_activity_max_tx_duration{datname!~"template.*|postgres|cloudsqladmin"})`,
						DataMayNotExist:   true,
						DataMayNotBeNaN:   false,
						Warning:           monitoring.Alert().GreaterOrEqual(300).For(5 * time.Minute),
						Critical:          monitoring.Alert().GreaterOrEqual(500).For(5 * time.Minute),
						PanelOptions:      monitoring.PanelOptions().Unit(monitoring.Milliseconds),
						PossibleSolutions: "none",
					},
				},
				},
			},
			{
				Title: "Database and collector status", Hidden: true,
				Rows: []monitoring.Row{
					{
						monitoring.Observable{
							Name:              "postgres_up",
							Description:       "current db status",
							Owner:             monitoring.ObservableOwnerCloud,
							Query:             "pg_up",
							DataMayNotExist:   true,
							DataMayNotBeNaN:   true,
							Critical:          monitoring.Alert().LessOrEqual(0).For(5 * time.Minute),
							PossibleSolutions: "none",
						},
						monitoring.Observable{
							Name:              "pg_exporter_err",
							Description:       "errors scraping postgres exporter",
							Owner:             monitoring.ObservableOwnerCloud,
							Query:             "pg_exporter_last_scrape_error",
							DataMayNotExist:   true,
							DataMayNotBeNaN:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(1).For(5 * time.Minute),
							PossibleSolutions: "none",
						},
						monitoring.Observable{
							Name:            "migration_in_progress",
							Description:     "schema migration status (where 0 is no migration in progress)",
							Owner:           monitoring.ObservableOwnerCloud,
							Query:           "pg_sg_migration_status",
							DataMayNotExist: true,
							DataMayNotBeNaN: true,
							Critical:        monitoring.Alert().GreaterOrEqual(1).For(5 * time.Minute),
							PossibleSolutions: "The database migration has been in progress for 5 or more minutes, " +
								"please contact Sourcegraph if this persists",
						},
						// TODO(Dax): Blocked by https://github.com/sourcegraph/sourcegraph/issues/13300,  need to enable `pg_stat_statements` in Postgres conf
						//monitoring.Observable{
						//	Name:            "cache_hit_ratio",
						//	Description:     "ratio of cache hits over 5m",
						//	Owner:           monitoring.ObservableOwnerCloud,
						//	Query:           `avg(rate(pg_stat_database_blks_hit{datname!~"template.*|postgres|cloudsqladmin"}[5m]) / (rate(pg_stat_database_blks_hit{datname!~"template.*|postgres|cloudsqladmin"}[5m]) + rate(pg_stat_database_blks_read{datname!~"template.*|postgres|cloudsqladmin"}[5m]))) by (datname) * 100`,
						//	DataMayNotExist: true,
						//	DataMayNotBeNaN: true,
						//	Warning:         monitoring.Alert().LessOrEqual(0.98).For(5 * time.Minute),
						//	PossibleSolutions: "Cache hit ratio should be at least 99%, please [open an issue](https://github.com/sourcegraph/sourcegraph/issues/new/choose) " +
						//		"to add additional indexes",
						//	PanelOptions: monitoring.PanelOptions().Unit(monitoring.Percentage)},
					},
				},
			},

			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm(dbSourcegraph, monitoring.ObservableOwnerCloud),
						shared.ProvisioningMemoryUsageLongTerm(dbSourcegraph, monitoring.ObservableOwnerCloud),
					},
					{
						shared.ProvisioningCPUUsageShortTerm(dbSourcegraph, monitoring.ObservableOwnerCloud),
						shared.ProvisioningMemoryUsageShortTerm(dbSourcegraph, monitoring.ObservableOwnerCloud),
					},
					{
						shared.ProvisioningCPUUsageLongTerm(dbCodeIntel, monitoring.ObservableOwnerCodeIntel),
						shared.ProvisioningMemoryUsageLongTerm(dbCodeIntel, monitoring.ObservableOwnerCodeIntel),
					},
					{
						shared.ProvisioningCPUUsageShortTerm(dbCodeIntel, monitoring.ObservableOwnerCodeIntel),
						shared.ProvisioningMemoryUsageShortTerm(dbCodeIntel, monitoring.ObservableOwnerCodeIntel),
					},
				},
			},
		},
	}
}
