package codeownership

import (
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func ForResult(r result.FileMatch) []string {
	if r.Repo.Name != "github.com/philipp-spiess/codeowners-test" {
		return []string{}
	}

	switch r.File.Path {
	case "backend/backend-code.go":
		return []string{"@nicolasdular"}
	case "frontend/frontend-code.ts":
		return []string{"@philipp-spiess"}
	default:
		return []string{"@philipp-spiess", "@nicolasdular"}
	}
}

func ForOwner(owner string) []string {
	switch owner {
	case "@philipp-spiess":
		return []string{"frontend/frontend-code.ts", "unowned/legacy-code.php", "CODEOWNERS", "README.md"}
	case "@nicolasdular":
		return []string{"backend/backend-code.go", "unowned/legacy-code.php", "CODEOWNERS", "README.md"}
	default:
		panic("unexpected owner")
	}
}
