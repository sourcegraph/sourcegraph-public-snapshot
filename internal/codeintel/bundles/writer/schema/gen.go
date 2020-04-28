package schema

//go:generate env GO111MODULE=on go run stringdata.go -i 0.1.0-index_definitions.sql -name IndexDefinitions -pkg schema -o 0.1.0-index_definitions_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i 0.1.0-table_definitions.sql -name TableDefinitions -pkg schema -o 0.1.0-table_definitions_stringdata.go
