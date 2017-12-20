package schema

//go:generate go build -o .bin/schema-generate sourcegraph.com/sourcegraph/sourcegraph/vendor/github.com/sqs/generate/cmd/schema-generate
//go:generate .bin/schema-generate -o schema.go -p schema site.schema.json
//go:generate gofmt -w schema.go

//go:generate go run stringdata.go -i site.schema.json -name SiteSchemaJSON -pkg schema -o site_stringdata.go
