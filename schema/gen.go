package schema

//go:generate env GOBIN=$PWD/.bin GO111MODULE=on go install github.com/sourcegraph/go-jsonschema/cmd/go-jsonschema-compiler
//go:generate $PWD/.bin/go-jsonschema-compiler -o schema.go -pkg schema aws_codecommit.schema.json bitbucket_cloud.schema.json bitbucket_server.schema.json site.schema.json settings.schema.json github.schema.json gitlab.schema.json gitolite.schema.json other_external_service.schema.json phabricator.schema.json
//go:generate $PWD/.bin/go-jsonschema-compiler -o critical/schema.go -pkg critical critical/critical.schema.json

//go:generate env GO111MODULE=on go run stringdata.go -i aws_codecommit.schema.json -name AWSCodeCommitSchemaJSON -pkg schema -o aws_codecommit_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i bitbucket_cloud.schema.json -name BitbucketCloudSchemaJSON -pkg schema -o bitbucket_cloud_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i bitbucket_server.schema.json -name BitbucketServerSchemaJSON -pkg schema -o bitbucket_server_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i critical/critical.schema.json -name CriticalSchemaJSON -pkg critical -o critical/critical_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i site.schema.json -name SiteSchemaJSON -pkg schema -o site_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i settings.schema.json -name SettingsSchemaJSON -pkg schema -o settings_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i github.schema.json -name GitHubSchemaJSON -pkg schema -o github_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i gitlab.schema.json -name GitLabSchemaJSON -pkg schema -o gitlab_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i gitolite.schema.json -name GitoliteSchemaJSON -pkg schema -o gitolite_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i other_external_service.schema.json -name OtherExternalServiceSchemaJSON -pkg schema -o other_external_service_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i phabricator.schema.json -name PhabricatorSchemaJSON -pkg schema -o phabricator_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i campaign-types/comby.schema.json -name CombyCampaignTypeSchemaJSON -pkg schema -o campaign-types/comby_stringdata.go
//go:generate env GO111MODULE=on go run stringdata.go -i campaign-types/credentials.schema.json -name CredentialsCampaignTypeSchemaJSON -pkg schema -o campaign-types/credentials_stringdata.go
//go:generate gofmt -s -w critical/critical_stringdata.go site_stringdata.go settings_stringdata.go
