package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol/reader"
)

func main() {
	// parse file into LsifValues proto
	file := os.Args[1]
	if strings.HasSuffix(file, ".lsif-typed") {
		data, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		index := lsif_typed.Index{}
		err = proto.Unmarshal(data, &index)
		if err != nil {
			panic(errors.Wrapf(err, "failed to parse protobuf file '%s'", file))
		}
		els, err := reader.ConvertTypedIndexToGraphIndex(&index)
		if err != nil {
			panic(errors.Wrapf(err, "failed reader.ConvertTypedIndexToGraphIndex"))
		}
		err = reader.WriteNDJSON(reader.ElementsToEmptyInterfaces(els), os.Stdout)
		if err != nil {
			panic(err)
		}
	} else {
		panic(fmt.Sprintf("unexpected file format (must have extension .lsif-typed): %s\n", file))
	}
}
