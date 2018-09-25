package schema

//go:generate go build -o .bin/go-jsonschema-compiler github.com/sourcegraph/sourcegraph/vendor/github.com/sourcegraph/go-jsonschema/cmd/go-jsonschema-compiler
//go:generate .bin/go-jsonschema-compiler -o schema.go -pkg schema site.schema.json settings.schema.json extension.schema.json

//go:generate go run stringdata.go -i site.schema.json -name SiteSchemaJSON -pkg schema -o site_stringdata.go
//go:generate go run stringdata.go -i settings.schema.json -name SettingsSchemaJSON -pkg schema -o settings_stringdata.go
//go:generate go run stringdata.go -i extension.schema.json -name ExtensionSchemaJSON -pkg schema -o extension_stringdata.go
//go:generate gofmt -w site_stringdata.go settings_stringdata.go extension_stringdata.go
