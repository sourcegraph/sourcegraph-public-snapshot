module github.com/sourcegraph/sourcegraph/dev/sg

go 1.17

require (
	github.com/Masterminds/semver v1.5.0
	github.com/buildkite/go-buildkite/v3 v3.0.0
	github.com/cockroachdb/errors v1.8.6
	github.com/gen2brain/beeep v0.0.0-20210529141713-5586760f0cc1
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/google/go-cmp v0.5.6
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jackc/pgx/v4 v4.11.0
	github.com/peterbourgon/ff/v3 v3.0.0
	github.com/rjeczalik/notify v0.9.3-0.20210809113154-3472d85e95cd
	github.com/slack-go/slack v0.9.5
	github.com/sourcegraph/sourcegraph/lib v0.0.0-20210906140940-dd601b549e29
	golang.org/x/mod v0.4.2
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	google.golang.org/api v0.50.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	cloud.google.com/go v0.84.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cockroachdb/logtags v0.0.0-20190617123548-eb05cc24525f // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/cockroachdb/sentry-go v0.6.1-cockroachdb.2 // indirect
	github.com/go-toast/toast v0.0.0-20190211030409-01e6764cf0a4 // indirect
	github.com/godbus/dbus/v5 v5.0.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/gopherjs/gopherwasm v1.1.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.8.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.0.6 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.7.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.8.0 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.12 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.1.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/tadvi/systray v0.0.0-20190226123456-11a2b8fa57af // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/sys v0.0.0-20211007075335-d3039528d8ac // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210701133433-6b8dcf568a95 // indirect
	google.golang.org/grpc v1.39.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

replace github.com/sourcegraph/sourcegraph/lib => ./../../lib
