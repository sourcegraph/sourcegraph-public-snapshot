package main

import "sort"

type Database struct {
	Name           string
	MigrationTable string
	DataTables     []string
}

var (
	frontendDatabase = Database{
		Name:           "frontend",
		MigrationTable: "schema_migrations",
		DataTables:     []string{"out_of_band_migrations"},
	}

	codeIntelDatabase = Database{
		Name:           "codeintel",
		MigrationTable: "codeintel_schema_migrations",
		DataTables:     nil,
	}

	codeInsightsDatabase = Database{
		Name:           "codeinsights",
		MigrationTable: "codeinsights_schema_migrations",
		DataTables:     nil,
	}

	databases = []Database{
		frontendDatabase,
		codeIntelDatabase,
		codeInsightsDatabase,
	}

	defaultDatabase = databases[0]
)

func databaseNames() []string {
	databaseNames := make([]string, 0, len(databases))
	for _, database := range databases {
		databaseNames = append(databaseNames, database.Name)
	}
	sort.Strings(databaseNames)

	return databaseNames
}

func databaseByName(name string) (Database, bool) {
	for _, database := range databases {
		if database.Name == name {
			return database, true
		}
	}

	return Database{}, false
}
