module github.com/sourcegraph/src-cli

go 1.16

require (
	github.com/Masterminds/semver v1.5.0
	github.com/dustin/go-humanize v1.0.0
	github.com/gobwas/glob v0.2.3
	github.com/google/go-cmp v0.5.5
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jig/teereadcloser v0.0.0-20181016160506-953720c48e05
	github.com/json-iterator/go v1.1.11
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51
	github.com/mattn/go-isatty v0.0.12
	github.com/mitchellh/copystructure v1.2.0
	github.com/neelance/parallel v0.0.0-20160708114440-4de9ce63d14c
	github.com/olekukonko/tablewriter v0.0.4 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sourcegraph/batch-change-utils v0.0.0-20210708162152-c9f35b905d94
	github.com/sourcegraph/go-diff v0.6.1
	github.com/sourcegraph/jsonx v0.0.0-20200629203448-1a936bd500cf
	github.com/sourcegraph/sourcegraph/lib v0.0.0-20210628190057-faa35110fad5
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	jaytaylor.com/html2text v0.0.0-20200412013138-3577fbdbcff7
)

replace github.com/gosuri/uilive v0.0.4 => github.com/mrnugget/uilive v0.0.4-fix-escape

// See: https://github.com/ghodss/yaml/pull/65
replace github.com/ghodss/yaml => github.com/sourcegraph/yaml v1.0.1-0.20200714132230-56936252f152
