package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

const (
	dbCodeIntel   = "codeintel-db"
	dbSourcegraph = "pgsql"
)

func Postgres() *Container {
	return &Container{
		Name:        "postgres",
		Title:       "Postgres",
		Description: "Metrics from postgres_exporter.",
		Groups: []Group{
			{
				Title: "Default postgres dashboard",
				Rows: []Row{
					{
						Observable{
							Name:              "connections",
							Description:       "connections",
							Owner:             ObservableOwnerCloud,
							Query:             "sum by (datname) (pg_stat_activity_count{datname!~\"template.*|postgres|cloudsqladmin\"})",
							DataMayNotExist:   true,
							DataMayNotBeNaN:   false,
							Warning:           Alert().LessOrEqual(5).For(5 * time.Minute),
							PossibleSolutions: "none",
						},
						Observable{
							Name:              "transactions",
							Description:       "transaction_durations",
							Owner:             ObservableOwnerCloud,
							Query:             "sum by (datname) (pg_stat_activity_max_tx_duration{datname!~\"template.*|postgres|cloudsqladmin\"})",
							DataMayNotExist:   true,
							DataMayNotBeNaN:   false,
							Warning:           Alert().GreaterOrEqual(300).For(5 * time.Minute),
							Critical:          Alert().GreaterOrEqual(500).For(5 * time.Minute),
							PanelOptions:      PanelOptions().Unit(Milliseconds),
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  "Database and collector status",
				Hidden: true,
				Rows: []Row{{
					Observable{
						Name:              "postgres_up",
						Description:       "current db status",
						Owner:             ObservableOwnerCloud,
						Query:             "pg_up",
						DataMayNotExist:   true,
						DataMayNotBeNaN:   true,
						Critical:          Alert().LessOrEqual(0).For(5 * time.Minute),
						PossibleSolutions: "none",
					},
					Observable{
						Name:              "pg_exporter_err",
						Description:       "error scraping postgres exporter",
						Owner:             ObservableOwnerCloud,
						Query:             "pg_exporter_last_scrape_error",
						DataMayNotExist:   true,
						DataMayNotBeNaN:   true,
						Warning:           Alert().GreaterOrEqual(1).For(5 * time.Minute),
						PossibleSolutions: "none",
					},
					Observable{
						Name:            "database_migration_status",
						Description:     "schema migration status",
						Owner:           ObservableOwnerCloud,
						Query:           "pg_sg_db_migrations",
						DataMayNotExist: true,
						Critical:        Alert().GreaterOrEqual(1).For(5 * time.Minute),
						PossibleSolutions: "A database migration has been in progress for 5 minutes," +
							" ensure that the migration has succeeded or contact Sourcegraph",
					},
					Observable{
						Name:              "cache_hit_ratio",
						Description:       "ratio of cache hits (should be 99%)",
						Owner:             ObservableOwnerCloud,
						Query:             "avg(rate(pg_stat_database_blks_hit{datname!~\"template.*|postgres|cloudsqladmin\"}[5m]) / (rate(pg_stat_database_blks_hit{datname!~\"template.*|postgres|cloudsqladmin\"}[5m]) + rate(pg_stat_database_blks_read{datname!~\"template.*|postgres|cloudsqladmin\"}[5m]))) by (datname)",
						DataMayNotExist:   false,
						DataMayNotBeNaN:   false,
						Warning:           Alert().LessOrEqual(0.98).For(5 * time.Minute),
						PossibleSolutions: "none",
					},
				}},
			},

			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsageLongTerm(db, ObservableOwnerCloud),
						sharedProvisioningMemoryUsageLongTerm(db, ObservableOwnerCloud),
					},
					{
						sharedProvisioningCPUUsageShortTerm(db, ObservableOwnerCloud),
						sharedProvisioningMemoryUsageShortTerm(db, ObservableOwnerCloud),
					},
					{
						sharedProvisioningCPUUsageLongTerm(codeintel, ObservableOwnerCodeIntel),
						sharedProvisioningMemoryUsageLongTerm(codeintel, ObservableOwnerCodeIntel),
					},
					{
						sharedProvisioningCPUUsageShortTerm(codeintel, ObservableOwnerCodeIntel),
						sharedProvisioningMemoryUsageShortTerm(codeintel, ObservableOwnerCodeIntel),
					},
				},
			},
		},
	}
}
