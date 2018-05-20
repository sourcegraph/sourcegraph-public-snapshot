// +build ignore

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/sourcegraph/sourcegraph/pkg/conf"

	// Import packages that contribute validators.
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/httpheader"
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/openidconnect"
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/saml"
	_ "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/userpasswd"
)

func main() {
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "\tvalidate-config files...")
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
	}
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "validate-config: no files listed.")
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(2)
	}

	for _, filename := range flag.Args() {
		data, err := readFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "validate-config: error reading file: %s.\n", err)
			os.Exit(2)
		}
		messages, err := conf.Validate(string(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "validate-config: error validating %s: %s.\n", filename, err)
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "# %s", filename)
		if len(messages) > 0 {
			fmt.Fprintln(os.Stderr, ": FAIL")
			for _, verr := range messages {
				fmt.Fprintf(os.Stderr, " - %s\n", verr)
			}
		} else {
			fmt.Fprintln(os.Stderr, ": OK")
		}
	}
}

func readFile(filename string) ([]byte, error) {
	var f io.ReadCloser
	if filename == "-" {
		f = os.Stdin
	} else {
		var err error
		f, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}
