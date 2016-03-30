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
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/syntaxhighlight"
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
	flag.StringVar(&ext, "ext", "", "Force specific file extension")
	flag.StringVar(&mime, "mime", "", "Force specific MIME type")
	flag.StringVar(&htmlConfig, "config", "", "Force specific HTML config")
}

// Scans source code src and produces array of annotations.
// src is source code to scan.
// fileName is source code file name, used to determine lexer to use. It has precedence over mimeType.
// mimeType is MIME type of source code, also used to determine lexer to use.
// annotator transforms tokens to annotations.
func Annotate(src []byte, fileName string, mimeType string, annotator syntaxhighlight.Annotator) (annotate.Annotations, error) {
	var lexer syntaxhighlight.Lexer
	if fileName != "" {
		lexer = syntaxhighlight.NewLexerByExtension(filepath.Ext(fileName))
	} else {
		lexer = syntaxhighlight.NewLexerByMimeType(mimeType)
	}
	if lexer == nil {
		// falling back
		lexer = &syntaxhighlight.FallbackLexer{}
	}

	return syntaxhighlight.Annotate(src, lexer, annotator)
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

	if ext == "" && mime == "" {
		ext = filepath.Ext(sourceFile)
	}

	var annotator syntaxhighlight.Annotator
	if jsonOutput {
		annotator = syntaxhighlight.NewJSONAnnotator(true, os.Stdout)
		Annotate(source, ext, mime, annotator)
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
		annotations, _ := Annotate(source, ext, mime, annotator)
		content, _ := annotate.Annotate(source, annotations, template.HTMLEscape)
		fmt.Println(string(content))
		if !raw {
			fmt.Println(`</pre></div>`)
		}
	}
}
