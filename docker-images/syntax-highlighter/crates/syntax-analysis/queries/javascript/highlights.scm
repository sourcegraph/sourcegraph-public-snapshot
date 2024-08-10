;; This file is adjusted from te original queries in
;; https://sourcegraph.com/github.com/tree-sitter/tree-sitter-javascript@15e85e80b851983fab6b12dce5a535f5a0df0f9c/-/blob/queries/highlights.scm

; Variables
;----------
(identifier) @variable

; Properties
;----------
(property_identifier) @property

(object (shorthand_property_identifier) @identifier)
(shorthand_property_identifier_pattern) @identifier
(shorthand_property_identifier) @identifier.attribute
(pair key: (property_identifier) @identifier.attribute)

; Function and method definitions
;--------------------------------

(function_expression
  name: (identifier) @identifier.function)
(function_declaration
  name: (identifier) @identifier.function)
(method_definition
  name: (property_identifier) @identifier.function)

(pair
  key: (property_identifier) @identifier.function
  value: [(function_expression) (arrow_function)])

(assignment_expression
  left: (member_expression
         property: (property_identifier) @identifier.function)
  right: [(function_expression) (arrow_function)])

(variable_declarator
  name: (identifier) @identifier.function
  value: [(function_expression) (arrow_function)])

(assignment_expression
  left: (identifier) @identifier.function
  right: [(function_expression) (arrow_function)])

; Function and method calls
;--------------------------

(call_expression
  function: (identifier) @identifier.function)

(call_expression
  function: (member_expression
             property: (property_identifier) @identifier.function))

; Literals
;---------

(this) @variable.builtin
(super) @variable.builtin

[
  (true)
  (false)
  (null)
  (undefined)]
@constant.builtin

(comment) @comment

(string ("\"" @string))
(string ("'" @string))
[
  (string_fragment)
  (template_string)
] @string

(regex) @string.special
(number) @number

; Tokens
;-------

(template_substitution
  "${" @string.escape
  "}" @string.escape)

[
  "as"
  "async"
  "await"
  "break"
  "case"
  "catch"
  "class"
  "const"
  "continue"
  "debugger"
  "default"
  "delete"
  "do"
  "else"
  "export"
  "extends"
  "finally"
  "for"
  "from"
  "function"
  "get"
  "if"
  "import"
  "in"
  "instanceof"
  "let"
  "new"
  "of"
  "return"
  "set"
  "static"
  "switch"
  "target"
  "throw"
  "try"
  "typeof"
  "var"
  "void"
  "while"
  "with"
  "yield"]
@keyword
