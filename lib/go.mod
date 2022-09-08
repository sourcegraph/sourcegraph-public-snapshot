module github.com/sourcegraph/sourcegraph/lib

go 1.18

require (
	github.com/charmbracelet/glamour v0.5.0
	github.com/cockroachdb/errors v1.9.0
	github.com/derision-test/go-mockgen v1.1.2
	github.com/fatih/color v1.13.0
	github.com/ghodss/yaml v1.0.0
	github.com/gobwas/glob v0.2.3
	github.com/google/go-cmp v0.5.8
	github.com/grafana/regexp v0.0.0-20220304100321-149c8afcd6cb
	github.com/hexops/autogold v1.3.0
	github.com/json-iterator/go v1.1.12
	github.com/klauspost/pgzip v1.2.5
	github.com/mattn/go-isatty v0.0.16
	github.com/mattn/go-runewidth v0.0.13
	github.com/mitchellh/copystructure v1.2.0
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6
	github.com/muesli/termenv v0.12.0
	github.com/sourcegraph/go-diff v0.6.1
	github.com/sourcegraph/jsonx v0.0.0-20200629203448-1a936bd500cf
	github.com/sourcegraph/log v0.0.0-20220901143117-fc0516a694c9
	github.com/stretchr/testify v1.8.0
	github.com/xeipuuv/gojsonschema v1.2.0
	go.uber.org/atomic v1.10.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/alecthomas/chroma v0.10.0 // indirect
	github.com/alecthomas/kingpin v2.2.6+incompatible // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20211118104740-dabe8e521a4f // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/dave/jennifer v1.4.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.4.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/getsentry/sentry-go v0.13.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/hexops/gotextdiff v1.0.3 // indirect
	github.com/hexops/valast v1.4.1 // indirect
	github.com/klauspost/compress v1.14.2 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/microcosm-cc/bluemonday v1.0.17 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/nightlyone/lockfile v1.0.0 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/shurcooL/go-goon v0.0.0-20210110234559-7585751d9a17 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/yuin/goldmark v1.4.4 // indirect
	github.com/yuin/goldmark-emoji v1.0.1 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.23.0 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
	golang.org/x/net v0.0.0-20220526153639-5463443f8c37 // indirect
	golang.org/x/tools v0.1.10 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	mvdan.cc/gofumpt v0.2.1 // indirect
)

// See: https://github.com/ghodss/yaml/pull/65
replace github.com/ghodss/yaml => github.com/sourcegraph/yaml v1.0.1-0.20200714132230-56936252f152
