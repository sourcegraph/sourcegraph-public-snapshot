package template

var testRepo1 = &TemplatingRepository{
	ID:            "src-cli",
	Name:          "github.com/sourcegraph/src-cli",
	DefaultBranch: TemplatingBranch{Name: "main", TargetOID: "d34db33f"},
	FileMatches: map[string]bool{
		"README.md": true,
		"main.go":   true,
	},
}
