// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"dmitri.shuralyov.com/text/kebabcase"
	"github.com/shurcooL/go-goon"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var oFlag = flag.String("o", "", "write output to `file` (default standard output)")

func main() {
	flag.Parse()

	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	f, err := os.Open(filepath.Join("_data", "data.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	var octicons map[string]octicon
	err = json.NewDecoder(f).Decode(&octicons)
	if err != nil {
		return err
	}

	var names []string
	for name := range octicons {
		names = append(names, name)
	}
	sort.Strings(names)

	var buf bytes.Buffer
	fmt.Fprint(&buf, `package octicon

import (
	"strconv"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Icon returns the named Octicon SVG node.
// It returns nil if name is not a valid Octicon symbol name.
func Icon(name string) *html.Node {
	switch name {
`)
	for _, name := range names {
		fmt.Fprintf(&buf, "	case %q:\n		return %v()\n", name, kebabcase.Parse(name).ToMixedCaps())
	}
	fmt.Fprint(&buf, `	default:
		return nil
	}
}

// SetSize sets size of icon, and returns a reference to it.
func SetSize(icon *html.Node, size int) *html.Node {
	icon.Attr[`, widthAttrIndex, `].Val = strconv.Itoa(size)
	icon.Attr[`, heightAttrIndex, `].Val = strconv.Itoa(size)
	return icon
}
`)

	// Write all individual Octicon functions.
	for _, name := range names {
		generateAndWriteOcticon(&buf, octicons, name)
	}

	var w io.Writer
	switch *oFlag {
	case "":
		w = os.Stdout
	default:
		f, err := os.Create(*oFlag)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	_, err = w.Write(buf.Bytes())
	return err
}

type octicon struct {
	Path   string
	Width  int
	Height int
}

func generateAndWriteOcticon(w io.Writer, octicons map[string]octicon, name string) {
	svgXML := generateOcticon(octicons[name])

	svg := parseOcticon(svgXML)
	// Clear these fields to remove cycles in the data structure, since go-goon
	// cannot print those in a way that's valid Go code. The generated data structure
	// is not a proper *html.Node with all fields set, but it's enough for rendering
	// to be successful.
	svg.LastChild = nil
	svg.FirstChild.Parent = nil

	fmt.Fprintln(w)
	fmt.Fprintf(w, "// %s returns an %q Octicon SVG node.\n", kebabcase.Parse(name).ToMixedCaps(), name)
	fmt.Fprintf(w, "func %s() *html.Node {\n", kebabcase.Parse(name).ToMixedCaps())
	fmt.Fprint(w, "	return ")
	goon.Fdump(w, svg)
	fmt.Fprintln(w, "}")
}

// These constants are used during generation of SetSize function.
// Keep them in sync with generateOcticon below.
const (
	widthAttrIndex  = 1
	heightAttrIndex = 2
)

func generateOcticon(o octicon) (svgXML string) {
	path := o.Path
	if strings.HasPrefix(path, `<path fill-rule="evenodd" `) {
		// Skip fill-rule, if present. It has no effect on displayed SVG, but takes up space.
		path = `<path ` + path[len(`<path fill-rule="evenodd" `):]
	}
	// Note, SetSize relies on the absolute position of the width, height attributes.
	// Keep them in sync with widthAttrIndex and heightAttrIndex.
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">%s</svg>`,
		o.Width, o.Height, o.Width, o.Height, path)
}

func parseOcticon(svgXML string) *html.Node {
	e, err := html.ParseFragment(strings.NewReader(svgXML), nil)
	if err != nil {
		panic(fmt.Errorf("internal error: html.ParseFragment failed: %v", err))
	}
	svg := e[0].LastChild.FirstChild // TODO: Is there a better way to just get the <svg>...</svg> element directly, skipping <html><head></head><body><svg>...</svg></body></html>?
	svg.Parent.RemoveChild(svg)
	for i, attr := range svg.Attr {
		if attr.Namespace == "" && attr.Key == "width" {
			svg.Attr[i].Val = "16"
			break
		}
	}
	svg.Attr = append(svg.Attr, html.Attribute{Key: atom.Style.String(), Val: `fill: currentColor; vertical-align: top;`})
	return svg
}
