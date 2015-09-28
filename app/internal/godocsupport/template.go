package godocsupport

import (
	"bytes"
	"fmt"
	godoc "go/doc"
	"html"
	htemp "html/template"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"github.com/sourcegraph/gddo/doc"
	"github.com/sourcegraph/gddo/gosrc"
)

// TemplateFuncMap exposes godoc-related template funcs defined in
// this package.
var TemplateFuncMap = htemp.FuncMap{
	"godoc_comment": commentFn,
	"godoc_code":    codeFn,
}

type TDoc struct {
	*doc.Package
	allExamples []*texample
}

type texample struct {
	ID      string
	Label   string
	Example *doc.Example
	obj     interface{}
}

func NewTDoc(pdoc *doc.Package) *TDoc {
	return &TDoc{Package: pdoc}
}

func (pdoc *TDoc) Code(pos doc.Pos, repoRevSpec sourcegraph.RepoRevSpec) htemp.HTML {
	v := repoRevSpec.RouteVars()

	var subdir string
	if pdoc.ImportPath != pdoc.ProjectRoot && strings.HasPrefix(pdoc.ImportPath, pdoc.ProjectRoot) {
		subdir = relativePathFn(pdoc.ImportPath, pdoc.ProjectRoot)
	} else if repoRevSpec.URI == "github.com/golang/go" {
		subdir = filepath.Join("src", pdoc.ImportPath)
	}

	return htemp.HTML(fmt.Sprintf(`<script type="text/javascript" src="/%s@%s/.tree/%s/.sourcebox.js?StartLine=%d&EndLine=%d"></script>`, html.EscapeString(v["Repo"]), html.EscapeString(v["Rev"]), html.EscapeString(path.Join(subdir, pdoc.Files[pos.File].Name)), pos.Line, pos.Line+int32(pos.N)))
}

func (pdoc *TDoc) PageName() string {
	if pdoc.Name != "" && !pdoc.IsCmd {
		return pdoc.Name
	}
	_, name := path.Split(pdoc.ImportPath)
	return name
}

func (pdoc *TDoc) DefSpec(obj, parent interface{}, repoRevSpec sourcegraph.RepoRevSpec) sourcegraph.DefSpec {
	defSpec := sourcegraph.DefSpec{
		Repo:     pdoc.Package.ProjectRoot,
		CommitID: repoRevSpec.CommitID,
		UnitType: "GoPackage",
		Unit:     pdoc.Package.ImportPath,
	}
	objName := func(v interface{}) string {
		switch v := v.(type) {
		case *doc.Func:
			return v.Name
		case *doc.Type:
			return v.Name
		}
		return ""
	}
	if parent == nil {
		defSpec.Path = objName(obj)
	} else {
		defSpec.Path = objName(parent) + "/" + objName(obj)
	}
	return defSpec
}

func (pdoc *TDoc) addExamples(obj interface{}, export, method string, examples []*doc.Example) {
	label := export
	id := export
	if method != "" {
		label += "." + method
		id += "-" + method
	}
	for _, e := range examples {
		te := &texample{Label: label, ID: id, Example: e, obj: obj}
		if e.Name != "" {
			te.Label += " (" + e.Name + ")"
			if method == "" {
				te.ID += "-"
			}
			te.ID += "-" + e.Name
		}
		pdoc.allExamples = append(pdoc.allExamples, te)
	}
}

type byExampleID []*texample

func (e byExampleID) Len() int           { return len(e) }
func (e byExampleID) Less(i, j int) bool { return e[i].ID < e[j].ID }
func (e byExampleID) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

func (pdoc *TDoc) AllExamples() []*texample {
	if pdoc.allExamples != nil {
		return pdoc.allExamples
	}
	pdoc.allExamples = make([]*texample, 0)
	pdoc.addExamples(pdoc, "package", "", pdoc.Examples)
	for _, f := range pdoc.Funcs {
		pdoc.addExamples(f, f.Name, "", f.Examples)
	}
	for _, t := range pdoc.Types {
		pdoc.addExamples(t, t.Name, "", t.Examples)
		for _, f := range t.Funcs {
			pdoc.addExamples(f, f.Name, "", f.Examples)
		}
		for _, m := range t.Methods {
			if len(m.Examples) > 0 {
				pdoc.addExamples(m, t.Name, m.Name, m.Examples)
			}
		}
	}
	sort.Sort(byExampleID(pdoc.allExamples))
	return pdoc.allExamples
}

func (pdoc *TDoc) ObjExamples(obj interface{}) []*texample {
	var examples []*texample
	for _, e := range pdoc.allExamples {
		if e.obj == obj {
			examples = append(examples, e)
		}
	}
	return examples
}

func (pdoc *TDoc) Breadcrumbs(templateName string) htemp.HTML {
	if !strings.HasPrefix(pdoc.ImportPath, pdoc.ProjectRoot) {
		return ""
	}
	var buf bytes.Buffer
	i := 0
	j := len(pdoc.ProjectRoot)
	if j == 0 {
		j = strings.IndexRune(pdoc.ImportPath, '/')
		if j < 0 {
			j = len(pdoc.ImportPath)
		}
	}
	for {
		if i != 0 {
			buf.WriteString(`<span class="text-muted">/</span>`)
		}
		link := j < len(pdoc.ImportPath) ||
			(templateName != "dir.html" && templateName != "cmd.html" && templateName != "pkg.html")
		if link {
			buf.WriteString(`<a href="`)
			buf.WriteString(formatPathFrag(pdoc.ImportPath[:j], ""))
			buf.WriteString(`">`)
		} else {
			buf.WriteString(`<span class="text-muted">`)
		}
		buf.WriteString(htemp.HTMLEscapeString(pdoc.ImportPath[i:j]))
		if link {
			buf.WriteString("</a>")
		} else {
			buf.WriteString("</span>")
		}
		i = j + 1
		if i >= len(pdoc.ImportPath) {
			break
		}
		j = strings.IndexRune(pdoc.ImportPath[i:], '/')
		if j < 0 {
			j = len(pdoc.ImportPath)
		} else {
			j += i
		}
	}
	return htemp.HTML(buf.String())
}

func formatPathFrag(path, fragment string) string {
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	u := url.URL{Path: path, Fragment: fragment}
	return u.String()
}

var (
	h3Pat      = regexp.MustCompile(`<h3 id="([^"]+)">([^<]+)</h3>`)
	rfcPat     = regexp.MustCompile(`RFC\s+(\d{3,4})`)
	packagePat = regexp.MustCompile(`\s+package\s+([-a-z0-9]\S+)`)
)

func replaceAll(src []byte, re *regexp.Regexp, replace func(out, src []byte, m []int) []byte) []byte {
	var out []byte
	for len(src) > 0 {
		m := re.FindSubmatchIndex(src)
		if m == nil {
			break
		}
		out = append(out, src[:m[0]]...)
		out = replace(out, src, m)
		src = src[m[1]:]
	}
	if out == nil {
		return src
	}
	return append(out, src...)
}

// commentFn formats a source code comment as HTML.
func commentFn(v string) htemp.HTML {
	var buf bytes.Buffer
	godoc.ToHTML(&buf, v, nil)
	p := buf.Bytes()
	p = replaceAll(p, h3Pat, func(out, src []byte, m []int) []byte {
		out = append(out, `<h4 id="`...)
		out = append(out, src[m[2]:m[3]]...)
		out = append(out, `">`...)
		out = append(out, src[m[4]:m[5]]...)
		out = append(out, ` <a class="permalink" href="#`...)
		out = append(out, src[m[2]:m[3]]...)
		out = append(out, `">&para</a></h4>`...)
		return out
	})
	p = replaceAll(p, rfcPat, func(out, src []byte, m []int) []byte {
		out = append(out, `<a href="http://tools.ietf.org/html/rfc`...)
		out = append(out, src[m[2]:m[3]]...)
		out = append(out, `">`...)
		out = append(out, src[m[0]:m[1]]...)
		out = append(out, `</a>`...)
		return out
	})
	p = replaceAll(p, packagePat, func(out, src []byte, m []int) []byte {
		path := bytes.TrimRight(src[m[2]:m[3]], ".!?:")
		if !gosrc.IsValidPath(string(path)) {
			return append(out, src[m[0]:m[1]]...)
		}
		out = append(out, src[m[0]:m[2]]...)
		out = append(out, `<a href="/`...)
		out = append(out, path...)
		out = append(out, `">`...)
		out = append(out, path...)
		out = append(out, `</a>`...)
		out = append(out, src[m[2]+len(path):m[1]]...)
		return out
	})
	return htemp.HTML(p)
}

var period = []byte{'.'}

func codeFn(c doc.Code, typ *doc.Type) htemp.HTML {
	var buf bytes.Buffer
	last := 0
	src := []byte(c.Text)
	for _, a := range c.Annotations {
		htemp.HTMLEscape(&buf, src[last:a.Pos])
		switch a.Kind {
		case doc.PackageLinkAnnotation:
			buf.WriteString(`<a href="`)
			buf.WriteString(formatPathFrag(c.Paths[a.PathIndex], ""))
			buf.WriteString(`">`)
			htemp.HTMLEscape(&buf, src[a.Pos:a.End])
			buf.WriteString(`</a>`)
		case doc.LinkAnnotation, doc.BuiltinAnnotation:
			var p string
			if a.Kind == doc.BuiltinAnnotation {
				p = "builtin"
			} else if a.PathIndex >= 0 {
				p = c.Paths[a.PathIndex]
			}
			n := src[a.Pos:a.End]
			n = n[bytes.LastIndex(n, period)+1:]
			buf.WriteString(`<a href="`)
			buf.WriteString(formatPathFrag(p, string(n)))
			buf.WriteString(`">`)
			htemp.HTMLEscape(&buf, src[a.Pos:a.End])
			buf.WriteString(`</a>`)
		case doc.CommentAnnotation:
			buf.WriteString(`<span class="com">`)
			htemp.HTMLEscape(&buf, src[a.Pos:a.End])
			buf.WriteString(`</span>`)
		case doc.AnchorAnnotation:
			buf.WriteString(`<span id="`)
			if typ != nil {
				htemp.HTMLEscape(&buf, []byte(typ.Name))
				buf.WriteByte('.')
			}
			htemp.HTMLEscape(&buf, src[a.Pos:a.End])
			buf.WriteString(`">`)
			htemp.HTMLEscape(&buf, src[a.Pos:a.End])
			buf.WriteString(`</span>`)
		default:
			htemp.HTMLEscape(&buf, src[a.Pos:a.End])
		}
		last = int(a.End)
	}
	htemp.HTMLEscape(&buf, src[last:])
	return htemp.HTML(buf.String())
}

func relativePathFn(path string, parentPath interface{}) string {
	if p, ok := parentPath.(string); ok && p != "" && strings.HasPrefix(path, p) {
		path = path[len(p)+1:]
	}
	return path
}
