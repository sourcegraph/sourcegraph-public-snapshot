module github.com/sourcegraph/sourcegraph/enterprise/dev/ci

go 1.17

require (
	github.com/cockroachdb/errors v1.8.6
	github.com/ghodss/yaml v1.0.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/sourcegraph/sourcegraph v0.0.0-20211029232645-dd300fa4351e
	github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images v0.0.0-20211020041242-9f6088e5b163
)

require (
	github.com/cockroachdb/logtags v0.0.0-20190617123548-eb05cc24525f // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/cockroachdb/sentry-go v0.6.1-cockroachdb.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/sourcegraph/sourcegraph => ./../../../
