package docgen

import (
	"bytes"

	"github.com/urfave/cli/v2"
)

func Default(app *cli.App) (string, error) {
	var w bytes.Buffer
	cli.HelpPrinterCustom(&w, app.CustomAppHelpTemplate, app, nil)
	return w.String(), nil
}
