package gosyntect

// This file

// Make sure all names are lowercase here, since they are normalized
var enryLanguageMappings = map[string]string{
	"c#": "c_sharp",
}

var treesitterSupportedFiletypes = map[string]struct{}{
	"c":          {},
	"c++":        {},
	"c_sharp":    {},
	"cpp":        {},
	"go":         {},
	"java":       {},
	"javascript": {},
	"jsonnet":    {},
	"jsx":        {},
	"nickel":     {},
	"perl":       {},
	"python":     {},
	"ruby":       {},
	"rust":       {},
	"scala":      {},
	"tsx":        {},
	"typescript": {},
	"xlsg":       {},
	"zig":        {},
}
