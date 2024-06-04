; Based on https://github.com/nvim-treesitter/nvim-treesitter/blob/master/queries/hack/highlights.scm
; To test the syntax of a Hack file you can run the compiler with the following steps
; > docker pull hhvm/hhvm
; > docker run  -v /local/path:/mount/path --tty --interactive hhvm/hhvm:latest /bin/bash -l
; > hhvm --version

((variable) @variable.builtin
  (#eq? @variable.builtin "$this"))

[
  (comment)
  (xhp_comment)
] @comment

((comment) @comment
  (#match? @comment "^[*][*][^*].*[*]$"))

(scope_identifier) @keyword
(visibility_modifier) @keyword

(xhp_open
  [
    "<"
    ">"
  ] @tag.delimiter)

(xhp_close
  [
    "</"
    ">"
  ] @tag.delimiter)


(xhp_open_close
  [
    "<"
    "/>"
  ] @tag.delimiter)

[
  "implements"
  "using"
  "attribute"
  "const"
  "extends"
  "insteadof"
  "trait"
  "throw"
  "yield"
  "is"
  "as"
  "?as"
  "super"
  "where"
  "list"
  "function"
  "use"
  "include"
  "include_once"
  "require"
  "require_once"
  "class"
  "type"
  "interface"
  "namespace"
  "enum"
  "new"
  "print"
  "echo"
  "newtype"
  "clone"
  "as"
  "concurrent"
  "async"
  "await"
  "return"
  "if"
  "else"
  "elseif"
  "switch"
  "case"
  "try"
  "catch"
  "finally"
  "for"
  "while"
  "foreach"
  "do"
  "continue"
  "break"
  (abstract_modifier)
  (final_modifier)
  (static_modifier)
  (visibility_modifier)
  (xhp_modifier)
  (inout_modifier)
  (reify_modifier)
  ; "nameof" This is missing in the grammar
] @keyword

(new_expression
  . (_) @type)

(new_expression
  (qualified_identifier
    (identifier) @type))

(scoped_identifier
  (qualified_identifier
    (identifier) @type))

(xhp_class_attribute
    name: (xhp_identifier) @variable
)

[
  "shape" ; also a keyword, but prefer highlighting as a type
  "tuple" ; also a keyword, but prefer highlighting as a type
  (array_type)
  "bool"
  "float"
  "int"
  "string"
  "arraykey"
  "void"
  "nonnull"
  "mixed"
  "dynamic"
  "noreturn"
  "nothing"
  "num"
] @type.builtin

(shape_type_specifier (open_modifier)) @type.builtin

(field_specifier (optional_modifier) @identifier.operator)

(null) @constant.null

[
  (true)
  (false)
] @boolean

[
  (nullable_modifier)
  (soft_modifier)
  (like_modifier)
] @operator

(alias_declaration
  ["newtype" "type"]
  .
  (identifier) @type)

(class_declaration
  name: (identifier) @type)

(interface_declaration
  name: (identifier) @type)

(type_parameter
  name: (identifier) @type)

(collection
  (qualified_identifier
    (identifier) @type .))

(attribute_modifier (qualified_identifier (identifier) @identifier.attribute))

[
  "@required"
  "@lateinit"
] @identifier.attribute

[
  "="
  "??="
  ".="
  "|="
  "^="
  "&="
  "<<="
  ">>="
  "+="
  "-="
  "*="
  "/="
  "%="
  "**="
  "==>"
  "|>"
  "??"
  "||"
  "&&"
  "|"
  "^"
  "&"
  "=="
  "!="
  "==="
  "!=="
  "<"
  ">"
  "<="
  ">="
  "<=>"
  "<<"
  ">>"
  "->"
  "+"
  "-"
  "."
  "*"
  "/"
  "%"
  "**"
  "++"
  "--"
  "!"
  "?:"
  "="
  "??="
  ".="
  "|="
  "^="
  "&="
  "<<="
  ">>="
  "+="
  "-="
  "*="
  "/="
  "%="
  "**="
  "=>"
  ; type modifiers
  "@"
  "?"
  "~"
  "?->"
  "->"
] @operator

[
  (integer)
  (float)
] @number

(parameter
  (variable) @variable.parameter)

(call_expression
  function: (qualified_identifier
    (identifier) @keyword .)
    (#eq? @keyword "invariant"))

(call_expression
  function: (qualified_identifier
    (identifier) @keyword .)
    (#eq? @keyword "exit"))

(call_expression
  function: (qualified_identifier
    (identifier) @function.builtin .)
    (#eq? @function.builtin "idx"))

(call_expression
  function: (qualified_identifier
    (identifier) @function.builtin .)
    (#eq? @function.builtin "is_int"))

(call_expression
  function: (qualified_identifier
    (identifier) @function.builtin .)
    (#eq? @function.builtin "is_bool"))

(call_expression
  function: (qualified_identifier
    (identifier) @function.builtin .)
    (#eq? @function.builtin "is_string"))

(call_expression
  function: (qualified_identifier
    (identifier) @function.call .))

(call_expression
  function: (scoped_identifier
    (identifier) @function.call .))

(call_expression
  function: (selection_expression
    (qualified_identifier
      (identifier) @function.call .)))

 (xhp_class_identifier) @variable

(qualified_identifier
  (identifier) @identifier.module
  .
  (identifier))

; Handle built-ins not named in the tree sitter grammar
(qualified_identifier
  . (identifier) @constant.builtin  .
  (#eq? @constant.builtin "NAN"))
(qualified_identifier
  . (identifier) @constant.builtin  .
  (#eq? @constant.builtin "INF"))
(qualified_identifier
  . (identifier) @constant.builtin  .
  (#eq? @constant.builtin "__CLASS__"))
(qualified_identifier
  . (identifier) @constant.builtin  .
  (#eq? @constant.builtin "__DIR__"))
(qualified_identifier
  . (identifier) @constant.builtin  .
  (#eq? @constant.builtin "__FILE__"))
(qualified_identifier
  . (identifier) @constant.builtin  .
  (#eq? @constant.builtin "__FUNCTION__"))
(qualified_identifier
  . (identifier) @constant.builtin  .
  (#eq? @constant.builtin "__LINE__"))
(qualified_identifier
  . (identifier) @constant.builtin  .
  (#eq? @constant.builtin "__METHOD__"))
(qualified_identifier
  . (identifier) @constant.builtin  .
  (#eq? @constant.builtin "__NAMESPACE__"))
(qualified_identifier
  . (identifier) @constant.builtin  .
  (#eq? @constant.builtin "__TRAIT__"))

; Explicitly handle internal amd module since they are not
; not mentioned in grammar
(qualified_identifier
  . (identifier) @keyword  .
  (#eq? @keyword "internal"))

(qualified_identifier
  . (identifier) @keyword  .
  (#eq? @keyword "module"))

(namespace_declaration
    name:  (qualified_identifier (identifier) @identifier.module))

;; Mark import path components as namespaces by default
(use_statement
  (qualified_identifier
    (identifier) @identifier.module)
  (use_clause))

(use_statement
  (use_type
    "namespace")
  (use_clause
    (qualified_identifier
      (identifier) @identifier.module .)
    alias: (identifier)? @identifier.module))

(use_statement
  (use_type
    "const")
  (use_clause
    (qualified_identifier
      (identifier) @constant .)
    alias: (identifier)? @constant))

(use_statement
  (use_type
    "function")
  (use_clause
    (qualified_identifier
      (identifier) @function .)
    alias: (identifier)? @function))

(use_statement
  (use_type
    "type")
  (use_clause
    (qualified_identifier
      (identifier) @type .)
    alias: (identifier)? @type))

(use_clause
  (use_type
    "namespace")
  (qualified_identifier
    (identifier) @identifier.module .)
  alias: (identifier)? @identifier.module)

(use_clause
  (use_type
    "function")
  (qualified_identifier
    (identifier) @function .)
  alias: (identifier)? @function)

(use_clause
  (use_type
    "const")
  (qualified_identifier
    (identifier) @constant .)
  alias: (identifier)? @constant)

(use_clause
  (use_type
    "type")
  (qualified_identifier
    (identifier) @type .)
  alias: (identifier)? @type)

(function_declaration
  name: (identifier) @function)

(method_declaration
  name: (identifier) @function.method)


(type_arguments
  [
    "<"
    ">"
  ] @punctuation.bracket)

[
  "("
  ")"
  "["
  "]"
  "{"
  "}"
  "<<"
  ">>"
] @punctuation.bracket

(ternary_expression
  [
    "?"
    ":"
  ] @identifier.operator)

[
  "."
  ";"
  "::"
  ":"
  ","
] @punctuation.delimiter

(qualified_identifier
  "\\" @punctuation.delimiter)

[
  (string)
  (xhp_string)
  (heredoc)
] @string

(xhp_open (xhp_identifier) @tag  (xhp_attribute)? @tag.attribute)
(xhp_close (xhp_identifier) @tag  (xhp_attribute)? @tag.attribute)
(xhp_open_close (xhp_identifier) @tag  (xhp_attribute)? @tag.attribute)

(type_specifier (qualified_identifier (identifier) @type))
(type_specifier) @type

(variable) @variable
(identifier) @variable
(pipe_variable) @variable

(braced_expression) @none
