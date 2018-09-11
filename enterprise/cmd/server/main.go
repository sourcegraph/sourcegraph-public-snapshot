package main

import (
	"log"

	"github.com/sourcegraph/sourcegraph/cmd/server/shared"
)

func main() {
	shared.ProcfileAdditions = []string{
		`indexer: indexer`,
	}
	shared.SrcProfServices = append(shared.SrcProfServices, map[string]string{
		"Name": "indexer",
		"Host": "127.0.0.1:6073",
	})

	if lines, err := maybeZoektProcfile(shared.DataDir); err != nil {
		log.Fatal(err)
	} else if lines != nil {
		shared.ProcfileAdditions = append(shared.ProcfileAdditions, lines...)
	}
	shared.Main()
}
