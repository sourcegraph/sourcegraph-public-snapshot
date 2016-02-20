package ace

import (
	"bytes"
	"fmt"
	"html/template"
)

// Actions
const (
	actionDefine               = `%sdefine "%s"%s`
	actionEnd                  = "%send%s"
	actionTemplate             = `%stemplate "%s"%s`
	actionTemplateWithPipeline = `%stemplate "%s" %s%s`
)

// PreDefinedFuncs
const (
	preDefinedFuncNameHTML = "HTML"
)

// CompileResult compiles the parsed result to the template.Template.
func CompileResult(name string, rslt *result, opts *Options) (*template.Template, error) {
	// Initialize the options.
	opts = InitializeOptions(opts)

	// Create a template.
	t := template.New(name)

	return CompileResultWithTemplate(t, rslt, opts)
}

// CompileResultWithTemplate compiles the parsed result and associates it with t.
func CompileResultWithTemplate(t *template.Template, rslt *result, opts *Options) (*template.Template, error) {
	// Initialize the options.
	opts = InitializeOptions(opts)

	var err error

	// Create a buffer.
	baseBf := bytes.NewBuffer(nil)
	innerBf := bytes.NewBuffer(nil)
	includeBfs := make(map[string]*bytes.Buffer)

	// Write data to the buffer.
	for _, e := range rslt.base {
		if _, err := e.WriteTo(baseBf); err != nil {
			return nil, err
		}
	}

	for _, e := range rslt.inner {
		if _, err = e.WriteTo(innerBf); err != nil {
			return nil, err
		}
	}

	for path, elements := range rslt.includes {
		bf := bytes.NewBuffer(nil)

		// Write a define action.
		bf.WriteString(fmt.Sprintf(actionDefine, opts.DelimLeft, path, opts.DelimRight))

		for _, e := range elements {
			if _, err = e.WriteTo(bf); err != nil {
				return nil, err
			}
		}

		// Write an end action.
		bf.WriteString(fmt.Sprintf(actionEnd, opts.DelimLeft, opts.DelimRight))

		includeBfs[path] = bf
	}

	// Set Delimiters.
	t.Delims(opts.DelimLeft, opts.DelimRight)

	// Set FuncMaps.
	t.Funcs(template.FuncMap{
		preDefinedFuncNameHTML: func(s string) template.HTML {
			return template.HTML(s)
		},
	})

	t.Funcs(opts.FuncMap)

	// Parse a string to the template.
	t, err = t.Parse(baseBf.String())
	if err != nil {
		return nil, err
	}

	t, err = t.Parse(innerBf.String())
	if err != nil {
		return nil, err
	}

	for _, bf := range includeBfs {
		t, err = t.Parse(bf.String())
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}
