package legacyschema

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/sourcegraph/go-jsonschema/cmd/go-jsonschema-compiler
//go:generate $PWD/.bin/go-jsonschema-compiler -o schema.go -pkg legacyschema site.schema.json

//go:generate env GO111MODULE=on go run ../schema/stringdata.go -i site.schema.json -name SiteSchemaJSON -pkg legacyschema -o site_stringdata.go
//go:generate gofmt -w site_stringdata.go
