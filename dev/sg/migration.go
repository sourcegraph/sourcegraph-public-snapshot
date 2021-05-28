package main

var (
	databaseNames = []string{
		"frontend",
		"codeintel-db",
		"codeinsights-db",
	}

	defaultDatabaseName = databaseNames[0]
)

func isValidDatabaseName(name string) bool {
	for _, candidate := range databaseNames {
		if candidate == name {
			return true
		}
	}

	return false
}
