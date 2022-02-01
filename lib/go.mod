module github.com/sourcegraph/sourcegraph/lib

go 1.16

require (
	github.com/alecthomas/kingpin v2.2.6+incompatible
	github.com/alecthomas/units v0.0.0-20210927113745-59d0afb8317a // indirect
	github.com/cockroachdb/errors v1.8.6
	github.com/cockroachdb/logtags v0.0.0-20211118104740-dabe8e521a4f // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/derision-test/go-mockgen v1.1.2
	github.com/fatih/color v1.13.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gobwas/glob v0.2.3
	github.com/google/go-cmp v0.5.6
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hexops/autogold v1.3.0
	github.com/hexops/valast v1.4.1 // indirect
	github.com/inconshreveable/log15 v0.0.0-20201112154412-8562bdadbbac
	github.com/json-iterator/go v1.1.12
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/klauspost/pgzip v1.2.5
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14
	github.com/mattn/go-runewidth v0.0.13
	github.com/mitchellh/copystructure v1.2.0
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6
	github.com/sourcegraph/go-diff v0.6.1
	github.com/sourcegraph/jsonx v0.0.0-20200629203448-1a936bd500cf
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/sys v0.0.0-20211205182925-97ca703d548d
	golang.org/x/tools v0.1.8 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

// See: https://github.com/ghodss/yaml/pull/65
replace github.com/ghodss/yaml => github.com/sourcegraph/yaml v1.0.1-0.20200714132230-56936252f152
