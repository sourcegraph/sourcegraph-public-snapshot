package schema

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/sourcegraph/go-jsonschema/cmd/go-jsonschema-compiler
//go:generate $PWD/.bin/go-jsonschema-compiler -o schema.go -pkg schema site.schema.json settings.schema.json extension.schema.json

//go:generate env GO111MODULE=on go run stringdata.go -i site.schema.json -name SiteSchemaJSON -pkg schema -o site_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i settings.schema.json -name SettingsSchemaJSON -pkg schema -o settings_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i extension.schema.json -name ExtensionSchemaJSON -pkg schema -o extension_stringdata.go
//go:generate gofmt -w site_stringdata.go settings_stringdata.go extension_stringdata.go
