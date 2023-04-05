package main

import (
	"fmt"
	"os"
	"text/template"

	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/staticcheck"
)

var BazelBuildTemplate = `load("@io_bazel_rules_go//go:def.bzl", "go_library")

{{- range .Analyzers}}
go_library(
    name = "{{.Analyzer.Name}}",
    srcs = ["staticcheck.go"],
    importpath = "github.com/sourcegraph/sourcegraph/dev/linters/staticcheck/{{.Analyzer.Name}}",
    visibility = ["//visibility:public"],
	x_defs = {"AnalyzerName": "{{.Analyzer.Name}}"},
    deps = [
    "@org_golang_x_tools//go/analysis",
    "@co_honnef_go_tools//staticcheck",
    "@co_honnef_go_tools//analysis/lint",
    ],
)

{{- end}}
`

var BazelDefTemplate = `
STATIC_CHECK_ANALYZERS = [
{{- range .Analyzers}}
	"//dev/linters/staticcheck:{{.Analyzer.Name}}",
{{- end}}
]
`

var analyzers []*lint.Analyzer = staticcheck.Analyzers

func writeTemplate(targetFile, templateDef string) error {
	name := targetFile
	tmpl := template.Must(template.New(name).Parse(templateDef))

	f, err := os.OpenFile(targetFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tmpl.Execute(f, struct {
		Analyzers []*lint.Analyzer
	}{
		Analyzers: analyzers,
	})
	if err != nil {
		return err
	}

	return nil
}

// We support two position arguments:
// 1: build bazel path - file where the analyzer targets should be generated to
// 2: analyzer definiton path - file where a convienience analyzer array is generated that contains all the targets
func main() {
	targetFile := "BUILD.bazel"
	if len(os.Args) > 1 {
		targetFile = os.Args[1]
	}

	// Generate targets for all the analyzers
	if err := writeTemplate(targetFile, BazelBuildTemplate); err != nil {
		fmt.Fprintln(os.Stderr, "failed to render Bazel buildfile template")
		panic(err)
	}

	// Generate a file where we can import the list of analyzers into our bazel scripts
	targetFile = "analyzers.bzl"
	if len(os.Args) > 2 {
		targetFile = os.Args[2]
	}
	if err := writeTemplate(targetFile, BazelDefTemplate); err != nil {
		fmt.Fprintln(os.Stderr, "failed to render Anazlyers definiton template")
		panic(err)
	}

}
