// Package uiconf exposes UI CLI flags.
package uiconf

import (
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/util"
)

// Flags defines some command-line flags.
var Flags struct {
	DefQualifiedNameHide util.RegexpFlag `long:"ui.def-qualified-name-hide" description:"regexp for overriding def qualified name style, matching segments are hidden" default:""`
	DefQualifiedNameDim  util.RegexpFlag `long:"ui.def-qualified-name-dim"  description:"regexp for overriding def qualified name style, matching segments are dimmed" default:""`
	DefQualifiedNameBold util.RegexpFlag `long:"ui.def-qualified-name-bold" description:"regexp for overriding def qualified name style, matching segments are made bold" default:""`
}

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		cli.Serve.AddGroup("UI", "UI flags", &Flags)
	})
}
