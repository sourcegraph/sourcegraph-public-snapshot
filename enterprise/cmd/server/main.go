package main

import (
	"log"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/server/shared"
)

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}

	shared.ProcfileAdditions = append(
		shared.ProcfileAdditions,
		`precise-code-intel-worker: precise-code-intel-worker`,
	)

	shared.SrcProfServices = append(
		shared.SrcProfServices,
		map[string]string{"Name": "precise-code-intel-worker", "Host": "127.0.0.1:6088"},
	)

	enableEmbeddings, _ := strconv.ParseBool(os.Getenv("SRC_ENABLE_EMBEDDINGS"))
	if enableEmbeddings {
		shared.ProcfileAdditions = append(
			shared.ProcfileAdditions,
			`embeddings: embeddings`,
		)
		shared.SrcProfServices = append(
			shared.SrcProfServices,
			map[string]string{"Name": "embeddings", "Host": "127.0.0.1:6099"},
		)
	}

	shared.Main()
}
