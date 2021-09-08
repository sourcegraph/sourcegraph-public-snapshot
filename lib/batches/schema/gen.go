package schema

//go:generate env GO111MODULE=on go run stringdata.go -i ../../../schema/batch_spec.schema.json -name BatchSpecJSON -pkg schema -o batch_spec_stringdata.go
//go:generate gofmt -s -w batch_spec_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i ../../../schema/changeset_spec.schema.json -name ChangesetSpecJSON -pkg schema -o changeset_spec_stringdata.go
//go:generate gofmt -s -w changeset_spec_stringdata.go
