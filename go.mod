module github.com/sourcegraph/sourcegraph

go 1.14

require (
	cloud.google.com/go v0.56.0 // indirect
	cloud.google.com/go/pubsub v1.3.1
	github.com/Masterminds/semver v1.5.0
	github.com/NYTimes/gziphandler v1.1.1
	github.com/RoaringBitmap/roaring v0.4.23
	github.com/avelino/slugify v0.0.0-20180501145920-855f152bd774
	github.com/aws/aws-sdk-go-v2 v0.20.0
	github.com/beevik/etree v1.1.0
	github.com/boj/redistore v0.0.0-20180917114910-cd5dcc76aeff
	github.com/certifi/gocertifi v0.0.0-20200211180108-c7c1fbc02894 // indirect
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/coreos/go-semver v0.3.0
	github.com/crewjam/saml v0.4.0
	github.com/davecgh/go-spew v1.1.1
	github.com/daviddengcn/go-colortext v1.0.0
	github.com/dghubble/gologin v2.2.0+incompatible
	github.com/dgraph-io/ristretto v0.0.2
	github.com/dnaeon/go-vcr v1.0.1
	github.com/docker/docker v1.4.2-0.20200213202729-31a86c4ab209
	github.com/efritz/go-mockgen v0.0.0-20200420163638-0338f3dfc81c
	github.com/efritz/pentimento v0.0.0-20190429011147-ade47d831101
	github.com/emersion/go-imap v1.0.4
	github.com/ericchiang/k8s v1.2.0
	github.com/fatih/astrewrite v0.0.0-20191207154002-9094e544fcef
	github.com/fatih/color v1.9.0
	github.com/felixfbecker/stringscore v0.0.0-20170928081130-e71a9f1b0749
	github.com/felixge/httpsnoop v1.0.1
	github.com/gchaincl/sqlhooks v1.3.0
	github.com/getsentry/raven-go v0.2.0
	github.com/ghodss/yaml v1.0.0
	github.com/gin-gonic/gin v1.6.2 // indirect
	github.com/gitchander/permutation v0.0.0-20181107151852-9e56b92e9909
	github.com/glycerine/go-unsnap-stream v0.0.0-20190901134440-81cf024a9e0a // indirect
	github.com/go-redsync/redsync v1.4.1
	github.com/gobwas/glob v0.2.3
	github.com/golang-migrate/migrate/v4 v4.10.0
	github.com/golang/gddo v0.0.0-20200324184333-3c2cc9a6329d
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/gomodule/oauth1 v0.0.0-20181215000758-9a59ed3b0a84
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-cmp v0.4.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v28 v28.1.1
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.1.1
	github.com/google/zoekt v0.0.0-20200324103759-172db892e9a2
	github.com/gopherjs/gopherjs v0.0.0-20200217142428-fce0ec30dd00 // indirect
	github.com/gorilla/context v1.1.1
	github.com/gorilla/csrf v1.6.2
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/schema v1.1.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.0
	github.com/goware/urlx v0.3.1
	github.com/grafana-tools/sdk v0.0.0-00010101000000-000000000000
	github.com/graph-gophers/graphql-go v0.0.0-20200309224638-dae41bde9ef9
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/hashicorp/go-hclog v0.12.2 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-retryablehttp v0.6.5 // indirect
	github.com/honeycombio/libhoney-go v1.12.4
	github.com/inconshreveable/log15 v0.0.0-20200109203555-b30bc20e4fd1
	github.com/jmespath/go-jmespath v0.3.0 // indirect
	github.com/jmoiron/sqlx v1.2.1-0.20190826204134-d7d95172beb5
	github.com/joho/godotenv v1.3.0
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/karlseguin/expect v1.0.6 // indirect
	github.com/karlseguin/typed v1.1.7 // indirect
	github.com/karrick/godirwalk v1.15.5
	github.com/karrick/tparse/v2 v2.8.0
	github.com/keegancsmith/sqlf v1.1.0
	github.com/keegancsmith/tmpfriend v0.0.0-20180423180255-86e88902a513
	github.com/kevinburke/go-bindata v3.19.0+incompatible
	github.com/kr/text v0.2.0
	github.com/kylelemons/godebug v1.1.0
	github.com/leanovate/gopter v0.2.7
	github.com/lib/pq v1.3.0
	github.com/lightstep/lightstep-tracer-common/golang/gogo v0.0.0-20200310182322-adf4263e074b // indirect
	github.com/lightstep/lightstep-tracer-go v0.20.0
	github.com/machinebox/graphql v0.2.2
	github.com/matryer/is v1.3.0 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/mattn/goreman v0.0.0-00010101000000-000000000000 // indirect
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/microcosm-cc/bluemonday v1.0.2
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f
	github.com/neelance/parallel v0.0.0-20160708114440-4de9ce63d14c
	github.com/opentracing-contrib/go-stdlib v0.0.0-20190519235532-cf7a6c988dc9
	github.com/opentracing/opentracing-go v1.1.0
	github.com/peterbourgon/ff v1.7.0
	github.com/peterhellberg/link v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be
	github.com/russellhaering/gosaml2 v0.4.0
	github.com/russellhaering/goxmldsig v0.0.0-20180430223755-7acd5e4a6ef7
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/segmentio/fasthash v1.0.1
	github.com/sergi/go-diff v1.1.0
	github.com/shirou/gopsutil v2.20.3+incompatible // indirect
	github.com/shurcooL/github_flavored_markdown v0.0.0-20181002035957-2122de532470
	github.com/shurcooL/go v0.0.0-20191216061654-b114cc39af9f // indirect
	github.com/shurcooL/highlight_diff v0.0.0-20181222201841-111da2e7d480 // indirect
	github.com/shurcooL/highlight_go v0.0.0-20191220051317-782971ddf21b // indirect
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/httpgzip v0.0.0-20190720172056-320755c1c1b0 // indirect
	github.com/shurcooL/octicon v0.0.0-20191102190552-cbb32d6a785c // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/sirupsen/logrus v1.5.0 // indirect
	github.com/sloonz/go-qprintable v0.0.0-20160203160305-775b3a4592d5 // indirect
	github.com/sourcegraph/annotate v0.0.0-20160123013949-f4cad6c6324d // indirect
	github.com/sourcegraph/ctxvfs v0.0.0-20180418081416-2b65f1b1ea81
	github.com/sourcegraph/go-diff v0.5.2
	github.com/sourcegraph/go-jsonschema v0.0.0-20191222043427-cdbee60427af
	github.com/sourcegraph/go-langserver v2.0.1-0.20181108233942-4a51fa2e1238+incompatible
	github.com/sourcegraph/go-lsp v0.0.0-20200117082640-b19bb38222e2
	github.com/sourcegraph/gosyntect v0.0.0-20200331033347-c35e64c39373
	github.com/sourcegraph/jsonx v0.0.0-20190114210550-ba8cb36a8614
	github.com/sourcegraph/syntaxhighlight v0.0.0-20170531221838-bd320f5d308e // indirect
	github.com/sqs/httpgzip v0.0.0-20180622165210-91da61ed4dff
	github.com/src-d/enry/v2 v2.1.0
	github.com/stripe/stripe-go v70.11.0+incompatible
	github.com/temoto/robotstxt v1.1.1
	github.com/tinylib/msgp v1.1.2 // indirect
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80
	github.com/uber/gonduit v0.6.1
	github.com/uber/jaeger-client-go v2.22.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/xeonx/timeago v1.0.0-rc4
	go.uber.org/atomic v1.6.0
	go.uber.org/automaxprocs v1.3.0
	golang.org/x/crypto v0.0.0-20200403201458-baeed622b8d8
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	golang.org/x/sys v0.0.0-20200331124033-c3d80250170d
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/tools v0.0.0-20200420001825-978e26b7c37c
	google.golang.org/api v0.21.0 // indirect
	google.golang.org/genproto v0.0.0-20200403120447-c50568487044 // indirect
	gopkg.in/jpoehls/gophermail.v0 v0.0.0-20160410235621-62941eab772c
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.2.8
)

replace (
	github.com/google/zoekt => github.com/sourcegraph/zoekt v0.0.0-20200511113954-b56036a3b745
	github.com/mattn/goreman => github.com/sourcegraph/goreman v0.1.2-0.20180928223752-6e9a2beb830d
	github.com/russellhaering/gosaml2 => github.com/sourcegraph/gosaml2 v0.3.2-0.20200109173551-5cfddeb48b17
	github.com/uber/gonduit => github.com/sourcegraph/gonduit v0.4.0
)

// https://github.com/grafana-tools/sdk/pull/80
replace github.com/grafana-tools/sdk => github.com/slimsag/sdk v0.0.0-20200402190125-fc52c0aed0b7

replace github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2

replace github.com/dghubble/gologin => github.com/sourcegraph/gologin v1.0.2-0.20181110030308-c6f1b62954d8

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f
