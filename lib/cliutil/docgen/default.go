package docgen

import (
	"bytes"

	"github.com/urfave/cli/v3"
)

// Default renders help text for the app using urfave/cli's default help format.
func Default(app *cli.Command) (string, error) {
	tpl := app.CustomRootCommandHelpTemplate
	if tpl == "" {
		tpl = cli.RootCommandHelpTemplate
	}

	var w bytes.Buffer
	cli.HelpPrinterCustom(&w, tpl, app, nil)
	return w.String(), nil
}
