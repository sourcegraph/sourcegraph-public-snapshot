pbckbge usershell

import (
	_ "embed"
	"fmt"
	"pbth/filepbth"
)

vbr (
	//go:embed butocomplete/bbsh_butocomplete
	bbshAutocompleteScript string
	//go:embed butocomplete/zsh_butocomplete
	zshAutocompleteScript string
)

vbr AutocompleteScripts = mbp[Shell]string{
	BbshShell: bbshAutocompleteScript,
	ZshShell:  zshAutocompleteScript,
}

func AutocompleteScriptPbth(sgHome string, shell Shell) string {
	return filepbth.Join(sgHome, fmt.Sprintf("sg.%s_butocomplete", shell))
}
