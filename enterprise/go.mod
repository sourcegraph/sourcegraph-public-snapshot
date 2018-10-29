module github.com/sourcegraph/enterprise

require (
	github.com/beevik/etree v0.0.0-20180609182452-90dafc1e1f11
	github.com/coreos/go-oidc v0.0.0-20171002155002-a93f71fdfe73
	github.com/crewjam/saml v0.0.0-20180831135026-ebc5f787b786
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/uuid v1.0.0
	github.com/google/zoekt v0.0.0-20180530125106-8e284ca7e964
	github.com/gorilla/csrf v1.5.1
	github.com/gorilla/mux v1.6.2
	github.com/graph-gophers/graphql-go v0.0.0-20180806175703-94da0f0031f9
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/golang-lru v0.5.0 // indirect
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/keegancsmith/sqlf v1.1.0
	github.com/keegancsmith/tmpfriend v0.0.0-20180423180255-86e88902a513
	github.com/kevinburke/go-bindata v3.11.1-0.20180909202705-9b44e0539c2a+incompatible
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348 // indirect
	github.com/lib/pq v1.0.0
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/mattn/goreman v0.2.1-0.20180930133601-738cf1257bd3
	github.com/neelance/parallel v0.0.0-20160708114440-4de9ce63d14c
	github.com/opentracing/opentracing-go v1.0.2
	github.com/pelletier/go-toml v1.2.0
	github.com/pkg/errors v0.8.0
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/prometheus/client_golang v0.8.0
	github.com/russellhaering/gosaml2 v0.3.1
	github.com/russellhaering/goxmldsig v0.0.0-20180430223755-7acd5e4a6ef7
	github.com/shurcooL/vfsgen v0.0.0-20180915214035-33ae1944be3f
	github.com/slimsag/godocmd v0.0.0-20161025000126-a1005ad29fe3 // indirect
	github.com/sourcegraph/ctxvfs v0.0.0-20180418081416-2b65f1b1ea81
	github.com/sourcegraph/go-jsonschema v0.0.0-20180805125535-0e659b54484d
	github.com/sourcegraph/go-langserver v2.0.1-0.20181010102349-2d9d8f4a24da+incompatible
	github.com/sourcegraph/go-vcsurl v0.0.0-20131114132947-6b12603ea6fd
	github.com/sourcegraph/godockerize v0.0.0-20181029061954-5cf4e6d81720
	github.com/sourcegraph/jsonrpc2 v0.0.0-20180831160525-549eb959f029
	github.com/sourcegraph/jsonx v0.0.0-20180801091521-5a4ae5eb18cd
	github.com/sourcegraph/rpc v0.0.0-20180329203801-5eaf49b36f85 // indirect
	github.com/sourcegraph/sourcegraph v0.0.0-20181015141638-6a69e6bebaf9
	github.com/src-d/gcfg v1.3.0 // indirect
	github.com/stripe/stripe-go v0.0.0-20181003141555-9e2a36d584c4
	github.com/zenazn/goji v0.9.0 // indirect
	go4.org v0.0.0-20180809161055-417644f6feb5
	golang.org/x/crypto v0.0.0-20180910181607-0e37d006457b
	golang.org/x/net v0.0.0-20181011144130-49bb7cea24b1
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
	golang.org/x/tools v0.0.0-20181017151246-e94054f4104a
	google.golang.org/appengine v1.2.0 // indirect
	google.golang.org/grpc v1.15.0 // indirect
	gopkg.in/inconshreveable/log15.v2 v2.0.0-20180818164646-67afb5ed74ec
	gopkg.in/square/go-jose.v2 v2.1.9 // indirect
	gopkg.in/src-d/go-git.v4 v4.7.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.2.1
	honnef.co/go/tools v0.0.0-20180910201051-f1b53a58b022
)

replace (
	github.com/google/zoekt => github.com/sourcegraph/zoekt v0.0.0-20180814142946-6c42419fec1f
	github.com/russellhaering/gosaml2 => github.com/sourcegraph/gosaml2 v0.0.0-20180820053343-1b78a6b41538
)

replace github.com/graph-gophers/graphql-go => github.com/sourcegraph/graphql-go v0.0.0-20180929065141-c790ffc3c46a

replace github.com/sourcegraph/sourcegraph => ../
