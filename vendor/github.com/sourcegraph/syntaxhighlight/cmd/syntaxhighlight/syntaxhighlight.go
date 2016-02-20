package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/sourcegraph/syntaxhighlight"
)

func main() {
	flag.Parse()

	log.SetFlags(0)

	if flag.NArg() != 1 {
		log.Fatal("Must specify exactly 1 filename argument.")
	}

	input, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	html, err := syntaxhighlight.AsHTML(input)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", html)
}
