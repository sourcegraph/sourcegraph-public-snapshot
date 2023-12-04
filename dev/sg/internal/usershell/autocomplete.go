package usershell

import (
	_ "embed"
	"fmt"
	"path/filepath"
)

var (
	//go:embed autocomplete/bash_autocomplete
	bashAutocompleteScript string
	//go:embed autocomplete/zsh_autocomplete
	zshAutocompleteScript string
)

var AutocompleteScripts = map[Shell]string{
	BashShell: bashAutocompleteScript,
	ZshShell:  zshAutocompleteScript,
}

func AutocompleteScriptPath(sgHome string, shell Shell) string {
	return filepath.Join(sgHome, fmt.Sprintf("sg.%s_autocomplete", shell))
}
