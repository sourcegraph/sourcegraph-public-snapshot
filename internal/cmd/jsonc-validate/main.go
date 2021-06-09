package main

import (
	"io"
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
)

func main() {
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	bs, err := jsonc.Parse(string(in))
	if err != nil {
		log.Fatal(err)
	}

	os.Stdout.Write(bs)
}
