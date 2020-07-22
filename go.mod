module github.com/sourcegraph/sourcegraph

go 1.14

require (
	cloud.google.com/go/bigquery v1.6.0 // indirect
	cloud.google.com/go/pubsub v1.3.1
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/NYTimes/gziphandler v1.1.1
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/RoaringBitmap/roaring v0.4.23
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/aphistic/sweet-junit v0.2.0 // indirect
	github.com/avelino/slugify v0.0.0-20180501145920-855f152bd774
	github.com/aws/aws-sdk-go-v2 v0.20.0
	github.com/beevik/etree v1.1.0
	github.com/boj/redistore v0.0.0-20180917114910-cd5dcc76aeff
	github.com/certifi/gocertifi v0.0.0-20200211180108-c7c1fbc02894 // indirect
	github.com/containerd/containerd v1.3.4 // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/coreos/go-semver v0.3.0
	github.com/crewjam/saml v0.4.0
	github.com/davecgh/go-spew v1.1.1
	github.com/daviddengcn/go-colortext v1.0.0
	github.com/dghubble/gologin v2.2.0+incompatible
	github.com/dgraph-io/ristretto v0.0.2
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/dineshappavoo/basex v0.0.0-20170425072625-481a6f6dc663
	github.com/dnaeon/go-vcr v1.0.1
	github.com/efritz/glock v0.0.0-20181228234553-f184d69dff2c
	github.com/efritz/go-genlib v0.0.0-20200616012750-c21aae2e13ac // indirect
	github.com/efritz/go-mockgen v0.0.0-20200524175724-37e2c732ee40
	github.com/efritz/pentimento v0.0.0-20190429011147-ade47d831101
	github.com/ericchiang/k8s v1.2.0
	github.com/fatih/astrewrite v0.0.0-20191207154002-9094e544fcef
	github.com/fatih/color v1.9.0
	github.com/felixge/fgprof v0.9.0
	github.com/felixge/httpsnoop v1.0.1
	github.com/gchaincl/sqlhooks v1.3.0
	github.com/getsentry/raven-go v0.2.0
	github.com/ghodss/yaml v1.0.0
	github.com/gin-gonic/gin v1.6.3 // indirect
	github.com/gitchander/permutation v0.0.0-20181107151852-9e56b92e9909
	github.com/gliderlabs/ssh v0.3.0 // indirect
	github.com/glycerine/go-unsnap-stream v0.0.0-20190901134440-81cf024a9e0a // indirect
	github.com/go-git/go-git/v5 v5.1.0 // indirect
	github.com/go-openapi/strfmt v0.19.5
	github.com/go-playground/validator/v10 v10.3.0 // indirect
	github.com/go-redsync/redsync v1.4.2
	github.com/gobwas/glob v0.2.3
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/golang/gddo v0.0.0-20200611223618-a4829ef13274
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/gomodule/oauth1 v0.0.0-20181215000758-9a59ed3b0a84
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-cmp v0.5.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v28 v28.1.1
	github.com/google/go-github/v31 v31.0.0
	github.com/google/go-querystring v1.0.0
	github.com/google/uuid v1.1.1
	github.com/google/zoekt v0.0.0-20200616181411-e6113752c10c
	github.com/gopherjs/gopherjs v0.0.0-20200217142428-fce0ec30dd00 // indirect
	github.com/gorilla/context v1.1.1
	github.com/gorilla/csrf v1.7.0
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/schema v1.1.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.0
	github.com/gosimple/slug v1.9.0 // indirect
	github.com/goware/urlx v0.3.1
	github.com/grafana-tools/sdk v0.0.0-20200627094057-622ba1f937bb
	github.com/graph-gophers/graphql-go v0.0.0-20200622220639-c1d9693c95a6
	github.com/graphql-go/graphql v0.7.9
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/hashicorp/go-hclog v0.14.1 // indirect
	github.com/hashicorp/go-multierror v1.1.0
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/honeycombio/libhoney-go v1.12.4
	github.com/inconshreveable/log15 v0.0.0-20200109203555-b30bc20e4fd1
	github.com/jmespath/go-jmespath v0.3.0 // indirect
	github.com/jmoiron/sqlx v1.2.1-0.20190826204134-d7d95172beb5
	github.com/joho/godotenv v1.3.0
	github.com/jordan-wright/email v0.0.0-20200602115436-fd8a7622303e
	github.com/json-iterator/go v1.1.10
	github.com/karlseguin/expect v1.0.7 // indirect
	github.com/karlseguin/typed v1.1.7 // indirect
	github.com/karrick/godirwalk v1.15.6
	github.com/keegancsmith/sqlf v1.1.0
	github.com/keegancsmith/tmpfriend v0.0.0-20180423180255-86e88902a513
	github.com/kevinburke/go-bindata v3.21.0+incompatible
	github.com/klauspost/compress v1.10.10 // indirect
	github.com/kr/text v0.2.0
	github.com/kylelemons/godebug v1.1.0
	github.com/leanovate/gopter v0.2.8
	github.com/lib/pq v1.7.0
	github.com/lightstep/lightstep-tracer-common/golang/gogo v0.0.0-20200310182322-adf4263e074b // indirect
	github.com/lightstep/lightstep-tracer-go v0.20.0
	github.com/machinebox/graphql v0.2.2
	github.com/matryer/is v1.3.0 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2
	github.com/microcosm-cc/bluemonday v1.0.3
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f
	github.com/neelance/parallel v0.0.0-20160708114440-4de9ce63d14c
	github.com/onsi/ginkgo v1.12.1 // indirect
	github.com/onsi/gomega v1.10.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opentracing-contrib/go-stdlib v1.0.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/peterbourgon/ff v1.7.0
	github.com/peterhellberg/link v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/alertmanager v0.21.0
	github.com/prometheus/client_golang v1.6.0
	github.com/prometheus/common v0.10.0
	github.com/prometheus/procfs v0.1.3 // indirect
	github.com/rainycape/unidecode v0.0.0-20150907023854-cb7f23ec59be
	github.com/russellhaering/gosaml2 v0.4.0
	github.com/russellhaering/goxmldsig v0.0.0-20180430223755-7acd5e4a6ef7
	github.com/russross/blackfriday v2.0.0+incompatible // indirect
	github.com/schollz/progressbar/v3 v3.3.4
	github.com/segmentio/fasthash v1.0.2
	github.com/sergi/go-diff v1.1.0
	github.com/shirou/gopsutil v2.20.5+incompatible // indirect
	github.com/shurcooL/github_flavored_markdown v0.0.0-20181002035957-2122de532470
	github.com/shurcooL/go v0.0.0-20200502201357-93f07166e636 // indirect
	github.com/shurcooL/highlight_diff v0.0.0-20181222201841-111da2e7d480 // indirect
	github.com/shurcooL/highlight_go v0.0.0-20191220051317-782971ddf21b // indirect
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/httpgzip v0.0.0-20190720172056-320755c1c1b0
	github.com/shurcooL/octicon v0.0.0-20191102190552-cbb32d6a785c // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/sourcegraph/annotate v0.0.0-20160123013949-f4cad6c6324d // indirect
	github.com/sourcegraph/codeintelutils v0.0.0-20200706141440-54ddac67b5b6
	github.com/sourcegraph/ctxvfs v0.0.0-20180418081416-2b65f1b1ea81
	github.com/sourcegraph/go-diff v0.5.3
	github.com/sourcegraph/go-jsonschema v0.0.0-20200429204646-54904d837db8
	github.com/sourcegraph/go-langserver v2.0.1-0.20181108233942-4a51fa2e1238+incompatible
	github.com/sourcegraph/go-lsp v0.0.0-20200429204803-219e11d77f5d
	github.com/sourcegraph/gosyntect v0.0.0-20200429204402-842ed26129d0
	github.com/sourcegraph/jsonx v0.0.0-20200625022044-c22a595bbad7
	github.com/sourcegraph/syntaxhighlight v0.0.0-20170531221838-bd320f5d308e // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/src-d/enry/v2 v2.1.0
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/stripe/stripe-go v70.15.0+incompatible
	github.com/temoto/robotstxt v1.1.1
	github.com/tinylib/msgp v1.1.2 // indirect
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80
	github.com/uber/gonduit v0.6.1
	github.com/uber/jaeger-client-go v2.24.0+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	github.com/vmihailenco/msgpack/v4 v4.3.12 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.2.0
	github.com/xeonx/timeago v1.0.0-rc4
	github.com/zenazn/goji v1.0.1 // indirect
	go.opencensus.io v0.22.4 // indirect
	go.uber.org/atomic v1.6.0
	go.uber.org/automaxprocs v1.3.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/mod v0.3.0 // indirect
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1
	golang.org/x/tools v0.0.0-20200626032829-bcbc01e07a20
	google.golang.org/api v0.28.0 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/grpc v1.30.0 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
)

replace (
	// protobuf v1.3.5+ causes issues - https://github.com/sourcegraph/sourcegraph/issues/11804
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.5

	// We need our fork until https://github.com/graph-gophers/graphql-go/pull/400 is merged upstream
	// Our change limits the number of goroutines spawned by resolvers which was causing memory spikes on our frontend
	github.com/graph-gophers/graphql-go => github.com/sourcegraph/graphql-go v0.0.0-20200626080007-7fa8b67cbb1d
	github.com/mattn/goreman => github.com/sourcegraph/goreman v0.1.2-0.20180928223752-6e9a2beb830d

	// prom-wrapper needs to be able to write alertmanager configuration with secrets - https://github.com/prometheus/alertmanager/pull/2316
	github.com/prometheus/alertmanager => github.com/bobheadxi/alertmanager v0.21.1-0.20200702014640-a941b136043b
	github.com/russellhaering/gosaml2 => github.com/sourcegraph/gosaml2 v0.3.2-0.20200109173551-5cfddeb48b17
	github.com/uber/gonduit => github.com/sourcegraph/gonduit v0.4.0
)

// We maintain our own fork of Zoekt. Update with ./dev/zoekt/update
replace github.com/google/zoekt => github.com/sourcegraph/zoekt v0.0.0-20200720095407-6597aebe357e

replace github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2

replace github.com/dghubble/gologin => github.com/sourcegraph/gologin v1.0.2-0.20181110030308-c6f1b62954d8

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f

// See: https://github.com/ghodss/yaml/pull/65
replace github.com/ghodss/yaml => github.com/sourcegraph/yaml v1.0.1-0.20200714132230-56936252f152
