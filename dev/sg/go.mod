module github.com/sourcegraph/sourcegraph/dev/sg

go 1.16

require (
	cloud.google.com/go v0.86.0 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/cockroachdb/errors v1.8.6
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/go-cmp v0.5.6
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jackc/pgx/v4 v4.11.0
	github.com/peterbourgon/ff/v3 v3.0.0
	github.com/rjeczalik/notify v0.9.2
	github.com/sourcegraph/sourcegraph/lib v0.0.0-20210906140940-dd601b549e29
	golang.org/x/mod v0.4.2
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	google.golang.org/api v0.50.0
	google.golang.org/genproto v0.0.0-20210708141623-e76da96a951f // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/sourcegraph/sourcegraph/lib => ./../../lib
