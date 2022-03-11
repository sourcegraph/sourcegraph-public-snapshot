package squirrel

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/grafana/regexp"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/lua"
	"github.com/smacker/go-tree-sitter/ocaml"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"
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
	query        string
	exportsQuery string
}

// Info about comments in a language.
type CommentStyle struct {
	placedBelow   bool
	ignoreRegex   *regexp.Regexp
	stripRegex    *regexp.Regexp
	skipNodeTypes []string
	nodeTypes     []string
	codeFenceName string
}

// Mapping from language name to language specification. Queries were copied from
// nvim-treesitter@5b6f6ae30c1cf8fceefe08a9bcf799870558a878 as a starting point.
var langToLangSpec = map[string]LangSpec{
	"cpp": {
		language:     cpp.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
		query: `
; inherits: c

;; Parameters
(variadic_parameter_declaration
  declarator: (variadic_declarator
                (identifier) @definition.parameter))
(optional_parameter_declaration
  declarator: (identifier) @definition.parameter)
;; Class / struct definitions
(class_specifier) @scope

(reference_declarator
  (identifier) @definition.var)

(variadic_declarator
  (identifier) @definition.var)

(struct_specifier
  name: (qualified_identifier
          name: (type_identifier) @definition.type))

(class_specifier
  name: (type_identifier) @definition.type)

(concept_definition
  name: (identifier) @definition.type)

(class_specifier
  name: (qualified_identifier
          name: (type_identifier) @definition.type))

(alias_declaration
  name: (type_identifier) @definition.type)

;template <typename T>
(type_parameter_declaration
  (type_identifier) @definition.type)
(template_declaration) @scope

;; Namespaces
(namespace_definition
  name: (identifier) @definition.namespace
  body: (_) @scope)

((namespace_identifier) @reference
                        (set! reference.kind "namespace"))

;; Function definitions
(template_function
  name: (identifier) @definition.function) @scope

(template_method
  name: (field_identifier) @definition.method) @scope

(function_declarator
  declarator: (qualified_identifier
                name: (identifier) @definition.function)) @scope

(field_declaration
        declarator: (function_declarator
                       (field_identifier) @definition.method))

(lambda_expression) @scope

;; Control structures
(try_statement
  body: (_) @scope)

(catch_clause) @scope

(requires_expression) @scope
`,
	},
	"csharp": {
		language:     csharp.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
		query: `
;; Definitions
(variable_declarator
  . (identifier) @definition.var)

(variable_declarator
  (tuple_pattern
    (identifier) @definition.var))

(declaration_expression
  name: (identifier) @definition.var)

(for_each_statement
  left: (identifier) @definition.var)

(for_each_statement
  left: (tuple_pattern
    (identifier) @definition.var))

(parameter
  (identifier) @definition.parameter)

(method_declaration
  name: (identifier) @definition.method)

(local_function_statement
  name: (identifier) @definition.method)

(property_declaration
  name: (identifier) @definition)

(type_parameter
  (identifier) @definition.type)

(class_declaration
  name: (identifier) @definition)

;; References
(identifier) @reference

;; Scope
(block) @scope
`,
	},
	"go": {
		language: golang.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"comment"},
			stripRegex:    regexp.MustCompile(`^//`),
			codeFenceName: "go",
		}, // TODO
		query: `
(
    (function_declaration
        name: (identifier) @definition.function) ;@function
)

(
    (method_declaration
        name: (field_identifier) @definition.method); @method
)

(short_var_declaration
  left: (expression_list
          (identifier) @definition.var))

(var_spec
  name: (identifier) @definition.var)

(parameter_declaration (identifier) @definition.var)
(variadic_parameter_declaration (identifier) @definition.var)

(for_statement
 (range_clause
   left: (expression_list
           (identifier) @definition.var)))

(const_declaration
 (const_spec
  name: (identifier) @definition.var))

(type_declaration
  (type_spec
    name: (type_identifier) @definition.type))

;; reference
(identifier) @reference
(type_identifier) @reference
(field_identifier) @reference
((package_identifier) @reference
  (set! reference.kind "namespace"))

(package_clause
   (package_identifier) @definition.namespace)

(import_spec_list
  (import_spec
    name: (package_identifier) @definition.namespace))

;; Call references
((call_expression
   function: (identifier) @reference)
 (set! reference.kind "call" ))

((call_expression
    function: (selector_expression
                field: (field_identifier) @reference))
 (set! reference.kind "call" ))


((call_expression
    function: (parenthesized_expression
                (identifier) @reference))
 (set! reference.kind "call" ))

((call_expression
   function: (parenthesized_expression
               (selector_expression
                 field: (field_identifier) @reference)))
 (set! reference.kind "call" ))

;; Scopes

(func_literal) @scope
(source_file) @scope
(function_declaration) @scope
(if_statement) @scope
(block) @scope
(expression_switch_statement) @scope
(for_statement) @scope
(method_declaration) @scope
`,
		exportsQuery: `
(source_file (var_declaration (var_spec name: (identifier) @exported)))
(source_file (const_declaration (const_spec name: (identifier) @exported)))
(source_file (function_declaration name: (identifier) @exported))
(source_file (method_declaration name: (field_identifier) @exported))
(source_file (type_declaration (type_spec name: (type_identifier) @exported)))
(source_file (type_declaration (type_alias name: (type_identifier) @exported)))
`,
	},
	"java": {
		language: java.GetLanguage(),
		commentStyle: CommentStyle{
			nodeTypes:     []string{"line_comment", "block_comment"},
			stripRegex:    regexp.MustCompile(`(^//|^\s*\*|^/\*\*|\*/$)`),
			ignoreRegex:   regexp.MustCompile(`^\s*(/\*\*|\*/)\s*$`),
			codeFenceName: "java",
			skipNodeTypes: []string{"modifiers"},
		}, // TODO
		query: `
; SCOPES
; declarations
(program) @scope
(class_declaration
  body: (_) @scope)
(record_declaration
  body: (_) @scope)
(enum_declaration
  body: (_) @scope)
(lambda_expression) @scope
(enhanced_for_statement) @scope

; block
(block) @scope

; if/else
(if_statement) @scope ; if+else
(if_statement
  consequence: (_) @scope) ; if body in case there are no braces
(if_statement
  alternative: (_) @scope) ; else body in case there are no braces

; try/catch
(try_statement) @scope ; covers try+catch, individual try and catch are covered by (block)
(catch_clause) @scope ; needed because "Exception" variable

; loops
(for_statement) @scope ; whole for_statement because loop iterator variable
(for_statement         ; "for" body in case there are no braces
  body: (_) @scope)
(do_statement
  body: (_) @scope)
(while_statement
  body: (_) @scope)

; Functions

(constructor_declaration) @scope
(method_declaration) @scope


; DEFINITIONS
(package_declaration
  (identifier) @definition.namespace)
(class_declaration
  name: (identifier) @definition.type)
(record_declaration
  name: (identifier) @definition.type)
(enum_declaration
  name: (identifier) @definition.enum)
(method_declaration
  name: (identifier) @definition.method)

(local_variable_declaration
  declarator: (variable_declarator
                name: (identifier) @definition.var))
(formal_parameter
  name: (identifier) @definition.var)
(catch_formal_parameter
  name: (identifier) @definition.var)
(inferred_parameters (identifier) @definition.var) ; (x,y) -> ...
(lambda_expression
    parameters: (identifier) @definition.var) ; x -> ...
(enhanced_for_statement ; for (var item : items) {
  name: (identifier) @definition.var)

((scoped_identifier
  (identifier) @definition.import)
 (#has-ancestor? @definition.import import_declaration))

(field_declaration
  declarator: (variable_declarator
                name: (identifier) @definition.field))

; REFERENCES
(identifier) @reference
(type_identifier) @reference
`,
	},
	"javascript": {
		language:     javascript.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
		query: `
; inherits: ecma,jsx

(formal_parameters
  (identifier) @definition.parameter)

; function(arg = []) {
(formal_parameters
  (assignment_pattern
    left: (identifier) @definition.parameter))

; x => x
(arrow_function
  parameter: (identifier) @definition.parameter)

;; ({ a }) => null
(formal_parameters
  (object_pattern
    (shorthand_property_identifier_pattern) @definition.parameter))

;; ({ a: b }) => null
(formal_parameters
  (object_pattern
    (pair_pattern
      value: (identifier) @definition.parameter)))

;; ([ a ]) => null
(formal_parameters
  (array_pattern
    (identifier) @definition.parameter))

(formal_parameters
  (rest_pattern
    (identifier) @definition.parameter))
`,
	},
	"lua": {
		language:     lua.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
		query: `
; Scopes

[
  (chunk)
  (do_statement)
  (while_statement)
  (repeat_statement)
  (if_statement)
  (for_statement)
  (function_declaration)
  (function_definition)
] @scope

; Definitions

(assignment_statement
  (variable_list
    (identifier) @definition.var))

(assignment_statement
  (variable_list
    (dot_index_expression . (_) @definition.associated (identifier) @definition.var)))

(function_declaration
  name: (identifier) @definition.function)
  (#set! definition.function.scope "parent")

(function_declaration
  name: (dot_index_expression
    . (_) @definition.associated (identifier) @definition.function))
  (#set! definition.method.scope "parent")

(function_declaration
  name: (method_index_expression
    . (_) @definition.associated (identifier) @definition.method))
  (#set! definition.method.scope "parent")

(for_generic_clause
  (variable_list
    (identifier) @definition.var))

(for_numeric_clause
  name: (identifier) @definition.var)

(parameters (identifier) @definition.parameter)

; References

[
  (identifier)
] @reference
`,
	},
	"ocaml": {
		language:     ocaml.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
		query: `
; Scopes
;-------

[
  (compilation_unit)
  (structure)
  (signature)
  (module_binding)
  (functor)
  (let_binding)
  (match_case)
  (class_binding)
  (class_function)
  (method_definition)
  (let_expression)
  (fun_expression)
  (for_expression)
  (let_class_expression)
  (object_expression)
  (attribute_payload)
] @scope

; Definitions
;------------

(value_pattern) @definition.var

(let_binding
  pattern: (value_name) @definition.var
  (set! definition.var.scope "parent"))

(let_binding
  pattern: (tuple_pattern (value_name) @definition.var)
  (set! definition.var.scope "parent"))

(let_binding
  pattern: (record_pattern (field_pattern (value_name) @definition.var))
  (set! definition.var.scope "parent"))

(external (value_name) @definition.var)

(type_binding (type_constructor) @definition.type)

(abstract_type (type_constructor) @definition.type)

(method_definition (method_name) @definition.method)

(module_binding
  (module_name) @definition.namespace
  (set! definition.namespace.scope "parent"))

(module_parameter (module_name) @definition.namespace)

(module_type_definition (module_type_name) @definition.type)

; References
;------------

(value_path .
  (value_name) @reference
  (set! reference.kind "var"))

(type_constructor_path .
  (type_constructor) @reference
  (set! reference.kind "type"))

(method_invocation
  (method_name) @reference
  (set! reference.kind "method"))

(module_path .
  (module_name) @reference
  (set! reference.kind "type"))

(module_type_path .
  (module_type_name) @reference
  (set! reference.kind "type"))
`,
	},
	"php": {
		language:     php.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
		query: `
; Scopes
;-------

((class_declaration
  name: (name) @definition.type) @scope
    (set! definition.type.scope "parent"))

((method_declaration
  name: (name) @definition.method) @scope
    (set! definition.method.scope "parent"))

((function_definition
  name: (name) @definition.function) @scope
    (set! definition.function.scope "parent"))

(anonymous_function_creation_expression
  (anonymous_function_use_clause
    (variable_name
      (name) @definition.var))) @scope

; Definitions
;------------

(simple_parameter
  (variable_name
    (name) @definition.var))

(foreach_statement
  (pair
    (variable_name
      (name) @definition.var)))

(foreach_statement
  (variable_name
    (name) @reference
      (set! reference.kind "var"))
  (variable_name
    (name) @definition.var))

(property_declaration
  (property_element
    (variable_name
      (name) @definition.field)))

(namespace_use_clause
  (qualified_name
    (name) @definition.type))

; References
;------------

(named_type
  (name) @reference
    (set! reference.kind "type"))

(named_type
  (qualified_name) @reference
    (set! reference.kind "type"))

(variable_name
  (name) @reference
    (set! reference.kind "var"))

(member_access_expression
  name: (name) @reference
    (set! reference.kind "field"))

(member_call_expression
  name: (name) @reference
    (set! reference.kind "method"))

(function_call_expression
  function: (qualified_name
    (name) @reference
      (set! reference.kind "function")))

(object_creation_expression
  (qualified_name
    (name) @reference
      (set! reference.kind "type")))

(scoped_call_expression
  scope: (qualified_name
    (name) @reference
      (set! reference.kind "type"))
  name: (name) @reference
    (set! reference.kind "method"))
`,
	},
	"python": {
		language:     python.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
		query: `
;;; Program structure
(module) @scope

(class_definition
  body: (block
          (expression_statement
            (assignment
              left: (identifier) @definition.field)))) @scope
(class_definition
  body: (block
          (expression_statement
            (assignment
              left: (_
                     (identifier) @definition.field))))) @scope

; Imports
(aliased_import
  alias: (identifier) @definition.import)
(import_statement
  name: (dotted_name ((identifier) @definition.import)))
(import_from_statement
  name: (dotted_name ((identifier) @definition.import)))

; Function with parameters, defines parameters
(parameters
  (identifier) @definition.parameter)

(default_parameter
  (identifier) @definition.parameter)

(typed_parameter
  (identifier) @definition.parameter)

(typed_default_parameter
  (identifier) @definition.parameter)

; *args parameter
(parameters
  (list_splat_pattern
    (identifier) @definition.parameter))

; **kwargs parameter
(parameters
  (dictionary_splat_pattern
    (identifier) @definition.parameter))

; Function defines function and scope
((function_definition
  name: (identifier) @definition.function) @scope
 (#set! definition.function.scope "parent"))


((class_definition
  name: (identifier) @definition.type) @scope
 (#set! definition.type.scope "parent"))

(class_definition
  body: (block
          (function_definition
            name: (identifier) @definition.method)))

;;; Loops
; not a scope!
(for_statement
  left: (pattern_list
          (identifier) @definition.var))
(for_statement
  left: (tuple_pattern
          (identifier) @definition.var))
(for_statement
  left: (identifier) @definition.var)

; not a scope!
;(while_statement) @scope

; for in list comprehension
(for_in_clause
  left: (identifier) @definition.var)
(for_in_clause
  left: (tuple_pattern
          (identifier) @definition.var))
(for_in_clause
  left: (pattern_list
          (identifier) @definition.var))

(dictionary_comprehension) @scope
(list_comprehension) @scope
(set_comprehension) @scope

;;; Assignments

(assignment
 left: (identifier) @definition.var)

(assignment
 left: (pattern_list
   (identifier) @definition.var))
(assignment
 left: (tuple_pattern
   (identifier) @definition.var))

(assignment
 left: (attribute
   (identifier)
   (identifier) @definition.field))

; Walrus operator  x := 1
(named_expression
  (identifier) @definition.var)

(as_pattern
  alias: (as_pattern_target) @definition.var)

;;; REFERENCES
(identifier) @reference
`,
	},
	"ruby": {
		language:     ruby.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
		query: `
; The MIT License (MIT)
;
; Copyright (c) 2016 Rob Rix
;
; Permission is hereby granted, free of charge, to any person obtaining a copy
; of this software and associated documentation files (the "Software"), to deal
; in the Software without restriction, including without limitation the rights
; to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
; copies of the Software, and to permit persons to whom the Software is
; furnished to do so, subject to the following conditions:
;
; The above copyright notice and this permission notice shall be included in all
; copies or substantial portions of the Software.
;
; THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
; IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
; FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
; AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
; LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
; OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
; SOFTWARE.

;;; DECLARATIONS AND SCOPES
(method) @scope
(class) @scope

[
 (block)
 (do_block)
 ] @scope

(identifier) @reference
(constant) @reference
(instance_variable) @reference

(module name: (constant) @definition.namespace)
(class name: (constant) @definition.type)
(method name: [(identifier) (constant)] @definition.function)
(singleton_method name: [(identifier) (constant)] @definition.function)

(method_parameters (identifier) @definition.var)
(lambda_parameters (identifier) @definition.var)
(block_parameters (identifier) @definition.var)
(splat_parameter (identifier) @definition.var)
(hash_splat_parameter (identifier) @definition.var)
(optional_parameter name: (identifier) @definition.var)
(destructured_parameter (identifier) @definition.var)
(block_parameter name: (identifier) @definition.var)
(keyword_parameter name: (identifier) @definition.var)

(assignment left: (_) @definition.var)

(left_assignment_list (identifier) @definition.var)
(rest_assignment (identifier) @definition.var)
(destructured_left_assignment (identifier) @definition.var)
`,
	},
	"rust": {
		language:     rust.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
		query: `
; Imports
(extern_crate_declaration
    name: (identifier) @definition.import)

(use_declaration
  argument: (scoped_identifier
              name: (identifier) @definition.import))

(use_as_clause
  alias: (identifier) @definition.import)

(use_list
    (identifier) @definition.import) ; use std::process::{Child, Command, Stdio};

; Functions
(function_item
    name: (identifier) @definition.function)

(function_item
  name: (identifier) @definition.method
  parameters: (parameters
                (self_parameter)))

; Variables
(parameter
  pattern: (identifier) @definition.var)

(let_declaration
  pattern: (identifier) @definition.var)

(const_item
  name: (identifier) @definition.var)

(tuple_pattern
  (identifier) @definition.var)

(if_let_expression
  pattern: (_
             (identifier) @definition.var))

(tuple_struct_pattern
  (identifier) @definition.var)

(closure_parameters
  (identifier) @definition.var)

(self_parameter
  (self) @definition.var)

(for_expression
  pattern: (identifier) @definition.var)

; Types
(struct_item
  name: (type_identifier) @definition.type)

(constrained_type_parameter
  left: (type_identifier) @definition.type) ; the P in  remove_file<P: AsRef<Path>>(path: P)

(enum_item
  name: (type_identifier) @definition.type)


; Fields
(field_declaration
  name: (field_identifier) @definition.field)

(enum_variant
  name: (identifier) @definition.field)

; References
(identifier) @reference
((type_identifier) @reference
                   (set! reference.kind "type"))
((field_identifier) @reference
                   (set! reference.kind "field"))


; Macros
(macro_definition
  name: (identifier) @definition.macro)

; Module
(mod_item
  name: (identifier) @definition.namespace)

; Scopes
[
 (block)
 (function_item)
 (closure_expression)
 (while_expression)
 (for_expression)
 (loop_expression)
 (if_expression)
 (if_let_expression)
 (match_expression)
 (match_arm)

 (struct_item)
 (enum_item)
 (impl_item)
] @scope
`,
	},
	"shell": {
		language:     bash.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
		query: `
; Scopes
(function_definition) @scope

; Definitions
(variable_assignment
  name: (variable_name) @definition.var)

(function_definition
  name: (word) @definition.function)

; References
(variable_name) @reference
(word) @reference
`,
	},
	"typescript": {
		language:     typescript.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
		query: `
; inherits: ecma
(required_parameter (identifier) @definition)
(optional_parameter (identifier) @definition)

; x => x
(arrow_function
  parameter: (identifier) @definition.parameter)

;; ({ a }) => null
(required_parameter
  (object_pattern
    (shorthand_property_identifier_pattern) @definition.parameter))

;; ({ a: b }) => null
(required_parameter
  (object_pattern
    (pair_pattern
      value: (identifier) @definition.parameter)))

;; ([ a ]) => null
(required_parameter
  (array_pattern
    (identifier) @definition.parameter))

(required_parameter
  (rest_pattern
    (identifier) @definition.parameter))
`,
	},
	"yaml": {
		language:     yaml.GetLanguage(),
		commentStyle: CommentStyle{}, // TODO
		query: `
[
 (stream)
 (document)
 (block_node)
] @scope

(anchor_name) @definition
(alias_name) @reference
`,
	},
}
