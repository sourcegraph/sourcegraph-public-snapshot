package main

import "time"

const (
	codeintel = "codeintel-db"
	db        = "pgsql"
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
							PanelOptions:      panelOptions{},
						},
						Observable{
							Name:              "transactions",
							Description:       "transaction_durations",
							Owner:             ObservableOwnerCloud,
							Query:             "sum by (datname) (pg_stat_activity_max_tx_duration{datname!~\"template.*|postgres|cloudsqladmin\"})",
							DataMayNotExist:   true,
							DataMayNotBeNaN:   false,
							Warning:           Alert().GreaterOrEqual(300),
							PossibleSolutions: "none",
							PanelOptions:      panelOptions{},
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
						PanelOptions:      panelOptions{},
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
						PanelOptions:      panelOptions{},
					},
				},
					{
						postgresVersion(db, ObservableOwnerCloud),
						postgresMaxConnections(db, ObservableOwnerCloud),
						postgresSharedBuffers(db, ObservableOwnerCloud),
						postgresEffectiveCacheBytes(db, ObservableOwnerCloud)},
				},
			},

			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsageLongTerm("pgsql", ObservableOwnerCloud),
						sharedProvisioningMemoryUsageLongTerm("pgsql", ObservableOwnerCloud),
					},
					{
						sharedProvisioningCPUUsageShortTerm("pgsql", ObservableOwnerCloud),
						sharedProvisioningMemoryUsageShortTerm("pgsql", ObservableOwnerCloud),
					},
				},
			},
		},
	}
}
