package squirrel

import (
	_ "embed"
	"fmt"

	"github.com/grafana/regexp"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/typescript/tsx"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
)

//go:embed language-file-extensions.json
var languageFileExtensionsJson string

// Mapping from langauge name to file extensions.
var langToExts = func() map[string][]string {
	var m map[string][]string
	err := jsonc.Unmarshal(languageFileExtensionsJson, &m)
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

// LangSpec contains info about a language.
type LangSpec struct {
	name         string
	language     *sitter.Language
	commentStyle CommentStyle
	// localsQuery is a tree-sitter localsQuery that finds scopes and defs.
	localsQuery          string
	topLevelSymbolsQuery string
}

// CommentStyle contains info about comments in a language.
type CommentStyle struct {
	ignoreRegex   *regexp.Regexp
	stripRegex    *regexp.Regexp
	skipNodeTypes []string
	nodeTypes     []string
	codeFenceName string
}

var javaStyleStripRegex = regexp.MustCompile(`^//|^\s*\*/?|^/\*\*|\*/$`)
var javaStyleIgnoreRegex = regexp.MustCompile(`^\s*(/\*\*|\*/)\s*$`)

// Mapping from language name to language specification.
var langToLangSpec = map[string]LangSpec{
	"java": {
		name:     "java",
		language: java.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"line_comment", "block_comment"},
			stripRegex:    javaStyleStripRegex,
			ignoreRegex:   javaStyleIgnoreRegex,
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
		topLevelSymbolsQuery: `
(program (class_declaration     name: (identifier) @symbol))
(program (enum_declaration      name: (identifier) @symbol))
(program (interface_declaration name: (identifier) @symbol))
`,
	},
	"go": {
		name:     "go",
		language: golang.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"comment"},
			stripRegex:    regexp.MustCompile(`^//`),
			codeFenceName: "go",
		},
		localsQuery: `
(block)                   @scope ; { ... }
(function_declaration)    @scope ; func f() { ... }
(method_declaration)      @scope ; func (r R) f() { ... }
(func_literal)            @scope ; func() { ... }
(if_statement)            @scope ; if true { ... }
(for_statement)           @scope ; for x := range xs { ... }
(expression_case)         @scope ; case "foo": ...
(communication_case)      @scope ; case x := <-ch: ...

(var_spec              name: (identifier) @definition)                   ; var x int = ...
(const_spec            name: (identifier) @definition)                   ; const x int = ...
(parameter_declaration name: (identifier) @definition)                   ; func(x int) { ... }
(short_var_declaration left: (expression_list (identifier) @definition)) ; x, y := ...
(range_clause          left: (expression_list (identifier) @definition)) ; for i := range ... { ... }
(receive_statement     left: (expression_list (identifier) @definition)) ; case x := <-ch: ...
`,
	},
	"csharp": {
		name:     "csharp",
		language: csharp.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"comment"},
			stripRegex:    javaStyleStripRegex,
			ignoreRegex:   javaStyleIgnoreRegex,
			codeFenceName: "csharp",
		},
		localsQuery: `
(block)                   @scope ; { ... }
(method_declaration)      @scope ; void f() { ... }
(for_statement)           @scope ; for (...) ...
(using_statement)         @scope ; using (...) ...
(lambda_expression)       @scope ; (x, y) => ...
(for_each_statement)      @scope ; foreach (int x in xs) ...
(catch_clause)            @scope ; try { ... } catch (Exception e) { ... }
(constructor_declaration) @scope ; public Foo() { ... }

(parameter           name: (identifier) @definition) ; void f(x int) { ... }
(variable_declarator (identifier) @definition)       ; int x = ...
(for_each_statement  left: (identifier) @definition) ; foreach (int x in xs) ...
(catch_declaration   name: (identifier) @definition) ; catch (Exception e) { ... }
`,
	},
	"python": {
		name:     "python",
		language: python.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"comment"},
			stripRegex:    regexp.MustCompile(`^#`),
			codeFenceName: "python",
		},
		localsQuery: pythonLocalsQuery,
	},
	"javascript": {
		name:     "javascript",
		language: javascript.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"comment"},
			stripRegex:    javaStyleStripRegex,
			ignoreRegex:   javaStyleIgnoreRegex,
			codeFenceName: "javascript",
		},
		localsQuery: `
(class_declaration)              @scope ; class C { ... }
(method_definition)              @scope ; class ... { f() { ... } }
(statement_block)                @scope ; { ... }
(for_statement)                  @scope ; for (let i = 0; ...) ...
(for_in_statement)               @scope ; for (const x of xs) ...
(catch_clause)                   @scope ; catch (e) ...
(function)                       @scope ; function(x) { ... }
(function_declaration)           @scope ; function f(x) { ... }
(generator_function)             @scope ; function*(x) { ... }
(generator_function_declaration) @scope ; function *f(x) { ... }
(arrow_function)                 @scope ; x => ...

(variable_declarator name: (identifier) @definition)                                   ; const x = ...
(function_declaration name: (identifier) @definition)                                  ; function f() { ... }
(generator_function_declaration name: (identifier) @definition)                        ; function *f() { ... }
(formal_parameters (identifier) @definition)                                           ; function(x) { ... }
(formal_parameters (rest_pattern (identifier) @definition))                            ; function(...x) { ... }
(formal_parameters (assignment_pattern left: (identifier) @definition))                ; function(x = 5) { ... }
(formal_parameters (assignment_pattern left: (rest_pattern (identifier) @definition))) ; function(...x = []) { ... }
(arrow_function parameter: (identifier) @definition)                                   ; x => ...
(for_in_statement left: (identifier) @definition)                                      ; for (const x of xs) ...
(catch_clause parameter: (identifier) @definition)                                     ; catch (e) ...
`,
	},
	"typescript": {
		name:     "typescript",
		language: tsx.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"comment"},
			stripRegex:    javaStyleStripRegex,
			ignoreRegex:   javaStyleIgnoreRegex,
			codeFenceName: "typescript",
		},
		localsQuery: `
(class_declaration)              @scope ; class C { ... }
(method_definition)              @scope ; class ... { f() { ... } }
(statement_block)                @scope ; { ... }
(for_statement)                  @scope ; for (let i = 0; ...) ...
(for_in_statement)               @scope ; for (const x of xs) ...
(catch_clause)                   @scope ; catch (e) ...
(function)                       @scope ; function(x) { ... }
(function_declaration)           @scope ; function f(x) { ... }
(generator_function)             @scope ; function*(x) { ... }
(generator_function_declaration) @scope ; function *f(x) { ... }
(arrow_function)                 @scope ; x => ...

(variable_declarator name: (identifier) @definition)            ; const x = ...
(function_declaration name: (identifier) @definition)           ; function f() { ... }
(generator_function_declaration name: (identifier) @definition) ; function *f() { ... }
(required_parameter (identifier) @definition)                   ; function(x) { ... }
(required_parameter (rest_pattern (identifier) @definition))    ; function(...x) { ... }
(optional_parameter (identifier) @definition)                   ; function(x?) { ... }
(optional_parameter (rest_pattern (identifier) @definition))    ; function(...x?) { ... }
(arrow_function parameter: (identifier) @definition)            ; x => ...
(for_in_statement left: (identifier) @definition)               ; for (const x of xs) ...
(catch_clause parameter: (identifier) @definition)              ; catch (e) ...
`,
	},
	"cpp": {
		name:     "cpp",
		language: cpp.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"comment"},
			stripRegex:    javaStyleStripRegex,
			ignoreRegex:   javaStyleIgnoreRegex,
			codeFenceName: "cpp",
		},
		localsQuery: `
(compound_statement)  @scope ; { ... }
(for_statement)       @scope ; for (int i = 0; ...) ...
(for_range_loop)      @scope ; for (int x : xs) ...
(catch_clause)        @scope ; catch (e) ...
(lambda_expression)   @scope ; [](auto x) { ... }
(function_definition) @scope ; void f() { ... }

(declaration                    declarator: (identifier) @definition)                        ; int x;
(init_declarator                declarator: (identifier) @definition)                        ; int x = 5;
(parameter_declaration          declarator: (identifier) @definition)                        ; [](auto x) { ... }
(parameter_declaration          declarator: (reference_declarator (identifier) @definition)) ; [](int& x) { ... }
(parameter_declaration          declarator: (pointer_declarator   (identifier) @definition)) ; [](int* x) { ... }
(optional_parameter_declaration declarator: (identifier) @definition)                        ; [](auto x = 5) { ... }
(for_range_loop declarator: (identifier) @definition)									     ; for (int x : xs) ...
`,
	},
	"ruby": {
		name:     "ruby",
		language: ruby.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"comment"},
			stripRegex:    regexp.MustCompile(`^#`),
			codeFenceName: "ruby",
		},
		localsQuery: `
(method)   @scope ; def f() ...
(block)    @scope ; { ... }
(do_block) @scope ; do ... end

(exception_variable   (identifier) @definition)          ; rescue ArgumentError => e
(method_parameters    (identifier) @definition)          ; def f(x) ...
(optional_parameter   name: (identifier) @definition)    ; def f(x = 5) ...
(splat_parameter      name: (identifier) @definition)    ; def f(*x) ...
(hash_splat_parameter name: (identifier) @definition)    ; def f(**x) ...
(block_parameters     (identifier) @definition)          ; |x| ...
(assignment           left: (identifier) @definition)    ; x = ...
(left_assignment_list (identifier) @definition)          ; x, y = ...
(for                  pattern: (identifier) @definition) ; for i in 1..5 ...
`,
	},
	"starlark": {
		name:     "starlark",
		language: python.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"comment"},
			stripRegex:    regexp.MustCompile(`^#`),
			codeFenceName: "starlark",
		},
		localsQuery: pythonLocalsQuery,
	},
}

var pythonLocalsQuery = `
(function_definition)     @scope ; def f(): ...
(lambda)                  @scope ; lambda ...: ...
(generator_expression)    @scope ; (x for x in xs)
(list_comprehension)      @scope ; [x for x in xs]

(parameters                    (identifier) @definition)                                   ; def f(x): ...
(typed_parameter               (identifier) @definition)                                   ; def f(x: bool): ...
(default_parameter       name: (identifier) @definition)                                   ; def f(x = False): ...
(typed_default_parameter name: (identifier) @definition)                                   ; def f(x: bool = False): ...
(except_clause
  (as_pattern (identifier) alias: (as_pattern_target (identifier) @definition)))           ; except Exception as e: ...
(expression_statement          (assignment left: (identifier) @definition))                ; x = ...
(expression_statement          (assignment left: (pattern_list (identifier) @definition))) ; x, y = ...
(for_statement           left: (identifier) @definition)                                   ; for x in ...: ...
(for_statement           left: (pattern_list (identifier) @definition))                    ; for x, y in ...: ...
(for_in_clause           left: (identifier) @definition)                                   ; (... for x in xs)
(for_in_clause           left: (pattern_list (identifier) @definition))                    ; (... for x, y in xs)
`
