package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsiftyped"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
)

func printHelpAndExit() {
	fmt.Println(
		`usage: lsif-typed [FILE]

FILE path to a single file that contains a Protobuf-encoded LSIF Typed payload.`,
	)
	os.Exit(0)
}

func main() {
	if len(os.Args) == 1 {
		printHelpAndExit()
	}
	switch os.Args[1] {
	case "help", "-help", "--help":
		printHelpAndExit()
	default:
		file := os.Args[1]
		if strings.HasSuffix(file, ".lsif-typed") {
			data, err := os.ReadFile(file)
			if err != nil {
				panic(err)
			}
			index := lsiftyped.Index{}
			err = proto.Unmarshal(data, &index)
			if err != nil {
				panic(errors.Wrapf(err, "failed to parse protobuf file '%s'", file))
			}
			els, err := reader.ConvertTypedIndexToGraphIndex(&index)
			if err != nil {
				panic(errors.Wrapf(err, "failed reader.ConvertTypedIndexToGraphIndex"))
			}
			err = reader.WriteNDJSON(reader.ElementsToJsonElements(els), os.Stdout)
			if err != nil {
				panic(err)
			}
		} else {
			panic(fmt.Sprintf("unexpected file format (must have extension .lsif-typed): %s\n", file))
		}
	}
}
