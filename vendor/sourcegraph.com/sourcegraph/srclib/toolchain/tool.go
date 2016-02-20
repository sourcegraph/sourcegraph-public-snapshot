package toolchain

import "sourcegraph.com/sourcegraph/srclib"

// ToolInfo describes a tool in a toolchain.
type ToolInfo struct {
	// Subcmd is the subcommand name of this tool.
	//
	// By convention, this is the same as Op in toolchains that only have one
	// tool that performs this operation (e.g., a toolchain's "graph" subcommand
	// performs the "graph" operation).
	Subcmd string

	// Op is the operation that this tool performs (e.g., "scan", "graph",
	// "deplist", etc.).
	Op string

	// SourceUnitTypes is a list of source unit types (e.g., "GoPackage") that
	// this tool can operate on.
	//
	// If this tool doesn't operate on source units (for example, it operates on
	// directories or repositories, such as the "blame" tools), then this will
	// be empty.
	//
	// TODO(sqs): determine how repository- or directory-level tools will be
	// defined.
	SourceUnitTypes []string `json:",omitempty"`
}

// ListTools lists all tools in all available toolchains (returned by List). If
// op is non-empty, only tools that perform that operation are returned.
func ListTools(op string) ([]*srclib.ToolRef, error) {
	tcs, err := List()
	if err != nil {
		return nil, err
	}

	var tools []*srclib.ToolRef
	for _, tc := range tcs {
		c, err := tc.ReadConfig()
		if err != nil {
			return nil, err
		}

		for _, tool := range c.Tools {
			if op == "" || tool.Op == op {
				tools = append(tools, &srclib.ToolRef{Toolchain: tc.Path, Subcmd: tool.Subcmd})
			}
		}
	}
	return tools, nil
}
