; Identifier conventions

; Assume all-caps names are constants
((identifier) @constant
 (#match? @constant "^[A-Z][A-Z\\d_]+$'"))

; Assume that uppercase names in paths are types
((scoped_identifier
  path: (identifier) @type)
 (#match? @type "^[A-Z]"))
((scoped_identifier
  path: (scoped_identifier
         name: (identifier) @type))
 (#match? @type "^[A-Z]"))
((scoped_type_identifier
  path: (identifier) @type)
 (#match? @type "^[A-Z]"))
((scoped_type_identifier
  path: (scoped_identifier
         name: (identifier) @type))
 (#match? @type "^[A-Z]"))

; Assume other uppercase names are enum constructors
((identifier) @constant
 (#match? @constant "^[A-Z]"))

; Assume all qualified names in struct patterns are enum constructors. (They're
; either that, or struct names; highlighting both as constructors seems to be
; the less glaring choice of error, visually.)
;; (struct_pattern
;;   type: (scoped_type_identifier
;;     name: (type_identifier) @identifier.function))

; Function calls

(call_expression
  function: (identifier) @identifier.function)
(call_expression
  function: (field_expression
             field: (field_identifier) @identifier.function))
(call_expression
  function: (scoped_identifier
             "::"
             name: (identifier) @identifier.function))

(generic_function
  function: (identifier) @identifier.function)
(generic_function
  function: (scoped_identifier
             name: (identifier) @identifier.function))
(generic_function
  function: (field_expression
             field: (field_identifier) @identifier.function))

(metavariable) @identifier.attribute
(fragment_specifier) @type

(macro_invocation
  macro: (identifier) @identifier.function
  "!" @identifier.builtin)

; Function definitions

(function_item (identifier) @identifier.function)
(function_signature_item (identifier) @identifier.function)

; Other identifiers

(type_identifier) @type
(primitive_type) @identifier.builtin
(field_identifier) @identifier.constant

(line_comment) @comment
(block_comment) @comment

;; "(" @punctuation.bracket
;; ")" @punctuation.bracket
;; "[" @punctuation.bracket
;; "]" @punctuation.bracket
;; "{" @punctuation.bracket
;; "}" @punctuation.bracket
;;
;; (type_arguments
;;   "<" @punctuation.bracket
;;   ">" @punctuation.bracket)
;; (type_parameters
;;   "<" @punctuation.bracket
;;   ">" @punctuation.bracket)

;; [
;;   "::"
;;   ":"
;;   "."
;;   ","
;;   ";"
;; ] @punctuation.delimiter

(parameter (identifier) @variable.parameter)

(lifetime (identifier) @identifier.attribute)

(identifier) @identifier

[
  "as"
  "async"
  "await"
  "break"
  "const"
  "continue"
  "default"
  "dyn"
  "else"
  "enum"
  "extern"
  "fn"
  "for"
  "if"
  "impl"
  "in"
  "let"
  "loop"
  "macro_rules!"
  "match"
  "mod"
  "move"
  "pub"
  "ref"
  "return"
  "static"
  "struct"
  "trait"
  "type"
  "union"
  "unsafe"
  "use"
  "where"
  "while"]
@keyword
(crate) @keyword
(mutable_specifier) @keyword
(use_list (self) @keyword)
(scoped_use_list (self) @keyword)
(scoped_identifier (self) @keyword)
(super) @keyword

(self) @identifier.builtin

(char_literal) @character
(string_literal) @string
(raw_string_literal) @string

(boolean_literal) @boolean
(integer_literal) @number
(float_literal) @number

(escape_sequence) @string.escape

;; (attribute_item) @identifier.attribute
;; (inner_attribute_item) @identifier.attribute

"*" @identifier.operator
"&" @identifier.operator
"'" @identifier.operator
