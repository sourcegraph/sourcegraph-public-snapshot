package toolchain

import (
	"fmt"

	"sourcegraph.com/sourcegraph/srclib"
)

// ChooseTool determines which toolchain and tool to use to run op (graph,
// depresolve, etc.) on a source unit of the given type. If no tools fit the
// criteria, an error is returned.
//
// The selection algorithm is currently very simplistic: if exactly one tool is
// found that can perform op on the source unit type, it is returned. If zero or
// more than 1 are found, then an error is returned. TODO(sqs): extend this to
// choose the "best" tool when multiple tools would suffice.
var ChooseTool = func(op, unitType string) (*srclib.ToolRef, error) {
	tcs, err := List()
	if err != nil {
		return nil, err
	}
	return chooseTool(op, unitType, tcs)
}

// chooseTool is like ChooseTool but the list of tools is provided as an
// argument instead of being obtained by calling List.
func chooseTool(op, unitType string, tcs []*Info) (*srclib.ToolRef, error) {
	var satisfying []*srclib.ToolRef
	for _, tc := range tcs {
		cfg, err := tc.ReadConfig()
		if err != nil {
			return nil, err
		}

		for _, tool := range cfg.Tools {
			if tool.Op == op {
				for _, u := range tool.SourceUnitTypes {
					if u == unitType {
						satisfying = append(satisfying, &srclib.ToolRef{Toolchain: tc.Path, Subcmd: tool.Subcmd})
					}
				}
			}
		}
	}

	if n := len(satisfying); n > 1 {
		return nil, fmt.Errorf("%d tools satisfy op %q for source unit type %q (refusing to choose between multiple possibilities)", n, op, unitType)
	} else if n == 0 {
		return nil, nil
	}
	return satisfying[0], nil
}
