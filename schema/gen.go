package schema

//go:generate env GO111MODULE=on go run stringdata.go -i batch_spec.schema.json -name BatchSpecJSON -pkg schema -o batch_spec_stringdata.go
//go:generate gofmt -s -w batch_spec_stringdata.go
