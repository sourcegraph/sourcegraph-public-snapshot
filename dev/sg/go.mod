module github.com/sourcegraph/sourcegraph/dev/sg

go 1.16

require (
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/peterbourgon/ff/v3 v3.0.0
	github.com/pkg/errors v0.9.1
	github.com/rjeczalik/notify v0.9.2
	github.com/sourcegraph/sourcegraph/lib v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/sourcegraph/sourcegraph/lib v0.0.0-00010101000000-000000000000 => ./../../lib
