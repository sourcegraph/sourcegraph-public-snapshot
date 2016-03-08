// Package tmpl implements a protobuf template-based generator.
package tmpl // import "sourcegraph.com/sourcegraph/prototools/tmpl"

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"path"

	gateway "github.com/gengo/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
	"github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

// Generator is the type whose methods generate the output, stored in the associated response structure.
type Generator struct {
	// FileMap is the map of template files to use for the generation process.
	FileMap FileMap

	// RootDir is the root directory path prefix to place onto URLs for generated
	// types.
	RootDir string

	// APIHost is the base URL to use for rendering grpc-gateway routes, e.g.:
	//
	//  http://api.mysite.com/
	//
	APIHost string

	// ReadFile if non-nil is used to read template files, otherwise
	// ioutil.ReadFile is used.
	ReadFile func(path string) ([]byte, error)

	// request from protoc compiler, which should be set by the user of this
	// package via the SetRequest method.
	request *plugin.CodeGeneratorRequest

	// Response to protoc compiler.
	response *plugin.CodeGeneratorResponse

	// grpc-gateway registry used to determine HTTP routes.
	registry *gateway.Registry
}

// ParseFileMap parses and executes a filemap template.
func (g *Generator) ParseFileMap(dir, data string) error {
	// Parse the template data.
	t, err := template.New("").Funcs(Preload).Parse(data)
	if err != nil {
		return err
	}

	// Execute the template.
	buf := bytes.NewBuffer(nil)
	err = t.Execute(buf, g.request)
	if err != nil {
		return err
	}

	// Parse the filemap.
	g.FileMap.Dir = dir
	err = xml.Unmarshal(buf.Bytes(), &g.FileMap)
	if err != nil {
		return err
	}
	if len(g.FileMap.Generate) == 0 {
		return errors.New("no generate elements found in file map")
	}
	return nil
}

// Generate generates a response for g.Request (which you should unmarshal data
// into using protobuf).
//
// If any error is encountered during generation, it is returned and should be
// considered fatal to the generation process (the response will be nil).
func (g *Generator) Generate() (response *plugin.CodeGeneratorResponse, err error) {
	// Reset the response to its initial state.
	g.response.Reset()

	// Execute each generator.
	errs := bytes.NewBuffer(nil)
	for _, gen := range g.FileMap.Generate {
		f, err := g.GenerateOutput(gen.Output, nil)
		if err != nil {
			fmt.Fprintf(errs, "%s\n", err)
			continue
		}
		g.response.File = append(g.response.File, f)
	}

	if errs.Len() > 0 {
		g.response.File = nil
		errsStr := errs.String()
		g.response.Error = &errsStr
	}
	return g.response, nil
}

// GenerateOutput generates a CodeGeneratorResponse_File for the output file
// name.
//
// The ctx parameter specifies an arbitrary context for which to execute the
// template with, it is exposed to the executed template file as "Ctx".
func (g *Generator) GenerateOutput(name string, ctx interface{}) (*plugin.CodeGeneratorResponse_File, error) {
	// Find the generator with that output filename.
	for _, gen := range g.FileMap.Generate {
		if gen.Output != name {
			continue
		}

		// Execute in whichever mode is correct.
		if gen.Target != "" {
			return g.genTarget(gen, ctx)
		} else {
			return g.genNoTarget(gen, ctx)
		}
	}

	var outputs []string
	for _, gen := range g.FileMap.Generate {
		outputs = append(outputs, gen.Output)
	}
	return nil, fmt.Errorf("no such generator with output file %q\nvalid outputs are: %q", name, outputs)
}

// SetRequest sets the request the generator is generating a response for. If an
// error is returned generation is not safe (the request is bad) until a
// different request object is set successfully through this method.
func (g *Generator) SetRequest(r *plugin.CodeGeneratorRequest) error {
	g.request = r

	// Load into the grpc-gateway registry.
	return g.registry.Load(g.request)
}

// New returns a new generator for the given template.
func New() *Generator {
	return &Generator{
		response: &plugin.CodeGeneratorResponse{},
		registry: gateway.NewRegistry(),
	}
}

// genTarget a filemap generator with a specific target (e.g. for individual doc
// pages).
func (g *Generator) genTarget(gen *FileMapGenerate, userCtx interface{}) (*plugin.CodeGeneratorResponse_File, error) {
	var (
		buf       = bytes.NewBuffer(nil)
		protoFile = g.request.GetProtoFile()
		f         *descriptor.FileDescriptorProto
	)

	// Find the target proto file.
	for _, v := range protoFile {
		if gen.Target == v.GetName() {
			f = v
			break
		}
	}
	if f == nil {
		return nil, fmt.Errorf("no input proto file for generator target %q", gen.Target)
	}

	// Prepare the generators template.
	tmpl, err := g.prepare(gen)
	if err != nil {
		return nil, err
	}

	// Execute the template with this context and generate a response
	// for the input file.
	ctx := &tmplFuncs{
		f:          f,
		outputFile: gen.Output,
		rootDir:    g.RootDir,
		protoFile:  protoFile,
		registry:   g.registry,
		apiHost:    g.APIHost,
	}
	err = tmpl.Funcs(ctx.funcMap()).Execute(buf, struct {
		*descriptor.FileDescriptorProto
		Generate *FileMapGenerate
		Data     map[string]string
		Request  *plugin.CodeGeneratorRequest
		Ctx      interface{}
	}{
		f,
		gen,
		gen.DataMap(),
		g.request,
		userCtx,
	})
	if err != nil {
		return nil, err
	}

	// Generate the response file with the rendered template.
	return &plugin.CodeGeneratorResponse_File{
		Name:    proto.String(gen.Output),
		Content: proto.String(buf.String()),
	}, nil
}

// genNoTarget executes a target-less filemap generator (e.g. for index pages
// rather than individual doc pages). It panics if gen.Target != "".
func (g *Generator) genNoTarget(gen *FileMapGenerate, userCtx interface{}) (*plugin.CodeGeneratorResponse_File, error) {
	buf := bytes.NewBuffer(nil)

	// Only running generators not on proto files (i.e. generators without
	// targets).
	if gen.Target != "" {
		panic("expected a generator without a target")
	}

	// Prepare the generators template.
	tmpl, err := g.prepare(gen)
	if err != nil {
		return nil, err
	}

	// Execute the template with this context and generate a response file.
	ctx := &tmplFuncs{
		outputFile: gen.Output,
		rootDir:    g.RootDir,
		registry:   g.registry,
		apiHost:    g.APIHost,
	}
	err = tmpl.Funcs(ctx.funcMap()).Execute(buf, struct {
		*plugin.CodeGeneratorRequest
		Generate *FileMapGenerate
		Data     map[string]string
		Ctx      interface{}
	}{
		g.request,
		gen,
		gen.DataMap(),
		userCtx,
	})
	if err != nil {
		return nil, err
	}

	// Generate the response file with the rendered template.
	return &plugin.CodeGeneratorResponse_File{
		Name:    proto.String(gen.Output),
		Content: proto.String(buf.String()),
	}, nil
}

// loadTemplate is responsible for loading a single template and associating it
// with t. It reads the template file from g.ReadFile as appropriate.
func (g *Generator) loadTemplate(t *template.Template, tmplPath string) (*template.Template, error) {
	// Make the filepath relative to the filemap.
	tmplPath = g.FileMap.relative(tmplPath)[0]

	// Determine the open function.
	readFile := g.ReadFile
	if readFile == nil {
		readFile = ioutil.ReadFile
	}

	// Read the file.
	data, err := readFile(tmplPath)
	if err != nil {
		return nil, err
	}

	// Create a new template and parse.
	_, name := path.Split(tmplPath)
	return t.New(name).Parse(string(data))
}

// prepare prepares the given filemap generators template for execution,
// handling parsing of both the relative-path templates and their includes.
func (g *Generator) prepare(gen *FileMapGenerate) (*template.Template, error) {
	// Preload the function map (or else the functions will fail when
	// called due to a lack of valid context).
	var (
		t   = template.New("").Funcs(Preload)
		err error
	)

	// Parse the included template files.
	for _, inc := range gen.Include {
		_, err := g.loadTemplate(t, inc)
		if err != nil {
			return nil, err
		}
	}

	// Parse the template file to execute.
	tmpl, err := g.loadTemplate(t, gen.Template)
	if err != nil {
		return nil, err
	}
	_, name := path.Split(gen.Template)
	tmpl = tmpl.Lookup(name)
	return tmpl, nil
}
