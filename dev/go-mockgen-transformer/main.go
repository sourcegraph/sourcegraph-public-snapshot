package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/dev/go-mockgen-transformer/config"
)

type sliceFlag []string

func (i *sliceFlag) String() string {
	return strings.Join(*i, ",")
}

func (i *sliceFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	payload, err := config.ReadManifest("mockgen.yaml")
	if err != nil {
		panic(err)
	}

	var sourceFiles sliceFlag
	var archives sliceFlag
	outfile := flag.String("outfile", "mockgen.yaml", "File to write the transformed config to")
	intermediaryGeneratedFile := flag.String("intermediary-generated-file", "", "Path to the intermediary generated file (before being copied to source tree)")
	finalGeneratedFile := flag.String("final-generated-file", "", "Path to the final generated file (in the source tree)")
	outputImportPath := flag.String("output-importpath", "", "The import path of the generated mock file")
	flag.StringVar(&payload.Goimports, "goimports", "./bin/.goimports", "Path to goimports binary")
	flag.StringVar(&payload.StdlibRoot, "stdlibroot", "", "Path to the root of the Go compiled stdlib")
	flag.Var(&sourceFiles, "source-files", "Values of the format IMPORTPATH=FILE, where IMPORTPATH is the import path for FILE")
	flag.Var(&archives, "archives", "Values of the format IMPORTPATH=ARCHIVE, where ARCHIVE is the path to the archive for the given IMPORTPATH")
	flag.Parse()

	payload.IncludeConfigPaths = []string{}

	f, err := os.Create(*outfile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// source files need to be grouped by their import path, so that when we process each (import path, source files)
	// tuple in go-mockgen, the packages declared at the top of the source files are all equal part import path grouping.
	importpathToSources := make(map[string][]string)
	for _, arg := range sourceFiles {
		split := strings.Split(arg, "=")
		importpathToSources[split[0]] = append(importpathToSources[split[0]], split[1])
	}

	outputPayload := payload
	outputPayload.Mocks = []config.YamlMock{}
	// extract the mock configuration for the specific file we want to generate mocks for.
	for _, mock := range payload.Mocks {
		if mock.Filename == *finalGeneratedFile {
			// the config declares the filepath for the _final_ generated file,
			// but we prefix the filename with _ so we can copy from the output base
			// into the source tree.
			mock.Filename = *intermediaryGeneratedFile
			outputPayload.Mocks = []config.YamlMock{mock}
			break
		}
	}
	if len(outputPayload.Mocks) == 0 {
		panic(fmt.Sprintf("Could not find mock for file %s", *finalGeneratedFile))
	}

	// TODO: why does this need to be set explicitly again?
	outputPayload.Mocks[0].ImportPath = *outputImportPath

	// now we have the right config section, we want to attach the right sources for each
	// import path that we're mocking interfaces from. For each dep in the go_mockgen rule,
	// there should be a corresponding section in the config extract.
	// Archive files are set globally, they don't need grouping.
	if len(outputPayload.Mocks[0].Sources) > 0 {
		newSources := outputPayload.Mocks[0].Sources[:0]
		for _, source := range outputPayload.Mocks[0].Sources {
			source.SourceFiles = importpathToSources[source.Path]
			outputPayload.Mocks[0].Archives = archives
			newSources = append(newSources, source)
		}
	} else {
		outputPayload.Mocks[0].Archives = archives
		outputPayload.Mocks[0].SourceFiles = importpathToSources[outputPayload.Mocks[0].Path]
	}

	out, err := yaml.Marshal(outputPayload)
	if err != nil {
		panic(err)
	}

	if _, err := f.Write(out); err != nil {
		panic(err)
	}
}
