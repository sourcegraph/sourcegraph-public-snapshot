package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	"sourcegraph.com/sourcegraph/go-flags"

	"sourcegraph.com/sourcegraph/srclib/cli"
	_ "sourcegraph.com/sourcegraph/srclib/dep"
	_ "sourcegraph.com/sourcegraph/srclib/scan"
)

func main() {
	if cpuprof := os.Getenv("CPUPROF"); cpuprof != "" {
		f, err := os.Create(cpuprof)
		if err != nil {
			log.Fatal("CPUPROF:", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("StartCPUProfile:", err)
		}
		defer func() {
			pprof.StopCPUProfile()
			f.Close()
		}()
	}

	if err := cli.Main(); err != nil {
		if _, ok := err.(*flags.Error); !ok {
			fmt.Fprintf(os.Stderr, "FAILED: %s\n", strings.Join(os.Args, " "))
		}
		os.Exit(1)
	}
}
