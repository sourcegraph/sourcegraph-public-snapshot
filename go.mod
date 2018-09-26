module github.com/sourcegraph/enterprise

require (
	github.com/beevik/etree v0.0.0-20180609182452-90dafc1e1f11
	github.com/coreos/go-oidc v0.0.0-20171002155002-a93f71fdfe73
	github.com/crewjam/saml v0.0.0-20180831135026-ebc5f787b786
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/gorilla/csrf v1.5.1
	github.com/joho/godotenv v1.3.0 // indirect
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/keegancsmith/sqlf v1.1.0 // indirect
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348 // indirect
	github.com/opentracing/opentracing-go v1.0.2
	github.com/pkg/errors v0.8.0
	github.com/pquerna/cachecontrol v0.0.0-20180517163645-1555304b9b35 // indirect
	github.com/russellhaering/gosaml2 v0.3.1
	github.com/russellhaering/goxmldsig v0.0.0-20180430223755-7acd5e4a6ef7
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20180915214035-33ae1944be3f
	github.com/sourcegraph/go-langserver v0.0.0-20180917104716-6b103664e059
	github.com/sourcegraph/go-vcsurl v0.0.0-20131114132947-6b12603ea6fd
	github.com/sourcegraph/jsonrpc2 v0.0.0-20180831160525-549eb959f029
	github.com/sourcegraph/rpc v0.0.0-20180329203801-5eaf49b36f85 // indirect
	github.com/sourcegraph/sourcegraph v0.0.0-20180926012248-635020a9ba62
	github.com/src-d/gcfg v1.3.0 // indirect
	github.com/zenazn/goji v0.9.0 // indirect
	golang.org/x/crypto v0.0.0-20180910181607-0e37d006457b
	golang.org/x/net v0.0.0-20180911220305-26e67e76b6c3
	golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be
	golang.org/x/sys v0.0.0-20180918153733-ee1b12c67af4 // indirect
	google.golang.org/appengine v1.2.0 // indirect
	google.golang.org/grpc v1.15.0 // indirect
	gopkg.in/inconshreveable/log15.v2 v2.0.0-20180818164646-67afb5ed74ec
	gopkg.in/square/go-jose.v2 v2.1.9 // indirect
	gopkg.in/src-d/go-git.v4 v4.7.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)

replace (
	github.com/google/zoekt => github.com/sourcegraph/zoekt v0.0.0-20180814142946-6c42419fec1f
	github.com/russellhaering/gosaml2 => github.com/sourcegraph/gosaml2 v0.0.0-20180820053343-1b78a6b41538
)

replace github.com/sourcegraph/sourcegraph => ../sourcegraph
