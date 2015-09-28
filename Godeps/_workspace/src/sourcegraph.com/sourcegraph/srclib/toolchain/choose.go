package toolchain

import (
	"fmt"
	"os"

	"strconv"

	"sourcegraph.com/sourcegraph/srclib"
)

var (
	// NoToolchains operates srclib in no-toolchain mode, where it
	// does not try to look for system toolchains in your SRCLIBPATH.
	NoToolchains, _ = strconv.ParseBool(os.Getenv("SRCLIB_NO_TOOLCHAINS"))

	// noneToolchain is returned by ChooseTool when NoToolchains is
	// true.
	noneToolchain = &srclib.ToolRef{Toolchain: "NONE", Subcmd: "NONE"}
)

// ChooseTool determines which toolchain and tool to use to run op (graph,
// depresolve, etc.) on a source unit of the given type. If no tools fit the
// criteria, an error is returned.
//
// The selection algorithm is currently very simplistic: if exactly one tool is
// found that can perform op on the source unit type, it is returned. If zero or
// more than 1 are found, then an error is returned. TODO(sqs): extend this to
// choose the "best" tool when multiple tools would suffice.
func ChooseTool(op, unitType string) (*srclib.ToolRef, error) {
	if NoToolchains {
		return noneToolchain, nil
	}
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

	if n := len(satisfying); n == 0 {
		return nil, fmt.Errorf("no tool satisfies op %q for source unit type %q", op, unitType)
	} else if n > 1 {
		return nil, fmt.Errorf("%d tools satisfy op %q for source unit type %q (refusing to choose between multiple possibilities)", n, op, unitType)
	}
	return satisfying[0], nil
}
