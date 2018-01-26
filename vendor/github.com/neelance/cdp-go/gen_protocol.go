// +build generate

package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"text/template"
)

type Protocol struct {
	Domains []*Domain
}

type Domain struct {
	Domain       string
	Dependencies []string
	Types        []*Type
	Commands     []*Command
	Events       []*Event
	Description  string
	Experimental bool

	Imports []string `json:"-"`
}

func (d *Domain) GoPackage() string {
	return strings.ToLower(d.Domain)
}

func (d *Domain) Doc() string {
	doc := d.Description
	if d.Experimental {
		doc += " (experimental)"
	}
	return strings.TrimSpace(doc)
}

func (d *Domain) lookupType(id string) *Type {
	for _, t := range d.Types {
		if t.ID == id {
			return t
		}
	}
	panic("type not found")
}

func (d *Domain) addImport(name string) {
	for _, imp := range d.Imports {
		if imp == name {
			return
		}
	}
	d.Imports = append(d.Imports, name)
}

type Type struct {
	ID           string
	Description  string
	Experimental bool
	Properties   []*Property
	TypeRef
}

func (t *Type) Doc() string {
	doc := t.Description
	if t.Experimental {
		doc += " (experimental)"
	}
	return strings.TrimSpace(doc)
}

type Command struct {
	Name         string
	Parameters   []*Property
	Returns      []*Property
	Description  string
	Experimental bool
}

func (c *Command) GoName() string {
	return strings.ToUpper(c.Name[:1]) + c.Name[1:]
}

func (c *Command) GoRequestType() string {
	return c.GoName() + "Request"
}

func (c *Command) GoResultType() string {
	return c.GoName() + "Result"
}

func (c *Command) Doc() string {
	doc := c.Description
	if c.Experimental {
		doc += " (experimental)"
	}
	return strings.TrimSpace(doc)
}

type Event struct {
	Name         string
	Parameters   []*Property
	Description  string
	Experimental bool
}

func (e *Event) GoName() string {
	return strings.ToUpper(e.Name[:1]) + e.Name[1:]
}

func (e *Event) GoType() string {
	return e.GoName() + "Event"
}

func (e *Event) Doc() string {
	doc := e.Description
	if e.Experimental {
		doc += " (experimental)"
	}
	return strings.TrimSpace(doc)
}

type Property struct {
	Name string
	TypeRef
	Description  string
	Optional     bool
	Experimental bool
}

func (p *Property) GoName() string {
	switch p.Name {
	case "url":
		return "URL"
	}
	return strings.ToUpper(p.Name[:1]) + p.Name[1:]
}

func (p *Property) Doc() string {
	doc := p.Description
	switch {
	case p.Optional && p.Experimental:
		doc += " (optional, experimental)"
	case p.Optional:
		doc += " (optional)"
	case p.Optional && p.Experimental:
		doc += " (experimental)"
	}
	return strings.TrimSpace(doc)
}

type TypeRef struct {
	Type  string
	Ref   string `json:"$ref"`
	Items *TypeRef
}

func goType(domains []*Domain, d *Domain, t *TypeRef) string {
	if ref := t.Ref; t.Ref != "" {
		if ref == "Page.FrameId" || ref == "Page.ResourceType" {
			return "string" // avoid circular dependency
		}

		pkg := ""
		if i := strings.Index(ref, "."); i != -1 {
			refD := findDomain(domains, ref[:i])
			d.addImport("github.com/neelance/cdp-go/protocol/" + refD.GoPackage())
			pkg = refD.GoPackage() + "."
			d = refD
			ref = ref[i+1:]
		}

		if d.lookupType(ref).Type == "object" {
			return "*" + pkg + ref
		}
		return pkg + ref
	}

	switch t.Type {
	case "string":
		return "string"
	case "boolean":
		return "bool"
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "array":
		return "[]" + goType(domains, d, t.Items)
	case "any", "object":
		return "interface{}"
	default:
		panic("unknown type: " + t.Type)
	}
}

func findDomain(domains []*Domain, name string) *Domain {
	for _, d := range domains {
		if d.Domain == name {
			return d
		}
	}
	panic("domain not found")
}

func main() {
	var domains []*Domain
	readProtocol := func(filename string) {
		in, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		var protocol Protocol
		if err := json.NewDecoder(in).Decode(&protocol); err != nil {
			panic(err)
		}
		domains = append(domains, protocol.Domains...)
		in.Close()
	}
	readProtocol("devtools-protocol/json/browser_protocol.json")
	readProtocol("devtools-protocol/json/js_protocol.json")

	sort.Slice(domains, func(i, j int) bool {
		return domains[i].Domain < domains[j].Domain
	})

	os.RemoveAll("protocol")
	os.Mkdir("protocol", 0777)

	for _, d := range domains {
		t := template.Must(template.New("").Funcs(template.FuncMap{
			"goType": func(t *TypeRef) string {
				return goType(domains, d, t)
			},
		}).Parse(domainTmpl))

		// collect imports
		if err := t.Execute(ioutil.Discard, d); err != nil {
			panic(err)
		}

		var buf bytes.Buffer
		if err := t.Execute(&buf, d); err != nil {
			panic(err)
		}

		dir := "protocol/" + d.GoPackage()
		os.Mkdir(dir, 0777)
		if err := ioutil.WriteFile(dir+"/"+d.GoPackage()+".go", buf.Bytes(), 0666); err != nil {
			panic(err)
		}
	}

	t := template.Must(template.New("").Parse(clientTmpl))
	var buf bytes.Buffer
	if err := t.Execute(&buf, domains); err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile("client.go", buf.Bytes(), 0666); err != nil {
		panic(err)
	}
}

const domainTmpl = `
{{if .Doc}}// {{.Doc}}{{end}}
package {{.GoPackage}}
{{$domain := .Domain}}
import (
	"github.com/neelance/cdp-go/rpc"

	{{range .Imports}}
		"{{.}}"
	{{- end}}
)

{{if .Doc}}// {{.Doc}}{{end}}
type Client struct {
	*rpc.Client
}

{{range .Types}}
	{{if .Doc}}// {{.Doc}}{{end}}
	{{if eq .Type "object"}}
		type {{.ID}} struct {
			{{- range .Properties}}
				{{if .Doc}}// {{.Doc}}{{end}}
				{{.GoName}} {{goType .TypeRef}} ` + "`" + `json:"{{.Name}}{{if .Optional}},omitempty{{end}}"` + "`" + `
			{{end}}
		}
	{{else}}
		type {{.ID}} {{goType .TypeRef}}
	{{end}}
{{end}}

{{range .Commands}}
	{{$reqType := .GoRequestType}}
	type {{$reqType}} struct {
		client *rpc.Client
		opts map[string]interface{}
	}

	{{if .Doc}}// {{.Doc}}{{end}}
	func (d *Client) {{.GoName}}() *{{$reqType}} {
		return &{{$reqType}}{opts: make(map[string]interface{}), client: d.Client}
	}

	{{- range .Parameters}}
		{{if .Doc}}// {{.Doc}}{{end}}
		func (r *{{$reqType}}) {{.GoName}}(v {{goType .TypeRef}}) *{{$reqType}} {
			r.opts["{{.Name}}"] = v
			return r
		}
	{{end}}

	{{if .Returns}}
		type {{.GoResultType}} struct {
			{{- range .Returns}}
				{{if .Doc}}// {{.Doc}}{{end}}
				{{.GoName}} {{goType .TypeRef}} ` + "`" + `json:"{{.Name}}"` + "`" + `
			{{end}}
		}

		func (r *{{.GoRequestType}}) Do() (*{{.GoResultType}}, error) {
			var result {{.GoResultType}}
			err := r.client.Call("{{$domain}}.{{.Name}}", r.opts, &result)
			return &result, err
		}
	{{else}}
		func (r *{{.GoRequestType}}) Do() error {
			return r.client.Call("{{$domain}}.{{.Name}}", r.opts, nil)
		}
	{{end}}
{{end}}

func init() {
	{{- range .Events}}
		rpc.EventTypes["{{$domain}}.{{.Name}}"] = func() interface{} { return new({{.GoType}}) }
	{{- end}}
}

{{range .Events}}
	{{if .Doc}}// {{.Doc}}{{end}}
	type {{.GoType}} struct {
		{{- range .Parameters}}
			{{if .Doc}}// {{.Doc}}{{end}}
			{{.GoName}} {{goType .TypeRef}} ` + "`" + `json:"{{.Name}}"` + "`" + `
		{{end}}
	}
{{end}}
`

const clientTmpl = `
package cdp

import (
	"golang.org/x/net/websocket"

	"github.com/neelance/cdp-go/rpc"

	{{range .}}
		"github.com/neelance/cdp-go/protocol/{{.GoPackage}}"
	{{- end}}
)

type Client struct {
	*rpc.Client

	{{range .}}
		{{.Domain}} {{.GoPackage}}.Client
	{{- end}}
}

func Dial(url string) *Client {
	conn, err := websocket.Dial(url, "", url)
	if err != nil {
		panic(err)
	}

	cl := rpc.NewClient(conn)
	return &Client{
		Client: cl,

		{{range .}}
			{{.Domain}}: {{.GoPackage}}.Client{Client: cl},
		{{- end}}
	}
}
`
