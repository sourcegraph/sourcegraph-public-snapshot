// Package gopherjs_http provides helpers for compiling Go using GopherJS and serving it over HTTP.
package gopherjs_http

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/shurcooL/gopherjslib"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// StaticHtmlFile returns a handler that statically serves the given .html file, with the "text/go" script tags compiled to JavaScript via GopherJS.
//
// It reads file from disk and recompiles "text/go" script tags on startup only.
func StaticHtmlFile(name string) http.Handler {
	file, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	return &staticHtmlFile{
		content: ProcessHtml(file).Bytes(),
	}
}

type staticHtmlFile struct {
	content []byte
}

func (this *staticHtmlFile) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(this.content)
}

// HtmlFile returns a handler that serves the given .html file, with the "text/go" script tags compiled to JavaScript via GopherJS.
//
// It reads file from disk and recompiles "text/go" script tags on every request.
func HtmlFile(name string) http.Handler {
	return &htmlFile{name: name}
}

type htmlFile struct {
	name string
}

func (this *htmlFile) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	file, err := os.Open(this.name)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.Copy(w, ProcessHtml(file))
}

// https://gist.github.com/the42/1956518
func compress(s string) string {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := io.WriteString(gw, s)
	if err != nil {
		panic(err)
	}
	err = gw.Close()
	if err != nil {
		panic(err)
	}
	return buf.String()
}

// StaticGoFiles returns a handler that serves the given .go files compiled to JavaScript via GopherJS.
//
// It reads files from disk and recompiles on startup only.
func StaticGoFiles(goFiles ...string) http.Handler {
	content := handleJsError(goFilesToJs(goFiles))
	return &staticGoFiles{
		gzipContent: strings.NewReader(compress(content)),
		modtime:     time.Now(),
	}
}

type staticGoFiles struct {
	gzipContent io.ReadSeeker
	modtime     time.Time
}

func (this *staticGoFiles) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	w.Header().Set("Content-Encoding", "gzip") // TODO: Check "Accept-Encoding"?
	http.ServeContent(w, req, "", this.modtime, this.gzipContent)
}

// GoFiles returns a handler that serves the given .go files compiled to JavaScript via GopherJS.
//
// It reads files from disk and recompiles on every request.
func GoFiles(files ...string) http.Handler {
	return &goFiles{goFiles: files}
}

type goFiles struct {
	goFiles []string
}

func (this *goFiles) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	if isGzipEncodingAccepted(req) {
		w.Header().Set("Content-Encoding", "gzip")
		_, err := io.WriteString(w, compress(handleJsError(goFilesToJs(this.goFiles))))
		if err != nil {
			panic(err)
		}
	} else {
		_, err := io.WriteString(w, handleJsError(goFilesToJs(this.goFiles)))
		if err != nil {
			panic(err)
		}
	}
}

// ProcessHtml takes HTML with "text/go" script tags and replaces them with compiled JavaScript script tags.
//
// TODO: Write into writer, no need for buffer (unless want to be able to abort on error?). Or, alternatively, parse html and serve minified version?
func ProcessHtml(r io.Reader) *bytes.Buffer {
	insideTextGo := false
	tokenizer := html.NewTokenizer(r)
	var buf bytes.Buffer

	for {
		if tokenizer.Next() == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				return &buf
			}

			return &bytes.Buffer{}
		}

		raw := string(tokenizer.Raw())

		token := tokenizer.Token()
		switch token.Type {
		case html.DoctypeToken:
			buf.WriteString(token.String())
		case html.CommentToken:
			buf.WriteString(token.String())
		case html.StartTagToken:
			if token.DataAtom == atom.Script && getType(token.Attr) == "text/go" {
				insideTextGo = true

				buf.WriteString(`<script type="text/javascript">`)

				if srcs := getSrcs(token.Attr); len(srcs) != 0 {
					buf.WriteString(handleJsError(goFilesToJs(srcs)))
				}
			} else {
				buf.WriteString(token.String())
			}
		case html.EndTagToken:
			if token.DataAtom == atom.Script && insideTextGo {
				insideTextGo = false
			}
			buf.WriteString(token.String())
		case html.SelfClosingTagToken:
			// TODO: Support <script type="text/go" src="..." />.
			buf.WriteString(token.String())
		case html.TextToken:
			if insideTextGo {
				buf.WriteString(handleJsError(goToJs(token.Data)))
			} else {
				buf.WriteString(raw)
			}
		default:
			panic("unknown token type")
		}
	}
}

func getType(attrs []html.Attribute) string {
	for _, attr := range attrs {
		if attr.Key == "type" {
			return attr.Val
		}
	}
	return ""
}

func getSrcs(attrs []html.Attribute) (srcs []string) {
	for _, attr := range attrs {
		if attr.Key == "src" {
			srcs = append(srcs, attr.Val)
		}
	}
	return srcs
}

func handleJsError(jsCode string, err error) string {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return `console.error("` + template.JSEscapeString(err.Error()) + `");`
	}
	return jsCode
}

// Needed to prevent race condition until https://github.com/go-on/gopherjslib/issues/2 is resolved.
var gopherjslibLock sync.Mutex

func goFilesToJs(goFiles []string) (jsCode string, err error) {
	started := time.Now()
	defer func() { fmt.Println("goFilesToJs taken:", time.Since(started)) }()
	gopherjslibLock.Lock()
	defer gopherjslibLock.Unlock()

	var out bytes.Buffer
	builder := gopherjslib.NewBuilder(&out, nil)

	for _, goFile := range goFiles {
		file, err := os.Open(goFile)
		if err != nil {
			return "", err
		}
		defer file.Close()

		builder.Add(goFile, file)
	}

	err = builder.Build()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func goReadersToJs(names []string, goReaders []io.Reader) (jsCode string, err error) {
	started := time.Now()
	defer func() { fmt.Println("goReadersToJs taken:", time.Since(started)) }()
	gopherjslibLock.Lock()
	defer gopherjslibLock.Unlock()

	var out bytes.Buffer
	builder := gopherjslib.NewBuilder(&out, &gopherjslib.Options{Minify: true})

	for i, goReader := range goReaders {
		builder.Add(names[i], goReader)
	}

	err = builder.Build()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func goToJs(goCode string) (jsCode string, err error) {
	started := time.Now()
	defer func() { fmt.Println("goToJs taken:", time.Since(started)) }()
	gopherjslibLock.Lock()
	defer gopherjslibLock.Unlock()

	code := strings.NewReader(goCode)

	var out bytes.Buffer
	err = gopherjslib.Build(code, &out, nil)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

// isGzipEncodingAccepted returns true if the request includes "gzip" under Accept-Encoding header.
func isGzipEncodingAccepted(req *http.Request) bool {
	for _, v := range strings.Split(req.Header.Get("Accept-Encoding"), ",") {
		if strings.TrimSpace(v) == "gzip" {
			return true
		}
	}
	return false
}
