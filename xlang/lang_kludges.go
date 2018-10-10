package xlang

import (
	"go/ast"

	"github.com/sourcegraph/go-lsp/lspext"
	xlangext "github.com/sourcegraph/sourcegraph/xlang/lspext"
)

// SymbolPackageDescriptor extracts the package descriptor from the
// symbol descriptor for supported languages. Returns true for the
// second return value if and only if the language is supported.
func SymbolPackageDescriptor(sym lspext.SymbolDescriptor, lang string) (xlangext.PackageDescriptor, bool) {
	subSelector, ok := subSelectors[lang]
	if !ok {
		return nil, false
	}
	return subSelector(sym), true
}

// DependencySymbolQuery returns a symbol descriptor that matches all symbols defined in
// the dependency described by depData.
func DependencySymbolQuery(depData map[string]interface{}, lang string) (lspext.SymbolDescriptor, bool) {
	f, ok := dependencySymbolQueries[lang]
	if !ok {
		return nil, false
	}
	return f(depData), true
}

// SymbolsInPackage returns a symbol descriptor that matches all symbols defined in the
// package.
func SymbolsInPackage(pkg xlangext.PackageDescriptor, lang string) (lspext.SymbolDescriptor, bool) {
	f, ok := symbolsInPackage[lang]
	if !ok {
		return nil, false
	}
	return f(pkg), true
}

// PackageIdentifier extracts the part of the PackageDescriptor that
// should be used to quasi-uniquely identify a package. Typically, it
// leaves out things like package version.
func PackageIdentifier(pkgDescriptor xlangext.PackageDescriptor, lang string) (xlangext.PackageDescriptor, bool) {
	pkgIDFn, ok := packageIdentifiers[lang]
	if !ok {
		return nil, false
	}
	return pkgIDFn(pkgDescriptor), true
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

// HasXDefinitionAndXPackages is the hardcoded list of languages that provide
// textDocument/xdefinition and workspace/xpackages. We cannot rely on the
// value returned from the LSP proxy, because that does not pass through the
// value of the initialize result.
var HasXDefinitionAndXPackages = map[string]struct{}{
	"go":         struct{}{},
	"javascript": struct{}{},
	"typescript": struct{}{},
	"java":       struct{}{},
	"python":     struct{}{},
}

// HasCrossRepoHover records the languages for which we support cross-repo
// hovers. In theory, this should be identical to
// HasXDefinitionAndXPackages, but cross-repo hover has the additional
// requirement that locations returned by workspace/symbol must
// correspond to the location of the *ident*, rather than the entire
// body AST node. This is not the case for TypeScript.
var HasCrossRepoHover = map[string]struct{}{"java": struct{}{}}

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
var subSelectors = map[string]func(lspext.SymbolDescriptor) xlangext.PackageDescriptor{
	"go": func(symbol lspext.SymbolDescriptor) xlangext.PackageDescriptor {
		return xlangext.PackageDescriptor{
			"package": symbol["package"],
		}
	},
	"php": func(symbol lspext.SymbolDescriptor) xlangext.PackageDescriptor {
		if _, ok := symbol["package"]; !ok {
			// package can be missing if the symbol did not belong to a package, e.g. a project without
			// a composer.json file. In this case, there are no external references to this symbol.
			return nil
		}
		return packageIdentifiers["php"](symbol["package"].(map[string]interface{}))
	},
	"typescript": func(symbol lspext.SymbolDescriptor) xlangext.PackageDescriptor {
		if _, ok := symbol["package"]; !ok {
			return nil
		}
		return packageIdentifiers["typescript"](symbol["package"].(map[string]interface{}))
	},
	"java": func(symbol lspext.SymbolDescriptor) xlangext.PackageDescriptor {
		if _, ok := symbol["package"].(map[string]interface{}); !ok {
			return nil
		}
		return packageIdentifiers["java"](symbol["package"].(map[string]interface{}))
	},
	"python": func(symbol lspext.SymbolDescriptor) xlangext.PackageDescriptor {
		if _, ok := symbol["package"].(map[string]interface{}); !ok {
			return nil
		}
		return packageIdentifiers["python"](symbol["package"].(map[string]interface{}))
	},
}

var dependencySymbolQueries = map[string]func(map[string]interface{}) lspext.SymbolDescriptor{
	"go": func(depData map[string]interface{}) lspext.SymbolDescriptor {
		return lspext.SymbolDescriptor{
			"package": depData["package"],
		}
	},
	// TODO(sqs): Support these for TypeScript, JavaScript, PHP, Java, and Python.
}

var symbolsInPackage = map[string]func(xlangext.PackageDescriptor) lspext.SymbolDescriptor{
	"go": func(pkg xlangext.PackageDescriptor) lspext.SymbolDescriptor {
		return lspext.SymbolDescriptor{
			"package": pkg["package"],
		}
	},
	// TODO(sqs): Support these for TypeScript, JavaScript, PHP, Java, and Python.
}

var packageIdentifiers = map[string]func(xlangext.PackageDescriptor) xlangext.PackageDescriptor{
	"go": func(pkg xlangext.PackageDescriptor) xlangext.PackageDescriptor {
		return xlangext.PackageDescriptor{
			"package": pkg["package"],
		}
	},
	"php": func(pkg xlangext.PackageDescriptor) xlangext.PackageDescriptor {
		return xlangext.PackageDescriptor{
			"name": pkg["name"],
		}
	},
	"typescript": func(pkg xlangext.PackageDescriptor) xlangext.PackageDescriptor {
		return xlangext.PackageDescriptor{
			"name": pkg["name"],
		}
	},
	"java": func(pkg xlangext.PackageDescriptor) xlangext.PackageDescriptor {
		return xlangext.PackageDescriptor{
			"id":   pkg["id"],
			"type": pkg["type"],
		}
	},
	"python": func(pkg xlangext.PackageDescriptor) xlangext.PackageDescriptor {
		return xlangext.PackageDescriptor{
			"name": pkg["name"],
		}
	},
}
