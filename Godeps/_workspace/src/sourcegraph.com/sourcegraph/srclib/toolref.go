package srclib

import (
	"errors"
	"fmt"
	"strings"
)

// A ToolRef identifies a tool inside a specific toolchain. It can be used to
// look up the tool.
type ToolRef struct {
	// Toolchain is the toolchain path of the toolchain that contains this tool.
	Toolchain string

	// Subcmd is the name of the toolchain subcommand that runs this tool.
	Subcmd string
}

func (t ToolRef) String() string { return fmt.Sprintf("%s %s", t.Toolchain, t.Subcmd) }

func (t *ToolRef) UnmarshalFlag(value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return errors.New("expected format 'TOOLCHAIN:TOOL' (separated by 1 colon)")
	}
	t.Toolchain = parts[0]
	t.Subcmd = parts[1]
	return nil
}

func (t ToolRef) MarshalFlag() (string, error) {
	return t.Toolchain + ":" + t.Subcmd, nil
}
