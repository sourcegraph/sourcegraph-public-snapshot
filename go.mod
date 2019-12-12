module github.com/sourcegraph/sourcegraph

go 1.13

require (
	cloud.google.com/go v0.49.0 // indirect
	cloud.google.com/go/bigquery v1.3.0 // indirect
	cloud.google.com/go/pubsub v1.1.0
	cloud.google.com/go/storage v1.4.0 // indirect
	github.com/DataDog/zstd v1.4.4 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/NYTimes/gziphandler v1.1.1
	github.com/RoaringBitmap/roaring v0.4.21
	github.com/avelino/slugify v0.0.0-20180501145920-855f152bd774
	github.com/aws/aws-sdk-go-v2 v0.17.0
	github.com/beevik/etree v1.1.0
	github.com/boj/redistore v0.0.0-20180917114910-cd5dcc76aeff
	github.com/bombsimon/wsl v1.2.8 // indirect
	github.com/certifi/gocertifi v0.0.0-20191021191039-0944d244cd40 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/containerd/containerd v1.3.1 // indirect
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/coreos/go-semver v0.3.0
	github.com/cosiner/argv v0.0.1 // indirect
	github.com/crewjam/saml v0.3.1
	github.com/davecgh/go-spew v1.1.1
	github.com/daviddengcn/go-colortext v0.0.0-20180409174941-186a3d44e920
	github.com/dghubble/gologin v2.2.0+incompatible
	github.com/dhui/dktest v0.3.1 // indirect
	github.com/dnaeon/go-vcr v1.0.1
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.7.3-0.20190817195342-4760db040282
	github.com/docker/go-units v0.4.0 // indirect
	github.com/emersion/go-imap v1.0.1
	github.com/emersion/go-sasl v0.0.0-20190817083125-240c8404624e // indirect
	github.com/ericchiang/k8s v1.2.0
	github.com/etdub/goparsetime v0.0.0-20160315173935-ea17b0ac3318 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/facebookgo/ensure v0.0.0-20160127193407-b4ab57deab51 // indirect
	github.com/facebookgo/limitgroup v0.0.0-20150612190941-6abd8d71ec01 // indirect
	github.com/facebookgo/muster v0.0.0-20150708232844-fd3d7953fd52 // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/subset v0.0.0-20150612182917-8dac2c3c4870 // indirect
	github.com/fatih/astrewrite v0.0.0-20190527122930-f5295d6854fb
	github.com/fatih/color v1.7.0
	github.com/felixfbecker/stringscore v0.0.0-20170928081130-e71a9f1b0749
	github.com/felixge/httpsnoop v1.0.1
	github.com/gchaincl/sqlhooks v1.3.0
	github.com/getsentry/raven-go v0.2.0
	github.com/ghodss/yaml v1.0.0
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.4.0 // indirect
	github.com/gitchander/permutation v0.0.0-20181107151852-9e56b92e9909
	github.com/glycerine/go-unsnap-stream v0.0.0-20190901134440-81cf024a9e0a // indirect
	github.com/go-delve/delve v1.3.2
	github.com/go-redsync/redsync v1.3.1
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang-migrate/migrate/v4 v4.7.0
	github.com/golang/gddo v0.0.0-20190904175337-72a348e765d2
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9
	github.com/golangci/gocyclo v0.0.0-20180528144436-0a533e8fa43d // indirect
	github.com/golangci/golangci-lint v1.21.0
	github.com/golangci/revgrep v0.0.0-20180812185044-276a5c0a1039 // indirect
	github.com/golangplus/bytes v0.0.0-20160111154220-45c989fe5450 // indirect
	github.com/golangplus/fmt v0.0.0-20150411045040-2a5d6d7d2995 // indirect
	github.com/golangplus/testing v0.0.0-20180327235837-af21d9c3145e // indirect
	github.com/gomodule/oauth1 v0.0.0-20181215000758-9a59ed3b0a84
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-cmp v0.3.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v28 v28.1.1
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.1.1
	github.com/google/zoekt v0.0.0-20191031082336-5851b63aed4d
	github.com/gopherjs/gopherjs v0.0.0-20191106031601-ce3c9ade29de // indirect
	github.com/gorilla/context v1.1.1
	github.com/gorilla/csrf v1.6.2
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/schema v1.1.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.0
	github.com/gostaticanalysis/analysisutil v0.0.3 // indirect
	github.com/goware/urlx v0.3.1
	github.com/graph-gophers/graphql-go v0.0.0-20191115155744-f33e81362277
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/hashicorp/go-hclog v0.10.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-retryablehttp v0.6.3 // indirect
	github.com/honeycombio/libhoney-go v1.12.2
	github.com/inconshreveable/log15 v0.0.0-20180818164646-67afb5ed74ec
	github.com/jmoiron/sqlx v1.2.0
	github.com/joho/godotenv v1.3.0
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/karlseguin/expect v1.0.1 // indirect
	github.com/karlseguin/typed v1.1.7 // indirect
	github.com/karrick/godirwalk v1.13.0
	github.com/karrick/tparse/v2 v2.7.1
	github.com/keegancsmith/sqlf v1.1.0
	github.com/keegancsmith/tmpfriend v0.0.0-20180423180255-86e88902a513
	github.com/kevinburke/go-bindata v3.16.0+incompatible
	github.com/klauspost/compress v1.9.2 // indirect
	github.com/kr/text v0.1.0
	github.com/kylelemons/godebug v1.1.0
	github.com/leanovate/gopter v0.2.4
	github.com/lib/pq v1.2.0
	github.com/lightstep/lightstep-tracer-go v0.18.1
	github.com/mattn/go-runewidth v0.0.6 // indirect
	github.com/mattn/go-sqlite3 v1.13.0
	github.com/mattn/goreman v0.3.4
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/microcosm-cc/bluemonday v1.0.2
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/neelance/parallel v0.0.0-20160708114440-4de9ce63d14c
	github.com/opentracing-contrib/go-stdlib v0.0.0-20190519235532-cf7a6c988dc9
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/peterbourgon/ff v1.6.1-0.20190916204019-6cd704ec2eeb
	github.com/peterh/liner v1.1.0 // indirect
	github.com/peterhellberg/link v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/pkg/profile v1.4.0 // indirect
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/client_golang v1.2.1
	github.com/prometheus/procfs v0.0.7 // indirect
	github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be
	github.com/russellhaering/gosaml2 v0.3.2-0.20190403162508-649841e7f48a
	github.com/russellhaering/goxmldsig v0.0.0-20180430223755-7acd5e4a6ef7
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/securego/gosec v0.0.0-20191119104125-df484bfa9e9f // indirect
	github.com/segmentio/fasthash v1.0.1
	github.com/sergi/go-diff v1.0.0
	github.com/shurcooL/github_flavored_markdown v0.0.0-20181002035957-2122de532470
	github.com/shurcooL/go v0.0.0-20190704215121-7189cc372560 // indirect
	github.com/shurcooL/highlight_diff v0.0.0-20181222201841-111da2e7d480 // indirect
	github.com/shurcooL/highlight_go v0.0.0-20181215221002-9d8641ddf2e1 // indirect
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/httpgzip v0.0.0-20190720172056-320755c1c1b0 // indirect
	github.com/shurcooL/octicon v0.0.0-20191102190552-cbb32d6a785c // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/sloonz/go-qprintable v0.0.0-20160203160305-775b3a4592d5 // indirect
	github.com/sourcegraph/annotate v0.0.0-20160123013949-f4cad6c6324d // indirect
	github.com/sourcegraph/ctxvfs v0.0.0-20180418081416-2b65f1b1ea81
	github.com/sourcegraph/docsite v1.3.1-0.20191122075004-8df06ae22b02
	github.com/sourcegraph/go-diff v0.5.1
	github.com/sourcegraph/go-jsonschema v0.0.0-20191016093751-6a4f2b621f5d
	github.com/sourcegraph/go-langserver v2.0.1-0.20181108233942-4a51fa2e1238+incompatible
	github.com/sourcegraph/go-lsp v0.0.0-20181119182933-0c7d621186c1
	github.com/sourcegraph/gosyntect v0.0.0-20191003053245-e91d603ba4eb
	github.com/sourcegraph/jsonx v0.0.0-20190114210550-ba8cb36a8614
	github.com/sourcegraph/syntaxhighlight v0.0.0-20170531221838-bd320f5d308e // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.5.0 // indirect
	github.com/sqs/httpgzip v0.0.0-20180622165210-91da61ed4dff
	github.com/src-d/enry/v2 v2.1.0
	github.com/stripe/stripe-go v67.9.0+incompatible
	github.com/temoto/robotstxt v1.1.1
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80
	github.com/uber/gonduit v0.6.1
	github.com/uber/jaeger-client-go v2.20.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	github.com/ugorji/go v1.1.7 // indirect
	github.com/uudashr/gocognit v1.0.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/wsxiaoys/terminal v0.0.0-20160513160801-0940f3fc43a0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/xeonx/timeago v1.0.0-rc4
	go.opencensus.io v0.22.2 // indirect
	go.starlark.net v0.0.0-20191113183327-aaf7be003892 // indirect
	go.uber.org/atomic v1.5.1 // indirect
	go.uber.org/automaxprocs v1.2.0
	golang.org/x/arch v0.0.0-20191101135251-a0d8588395bd // indirect
	golang.org/x/crypto v0.0.0-20191128160524-b544559bb6d1
	golang.org/x/net v0.0.0-20191126235420-ef20fe5d7933
	golang.org/x/oauth2 v0.0.0-20191122200657-5d9234df094c
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20191128015809-6d18c012aee9
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/tools v0.0.0-20191130070609-6e064ea0cf2d
	golang.org/x/xerrors v0.0.0-20191011141410-1b5146add898 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20191115221424-83cc0476cb11 // indirect
	google.golang.org/grpc v1.25.1 // indirect
	gopkg.in/alexcesaro/statsd.v2 v2.0.0 // indirect
	gopkg.in/inconshreveable/log15.v2 v2.0.0-20180818164646-67afb5ed74ec
	gopkg.in/jpoehls/gophermail.v0 v0.0.0-20160410235621-62941eab772c
	gopkg.in/karlseguin/expect.v1 v1.0.1 // indirect
	gopkg.in/square/go-jose.v2 v2.4.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.2.7
	mvdan.cc/unparam v0.0.0-20191111180625-960b1ec0f2c2 // indirect
	sourcegraph.com/sqs/pbtypes v1.0.0 // indirect
)

replace (
	github.com/google/zoekt => github.com/sourcegraph/zoekt v0.0.0-20191204145050-b470e5f5cf39
	github.com/mattn/goreman => github.com/sourcegraph/goreman v0.1.2-0.20180928223752-6e9a2beb830d
	github.com/russellhaering/gosaml2 => github.com/sourcegraph/gosaml2 v0.0.0-20190712190530-f05918046bab
	github.com/uber/gonduit => github.com/sourcegraph/gonduit v0.4.0
)

replace github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2

replace github.com/dghubble/gologin => github.com/sourcegraph/gologin v1.0.2-0.20181110030308-c6f1b62954d8

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f
