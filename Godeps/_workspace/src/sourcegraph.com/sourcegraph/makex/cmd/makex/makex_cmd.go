package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"sourcegraph.com/sourcegraph/makex"
)

var expand = flag.Bool("x", true, "expand globs in makefile prereqs")
var cwd = flag.String("C", "", "change to this directory before doing anything")
var file = flag.String("f", "Makefile", "path to Makefile")

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `makex is an experimental, incomplete implementation of make in Go.

Usage:

        makex [options] [target] ...

If no targets are specified, the first target that appears in the makefile (not
beginning with ".") is used.

The options are:
`)
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.SetFlags(0)

	conf := makex.Default
	makex.Flags(nil, &conf, "")
	flag.Parse()

	data, err := ioutil.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
	}

	if *cwd != "" {
		err := os.Chdir(*cwd)
		if err != nil {
			log.Fatal(err)
		}
	}

	mf, err := makex.Parse(data)
	if err != nil {
		log.Fatal(err)
	}

	goals := flag.Args()
	if len(goals) == 0 {
		// Find the first rule that doesn't begin with a ".".
		if defaultRule := mf.DefaultRule(); defaultRule != nil {
			goals = []string{defaultRule.Target()}
		}
	}

	if *expand {
		mf, err = conf.Expand(mf)
		if err != nil {
			log.Fatal(err)
		}
	}

	mk := conf.NewMaker(mf, goals...)

	targetSets, err := mk.TargetSetsNeedingBuild()
	if err != nil {
		log.Fatal(err)
	}

	if len(targetSets) == 0 {
		fmt.Println("Nothing to do.")
	}

	if conf.DryRun {
		mk.DryRun(os.Stdout)
		return
	}

	err = mk.Run()
	if err != nil {
		log.Fatal(err)
	}
}
