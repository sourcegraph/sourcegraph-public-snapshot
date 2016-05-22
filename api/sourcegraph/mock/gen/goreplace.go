// +build generate

// Command goreplace replaces all instances of "from" string with "to" string in specified files.
// Like basic string replacement functionality of sed, but works on OS X, Linux and Windows.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	var fromFlag = flag.String("from", "", "source string to replace")
	var toFlag = flag.String("to", "", "string to replace with")
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		log.Fatalln("no files to process")
	}

	for _, path := range files {
		fmt.Fprintln(os.Stderr, "#", path)
		err := processFile(path, *fromFlag, *toFlag)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func processFile(path, from, to string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	b = bytes.Replace(b, []byte(from), []byte(to), -1)

	err = ioutil.WriteFile(path, b, 0644)
	if err != nil {
		return err
	}

	return nil
}
