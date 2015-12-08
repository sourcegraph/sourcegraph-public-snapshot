package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/sourcegraph/annotate"
	"src.sourcegraph.com/syntaxhighlight"
)

var (
	jsonOutput bool
	raw        bool
	stylesheet string
	ext        string
	mime       string
	htmlConfig string
)

func init() {
	flag.BoolVar(&jsonOutput, "json", false, "Print JSON representation of tokens instead of formatting them")
	flag.BoolVar(&raw, "raw", false, "Do not output stylesheet and wrapping HTML")
	flag.StringVar(&stylesheet, "stylesheet", "default.css", "Use specific stylesheet")
	flag.StringVar(&ext, "ext", ``, "Force specific file extension")
	flag.StringVar(&mime, "mime", ``, "Force specific MIME type")
	flag.StringVar(&htmlConfig, "config", ``, "Force specific HTML config")
}

func main() {
	flag.Parse()
	sourceFile := flag.Arg(0)
	source, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	re := regexp.MustCompile("\r\n|\r")
	source = re.ReplaceAll(source, []byte{'\n'})

	if ext == `` && mime == `` {
		ext = filepath.Ext(sourceFile)
	}

	var annotator syntaxhighlight.Annotator
	if jsonOutput {
		annotator = syntaxhighlight.NewJSONAnnotator(true, os.Stdout)
		syntaxhighlight.Annotate(source, ext, mime, annotator)
	} else {
		var cfg syntaxhighlight.HTMLConfig
		if htmlConfig == `pygments` {
			cfg = syntaxhighlight.PygmentsHTMLConfig
		} else if htmlConfig == `prettify` {
			cfg = syntaxhighlight.GooglePrettifyHTMLConfig
		} else {
			cfg = syntaxhighlight.DefaultHTMLConfig
		}
		annotator = syntaxhighlight.NewHTMLAnnotator(cfg)
		if !raw {
			fmt.Println(`<link rel="stylesheet" href="` + stylesheet + `"/>`)
			fmt.Println(`<div class="codehilite"><pre>`)
		}
		annotations, _ := syntaxhighlight.Annotate(source, ext, mime, annotator)
		content, _ := annotate.Annotate(source, annotations, template.HTMLEscape)
		fmt.Println(string(content))
		if !raw {
			fmt.Println(`</pre></div>`)
		}
	}
}
