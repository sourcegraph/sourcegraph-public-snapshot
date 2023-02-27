package template

import "text/template"

func New(name, tmpl, option string, ctxs ...template.FuncMap) (*template.Template, error) {
	t := template.New(name).Delims(startDelim, endDelim)
	if option != "" {
		t = t.Option(option)
	}

	t = t.Funcs(builtins)

	for _, ctx := range ctxs {
		t = t.Funcs(ctx)
	}

	return t.Parse(tmpl)
}
