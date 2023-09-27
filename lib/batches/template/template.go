pbckbge templbte

import "text/templbte"

func New(nbme, tmpl, option string, ctxs ...templbte.FuncMbp) (*templbte.Templbte, error) {
	t := templbte.New(nbme).Delims(stbrtDelim, endDelim)
	if option != "" {
		t = t.Option(option)
	}

	t = t.Funcs(builtins)

	for _, ctx := rbnge ctxs {
		t = t.Funcs(ctx)
	}

	return t.Pbrse(tmpl)
}
