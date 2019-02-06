package schema

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/sourcegraph/go-jsonschema/cmd/go-jsonschema-compiler
//go:generate $PWD/.bin/go-jsonschema-compiler -o schema.go -pkg schema critical.schema.json site.schema.json settings.schema.json github.schema.json

//go:generate env GO111MODULE=on go run stringdata.go -i critical.schema.json -name CriticalSchemaJSON -pkg schema -o critical_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i site.schema.json -name SiteSchemaJSON -pkg schema -o site_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i settings.schema.json -name SettingsSchemaJSON -pkg schema -o settings_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i github.schema.json -name GitHubSchemaJSON -pkg schema -o github_stringdata.go
//go:generate gofmt -s -w critical_stringdata.go site_stringdata.go settings_stringdata.go
