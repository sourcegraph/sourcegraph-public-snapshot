package squirrel

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/grafana/regexp"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
)

//go:embed language-file-extensions.json
var languageFileExtensionsJson string

// Mapping from langauge name to file extensions.
var langToExts = func() map[string][]string {
	var m map[string][]string
	err := json.Unmarshal([]byte(languageFileExtensionsJson), &m)
	if err != nil {
		panic(err)
	}
	return m
}()

// Mapping from file extension to language name.
var extToLang = func() map[string]string {
	m := map[string]string{}
	for lang, exts := range langToExts {
		for _, ext := range exts {
			if _, ok := m[ext]; ok {
				panic(fmt.Sprintf("duplicate file extension %s", ext))
			}
			m[ext] = lang
		}
	}
	return m
}()

// Info about a language.
type LangSpec struct {
	language     *sitter.Language
	commentStyle CommentStyle
	// localsQuery is a tree-sitter localsQuery that finds scopes and defs.
	localsQuery string
}

// Info about comments in a language.
type CommentStyle struct {
	ignoreRegex   *regexp.Regexp
	stripRegex    *regexp.Regexp
	skipNodeTypes []string
	nodeTypes     []string
	codeFenceName string
}

// Mapping from language name to language specification. Queries were copied from
// nvim-treesitter@5b6f6ae30c1cf8fceefe08a9bcf799870558a878 as a starting point.
var langToLangSpec = map[string]LangSpec{
	"java": {
		language: java.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"comment"},
			stripRegex:    regexp.MustCompile(`(^//|^\s*\*|^/\*\*|\*/$)`),
			ignoreRegex:   regexp.MustCompile(`^\s*(/\*\*|\*/)\s*$`),
			codeFenceName: "java",
			skipNodeTypes: []string{"modifiers"},
		},
		localsQuery: `
(block)                   @scope ; { ... }
(lambda_expression)       @scope ; (x, y) -> ...
(catch_clause)            @scope ; try { ... } catch (Exception e) { ... }
(enhanced_for_statement)  @scope ; for (var item : items) ...
(for_statement)           @scope ; for (var i = 0; i < 5; i++) ...
(constructor_declaration) @scope ; public Foo() { ... }
(method_declaration)      @scope ; public void f() { ... }

(local_variable_declaration declarator: (variable_declarator name: (identifier) @definition)) ; int x = ...;
(formal_parameter           name:       (identifier) @definition)                             ; public void f(int x) { ... }
(catch_formal_parameter     name:       (identifier) @definition)                             ; try { ... } catch (Exception e) { ... }
(lambda_expression          parameters: (inferred_parameters (identifier) @definition))       ; (x, y) -> ...
(lambda_expression          parameters: (identifier) @definition)                             ; x -> ...
(enhanced_for_statement     name:       (identifier) @definition)                             ; for (var item : items) ...
`,
	},
}
