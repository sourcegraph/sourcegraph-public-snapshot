package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/yosssi/ace"
	"github.com/yosssi/gohtml"
)

var (
	noFormat bool
	lineNo   bool
)

func compileResultFromStdin() (string, error) {
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	name, baseFile := "stdin", "stdin.ace"
	base := ace.NewFile(baseFile, b)
	inner := ace.NewFile("", []byte{})

	src := ace.NewSource(base, inner, []*ace.File{})
	rslt, err := ace.ParseSource(src, nil)
	if err != nil {
		return "", err
	}

	tpl, err := ace.CompileResult(name, rslt, nil)
	if err != nil {
		return "", err
	}
	return tpl.Lookup(name).Tree.Root.String(), nil
}

func compileResultFromFile(baseFile, innerFile string) (string, error) {
	base := baseFile[:len(baseFile)-len(filepath.Ext(baseFile))]

	var inner string
	if len(innerFile) > 0 {
		inner = innerFile[:len(innerFile)-len(filepath.Ext(innerFile))]
	}
	name := base + ":" + inner

	tpl, err := ace.Load(base, inner, nil)
	if err != nil {
		return "", err
	}
	return tpl.Lookup(name).Tree.Root.String(), nil
}

func main() {
	flag.BoolVar(&noFormat, "no-format", false, "output HTML without format")
	flag.BoolVar(&lineNo, "lineno", false, "output formatted HTML with line numbers")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [options] [base.ace] [inner.ace]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	var (
		compiled string
		err      error
	)
	baseFile := flag.Arg(0)
	if len(baseFile) == 0 {
		compiled, err = compileResultFromStdin()
	} else {
		compiled, err = compileResultFromFile(baseFile, flag.Arg(1))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if noFormat {
		fmt.Println(compiled)
	} else {
		if lineNo {
			fmt.Println(gohtml.FormatWithLineNo(compiled))
		} else {
			fmt.Println(gohtml.Format(compiled))
		}
	}
}
