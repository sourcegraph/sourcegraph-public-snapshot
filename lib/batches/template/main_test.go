package template

var testRepo1 = &Repository{
	Name: "github.com/sourcegraph/src-cli",
	FileMatches: map[string]bool{
		"README.md": true,
		"main.go":   true,
	},
}
