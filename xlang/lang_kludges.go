package xlang

import (
	"go/ast"

	"github.com/sourcegraph/go-langserver/pkg/lspext"
)

// SymbolPackageDescriptor extracts the package descriptor from the
// symbol descriptor for supported languages. Returns true for the
// second return value if and only if the language is supported.
func SymbolPackageDescriptor(sym lspext.SymbolDescriptor, lang string) (map[string]interface{}, bool) {
	subSelector, ok := subSelectors[lang]
	if !ok {
		return nil, false
	}
	return subSelector(sym), true
}

// SymbolRepoURL returns the repository URL extracted from the
// package metadata at the JSON path
// `symDescriptor.package.repoURL`. If that does not exist, it returns
// the empty string.
func SymbolRepoURL(symDescriptor lspext.SymbolDescriptor) string {
	pkgData := symDescriptor["package"]
	if pkgData, ok := pkgData.(map[string]interface{}); ok {
		repoURL := pkgData["repoURL"]
		if repoURL, ok := repoURL.(string); ok {
			return repoURL
		}
	}
	return ""
}

// IsSymbolReferenceable tells if the SymbolDescriptor is referenceable
// according to the language semantics defined by the mode.
func IsSymbolReferenceable(mode string, symbolDescriptor lspext.SymbolDescriptor) bool {
	switch mode {
	case "go":
		if name, ok := symbolDescriptor["name"]; ok {
			if !ast.IsExported(name.(string)) {
				return false
			}
		}
		if recv, ok := symbolDescriptor["recv"]; ok && recv.(string) != "" {
			if !ast.IsExported(recv.(string)) {
				return false
			}
		}
		return true
	default:
		return true
	}
}

// subSelectors is a map of language-specific data selectors. The
// input data is from the language server's workspace/xdefinition
// request, and the output data should be something that can be
// matched (using the jsonb containment operator) against the
// `attributes` field of `DependenceReference` (output of
// workspace/xdependencies).
var subSelectors = map[string]func(lspext.SymbolDescriptor) map[string]interface{}{
	"go": func(symbol lspext.SymbolDescriptor) map[string]interface{} {
		return map[string]interface{}{
			"package": symbol["package"],
		}
	},
	"php": func(symbol lspext.SymbolDescriptor) map[string]interface{} {
		if _, ok := symbol["package"]; !ok {
			// package can be missing if the symbol did not belong to a package, e.g. a project without
			// a composer.json file. In this case, there are no external references to this symbol.
			return nil
		}
		return map[string]interface{}{
			"name": symbol["package"].(map[string]interface{})["name"],
		}
	},
	"typescript": func(symbol lspext.SymbolDescriptor) map[string]interface{} {
		if _, ok := symbol["package"]; !ok {
			return nil
		}
		return map[string]interface{}{
			"name": symbol["package"].(map[string]interface{})["name"],
		}
	},
}
