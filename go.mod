module github.com/sourcegraph/sourcegraph

require (
	github.com/NYTimes/gziphandler v1.0.1
	github.com/aws/aws-sdk-go-v2 v2.0.0-preview.4+incompatible
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973 // indirect
	github.com/boj/redistore v0.0.0-20160128113310-fc113767cd6b
	github.com/certifi/gocertifi v0.0.0-20180118203423-deb3ae2ef261 // indirect
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/coreos/go-semver v0.2.0
	github.com/davecgh/go-spew v1.1.1
	github.com/daviddengcn/go-colortext v0.0.0-20171126034257-17e75f6184bc
	github.com/emersion/go-imap v1.0.0-beta.1
	github.com/emersion/go-sasl v0.0.0-20161116183048-7e096a0a6197 // indirect
	github.com/ericchiang/k8s v1.2.0
	github.com/etdub/goparsetime v0.0.0-20160315173935-ea17b0ac3318 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/facebookgo/ensure v0.0.0-20160127193407-b4ab57deab51 // indirect
	github.com/facebookgo/limitgroup v0.0.0-20150612190941-6abd8d71ec01 // indirect
	github.com/facebookgo/muster v0.0.0-20150708232844-fd3d7953fd52 // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/subset v0.0.0-20150612182917-8dac2c3c4870 // indirect
	github.com/fatih/astrewrite v0.0.0-20180730114054-bef4001d457f
	github.com/fatih/color v1.7.0
	github.com/felixfbecker/stringscore v0.0.0-20170928081130-e71a9f1b0749
	github.com/felixge/httpsnoop v1.0.0
	github.com/garyburd/redigo v1.6.0
	github.com/gchaincl/sqlhooks v1.1.0
	github.com/getsentry/raven-go v0.0.0-20180903072508-084a9de9eb03
	github.com/ghodss/yaml v1.0.0
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gobwas/glob v0.0.0-20180809073612-f756513aec94
	github.com/gogo/protobuf v0.0.0-20170330071051-c0656edd0d9e // indirect
	github.com/golang-migrate/migrate v0.0.0-20180905051849-93d53a5ae84d
	github.com/golang/groupcache v0.0.0-20180513044358-24b0969c4cb7
	github.com/golangplus/bytes v0.0.0-20160111154220-45c989fe5450 // indirect
	github.com/golangplus/fmt v0.0.0-20150411045040-2a5d6d7d2995 // indirect
	github.com/golangplus/testing v0.0.0-20180327235837-af21d9c3145e // indirect
	github.com/google/go-querystring v0.0.0-20170111101155-53e6ce116135
	github.com/google/uuid v1.0.0
	github.com/google/zoekt v0.0.0-20180530125106-8e284ca7e964
	github.com/gorilla/context v1.1.1
	github.com/gorilla/csrf v1.5.1
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.6.2
	github.com/gorilla/schema v1.0.2
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.1.2
	github.com/gorilla/websocket v1.4.0
	github.com/graph-gophers/graphql-go v0.0.0-20180806175703-94da0f0031f9
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7
	github.com/hashicorp/go-multierror v1.0.0
	github.com/honeycombio/libhoney-go v1.7.0
	github.com/joho/godotenv v1.3.0
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1
	github.com/karrick/tparse v2.4.2+incompatible
	github.com/keegancsmith/sqlf v1.0.0
	github.com/keegancsmith/tmpfriend v0.0.0-20180423180255-86e88902a513
	github.com/kevinburke/differ v0.0.0-20181006040839-bdfd927653c8
	github.com/kevinburke/go-bindata v3.11.1-0.20180909202705-9b44e0539c2a+incompatible
	github.com/kisielk/gotool v1.0.0 // indirect
	github.com/kr/text v0.1.0
	github.com/lib/pq v1.0.0
	github.com/lightstep/lightstep-tracer-go v0.15.4
	github.com/mattn/goreman v0.1.2-0.20180926031137-83bee30f0a15
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/microcosm-cc/bluemonday v1.0.1
	github.com/neelance/parallel v0.0.0-20160708114440-4de9ce63d14c
	github.com/opentracing-contrib/go-stdlib v0.0.0-20180702182724-07a764486eb1
	github.com/opentracing/basictracer-go v1.0.0
	github.com/opentracing/opentracing-go v1.0.2
	github.com/peterhellberg/link v1.0.0
	github.com/pkg/errors v0.8.0
	github.com/prometheus/client_golang v0.8.0
	github.com/prometheus/client_model v0.0.0-20171117100541-99fa1f4be8e5 // indirect
	github.com/prometheus/common v0.0.0-20180801064454-c7de2306084e // indirect
	github.com/prometheus/procfs v0.0.0-20180725123919-05ee40e3a273 // indirect
	github.com/russross/blackfriday v0.0.0-20180829180401-f1f45ab762c2 // indirect
	github.com/shurcooL/github_flavored_markdown v0.0.0-20180602233135-8913699a52e3
	github.com/shurcooL/go v0.0.0-20180423040247-9e1955d9fb6e // indirect
	github.com/shurcooL/go-goon v0.0.0-20170922171312-37c2f522c041
	github.com/shurcooL/highlight_diff v0.0.0-20170515013008-09bb4053de1b // indirect
	github.com/shurcooL/highlight_go v0.0.0-20170515013102-78fb10f4a5f8 // indirect
	github.com/shurcooL/httpfs v0.0.0-20171119174359-809beceb2371
	github.com/shurcooL/httpgzip v0.0.0-20180522190206-b1c53ac65af9 // indirect
	github.com/shurcooL/octicon v0.0.0-20180602230221-c42b0e3b24d9 // indirect
	github.com/shurcooL/sanitized_anchor_name v0.0.0-20170918181015-86672fcb3f95 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20180825020608-02ddb050ef6b
	github.com/sloonz/go-qprintable v0.0.0-20160203160305-775b3a4592d5 // indirect
	github.com/sourcegraph/annotate v0.0.0-20160123013949-f4cad6c6324d // indirect
	github.com/sourcegraph/ctxvfs v0.0.0-20180418081416-2b65f1b1ea81
	github.com/sourcegraph/go-jsonschema v0.0.0-20180805125535-0e659b54484d
	github.com/sourcegraph/go-langserver v0.0.0-20180817131207-7df19dc017ef
	github.com/sourcegraph/godockerize v0.0.0-20180919081208-f4fb18a2ab18
	github.com/sourcegraph/gosyntect v0.0.0-20180604231642-c01be3625b10
	github.com/sourcegraph/httpcache v0.0.0-20160524185540-16db777d8ebe
	github.com/sourcegraph/jsonrpc2 v0.0.0-20180831160525-549eb959f029
	github.com/sourcegraph/jsonx v0.0.0-20180801091521-5a4ae5eb18cd
	github.com/sourcegraph/syntaxhighlight v0.0.0-20170531221838-bd320f5d308e // indirect
	github.com/sqs/httpgzip v0.0.0-20180622165210-91da61ed4dff
	github.com/stvp/tempredis v0.0.0-20160122230306-83f7aae7ea49 // indirect
	github.com/temoto/robotstxt-go v0.0.0-20180810133444-97ee4a9ee6ea
	github.com/uber-go/atomic v1.3.2 // indirect
	github.com/uber/jaeger-client-go v2.14.0+incompatible
	github.com/uber/jaeger-lib v1.5.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20180816142147-da425ebb7609
	go.uber.org/atomic v1.3.2 // indirect
	golang.org/x/crypto v0.0.0-20180904163835-0709b304e793
	golang.org/x/net v0.0.0-20180906233101-161cd47e91fd
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys v0.0.0-20180925112736-b09afc3d579e
	golang.org/x/time v0.0.0-20180412165947-fbb02b2291d2
	golang.org/x/tools v0.0.0-20181001162950-8deeabbe2e53
	google.golang.org/grpc v1.14.0 // indirect
	gopkg.in/alexcesaro/statsd.v2 v2.0.0 // indirect
	gopkg.in/inconshreveable/log15.v2 v2.0.0-20180818164646-67afb5ed74ec
	gopkg.in/jpoehls/gophermail.v0 v0.0.0-20160410235621-62941eab772c
	gopkg.in/redsync.v1 v1.0.1
	gopkg.in/src-d/go-git.v4 v4.6.0
	gopkg.in/urfave/cli.v2 v2.0.0-20180128182452-d3ae77c26ac8 // indirect
	gopkg.in/yaml.v2 v2.2.1
	honnef.co/go/tools v0.0.0-20180728063816-88497007e858
	sourcegraph.com/sourcegraph/go-diff v0.0.0-20171119081133-3f415a150aec
	sourcegraph.com/sqs/pbtypes v0.0.0-20180604144634-d3ebe8f20ae4 // indirect
)

replace (
	github.com/google/zoekt => github.com/sourcegraph/zoekt v0.0.0-20180925141536-852b3842c11d
	github.com/russellhaering/gosaml2 => github.com/sourcegraph/gosaml2 v0.0.0-20180820053343-1b78a6b41538
)

replace github.com/shurcooL/vfsgen => github.com/beyang/vfsgen v0.0.0-20180926055532-04927b934b6e

replace github.com/mattn/goreman => github.com/sourcegraph/goreman v0.1.2-0.20180928223752-6e9a2beb830d

replace github.com/graph-gophers/graphql-go => github.com/sourcegraph/graphql-go v0.0.0-20180929065141-c790ffc3c46a
