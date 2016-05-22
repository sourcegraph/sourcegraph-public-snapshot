package dbutil2

import "strings"

// getPGEnvsFromDataSource converts key value pairs from a PostgreSQL
// connection string into PG* env variables recognized by PostgreSQL
// utilites, like 'createdb'. This is useful when shelling out to a
// utility like 'createdb' which does not accept a connection string
// but reads the PG* values from the environment.
//
// The connection params currently supported are the ones supported in
// 'parseEnviron' in 'lib/pq', which is the inverse of this function and
// is used for connecting to PostgreSQL directly via the 'database/sql' API.
// (https://sourcegraph.com/github.com/lib/pq@master/-/def/GoPackage/github.com/lib/pq/-/parseEnviron)
func getPGEnvsFromDataSource(dataSource string) map[string]string {
	pgEnvs := make(map[string]string)

	for _, v := range strings.Fields(dataSource) {
		parts := strings.SplitN(v, "=", 2)

		switch parts[0] {
		case "host":
			pgEnvs["PGHOST"] = parts[1]
		case "port":
			pgEnvs["PGPORT"] = parts[1]
		case "dbname":
			pgEnvs["PGDATABASE"] = parts[1]
		case "user":
			pgEnvs["PGUSER"] = parts[1]
		case "password":
			pgEnvs["PGPASSWORD"] = parts[1]
		case "options":
			pgEnvs["PGOPTIONS"] = parts[1]
		case "application_name":
			pgEnvs["PGAPPNAME"] = parts[1]
		case "sslmode":
			pgEnvs["PGSSLMODE"] = parts[1]
		case "sslcert":
			pgEnvs["PGSSLCERT"] = parts[1]
		case "sslkey":
			pgEnvs["PGSSLKEY"] = parts[1]
		case "sslrootcert":
			pgEnvs["PGSSLROOTCERT"] = parts[1]
		case "connect_timeout":
			pgEnvs["PGCONNECT_TIMEOUT"] = parts[1]
		case "client_encoding":
			pgEnvs["PGCLIENTENCODING"] = parts[1]
		case "datestyle":
			pgEnvs["PGDATESTYLE"] = parts[1]
		case "timezone":
			pgEnvs["PGTZ"] = parts[1]
		case "geqo":
			pgEnvs["PGGEQO"] = parts[1]
		}
	}

	return pgEnvs
}
