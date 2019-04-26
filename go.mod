module github.com/sourcegraph/sourcegraph

require (
	cloud.google.com/go v0.34.0
	github.com/NYTimes/gziphandler v1.1.1
	github.com/aws/aws-sdk-go-v2 v0.7.0
	github.com/beevik/etree v0.0.0-20180609182452-90dafc1e1f11
	github.com/boj/redistore v0.0.0-20160128113310-fc113767cd6b
	github.com/certifi/gocertifi v0.0.0-20190105021324-abcd57078448 // indirect
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/coreos/go-oidc v0.0.0-20171002155002-a93f71fdfe73
	github.com/coreos/go-semver v0.2.0
	github.com/crewjam/saml v0.0.0-20180831135026-ebc5f787b786
	github.com/davecgh/go-spew v1.1.1
	github.com/daviddengcn/go-colortext v0.0.0-20190211032704-186a3d44e920
	github.com/dghubble/gologin v1.0.2-0.20181013174641-0e442dd5bb73
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/dnaeon/go-vcr v1.0.1
	github.com/docker/docker v0.7.3-0.20190108045446-77df18c24acf
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
	github.com/getsentry/raven-go v0.2.0
	github.com/ghodss/yaml v1.0.0
	github.com/gin-contrib/sse v0.0.0-20190301062529-5545eab6dad3 // indirect
	github.com/gin-gonic/gin v1.3.0 // indirect
	github.com/gitchander/permutation v0.0.0-20181107151852-9e56b92e9909
	github.com/go-redsync/redsync v1.0.1
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang-migrate/migrate/v4 v4.2.3
	github.com/golang/gddo v0.0.0-20181116215533-9bd4a3295021
	github.com/golang/groupcache v0.0.0-20180513044358-24b0969c4cb7
	github.com/golangci/golangci-lint v1.16.0
	github.com/golangplus/bytes v0.0.0-20160111154220-45c989fe5450 // indirect
	github.com/golangplus/fmt v0.0.0-20150411045040-2a5d6d7d2995 // indirect
	github.com/golangplus/testing v0.0.0-20180327235837-af21d9c3145e // indirect
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/google/go-cmp v0.2.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.1.0
	github.com/google/zoekt v0.0.0-20180530125106-8e284ca7e964
	github.com/gorilla/context v1.1.1
	github.com/gorilla/csrf v1.5.1
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.7.0
	github.com/gorilla/schema v1.0.2
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.1.4-0.20181015005113-68d1edeb366b
	github.com/goware/urlx v0.2.0
	github.com/graph-gophers/graphql-go v0.0.0-20190204230732-e582242c92cc
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc
	github.com/hashicorp/go-multierror v1.0.0
	github.com/honeycombio/libhoney-go v1.8.1
	github.com/jmoiron/sqlx v1.2.0
	github.com/joho/godotenv v1.3.0
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1
	github.com/karlseguin/expect v1.0.1 // indirect
	github.com/karlseguin/typed v1.1.7 // indirect
	github.com/karrick/tparse v2.4.2+incompatible
	github.com/keegancsmith/rpc v1.1.0
	github.com/keegancsmith/sqlf v1.1.0
	github.com/keegancsmith/tmpfriend v0.0.0-20180423180255-86e88902a513
	github.com/kevinburke/differ v0.0.0-20181006040839-bdfd927653c8
	github.com/kevinburke/go-bindata v3.13.0+incompatible
	github.com/kr/text v0.1.0
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/leanovate/gopter v0.2.4
	github.com/lib/pq v1.0.0
	github.com/lightstep/lightstep-tracer-go v0.15.6
	github.com/mattn/go-colorable v0.1.0 // indirect
	github.com/mattn/go-sqlite3 v1.10.0
	github.com/mattn/goreman v0.2.1
	github.com/mcuadros/go-version v0.0.0-20180611085657-6d5863ca60fa
	github.com/microcosm-cc/bluemonday v1.0.2
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/neelance/parallel v0.0.0-20160708114440-4de9ce63d14c
	github.com/onsi/ginkgo v1.7.0 // indirect
	github.com/onsi/gomega v1.4.3 // indirect
	github.com/opentracing-contrib/go-stdlib v0.0.0-20190324214902-3020fec0e66b
	github.com/opentracing/opentracing-go v1.1.0
	github.com/peterhellberg/link v1.0.0
	github.com/pkg/errors v0.8.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/client_golang v0.9.2
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
	github.com/prometheus/common v0.2.0 // indirect
	github.com/prometheus/procfs v0.0.0-20190209105433-f8d8b3f739bd // indirect
	github.com/russellhaering/gosaml2 v0.3.2-0.20190403162508-649841e7f48a
	github.com/russellhaering/goxmldsig v0.0.0-20180430223755-7acd5e4a6ef7
	github.com/sergi/go-diff v1.0.0
	github.com/shurcooL/github_flavored_markdown v0.0.0-20181002035957-2122de532470
	github.com/shurcooL/go v0.0.0-20190121191506-3fef8c783dec // indirect
	github.com/shurcooL/go-goon v0.0.0-20170922171312-37c2f522c041
	github.com/shurcooL/highlight_diff v0.0.0-20181222201841-111da2e7d480 // indirect
	github.com/shurcooL/highlight_go v0.0.0-20181215221002-9d8641ddf2e1 // indirect
	github.com/shurcooL/httpfs v0.0.0-20181222201310-74dc9339e414
	github.com/shurcooL/httpgzip v0.0.0-20180522190206-b1c53ac65af9 // indirect
	github.com/shurcooL/octicon v0.0.0-20181222203144-9ff1a4cf27f4 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/sloonz/go-qprintable v0.0.0-20160203160305-775b3a4592d5 // indirect
	github.com/sourcegraph/annotate v0.0.0-20160123013949-f4cad6c6324d // indirect
	github.com/sourcegraph/ctxvfs v0.0.0-20180418081416-2b65f1b1ea81
	github.com/sourcegraph/docsite v0.0.0-20190303064655-ad3087aa6352
	github.com/sourcegraph/go-diff v0.5.1-0.20190324182542-c825d9a1a0ca
	github.com/sourcegraph/go-jsonschema v0.0.0-20190205151546-7939fa138765
	github.com/sourcegraph/go-langserver v2.0.1-0.20181108233942-4a51fa2e1238+incompatible
	github.com/sourcegraph/go-lsp v0.0.0-20181119182933-0c7d621186c1
	github.com/sourcegraph/gosyntect v0.0.0-20190422212758-2070e25f4131
	github.com/sourcegraph/jsonx v0.0.0-20190114210550-ba8cb36a8614
	github.com/sourcegraph/syntaxhighlight v0.0.0-20170531221838-bd320f5d308e // indirect
	github.com/sqs/httpgzip v0.0.0-20180622165210-91da61ed4dff
	github.com/stripe/stripe-go v0.0.0-20181128170521-1436b6008c5e
	github.com/stvp/tempredis v0.0.0-20190104202742-b82af8480203 // indirect
	github.com/temoto/robotstxt-go v0.0.0-20180810133444-97ee4a9ee6ea
	github.com/uber-go/atomic v1.3.2 // indirect
	github.com/uber/gonduit v0.3.2
	github.com/uber/jaeger-client-go v2.14.0+incompatible
	github.com/uber/jaeger-lib v1.5.0
	github.com/ugorji/go v1.1.4 // indirect
	github.com/wsxiaoys/terminal v0.0.0-20160513160801-0940f3fc43a0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20180816142147-da425ebb7609
	github.com/xeonx/timeago v1.0.0-rc3
	github.com/zenazn/goji v0.9.0 // indirect
	go.opencensus.io v0.19.0 // indirect
	go.uber.org/atomic v1.3.2 // indirect
	golang.org/x/crypto v0.0.0-20190320223903-b7391e95e576
	golang.org/x/net v0.0.0-20190322120337-addf6b3196f6
	golang.org/x/oauth2 v0.0.0-20190212230446-3e8b2be13635
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4
	golang.org/x/sys v0.0.0-20190322080309-f49334f85ddc
	golang.org/x/time v0.0.0-20190401211219-9d24e82272b4
	golang.org/x/tools v0.0.0-20190322203728-c1a832b0ad89
	google.golang.org/api v0.1.0 // indirect
	google.golang.org/genproto v0.0.0-20190215211957-bd968387e4aa // indirect
	google.golang.org/grpc v1.18.0 // indirect
	gopkg.in/alexcesaro/statsd.v2 v2.0.0 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/inconshreveable/log15.v2 v2.0.0-20180818164646-67afb5ed74ec
	gopkg.in/jpoehls/gophermail.v0 v0.0.0-20160410235621-62941eab772c
	gopkg.in/karlseguin/expect.v1 v1.0.1 // indirect
	gopkg.in/square/go-jose.v2 v2.1.9 // indirect
	gopkg.in/src-d/go-git.v4 v4.8.0
	gopkg.in/yaml.v2 v2.2.2
	sourcegraph.com/sqs/pbtypes v1.0.0 // indirect
)

replace (
	github.com/google/zoekt => github.com/sourcegraph/zoekt v0.0.0-20190116094554-c742ce874aa3
	github.com/graph-gophers/graphql-go => github.com/sourcegraph/graphql-go v0.0.0-20180929065141-c790ffc3c46a
	github.com/mattn/goreman => github.com/sourcegraph/goreman v0.1.2-0.20180928223752-6e9a2beb830d
	github.com/uber/gonduit => github.com/sourcegraph/gonduit v0.4.0
)

replace github.com/dghubble/gologin => github.com/sourcegraph/gologin v1.0.2-0.20181110030308-c6f1b62954d8

replace gopkg.in/russross/blackfriday.v2 v2.0.1 => github.com/russross/blackfriday/v2 v2.0.1
