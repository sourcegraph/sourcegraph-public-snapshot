package docgen

import (
	"bytes"

	"github.com/urfave/cli/v2"
)

// Default renders help text for the app using urfave/cli's default help format.
func Default(app *cli.App) (string, error) {
	tpl := app.CustomAppHelpTemplate
	if tpl == "" {
		tpl = cli.AppHelpTemplate
	}

	var w bytes.Buffer
	cli.HelpPrinterCustom(&w, tpl, app, nil)
	return w.String(), nil
}
