package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"

	"github.com/cockroachdb/errors"
	tree_sitter "github.com/smacker/go-tree-sitter"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/tree-sitter/highlight"
)

type Arguments struct {
	language       *tree_sitter.Language
	inputPath      string
	mode           string
	_deferredFuncs []func()
}

func (a *Arguments) cleanup() {
	for i := len(a._deferredFuncs) - 1; i >= 0; i-- {
		a._deferredFuncs[i]()
	}
}

func parseArguments() Arguments {
	var languageFlag = flag.String("language", "",
		fmt.Sprintf("Specify the language: support languages are %v.", highlight.SUPPORTED_LANGUAGES))
	var inputFilePathFlag = flag.String("input-filepath", "-",
		fmt.Sprintf("Specify the path to the file for highlighting. Defaults to stdin"))
	var mode = flag.String("mode", "highlight",
		fmt.Sprintf("'highlight' or 'parse-only'"))
	var cpuProfilePath = flag.String("cpuprofile", "", "write cpu profile to `file`")
	var memProfilePath = flag.String("memprofile", "", "write mem profile to `file`")
	flag.Parse()

	var args Arguments

	if !stringSliceContains(highlight.SUPPORTED_LANGUAGES, *languageFlag) {
		panic(fmt.Sprintf("Found unsupported language; supported languages are %v", highlight.SUPPORTED_LANGUAGES))
	}
	args.language = highlight.LANGUAGE_MAP[*languageFlag]
	args.inputPath = *inputFilePathFlag
	if *mode != "highlight" && *mode != "parse-only" {
		log.Fatal("illegal mode argument, expected 'highlight' or 'parse-only': ", mode)
	}
	args.mode = *mode

	if *cpuProfilePath != "" {
		f, err := os.Create(*cpuProfilePath)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		args._deferredFuncs = append(args._deferredFuncs, func() {
			pprof.StopCPUProfile()
			f.Close()
		})
	}
	args._deferredFuncs = append(args._deferredFuncs, func() {
		if *memProfilePath != "" {
			defer func() {
				f, err := os.Create(*memProfilePath)
				if err != nil {
					log.Fatal("could not create memory profile: ", err)
				}
				defer f.Close() // error handling omitted for example
				if err := pprof.WriteHeapProfile(f); err != nil {
					log.Fatal("failed to write memory profile: ", err)
				}
			}()
		}
	})
	return args
}

func main() {
	var args = parseArguments()
	defer args.cleanup()

	var inputBytes []byte
	var err error
	if args.inputPath == "-" {
		inputBytes, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal("failed to read input from stdin: ", err)
		}
	} else {
		inputBytes, err = ioutil.ReadFile(args.inputPath)
		if err != nil {
			log.Fatal("failed to read from file: ", err)
		}
	}
	parseTree, err := tree_sitter.ParseCtx(context.Background(), inputBytes, args.language)
	if err != nil {
		panic(errors.Wrap(err, "tree-sitter failed to parse input"))
	}
	if args.mode == "parse-only" {
		formatter := highlight.NewSexprFormatter(inputBytes)
		fmt.Println(formatter.Format(parseTree))
		return
	}
	var outputBuffer bytes.Buffer
	ctx := context.Background()
	highlighter := highlight.NewHighlightingContext(ctx, inputBytes, &outputBuffer, args.language)
	treeIterator := highlight.NewAllOrderIterator(parseTree, &highlighter)
	if node, err := treeIterator.VisitTree(); err != nil {
		panic(fmt.Sprintf("Failed to visit tree: node=%v err=%v", node, err))
	}
	fmt.Println(outputBuffer.String())
}

func stringSliceContains(array []string, s string) bool {
	for _, element := range array {
		if element == s {
			return true
		}
	}
	return false
}
